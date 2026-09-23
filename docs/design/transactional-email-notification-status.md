# 事务通知与售后物流实施状态

> **审计日期**：2026-09-22  
> **代码优先**：运行时行为以 `go-backend` 当前代码、迁移和测试为准；本文件只记录范围、完成度和明确未完成项。

## 已完成

| 范围 | 当前实现与验收证据 |
| --- | --- |
| 订单/售后领域事件 | 订单支付成功、过期、取消、发货、送达、完成、退款，以及售后状态变更均使用 canonical Outbox 事件；旧订单确认/发货邮件命令已删除。 |
| 售后邮件规则 | `requested`、`approved`、`awaiting_return`、`return_in_transit`、`received`、`resolving`、`completed`、`rejected` 已映射稳定模板；每个模板支持 `notification_<template_code>_enabled` 运营开关，缺失配置默认开启；`inspecting`、`cancelled`、`exception` 默认只审计不发客户邮件。 |
| 模板系统 | 数据库存储、英文种子、locale fallback、变量白名单、HTML/纯文本安全渲染、版本历史、回滚为新版本、后台编辑/预览已完成。 |
| SMTP Provider | 数据库 provider 配置、AES-GCM 密文、默认 provider 选择、环境变量回退、解密失败 fail-closed、测试发送和测试结果持久化已完成。 |
| 售后退货资料 | 退货仓库、承运商、追踪号、追踪链接、面单 URL、发出/签收时间和收货人已持久化；状态与退货包裹同事务写入。 |
| 售后承运商 webhook 基础闭环 | 复用已有承运商签名校验和追踪 webhook；当追踪号匹配现有售后退货包裹且承运商报告妥投时，事务内推进案件至 `received`，写入状态事件和 canonical Outbox；未知追踪号不会创建售后包裹。 |
| 退货通知快照 | `after_sales.status_changed` 事件携带退货包裹快照；模板变量支持 `return_label_url`，worker 不依赖当前数据库状态回查。 |
| 仓库反查 | `GET /api/admin/after-sales/return-shipments?tracking_number=...` 可按追踪号反查售后案件，供扫码收货匹配。 |
| 售后商品额度 | 买家必须显式选择商品行和数量；已完成/驳回/取消工单释放额度；客户申请快照不再吞噬整单额度。 |
| SEO Sitemap | Go 后端 Sitemap 已包含已发布商品和启用分类，不再只输出博客。 |
| 媒体衍生图 | 全局生成槽位已配置化，默认容量为 8，不再硬编码为 2；非法容量不会替换当前限流器。 |
| 退款 Webhook | 网关退款回调可修复超时留下的本地执行失败状态，并完成关联售后工单。 |
| Outbox 死信后台 | 管理端复用 `outbox_events` 提供失败/死信事件筛选、详情、人工立即重试和忽略操作；重试/忽略均要求 `system:manage`，并写入管理员审计日志。列表不返回 payload，详情仅在系统管理权限下返回脱敏后的事件元数据和错误信息。 |
| 通知规则运营开关 | 每个已启用模板对应 `notification_<template_code>_enabled` 设置；缺失时默认开启，管理员可通过设置 API 关闭单个订单/售后通知，canonical 事件和 Outbox 事实仍保留。 |

## 尚未完成

| 范围 | 当前边界 | 后续验收条件 |
| --- | --- | --- |
| 生产 SMTP 放量 | 代码和后台配置已完成，但当前不代表生产邮箱已绑定并可发信。 | 配置真实 SMTP 凭据，完成订单确认、发货、售后各状态的真实收件验收。 |
| 运营模板内容 | 模板机制和默认英文种子已完成，业务方仍需审核各语言标题、正文、品牌签名和退货文案。 | 每个启用 locale 完成内容审核并发布模板版本。 |
| 外部客服/站内订阅 | 当前只完成 Outbox 和邮件投递；Zendesk、邮件网关 API、站内通知订阅尚未接入。 | 复用 canonical 事件接入订阅器，并完成重试、幂等和权限验收。 |
| 退货承运商全量同步 | 已复用现有签名 webhook、追踪轮询租约、失败退避和追踪号反查；妥投/签收自动推进 `received`，在途、异常、退回只记录物流事实并保留人工接管边界。 | 生产环境完成各承运商状态样本回放、重复/乱序/并发验收，并按承运商补齐状态别名。 |

### 退货承运商全量同步设计冻结（下一阶段入口）

本阶段只冻结规则，不宣称代码已完成。退货物流事实与售后案件状态必须分开处理：物流状态可以完整记录，但只有明确、单向且不会误导买家的事实才自动推进案件状态。

| 承运商归一化状态 | 数据处理 | 是否自动推进售后案件 | 备注 |
| --- | --- | --- | --- |
| `pre_transit` / `label_created` | 记录追踪事件 | 否 | 面单创建不代表包裹已交运 |
| `in_transit` / `out_for_delivery` | 记录追踪事件，更新退货包裹最后物流时间 | 否 | 不重复触发 `return_in_transit` 邮件 |
| `delivered` / `signed` / `pod` | 记录妥投证据、签收时间和收货人 | 是，推进到 `received` | 仅匹配已存在的售后退货包裹 |
| `exception` / `delayed` / `address_issue` | 记录异常代码和原始描述 | 否，转人工队列 | 不自动把案件改为 `exception`，避免承运商短暂异常误结案 |
| `returned_to_sender` / `undeliverable` | 记录退回事实和时间 | 否，转人工队列 | 由客服决定重新寄送、取消或改案 |
| 未知/冲突状态 | 保留原始事件并标记不可归一化 | 否 | 不创建售后包裹，不改变案件状态 |

实施约束：

1. Webhook 与轮询必须共享同一归一化器、追踪号反查和幂等键；同一承运商事件重复到达不得重复写状态事件或邮件 Outbox。
2. 轮询复用现有 `tracking_shipments` 的 claim lease、`next_sync_at` 和失败退避；只选择已绑定售后退货包裹的追踪号，未知追踪号不得创建售后记录。
3. 事件乱序时按承运商事件时间和已持久化的最高可信状态处理；较旧的 `in_transit` 不得覆盖已确认的 `received`。
4. `exception`、`returned_to_sender` 和无法解析的状态只产生可审计物流事实，人工处理完成前不发送售后客户邮件。
5. 首轮验收必须覆盖：重复 webhook、乱序 webhook、轮询与 webhook 并发、未知追踪号、承运商 API 超时、妥投后重试，以及死信回放。

## 暂不纳入本阶段

- 不新增按状态字段直接绑定 SMTP 的旧链路。
- 不保留已删除的 `order.confirmation_email` / `order.shipping_notification_email` 兼容命令。
- 不在本阶段实现 Zendesk、站内信或多渠道通知。
- 不把生产凭据配置、模板内容审核误报为代码完成。
- 死信人工重试只重置现有 Outbox 事件并立即重新入队，不创建第二套邮件队列；未知结果事件仍必须先走既有反查恢复流程。

## 验证命令

```text
go test ./... -count=1
go build ./internal/service ./internal/api/admin ./internal/api/v1/order ./internal/app
npm run typecheck
npm run build
```
