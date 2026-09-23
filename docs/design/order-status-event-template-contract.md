# 订单状态、业务事实、事件与邮件模板契约

> **文档状态**：实现基线；运行时行为以当前代码、迁移和验收测试为准。  
> **最后审计**：2026-09-22  
> **关联规范**：[`transactional-email-and-notification-architecture.md`](./transactional-email-and-notification-architecture.md)
> **完成度清单**：[`transactional-email-notification-status.md`](./transactional-email-notification-status.md)

## 1. 目的与结论

本契约对应的代码实现已完成 canonical 事件、Outbox、通知规则、模板渲染、SMTP Provider、售后退货资料、追踪号反查，以及签名承运商 webhook 对妥投退件的自动收货闭环。生产 SMTP 凭据验收、运营模板内容审核、外部客服订阅和退货承运商全量状态同步仍未完成；全量同步的状态映射与人工接管边界已冻结在完成度清单中，待单独实现和验收。

订单目前不是一个单状态机，而是四个相互独立的状态维度：主订单、支付、物流、生产。邮件不能直接监听某个数据库字段，也不能把后台可修改的状态值当成完整业务事实。

本契约规定后续实现必须遵循以下链路：

```text
业务动作/外部事实
    -> 受保护的领域事件（说明发生了什么）
    -> 同事务 Outbox
    -> 通知规则（是否发送、收件人、模板编码）
    -> 模板渲染与邮件投递
```

因此，不做“状态下拉框绑定模板”或在业务请求中直接调用 SMTP。SMTP 解决的是“怎么发”，模板解决的是“发什么”，事件契约解决“什么时候发”。

## 2. 当前代码审计结果

### 2.1 已存在的入口

| 事实 | 当前入口 | 当前结果 |
| --- | --- | --- |
| 支付成功 | `PaymentService`/`PaymentTransactionService` 验证网关交易 | 同事务更新 `payment_status=paid`，并写入 `order.paid`、`order.payment_succeeded`；旧确认邮件命令不再由新支付流程写入 |
| 订单发货 | `OrderService.FulfillOrder` | 同事务写入包裹、主订单 `status/shipping_status=shipped`，并写入 `order.shipped`；旧发货邮件命令不再由新履约流程写入 |
| 物流送达 | `shipping_tracking_helpers.go` | 更新 `shipping_status=delivered` 和 `delivered_at`，并写入 `order.delivered` canonical 事件 |
| 未付款取消/支付过期 | `cancelOrderWithRollback`、`order_payment_expiration_service.go` | 更新主订单状态、回滚库存并写入对应 canonical 事件 |
| 全额退款 | `PaymentService` 退款执行/回调流程 | 更新 `payment_status=refunded`，并在退款事实提交后写入 `order.refunded` canonical 事件 |
| 拒付/风控审核 | `MarkDisputed`、`MarkPaymentLiabilityReviewHold` | 直接写主订单投影，没有统一买家通知事件 |
| 售后状态 | `AfterSalesService.UpdateStatusWithReturnShipment` | 状态与物流记录事务内写入 `after_sales.status_changed` 事件，由通知规则按状态选择模板 |

### 2.2 当前不能作为通知契约的内容

1. `OrderRepository.UpdateStatus`、`UpdatePaymentStatus`、`UpdateShippingStatus` 等方法主要负责持久化，没有统一的状态变化事件。
2. `order.paid` 是现有 Outbox 的业务集成事件（当前用于 webhook/ERP 类处理），不能直接等同于“给买家发送订单确认邮件”。买家确认/发货邮件现在由 `order.payment_succeeded`、`order.shipped` 事实经过通知规则驱动；旧邮件命令已删除，不再兼容消费。
3. 管理端主订单状态请求目前只接受 `pending|processing|shipped|completed|cancelled`，而后端还存在系统产生的 `payment_expired|refunded|disputed|needs_review`，两者不能直接共用一套“可编辑状态”列表。
4. `paid` 仍出现在主订单转换表中，但支付成功主流程通常将订单从 `pending` 直接推进为 `processing`，所以 `paid` 只能按兼容状态处理，不能作为新的邮件触发依据。

## 3. 权威状态维度

### 3.1 主订单状态 `Order.Status`

| 值 | 产生方 | 允许后台直接修改 | 业务含义 | 邮件绑定原则 |
| --- | --- | --- | --- | --- |
| `pending` | 下单 | 仅作为初始/兼容值 | 尚未完成支付 | 不发送“订单确认” |
| `paid` | 历史兼容/少量内部数据 | 否 | 旧数据中的已支付中间态 | 不直接触发邮件 |
| `processing` | 支付成功流程、审核释放 | 有限（必须经过服务层） | 已具备履约条件，正在处理 | 不因字段写入重复发确认邮件 |
| `shipped` | 履约服务 | 否，管理端直接改为 shipped 被拒绝 | 至少一个包裹已发出 | 使用包裹发货事实触发发货模板 |
| `completed` | 完成/积分结算流程 | 受服务层约束 | 订单履约完成 | 使用 `order.completed` 事实触发完成模板 |
| `cancelled` | 未付款取消流程 | 不能取消已支付订单 | 订单已取消 | 使用取消事实触发取消模板 |
| `payment_expired` | 支付过期任务 | 否 | 支付窗口过期 | 使用支付过期事实触发过期模板 |
| `refunded` | 全额退款流程 | 否 | 订单资金已全额退回；部分场景还受实物补货条件约束 | 使用退款完成事实触发退款模板 |
| `disputed` | Stripe/PayPal 拒付流程 | 否 | 存在活动中的资金争议，履约冻结 | 默认不自动暴露风控细节；若启用须单独批准模板文案 |
| `needs_review` | 高价值支付责任转移审核 | 否 | 支付已成功但履约需人工审核 | 默认不向买家发送内部审核原因 |

主订单的转换表只描述正常履约路径，不能约束支付、退款和拒付投影。系统事件必须显式记录旧值、新值和产生原因，不能依赖转换表推断事实。

### 3.2 支付状态 `Order.PaymentStatus`

当前持久化值为：

```text
unpaid -> paid -> refunded
unpaid -> expired
```

支付失败不是当前订单支付状态的稳定值；失败交易记录在支付交易/网关结果中。邮件事件应使用 `payment_succeeded`、`payment_expired` 和 `refund_completed` 等事实名，而不是监听 `payment_status` 的任意写入。

### 3.3 物流状态 `Order.ShippingStatus`

当前值为 `pending|processing|shipped|delivered`。`shipped` 必须由履约包裹创建流程产生，`delivered` 必须由承运商/物流同步在所有包裹满足条件后产生。新增包裹可能将聚合投影从 `delivered` 重新打开为 `shipped`，因此“送达邮件”必须按不可重复的送达事实或包裹事实幂等，而不是按当前字段值发送。

### 3.4 生产状态 `Order.ProductionStatus`

当前值为 `not_applicable|not_started|started|completed|cancelled`，只适用于定制/生产履约。生产状态用于履约闸门，不自动等价于买家邮件状态；如果未来需要“生产开始/完成”通知，应新增独立业务事实和模板，不监听字段保存。

## 4. 目标领域事件与通知映射

事件名称描述“已经发生的事实”，模板编码描述“选择哪封邮件”。同一个事实可以没有邮件，也可以在未来增加站内信、客服系统或 Webhook 订阅者。

| 领域事件 | 事实条件 | 默认邮件模板 | 当前阶段 | 幂等事实键 |
| --- | --- | --- | --- | --- |
| `order.payment_succeeded` | 网关交易验证成功且本地支付事务提交 | `order_confirmation` | canonical 通知 | `payment_transaction:{id}` |
| `order.payment_expired` | 未付款订单被过期任务成功认领 | `order_payment_expired` | 新增 | `order_expiration:{order_id}`（当前单次过期事实） |
| `order.cancelled` | 未付款订单取消并完成库存回滚 | `order_cancelled` | 新增 | `order_cancellation:{order_id}`（当前单次取消事实） |
| `order.shipped` | 一个包裹成功创建并取得追踪号 | `order_shipping_notification` | canonical 通知 | `shipment:{shipment_id}` |
| `order.delivered` | 订单所有相关包裹确认送达 | `order_delivered` | 新增 | `delivery_fact:{order_id}:{tracking_number}`（当前拆单兼容键） |
| `order.completed` | 订单完成流程成功提交 | `order_completed` | 新增 | `order_completion:{order_id}`（当前完成事实） |
| `order.refunded` | 网关退款已验证完成且本地退款状态提交 | `order_refunded` | 新增 | `refund:{refund_id}` |
| `order.disputed` | 活动拒付已写入并冻结履约 | 无（默认） | 仅审计/风控 | `dispute:{provider}:{provider_dispute_id}` |
| `order.payment_review_required` | 高价值支付责任审核被创建 | 无（默认） | 仅内部通知 | `payment_review:{review_id}` |
| `order.payment_review_released` | 审核通过且履约冻结解除 | 无（默认） | 仅内部审计 | `payment_review:{review_id}:released` |
| `after_sales.status_changed` | 售后状态机发生合法转换 | 按状态映射 | 新增 | `after_sales_case:{case_id}:transition:{transition_id}` |

`order.confirmation_email` 和 `order.shipping_notification_email` 已从代码和 Outbox 注册表移除。由于系统尚未正式运营，不保留旧命令兼容分支；确认/发货通知统一由 canonical 领域事实驱动。

## 5. 售后状态邮件映射

售后状态完整集合为 `requested|reviewing|approved|awaiting_return|return_in_transit|received|inspecting|resolving|completed|rejected|cancelled|exception`。不是每个中间状态都需要打扰买家，模板只绑定明确的客户可理解节点：

| 售后事实 | 模板编码 | 必要变量 | 说明 |
| --- | --- | --- | --- |
| 创建申请 | `after_sales_requested` | 案件号、订单号、售后类型、当前状态 | 告知已受理，不承诺批准；当前使用创建时生成的 transition ID 作为事实键 |
| 审核批准 | `after_sales_approved` | 案件号、处理方式、下一步 | 只在审核动作提交后发送 |
| 等待寄回 | `after_sales_awaiting_return` | 仓库名、地址、面单链接（如已生成）、截止时间 | 退货仓字段必须已经落库；面单链接来自同一退货包裹快照 |
| 退货在途 | `after_sales_return_in_transit` | 承运商、追踪号、追踪链接 | 与退货包裹记录同事务 |
| 已签收 | `after_sales_received` | 案件号、签收时间 | 不能在仅创建面单时发送 |
| 进入处理/解决 | `after_sales_resolving` | 案件号、处理说明（可选） | 支持 `notification_after_sales_resolving_enabled` 运营开关，缺失时默认发送 |
| 已完成 | `after_sales_completed` | 案件号、解决方案、退款金额（如适用） | 退款/补发事实必须已提交 |
| 已驳回 | `after_sales_rejected` | 案件号、驳回原因 | 原因必须经过内容安全与隐私审查 |
| 取消或异常 | 默认不自动发送 | — | 先建立客服/站内通知策略，避免误导 |

`return_in_transit`、`received` 等退货物流节点必须使用案件状态变更与 `AfterSalesReturnShipment` 同一事务写入；邮件 worker 不应再次查询易变的当前状态来猜测发送内容，而应使用事件中保存的快照变量。

退货包裹快照至少包含仓库名称/地址、承运商、追踪号、追踪链接和面单链接（`label_url`）。`after_sales.status_changed` 事件同步携带这些字段；模板渲染侧以 `return_label_url` 暴露面单地址，缺失时不得从当前案件或物流表回查补齐。

后台收货场景可使用 `GET /api/admin/after-sales/return-shipments?tracking_number=...` 按退货追踪号反查售后案件。该接口沿用订单查看权限，并返回案件、状态历史及退货包裹快照，供仓库扫码匹配；它不改变案件状态，收货确认仍必须通过状态机接口提交。

## 6. 事件 Envelope 与接口边界

后续新增事件统一采用以下最小 Envelope。当前系统的 canonical 持久化载体是 `outbox.Event`；版本和幂等字段必须沿用下面定义，不另造平行 Envelope：

```json
{
  "event_id": 123,
  "event_type": "order.shipped",
  "schema_version": 1,
  "aggregate_type": "order",
  "aggregate_id": "123",
  "occurred_at": "2026-09-19T08:00:00Z",
  "actor": {"kind": "system", "id": null},
  "idempotency_key": "shipment:456",
  "payload": {
    "previous_status": "processing",
    "new_status": "shipped",
    "order_number": "ORD-2026-001",
    "locale": "en"
  }
}
```

当前实现使用 `outbox_events.id` 的数据库自增整数作为 `event_id`；事件类型、聚合和幂等键由现有表列及版本化 payload 承载，不另建 UUID envelope 表。若未来需要改为 UUID，必须先提交独立架构决策并同步迁移消费者，不能在单个事件中混用两种 ID 契约。

约束：

1. 业务状态更新与 Outbox 插入必须在同一本地事务内完成。
2. Payload 或与事件同事务绑定的不可变投递上下文保存邮件需要的事实快照（订单号、金额、承运商、追踪号、语言、案件号等），避免 worker 依赖实时商品或状态查询重建历史内容。订单 canonical payload 现在可携带 `recipient_email`、`locale` 和 `customer_name` audience snapshot；旧事件或售后事件缺失时，投递计划 API 才接受显式上下文补充。
3. 同一事实的 `idempotency_key` 必须映射到 `outbox_events.event_key` 的唯一约束；不得仅使用当前时间戳去重。
4. 事件消费失败只影响通知，不回滚已提交的订单/退款/售后事实；worker 使用现有 Outbox 重试和死信机制。
5. 后台人工状态操作必须走领域服务接口，接口需要返回 `previous_state`、`new_state`、`transition_id`，禁止控制器直接更新表字段。

## 7. 模板规则

模板编码是稳定 API，不得把文件名、语言或状态展示标签作为业务调用参数。第一批订单模板编码为：

```text
order_confirmation
order_shipping_notification
order_payment_expired
order_cancelled
order_delivered
order_completed
order_refunded
```

售后模板编码见第 5 节。模板选择顺序为：事件指定的客户 locale 精确匹配 -> 同语言主区域匹配 -> `en` -> 内置纯文本兜底。模板变量必须由事件类型白名单校验，未知变量或必填变量缺失时事件进入可观测失败/死信，不允许静默发送空内容。

## 8. 实施顺序与验收门槛

### 阶段 0：契约冻结（已完成）

- 评审本文件的状态维度、事件名称、模板编码和幂等键。
- `order_confirmation`、`order_shipping_notification` 已统一为 canonical 模板编码；旧的订单邮件命令、接口方法和内嵌模板已删除，不保留兼容别名或双写分支。
- 为每个领域事件确定 payload schema 和可发送的收件人/语言来源。

### 阶段 1：事件基础设施（已完成）

- 增加事件常量、Envelope/schema 版本、领域服务接口和事务内 Outbox 写入辅助函数。
- 覆盖支付成功、支付过期、取消、发货、送达、退款完成、订单完成和售后状态变更；SMTP provider 管理后台已在阶段 2 交付。
- Outbox 消费者校验版本、幂等 Envelope 和事件类型所需的最小事实字段，并解析事件对应的稳定模板编码。
- 已增加纯规则层、模板定义注册表、必填/允许变量校验、locale fallback、内容渲染和异步邮件投递，不把状态字段直接绑定到 SMTP。
- 为每个事实增加重复调用、并发冲突、回放和 Outbox 唯一键测试。

### 阶段 2：模板与发件配置（已完成）

- 模板与版本表已由 `335_notification_templates` 创建，英文默认模板由 `336_seed_transactional_notification_templates` 写入；发件 provider 表由 `341_email_provider_configs` 创建。
- 模板保存、版本快照、locale fallback、受控变量渲染和英文默认种子已落地；SMTP 配置复用阶段 1 的模板编码与变量契约。
- 发件 provider 配置表已由 `341_email_provider_configs` 创建，后台已提供脱敏 CRUD、默认通道切换、AES-GCM 密文存储和显式目标地址的真实 SMTP 测试发送。
- Outbox 发送运行时每次动态读取数据库默认 provider；无默认 provider 时才回退环境变量 SMTP，数据库 provider 解密或配置失败时 fail-closed，不得静默回退。
- 后台模板编辑、HTML/纯文本预览、版本历史和“回滚为新版本”已接入；回滚不会覆盖既有历史快照。

### 阶段 3：模板渲染与业务扩展（已完成）

- canonical 领域事件已经接入投递计划、模板版本、locale fallback、变量白名单和异步邮件投递；发送层优先使用数据库默认 provider，无默认 provider 时才复用环境变量 SMTP。
- 订单完成事件继续由结算处理器负责积分结算，同时复用同一投递链路发送完成通知。
- canonical 领域事件已经接管订单确认、发货及售后通知；旧邮件命令的直接绑定已删除，后续不得重新引入按文件名或旧命令分流的兼容路径。

当前代码已落地一个不依赖 provider 的投递计划构建器
`BuildTransactionalNotificationDeliveryPlan`。它把一个 canonical Outbox
事实解析为 `template_code`、收件人、locale、幂等键和受白名单约束的渲染变量，
随后由 Outbox handler 解析模板、渲染 HTML/纯文本并调用现有邮件服务发送。计划或渲染
失败时必须让事件进入可观测
的重试/死信路径；`inspecting`、`cancelled`、`exception` 等默认静默状态则返回
空模板计划并正常确认事实。

投递计划的收件人和 locale 是显式上下文，不从当前订单状态反推：后续 worker 应从
订单/客户快照或事件 payload 提供 `RecipientEmail`、`Locale` 和必要的快照变量。
provider 选择和真实连通性测试已经落地。当前 SMTP 发送器按事件发送动态创建，管理员切换默认 provider 后新事件立即使用新配置；暂不引入跨事件共享连接池。

### 阶段 4：外部订阅

- 在不改变买家邮件事实的前提下，再将同一领域事件投递到 Zendesk、邮件网关或站内通知。

## 9. 冲突修正记录

| 文档/描述 | 修正 |
| --- | --- |
| 事务邮件架构文档把自身称为代码之前的绝对事实源 | 改为设计基线；在实现验收前，`docs/README.md` 规定的“代码优先”继续有效 |
| 事务邮件架构文档预占 Migration `332` | 已修正为：模板表使用 `335_notification_templates`；后续 provider 表使用下一个可用编号 |
| SMTP 运维文档把邮件范围写成 newsletter/warranty 已接通 | 保持“当前已接通范围”表述；订单/售后通知属于待验收能力，不得写成已上线 |
| `order.paid` 被误认为买家确认邮件事件 | 明确其与邮件命令不同；确认邮件必须由 `order.payment_succeeded` 通知规则驱动 |
