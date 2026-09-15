# 支付争议证据工作台

## 目标

拒付处理必须先保证证据真实、可审计、可人工复核。拒付发生和证据提交是两个不同的时刻：webhook 只记录拒付事实并进入待人工复核状态，运营人员可以在最终提交前查看、补录、修改或移除错误材料；只有点击最终提交后，系统才锁定本次渠道快照并发送给 Stripe / PayPal。当前已实现的是提交前编辑、首次最终提交创建快照，以及外部失败后的同快照重试；成功提交后的修订版创建和再次提交尚未形成可用的端到端能力。

## 时序规则

1. 拒付发生：只接收并保存支付渠道事件、金额、原因、状态、订单关联和原始 payload；不得自动提交证据。
2. 人工复核：后台打开当前订单证据包，查看结构化证据、发货照片、张力表、物流轨迹、签收/POD、沟通记录和附件。
3. 提交前编辑：当前订单证据包未锁定时，可以补传、替换、删除错误附件，修正物流号和证据项；系统不判断照片或张力表是否“合格”，只保存人工上传和填写的内容。
4. 最终提交：运营确认无误后点击提交。此时才组装当前最新证据、创建不可变的渠道提交快照，并调用对应支付渠道。
5. 外部失败重试（已实现）：复用同一渠道快照，不重新读取可能已经变化的订单或物流数据；重试请求中的新材料不会覆盖锁定版本。
6. 成功提交后的修订再提交（尚未实现）：已提交快照不能覆盖。订单证据域虽然支持创建证据包修订版，但当前拒付工作台尚未提供“修订版 -> 新渠道快照 -> 再次调用 Stripe/PayPal”的受支持闭环，不能对运营承诺该流程。

## 发货与物流纠正边界

- `POST /api/admin/orders/:id/fulfillment` 只表示首次确认发货。已发货订单再次提交完全相同的物流信息时可以幂等返回；如果带入不同物流信息，接口返回冲突。这个冲突只保护“重复发货”，不代表错误物流号无法修正。
- 物流号、承运商或物流服务填错后，必须使用独立的 `PATCH /api/admin/orders/:id/tracking` 物流纠正入口。该入口更新订单当前物流和当前唯一 tracking shipment，不重新发货，不改变原始 `shipped_at`。
- 物流来源发生变化时，先前 shipment 的 tracking events 不再作为当前证据链的物流事实；当前实现会清理当前事件集，避免错误物流轨迹继续进入证据包，新物流需要重新同步后才会产生新的轨迹事件。
- 物流纠正动作本身记录 before/after 审计，只保存订单号、物流号、服务商/承运商标识和状态等元数据，不保存 API Key、Webhook Secret 或证据文件内容。
- 如果某次拒付已经生成渠道提交快照，物流纠正不会覆盖旧快照。当前不能把物流纠正后的订单证据包修订版直接作为拒付重提入口；成功提交后的修订材料重提仍属于未实现能力。

## 当前数据链路

1. Stripe webhook 只负责记录拒付事实。
   - 入口：`internal/api/v1/payment/webhook_handler.go`
   - 本地表：`stripe_disputes`
   - 负责字段：Stripe dispute id、charge id、PaymentIntent、金额、币种、原因、状态、举证截止时间、原始 payload。
   - 同步写入支付风险事件与退款建议队列，但不自动退款。

2. 后台支付风控页展示退款建议。
   - API：`GET /api/admin/payment/risk/refund-recommendations`
   - API：`POST /api/admin/payment/risk/refund-recommendations/:id/pending-refund`
   - 本地表：`payment_refund_recommendations`
   - 建议来源：Stripe Early Fraud Warning、Stripe dispute、PayPal dispute。
   - 处理动作：`pending`、`accepted`、`dismissed`、`cancelled`。
   - `accepted` 只代表运营采纳建议；只有后台单独确认 `pending-refund` 动作后，才会生成一条本地 `pending` 退款草稿。
   - 本地退款草稿会通过 `linked_refund_id` 回写到建议记录，但不会调用 Stripe / PayPal 退款接口。
   - 如需真实退款，运营必须再次确认 `POST /api/admin/payment/refunds/:id/execute`；该动作会写入 `payment_refund_executions` 执行审计，并带幂等键调用支付渠道。

3. 后台支付风控页生成证据包。
   - API：`GET /api/admin/payment/disputes/:id/evidence`
   - 服务：`PaymentService.BuildStripeDisputeEvidencePackage`
   - 数据源：订单、支付交易、物流轨迹、候选客服沟通。
   - 客服沟通只作为候选预览，不默认提交。

4. 运营人工确认后提交 Stripe。
   - API：`POST /api/admin/payment/disputes/:id/evidence/submit`
   - 权限：订单编辑权限。
   - 服务：`PaymentService.SubmitStripeDisputeEvidence`
   - Stripe secret 只在后端读取，不返回前端。
   - 提交前必须传 `confirm=true`。

5. 提交审计独立保存。
   - 字段：`evidence_submitted_at`、`evidence_submission_payload`、`evidence_submission_error`
   - webhook 后续更新 dispute 状态时，只更新 webhook 负责字段，不能覆盖证据提交审计。
   - 只有最终提交动作才会创建 `order_evidence_submission_snapshots` 锁定版本，提交审计和接口结果都会记录快照 ID、版本和 SHA-256；拒付发生、页面预览和提交前编辑不会创建渠道锁定快照。
   - 外部提交失败后的重试优先读取同一锁定版本，不重新组装订单、证据、物流或客服沟通；PayPal 商业发票 URL 也从快照复用，不重复生成或上传。重试复用不等于成功提交后的修订重提；后者当前未实现。

6. 订单证据锁定导出独立保存。
   - API：`GET /api/admin/orders/:id/evidence/export`
   - 只有当前订单证据包为 `locked` 时，才允许首次生成导出清单。
   - 首次导出把订单、订单行、证据包、证据项、附件元数据/哈希以及当时匹配的物流 shipment/tracking events 固化到 `order_evidence_export_snapshots`。
   - 同一证据包版本的重复导出复用原 JSON 和 SHA-256，不重新读取实时订单证据或物流。
   - 创建新的证据包修订版后，新的锁定版本才会产生新的导出快照。
   - 导出清单不包含对象存储 Key；附件仍通过认证的订单证据附件接口访问。

7. PayPal dispute webhook 只记录拒付并进入人工复核。
   - 入口：`internal/api/v1/payment/paypal_payment_risk_webhook_handler.go`
   - 本地表：`paypal_disputes`
   - 触发：`CUSTOMER.DISPUTE.*`。
   - webhook 只保存拒付事实、订单关联和 `evidence_pending_review=true`，不调用 PayPal `ProvideEvidence`，不创建渠道提交快照，也不上传商业发票。
   - 后台拒付工作台通过 `GET /api/admin/payment/paypal-disputes/:id/evidence` 生成当前证据预览；“打开订单证据包”进入可编辑的订单证据域。
   - 运营检查并确认后，才调用 `POST /api/admin/payment/paypal-disputes/:id/evidence/submit`，由 `PaymentService.SubmitPayPalDisputeEvidence` 组装并提交 `PROOF_OF_FULFILLMENT`、物流、订单/发票摘要、送达/签收说明、沟通摘要和配置允许的商业发票 PDF。
   - `PAYPAL_DISPUTE_AUTO_ATTACH_INVOICE_PDF` 只控制最终人工提交时是否自动附加商业发票 PDF，不代表 webhook 自动提交。
   - 最终提交失败时，错误写入 `paypal_disputes.evidence_submission_error`；后续重试仍复用同一锁定快照，提交请求中的修正材料不会替换该快照。成功提交后的修订版重提当前未实现。

8. 后台可预览 PayPal 商业发票 PDF。
   - API：`GET /api/admin/payment/paypal-disputes/:id/evidence/invoice.pdf`
   - 返回：`application/pdf`，`Content-Disposition: inline`，浏览器可直接打开查看。
   - 作用：查看某个真实订单快照生成的 PDF。
   - 如果订单未关联、没有商品明细或卖方资料未配置，接口返回可读错误，不会生成空白或不完整 PDF。
   - 样式沙盒 API：`POST /api/admin/payment/paypal-invoice-preview.pdf`
   - 样式沙盒接收临时卖方、客户、地址、商品、金额和付款字段，只生成内存 PDF 返回后台，不创建争议、不上传 storage、不调用 PayPal。

## 当前支持的证据

- 客户姓名、邮箱、账单地址、收货地址。
- 商品描述、SKU、数量、订单金额。
- 物流承运商、发货日期、追踪号。
- 本地物流轨迹与 delivered/signed 事件摘要。
- 客服沟通候选记录预览。
- 人工补充说明。
- Stripe File ID：物流签收凭证、客服沟通 PDF、收据、其他附件。
- PayPal `PROOF_OF_FULFILLMENT`：tracking info 与结构化订单/发票/送达签收说明。
- PayPal 商业发票/订单收据 PDF：订单商品、SKU、数量、金额、运费、税费、折扣、总额、账单/收货地址、付款状态和 PayPal payment reference。该 PDF 只在人工最终提交阶段按配置生成和上传；后台另有独立样式沙盒，可用临时输入检查版式。

## PayPal 商业发票 PDF 当前需要什么

商业发票 PDF 由 `internal/pkg/invoice` 根据本地订单快照生成，再通过现有 storage 服务上传，最后转换成 PayPal `documents`。要让附件真正提交，需要同时满足：

- 订单必须有关联商品明细、账单/收货地址、币种、金额和 PayPal payment reference。
   - 后台专用设置优先配置卖方资料：`/api/admin/settings/paypal-invoice-seller-profile`，后台页面为“PayPal 发票卖方资料”。姓名和商业地址必填；邮箱、电话、网站、Tax ID / VAT ID / GST ID 可选。
   - 环境变量仍作为 fallback：`PAYPAL_DISPUTE_INVOICE_SELLER_NAME`、`PAYPAL_DISPUTE_INVOICE_SELLER_ADDRESS`、`PAYPAL_DISPUTE_INVOICE_SELLER_EMAIL`、`PAYPAL_DISPUTE_INVOICE_SELLER_PHONE`、`PAYPAL_DISPUTE_INVOICE_SELLER_WEBSITE`、`PAYPAL_DISPUTE_INVOICE_SELLER_TAX_ID`。
   - 当前 PDF 模板已经支持并在非空时生成 `Email`、`Phone`、`Website`、`Tax ID`；缺少这些可选字段不会阻止 PDF 生成。
- 样式沙盒预览只需要请求中的临时数据和卖方资料；不需要公开 storage，也不会提交给 PayPal。
- 真实争议 PDF 自动附件需要公开 storage 和 `PAYPAL_DISPUTE_AUTO_ATTACH_INVOICE_PDF=true`；该开关默认开启，只有明确设为 `false` 才关闭。
- `STORAGE_BASE_URL` 必须是 PayPal 可访问的公开 HTTPS URL；`localhost`、`127.0.0.1` 或非 HTTPS URL 会生成审计 warning，但不会把文件附给 PayPal。

当前 PDF 明确标注为 `Commercial invoice / order receipt prepared for payment dispute evidence`。它可以作为争议证据附件，但不等同于法定 VAT/GST 税务发票；正式税务发票仍需要卖方税务主体、税号、发票编号规则和司法辖区税务字段。

## 后续待做：PayPal 其他文件型证据

以下能力**当前尚未实现，属于后续迭代**：

- 从 DHL / FedEx 等承运商官方 API 自动获取带签名/收件人信息的 POD PDF。
- 获取客服沟通 PDF、官方签收证明和其他附件，并形成 PayPal 可访问的文档 URL。
- 将 POD、客服沟通 PDF 和其他附件通过 PayPal evidence `documents` 字段一并提交。
- 对官方文件获取、URL 过期和官方文件刷新建立独立审计链路。拒付结构化提交本身已经通过提交快照支持失败重试复用，但官方文件获取仍未接入。

当前仓库没有稳定的承运商官方签收 PDF API 适配器，也没有承运商认证的 POD 文档 URL 来源。因此当前 PayPal 文件附件只覆盖内部生成的商业发票/订单收据 PDF，不伪造“已附上官方签收 PDF”。

后续正式接入 PayPal 文件证据时建议新增独立适配层：

- `CarrierProofProvider`：按承运商和 tracking number 获取官方 POD PDF。
- `EvidenceDocumentStorage`：保存文件并生成 PayPal 可访问的短期 URL。
- `PayPalDisputeEvidenceAttachmentService`：把 POD、客服沟通 PDF 和其他附件转换成 PayPal `documents`，并与结构化 evidence 一起提交。

## 安全边界

- Stripe webhook 不自动提交 evidence，必须由后台人工确认。
- PayPal dispute webhook 不提交证据，只记录拒付并把订单置于待人工复核路径；商业发票 PDF 也不会在 webhook 阶段生成或上传。
- 拒付工作台允许运营在最终提交前查看当前证据包并按需编辑、补录、删除错误附件；提交前的材料变化不会被旧渠道快照锁住。
- 系统不会判定照片、张力表或人工证据内容是否合格；清单只提供缺失、待人工确认和未接入提示。
- `POST /api/admin/payment/paypal-invoice-preview.pdf` 只是版式预览工具，不改变真实争议提交流程。
- PayPal 商业发票 PDF 不是法定税务发票，不得把缺少税务主体/税号/正式编号规则的收据误标为 VAT/GST invoice。
- 拒付 webhook 不自动创建或执行退款；这条边界不适用于已验证的重复扣款 webhook，后者必须先落库并创建本地 `pending` 全额退款工单。
- 退款建议的“生成待处理退款”只创建本地草稿，不触发支付渠道 API。
- 真实渠道退款必须由后台退款权限账号显式确认执行，不能由风险 webhook 自动触发；退款完成或收到已验证的渠道退款事实后，系统还会按统一退款事务处理已发放积分扣回和下单抵扣积分返还。
- 前端不能接触 Stripe secret。
- 只有 `needs_response` / `warning_needs_response` 状态允许提交。
- 客服沟通不默认提交，避免误把无关聊天发给 Stripe。
- 证据文本按 Unicode 字符截断，避免多语言内容被截断成非法文本。
