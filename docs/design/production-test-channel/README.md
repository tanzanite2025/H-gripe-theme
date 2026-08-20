# Production Test Channel

Last updated: 2026-08-19

## 设计结论

生产环境需要一个受控的测试通道，但不应该开放一个拥有全站能力的“顶级账号”。建议单独建立 `productiontest` 业务域，专门管理生产测试账号、测试商品准入、测试订单标记、售后/退款验证和审计。

核心原则：

- 测试账号不是超级管理员，只是被授权进入生产测试通道的普通用户。
- 测试商品默认只允许白名单测试账号购买，不能出现在普通用户可购买路径里。
- 订单、支付、退款、售后尽量走真实生产链路，但必须全链路打 `production_test` 标记。
- 测试单默认不进入真实发货、真实库存消耗、营收统计、佣金、积分、优惠券发放等业务结果。
- 所有启用、购买、退款、售后和关闭动作都必须保留审计记录。

## 本次先落地的代码边界

本次只实现不会改变现有订单、支付、售后路径的纯判断：

- `auth.RoleTestUser`（值为 `test_user`）代表生产测试用的前台账号角色。
- `test_user` 是有效的 storefront 角色，但不是 backoffice 角色，不拥有任何后台权限。
- 测试账号仍然必须有独立的生产测试授权记录；角色本身不能直接获得测试商品购买资格。
- `productiontest.EvaluatePurchase` 判断当前用户是否可以买某个已解析的测试商品 gate。
- `productiontest.EvaluateOrder` 判断购物车是否允许提交，以及是否必须写入生产测试订单标记。

本次明确不做：

- 不新增或修改订单、支付、退款、售后模型和服务调用。
- 不把测试判断硬编码进 `CreateOrder`、支付回调或售后状态机。
- 不新增后台页面、后台 API、migration 或报表过滤。
- 不改变现有 Nuxt 登录路径；测试账号使用现有 `/api/v1/auth/login` 和用户域。
- 不把 `test_user` 加入现有 `/api/admin/users` 的后台员工账号表单；后续由独立的生产测试管理功能创建和绑定。

## 独立领域目录

后端领域目录建议独立为：

```text
go-backend/internal/domain/productiontest/
  test_account.go
  product_gate.go
  order_marker.go
  purchase_policy.go
  purchase_policy_test.go
```

后续接入时继续按现有仓库分层扩展：

```text
go-backend/internal/repository/production_test_account_repository.go
go-backend/internal/repository/production_test_product_gate_repository.go
go-backend/internal/repository/production_test_order_marker_repository.go
go-backend/internal/service/production_test_service.go
go-backend/internal/api/admin/production_test_handler.go
go-backend/web/admin/src/api/productionTest.ts
go-backend/web/admin/src/views/ProductionTest.vue
```

不要把这类规则散落在 `order`、`payment`、`aftersales` 或 `product` 域里。那些域只在关键节点调用 `productiontest` 做裁决和打标。

## 权限设计

建议拆成三类后台权限：

| 权限 | 用途 | 风险控制 |
| --- | --- | --- |
| `production_test.manage` | 新增/停用测试账号、配置测试商品、调整有效期 | 仅限高级运营或系统负责人 |
| `production_test.use` | 使用指定用户账号完成生产链路验证 | 不授予后台管理能力 |
| `production_test.audit` | 查看测试单、测试退款、测试售后和审计日志 | 只读 |

测试账号与后台账号必须是两个边界：

- 后台操作者可以在生产测试管理功能中创建或停用测试账号。
- 创建出来的用户使用 `test_user` storefront 角色，或由独立测试授权记录绑定到一个 storefront 用户。
- `test_user` 不属于 `IsBackofficeRole`，没有后台权限，不能通过 `/api/admin/auth/login` 或后台刷新令牌进入 admin console。
- 该账号只能通过 Nuxt 前台现有用户登录入口进入 `/api/v1/auth/login`，再访问用户域功能。
- `test_user` 会被视为客户账号，而不是员工账号；后台客户视图可以把它作为 storefront customer 展示。
- 仅有 `test_user` 角色不等于可以购买测试商品；必须同时满足有效、未过期、未停用的生产测试授权，以及商品 gate 的账号范围。

管理规则：

- 启用测试通道必须填写原因、负责人和截止时间。
- 测试账号必须绑定真实 `user_id`，不能使用匿名开关。
- 测试账号默认 24 小时或 7 天到期，到期后自动失效。
- 测试商品启用时必须二次确认，因为它会影响前台购买准入。
- 生产测试账号不应该绕过支付、订单金额、税费、退款金额计算和售后校验。

## 数据模型建议

### `production_test_accounts`

记录哪些用户可以进入生产测试通道。

| 字段 | 说明 |
| --- | --- |
| `id` | 测试账号授权 ID |
| `user_id` | 绑定的普通用户 ID |
| `label` | 账号说明，例如 `prod-checkout-smoke` |
| `purpose` | 使用目的 |
| `status` | `active` / `disabled` |
| `expires_at` | 到期时间 |
| `created_by` / `updated_by` | 后台操作者 |

### `production_test_product_gates`

记录哪些商品或 SKU 是测试用，只允许测试账号购买。

| 字段 | 说明 |
| --- | --- |
| `product_id` | 被保护的商品 |
| `variant_id` | 可选，限制到具体 SKU |
| `is_test_only` | 是否仅测试账号可购买 |
| `allowed_test_account_id` | 可选；为空表示所有有效测试账号可买 |
| `enabled` | 是否启用 |
| `starts_at` / `ends_at` | 生效窗口 |
| `hold_fulfillment` | 默认暂停真实发货 |
| `reason` | 为什么设为测试商品 |

### `production_test_order_markers`

订单创建后立即写入，用于后续支付、售后、报表、风控统一识别。

| 字段 | 说明 |
| --- | --- |
| `order_id` | 关联订单 |
| `order_number` | 订单号快照 |
| `user_id` | 下单用户 |
| `test_account_id` | 使用的测试授权 |
| `status` | `active` / `reconciled` / `voided` |
| `hold_fulfillment` | 是否暂停真实履约 |
| `exclude_from_revenue` | 是否排除营收 |
| `exclude_from_analytics` | 是否排除行为和转化统计 |
| `reason` | 本次测试目的 |

### `production_test_audit_events`

记录配置变更和关键链路动作，例如启用测试商品、下单、支付成功、申请售后、发起退款、关闭测试单。

## 订单链路

```text
用户进入商品页/购物车
  -> 商品查询时过滤 test-only 商品，普通用户不可见或不可加购
  -> 结算前调用 productiontest.EvaluatePurchase
  -> 普通用户购买 test-only 商品：阻止并返回业务错误
  -> 测试账号购买 test-only 商品：允许创建订单
  -> 创建订单后写 production_test_order_markers
  -> 后续支付、退款、售后、报表通过 marker 判断是否测试单
```

纯判断接口的输入/输出约定：

| 判断 | 输入 | 结果 |
| --- | --- | --- |
| `EvaluatePurchase` | 用户 ID、可选测试授权、商品/SKU gate、判断时间 | 是否允许购买、是否需要测试单标记、拒绝原因 |
| `EvaluateOrder` | 用户 ID、可选测试授权、购物车各行 gate、判断时间 | 任一测试商品通过时必须标记；任一行拒绝则整单拒绝 |

关键规则：

- 判断必须在后端执行，前端隐藏只能作为体验优化。
- 同一购物车里只要包含测试商品，订单就必须标记为生产测试单。
- 测试账号购买普通商品默认不算生产测试单，避免误把真实订单排除。
- 测试商品可以使用极小金额，但金额计算、币种、税费、支付回调必须走真实逻辑。
- 默认 `hold_fulfillment=true`，除非负责人明确允许发货验证。

## 支付链路

支付应走真实 provider，例如 PayPal、Stripe、支付宝或微信的生产配置，但要做三件事：

- 创建支付前把订单 marker 传入支付服务上下文。
- 支付交易、provider 回调、风控事件和争议证据都附带 `production_test` 标签。
- 支付成功后不触发真实发货、佣金、积分、营销自动化等不可逆副作用，除非该测试场景明确开启。

建议优先使用小额测试商品来验证：

- 支付按钮和弹窗样式。
- 金额、币种、税费、折扣展示。
- provider return/cancel 页面。
- 支付成功、失败、超时和回调重试。
- 风控拦截和人工复核提示。

## 售后和退款链路

售后应该允许对测试订单发起，因为这正是线上最难被测试环境覆盖的部分。

验证范围：

- 售后入口是否正确展示。
- 可申请商品、数量、原因、图片上传和提示文案是否正确。
- 退款金额上限、运费、税费、优惠抵扣是否正确。
- 后台审核、拒绝、补充资料、通过、退款执行的状态流转是否正确。
- 支付 provider 的真实退款回调是否能反写订单和售后状态。

限制规则：

- 测试退款必须只针对测试单。
- 测试售后附件应按普通附件安全规则扫描和访问控制。
- 测试退款完成后应进入对账/关闭流程，避免长期留在异常列表。

## 后台产品能力

后台商品管理建议新增一个“生产测试”面板，而不是把字段散落在普通商品编辑表单里。

建议控件：

- `测试用商品` 开关。
- 测试范围：整个商品 / 指定 SKU。
- 允许账号：所有有效测试账号 / 指定测试账号。
- 生效时间：开始时间、结束时间。
- 履约策略：默认暂停真实发货，可显式允许发货验证。
- 原因和负责人：必填。

后台订单、支付、售后列表都应该有明显的 `生产测试` 标签，并提供筛选条件。

## 报表与副作用隔离

测试单必须从以下业务结果里默认排除：

- 营收、客单价、转化率、复购率。
- 库存销量和补货建议。
- 会员积分、等级成长值、优惠券发放。
- 佣金、返利、联盟营销。
- 客服 SLA 和售后质量统计。
- 商品推荐、热销排序、SEO 结构化销量数据。

可以保留在技术观测里：

- 支付成功率。
- 回调耗时和失败率。
- 售后状态流转耗时。
- 前端错误、弹窗曝光、接口错误。

## 发布阶段

1. 领域层：新增 `productiontest` 域，完成准入判断、测试账号、商品 gate、订单 marker 类型。
2. 数据层：新增 migrations 和 repository，保存测试账号、测试商品 gate、订单 marker 和审计事件。
3. 后台层：新增生产测试管理页，支持账号、商品 gate、订单筛选和审计查看。
4. 订单层：商品查询、购物车、下单前接入准入判断；订单创建后写 marker。
5. 支付层：支付创建、回调、退款和风控事件都读取 marker 并打标。
6. 售后层：售后申请、审核、退款执行和附件链路支持测试单标记。
7. 验收层：补一条生产 smoke runbook，用固定账号和固定测试商品跑完整购买、退款、售后闭环。

## 验收标准

- 普通用户无法看到或购买 `test-only` 商品。
- 有效测试账号可以购买允许范围内的 `test-only` 商品。
- 过期或停用测试账号无法购买测试商品。
- 测试订单创建后一定有 `production_test_order_markers`。
- 支付交易、退款记录、售后记录、后台列表都能识别测试单。
- 测试单默认不进入发货、营收、积分、佣金和推荐统计。
- 每一次配置变更和测试执行都有审计记录。
