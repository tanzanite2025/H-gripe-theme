# Stripe 拒付证据工作台

## 目标

拒付处理必须先保证证据真实、可审计、可人工复核。当前实现把 Stripe webhook、ERP 本地证据聚合、后台人工确认、Stripe Dispute Evidence 提交拆成独立职责，避免 webhook 自动把不完整或不相关材料提交给 Stripe。

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

## 当前支持的证据

- 客户姓名、邮箱、账单地址、收货地址。
- 商品描述、SKU、数量、订单金额。
- 物流承运商、发货日期、追踪号。
- 本地物流轨迹与 delivered/signed 事件摘要。
- 客服沟通候选记录预览。
- 人工补充说明。
- Stripe File ID：物流签收凭证、客服沟通 PDF、收据、其他附件。

## 尚未实现的承运商官方 PDF 自动拉取

当前仓库没有 DHL / FedEx 官方签收 PDF API 适配器，也没有承运商凭证自动上传 Stripe File API 的稳定链路。因此第一阶段不伪造“自动拉取官方 PDF”。

后续正式接入时建议新增独立适配层：

- `CarrierProofProvider`：按承运商和 tracking number 获取官方 POD PDF。
- `StripeDisputeFileUploader`：用 `purpose=dispute_evidence` 上传 PDF，返回 `file_...`。
- `DisputeEvidenceAttachmentService`：把承运商凭证、订单收据、客服沟通 PDF 统一转换成 Stripe File ID，再交给证据工作台预览提交。

## 安全边界

- webhook 不自动提交 evidence。
- webhook 不自动创建或执行退款。
- 退款建议的“生成待处理退款”只创建本地草稿，不触发支付渠道 API。
- 真实渠道退款必须由后台退款权限账号显式确认执行，不能由风险 webhook 自动触发。
- 前端不能接触 Stripe secret。
- 只有 `needs_response` / `warning_needs_response` 状态允许提交。
- 客服沟通不默认提交，避免误把无关聊天发给 Stripe。
- 证据文本按 Unicode 字符截断，避免多语言内容被截断成非法文本。
