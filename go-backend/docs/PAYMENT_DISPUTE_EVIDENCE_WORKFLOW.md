# 支付争议证据工作台

## 目标

拒付处理必须先保证证据真实、可审计、可人工复核。当前实现同时覆盖 Stripe 的后台人工确认提交和 PayPal 的 webhook 自动证据提交，两条链路分别保留本地证据聚合与提交审计。

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

6. PayPal dispute webhook 自动提交结构化证据。
   - 入口：`internal/api/v1/payment/paypal_payment_risk_webhook_handler.go`
   - 本地表：`paypal_disputes`
   - 触发：`CUSTOMER.DISPUTE.*` 且争议仍处于卖家需要响应的状态。
   - 提交服务：`PaymentService.SubmitPayPalDisputeEvidence`
   - 当前提交内容：PayPal `PROOF_OF_FULFILLMENT`、物流承运商、tracking number、写入 `notes` 的订单/发票摘要、收货地址、商品明细、送达/签收事件和客服沟通摘要。
   - 后台提供独立的商业发票 PDF 样式预览，不会阻断真实 webhook 自动提交。配置卖方资料与公开 HTTPS 文件存储后，PayPal 会按自动流程通过 `documents` 字段提交 `commercial_invoice` 附件；`PAYPAL_DISPUTE_AUTO_ATTACH_INVOICE_PDF` 默认开启，显式设为 `false` 才关闭 PDF 附件。
   - 自动提交失败不会让 PayPal webhook 返回 500；失败原因写入 `paypal_disputes.evidence_submission_error`，并在 webhook 响应中返回。

7. 后台可预览 PayPal 商业发票 PDF。
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
- PayPal 商业发票/订单收据 PDF：订单商品、SKU、数量、金额、运费、税费、折扣、总额、账单/收货地址、付款状态和 PayPal payment reference。后台另有独立样式沙盒，可用临时输入检查版式。

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
- 对官方文件获取、URL 过期、失败重试和幂等提交建立独立审计链路。

当前仓库没有稳定的承运商官方签收 PDF API 适配器，也没有承运商认证的 POD 文档 URL 来源。因此当前 PayPal 文件附件只覆盖内部生成的商业发票/订单收据 PDF，不伪造“已附上官方签收 PDF”。

后续正式接入 PayPal 文件证据时建议新增独立适配层：

- `CarrierProofProvider`：按承运商和 tracking number 获取官方 POD PDF。
- `EvidenceDocumentStorage`：保存文件并生成 PayPal 可访问的短期 URL。
- `PayPalDisputeEvidenceAttachmentService`：把 POD、客服沟通 PDF 和其他附件转换成 PayPal `documents`，并与结构化 evidence 一起提交。

## 安全边界

- Stripe webhook 不自动提交 evidence，必须由后台人工确认。
- PayPal dispute webhook 会自动提交当前已具备的结构化 evidence；商业发票 PDF 也按配置自动附加，不要求运营额外上传。
- `POST /api/admin/payment/paypal-invoice-preview.pdf` 只是版式预览工具，不改变真实争议提交流程。
- PayPal 商业发票 PDF 不是法定税务发票，不得把缺少税务主体/税号/正式编号规则的收据误标为 VAT/GST invoice。
- webhook 不自动创建或执行退款。
- 退款建议的“生成待处理退款”只创建本地草稿，不触发支付渠道 API。
- 真实渠道退款必须由后台退款权限账号显式确认执行，不能由风险 webhook 自动触发。
- 前端不能接触 Stripe secret。
- 只有 `needs_response` / `warning_needs_response` 状态允许提交。
- 客服沟通不默认提交，避免误把无关聊天发给 Stripe。
- 证据文本按 Unicode 字符截断，避免多语言内容被截断成非法文本。
