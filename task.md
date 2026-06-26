# 任务列表

- [x] 在 `nuxt-i18n` 中安装 `pinia` 和 `@pinia/nuxt` 依赖。
- [x] 修改 `nuxt-i18n/nuxt.config.ts`，将 `'@pinia/nuxt'` 添加到 `modules` 数组中。
- [x] 重构 `nuxt-i18n/app/components/AuthModal.vue`，将 `modelValue` 和 `update:modelValue` 事件替换为 Vue 3.4+ 的 `defineModel` 宏。
- [x] 重构 `nuxt-i18n/app/components/WishlistDrawer.vue`，将 `modelValue` 和 `update:modelValue` 事件替换为 Vue 3.4+ 的 `defineModel` 宏。
- [x] 在终端运行 `npm run typecheck` 进行编译和类型检查，确保 `defineModel` 和 Pinia 运行正常。
- [x] 移除 `useAuth.ts` 中手动保存 JWT 到 cookie/localStorage 的逻辑，改为依赖后端设置的 `HttpOnly` cookie。
- [x] 在 `AuthModal.vue` 中引入 `zod`，为邮箱和密码添加严格的客户端验证。
- [x] 运行类型检查或构建，确保无编译错误。

## Level 7 Backend Plan

- [x] 在 `go-backend` 目录下，安装 `gorilla/websocket` 依赖。
- [x] 创建 `internal/api/v1/ticket/hub.go` 或在现有 handler 中添加 WebSocket 的升级逻辑，配置一个并发安全的 Hub。
- [x] 在 `go-backend` 目录下，安装 `github.com/sashabaranov/go-openai` 和 `github.com/pgvector/pgvector-go`。
- [x] 在 `internal/repository/product_repository.go` 中，添加 `SemanticSearchPublic(ctx context.Context, query string) ([]domain.Product, error)` 占位方法。
- [x] 在 `go-backend` 目录下运行 `go mod tidy` 和 `go build ./...` 验证编译。

## Level 9 Error & Health Plan

- [x] 在 `go-backend` 中创建 `internal/api/v1/apierror/error.go`。
- [x] 定义 `AppError` 结构体，包含 `Code`、`Message` 和 `StatusCode`。
- [x] 编写 `Send(c *gin.Context, err error)` 辅助函数，处理 `AppError` 并隐藏非 `AppError` 的内部细节（默认返回 500）。
- [x] 更新 `internal/api/v1/auth/handler.go`，使用新的 `apierror.Send` 替换原有的错误返回。
- [x] 更新 `internal/api/v1/router.go`，修改 `/health` 接口，增加数据库（`db.DB().Ping()`）和 Redis（`redis.Client.Ping()`）连通性检查，任意失败返回 503。
- [x] 在 `internal/api/v1/router.go` 中添加 `/ready` 接口，进行相同的连通性检查。
- [x] 运行 `go mod tidy` 和 `go build ./...` 验证编译。
## Level 9 Infrastructure Plan

- [x] 在 `go-backend` 目录，安装 `github.com/golang-migrate/migrate/v4`。
- [x] 更新 `internal/pkg/database/migrate.go`，禁用生产环境下的 GORM `AutoMigrate` 或提供警告，并配置使用 `migrate.New` 执行 `migrations/` 中的 SQL。
- [x] 在 `go-backend` 目录，安装 `github.com/hibiken/asynq`。
- [x] 创建 `internal/pkg/worker/worker.go`（Asynq Server）。
- [x] 创建 `internal/pkg/worker/client.go`（Asynq Client）。
- [x] 在 `cmd/server/main.go` 中集成 worker server 的启停逻辑，与 HTTP 服务一同实现优雅关闭。
- [x] 运行 `go mod tidy` 和 `go build ./...` 验证编译通过。
