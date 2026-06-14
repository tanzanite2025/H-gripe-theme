# PHP → Go 迁移工作流

## 当前原则

本项目仍处在 WordPress/PHP 到 Go 的迁移过程中。根目录 WordPress 主题壳文件已经删除；剩余 PHP 目录保留的意义是兼容旧线上能力和查询旧业务逻辑，不是新增功能入口。

硬性规则：

1. 不恢复根目录 WordPress 主题壳文件，不在 `wp-plugin/**` 添加、修改、回调核心业务逻辑。
2. 所有新业务逻辑进入 `go-backend`。
3. 当前管理后台主线是 `go-backend/web/admin`，不是 `go-backend/admin-panel`。
4. Nuxt 前台主线是 `nuxt-i18n`，最终只能走 Go API；现存 `/wp-json/**` 调用按模块逐步替换。
5. 做完一个模块就停下来开 PR，不一次性做完多个模块。

## 单模块 PR 契约

每个 PR 只能选择一个模块，模块边界按以下方式定义：

- 一个业务域：如 FAQ、Gallery、Subscription、Ticket、Wishlist、Product、Order。
- 或一个基础设施域：如 Go server 启动、配置、文档入口、CI、部署脚本。
- 或一个迁移准备域：如 API 矩阵、数据字段映射、前端调用清单。

每个 PR 必须写清：

- 本 PR 做什么。
- 本 PR 明确不做什么。
- 是否修改 PHP；如果修改，必须是删除/屏蔽旧路径或迁移辅助，不得新增 PHP 业务。
- 本地验证命令。
- 下一模块建议。

## 已完成基础模块

### P0 Go 后端基座

状态：已通过独立 PR 合入。

范围：

- `cmd/server/main.go`
- `/health`
- 配置默认值和环境变量覆盖
- 可选 `AutoMigrate`
- `go test ./...` 编译路径修正

不包含：

- 任何业务接口迁移
- Nuxt 调用切换
- PHP 删除
- API 迁移矩阵

## 下一步顺序

按以下顺序逐个模块推进，每步完成后停止并开独立 PR。

### D1a 文档入口与过期文档标记

目标：

- 固化当前文档入口。
- 标记过期/历史参考文档。
- 统一“单模块 PR”规则。

验收：

- 开发者能从文档入口知道哪些文档有效、哪些只能参考。
- 下一步顺序和防跑偏检查表清晰可见。

### D1b PHP API → Go API 迁移矩阵

状态：当前文档在 `docs/PHP_TO_GO_API_MATRIX.md`。

目标：

- 建 PHP API → Go API 迁移矩阵。
- 建 Nuxt `/wp-json/**` 调用清单。

验收：

- 能明确每个旧 PHP endpoint 的 Go 去向。
- 能明确下一个最小可迁移业务模块。

### M1 低风险只读模块

推荐顺序：

1. FAQ
2. Gallery
3. Site settings / quick-buy settings
4. Blog/content read APIs
5. Subscription submit/list

原因：读多写少，交易风险低，适合验证 Go API 与 Nuxt 切流流程。

### M2 用户轻交互模块

推荐顺序：

1. Wishlist
2. Review
3. Product registration / warranty
4. Feedback / suggestion
5. Ticket / customer service history

原因：涉及登录态和用户数据，但不直接影响支付下单。

### M3 交易链路模块

推荐顺序：

1. Product catalog
2. Cart
3. Coupon / gift card / loyalty
4. Shipping / tax
5. Order
6. Payment

原因：需要数据对账、端到端下单验证和回滚方案。

### M4 PHP 物理删除

前置条件：

- Nuxt 不再调用对应 `/wp-json/**`。
- Go API 覆盖旧能力。
- 数据迁移脚本 dry-run 和导入校验完成。
- 线上 Nginx/CDN 不再把该模块请求转发到 WordPress。

删除策略：

- 先屏蔽旧 route/menu。
- 再删除旧插件模块文件。
- 根目录主题壳已删除；剩余删除工作只针对仍需兼容的旧插件模块。

## 防跑偏检查表

开始任何模块前先回答：

- 当前模块名称是什么？
- 本 PR 是否只做这一个模块？
- 是否触碰 PHP？触碰原因是什么？
- Nuxt 是否还存在旧 `/wp-json/**` 调用？
- 需要数据迁移还是只切 API？
- 验证命令是什么？

如果回答不清楚，先补迁移矩阵或文档，不写业务代码。
