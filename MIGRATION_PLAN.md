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

## 立即下一步

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

推荐顺序：

1. Wishlist
2. Review
3. Product registration / warranty
4. Feedback / suggestion feedback
5. Ticket / customer service history

要求：

- 先补 Go 能力，再切 Nuxt。
- 必须核对登录态、用户 ID、权限和错误响应。
- customer-service 不能简单等同于 ticket；需要单独设计。

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
