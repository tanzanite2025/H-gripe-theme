# Tanzanite 下一步迁移规划

## 当前状态

项目已经明确从 WordPress/PHP 迁移到 Go + Nuxt + Vue Admin：

- C 端前台主线：`nuxt-i18n`
- B 端后台主线：`go-backend/web/admin`
- API 与业务逻辑主线：`go-backend`
- 根目录 WordPress 主题壳和 `style.css` 已删除，不再恢复。
- `wp-plugin/**` 只作为旧行为参考和过渡兼容来源，禁止新增 PHP 业务逻辑。

## 总原则

1. 一个模块一个 PR。
2. 每个 PR 必须只做一个明确模块，不能顺手混入别的迁移。
3. 不再恢复根目录 WordPress 主题入口。
4. Nuxt 前台最终只能调用 Go API。
5. PHP 只能在确认对应 Go/Nuxt 能力接管后删除，不能边迁边猜。

## 已完成

| 阶段 | 内容 | 状态 |
| --- | --- | --- |
| P0 | Go 后端启动基座、healthcheck、配置默认值 | 已完成 |
| D1a | 文档入口、过期文档边界、单模块 PR 规则 | 已完成 |
| D1b | PHP/WP API → Go API 迁移矩阵 | 已完成 |
| D1c | 删除根目录 WordPress 主题壳文件 | 已完成 |
| D1d | 删除根目录 `style.css` WordPress 主题元数据 | 已完成 |

## 当前迁移进度（2026-06-14）

按“一个模块一个 PR”推进，目前已合并到 `master`：

| 模块 | PR | 内容 | 状态 |
| --- | --- | --- | --- |
| M1.1 Settings / Quick-buy settings | #7 | Nuxt public settings、quick-buy settings 切到 Go `/api/v1/settings/*` | 已合并 |
| M1.2 Subscription submit | #8 | `SubscriptionOptIn.vue` 从 `/tanz/v1/subscribe` 切到 Go `/api/v1/subscriptions` | 已合并 |
| M1.3 Blog/content read APIs | #9 | 固定 Go content API 契约，Nuxt blog 读取不再回退旧 PHP | 已合并 |
| M2.1 Wishlist | #10 | Wishlist 前台读写切到 Go `/api/v1/wishlist` | 已合并 |
| M2.2 Review | #11 | 固定 review Go API 契约，修正 Go preload 关联 | 已合并 |
| M2.3 Feedback | #12 | Feedback submit / eligibility 切到 Go `/api/v1/feedback*` | 已合并 |
| M2.4 Suggestion feedback | #13 | Suggestion feedback submit / eligibility 切到 Go `/api/v1/suggestion-feedback*` | 已合并 |
| M2.5 Warranty / product registration | #14 | Warranty check、claim submission 切到 Go registrations/warranty API | 已合并 |
| M2.6 Spoke history / data export | #15 | Spoke history、spoke calculator data export 切到 Go `/api/v1/spoke/*` | 已合并 |
| M2.7 Agent customer service | #16 | Agent-side conversations/messages/read/transfer/status 切到 Go `/api/v1/customer-service/agent/*` ticket projection | 已合并 |

当前边界：

- `nuxt-i18n` 已完成 M1 与多数 M2 轻交互模块切流。
- customer-service 只完成 **agent-side** projection；公开聊天窗、agents 列表、auto-reply、订单历史仍未迁完。
- `wp-plugin/**` 仍作为旧行为参考与过渡兼容来源，不能因为某个前台调用已切 Go 就立即物理删除对应 PHP。

## 下一步建议

优先继续收敛仍在 `nuxt-i18n` 使用的旧 WordPress REST 入口，但继续保持一个业务域一个 PR：

1. **Public customer-service chat**：`has-conversation`、`messages`、`agents`、welcome/auto-reply 需要单独设计，不能直接等同 ticket。
2. **Orders read for chat drawer**：`/wp-json/mytheme-vue/v1/my-orders` 需要映射到 Go orders，并保持 WooCommerce 旧响应契约。
3. **M3 product catalog**：products/categories/filterable attributes 与 product SEO 分开做。
4. **M3 cart calculation**：shipping templates、tax rates、packaging rules、coupon/loyalty 逐项切流。

## 已执行模块说明

优先从低风险、只读或轻交互模块开始，先验证“Go API 接管 → Nuxt 切流 → PHP 路径删除”的流程。

### M1.1 Site settings / quick-buy settings

目标：

- 把 Nuxt 中站点配置、标题、社交链接、quick-buy 配置从 `/wp-json/**` 切到 Go。
- 使用 Go 路由：
  - `GET /api/v1/settings/site`
  - `GET /api/v1/settings/quick-buy`

不做：

- 不碰产品、购物车、订单。
- 不新增 PHP。
- 不删除 `wp-plugin/**`，除非确认没有任何调用。

验收：

- Nuxt 对站点公共设置不再依赖旧 `/wp-json/tanzanite/v1/settings*`。
- Go 返回字段覆盖当前页面需要。
- lint/typecheck 通过。

### M1.2 Subscription submit

目标：

- 把 `SubscriptionOptIn.vue` 从 `/tanz/v1/subscribe` 切到 Go subscription API。

不做：

- 不做订阅后台导出。
- 不做广播发送。
- 不迁移其它营销模块。

### M1.3 Blog/content read APIs

目标：

- 明确 Nuxt blog 当前到底走 Go content API 还是 Go 的 WP 兼容层。
- 固定文章列表、文章详情、翻译组接口契约。

不做：

- 不处理产品 SEO。
- 不处理全站 search。

## 中期迁移顺序

### M2 用户轻交互

执行进度：

1. Wishlist：已完成，PR #10。
2. Review：已完成，PR #11。
3. Product registration / warranty：已完成，PR #14。
4. Feedback / suggestion feedback：已完成，PR #12、#13。
5. Spoke history / data export：已完成，PR #15。
6. Agent customer-service：已完成 agent-side projection，PR #16。
7. Ticket / public customer-service history：未完成，需继续单独设计。

要求：

- 先补 Go 能力，再切 Nuxt。
- 必须核对登录态、用户 ID、权限和错误响应。
- customer-service 不能简单等同于 ticket；需要单独设计。
- public chat、auto-reply、agents 列表、order history 不能混入 agent customer-service PR。

### M3 交易链路

推荐顺序：

1. Product catalog
2. Cart
3. Coupon / gift card / loyalty
4. Shipping / tax / packaging
5. Order
6. Payment

要求：

- 必须做字段映射。
- 必须做数据对账。
- 必须做端到端下单验证。
- 不允许和其它模块混 PR。

### M4 PHP 物理删除

某个 PHP 模块只能在满足以下条件后删除：

1. Nuxt 搜索不到对应 `/wp-json/**` 调用。
2. Go API 已覆盖旧成功/失败响应。
3. 旧数据迁移或读取路径已验证。
4. 线上代理不再转发该模块请求到 WordPress。
5. 对应迁移 PR 已合入。

## 每个 PR 开始前检查

新开 PR 前必须先回答：

- 当前模块名称是什么？
- 这个 PR 是否只做这一个模块？
- 是否触碰 PHP？原因是什么？
- 是否涉及 Nuxt 切流？
- 是否需要数据迁移？
- 验证命令是什么？

如果回答不清楚，先补文档或矩阵，不写业务代码。

## 不要做的事

- 不要一次性迁多个模块。
- 不要恢复根目录 WordPress 主题文件或 `style.css`。
- 不要在 `wp-plugin/**` 增加新业务逻辑。
- 不要把 API 矩阵、Go 实现、Nuxt 切流、PHP 删除放在同一个 PR。
- 不要为了临时跑通把 Nuxt 调回 PHP。
