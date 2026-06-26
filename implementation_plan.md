# 实施计划

## 1. 安装依赖并配置模块
- 进入 `nuxt-i18n` 目录，通过命令 `npm install pinia @pinia/nuxt` 安装依赖。
- 修改 `nuxt-i18n/nuxt.config.ts` 文件，将 `@pinia/nuxt` 添加至 `modules` 数组中。

## 2. 引入 `defineModel` 宏进行重构
- **重构组件一：`AuthModal.vue`**
  - 在 `<script setup>` 中移除原有的 `props` 定义中的 `modelValue`，以及 `emit` 定义中的 `'update:modelValue'`。
  - 使用 Vue 3.4+ 的 `defineModel` 定义模型：`const modelValue = defineModel<boolean>({ default: false })`。
  - 将所有原有的 `emit('update:modelValue', false)` 改为直接修改 `modelValue.value = false`，并修改 `watch` 中的参数监控。
- **重构组件二：`WishlistDrawer.vue`**
  - 同样移除 `props.modelValue` 和对应的 `emit`。
  - 使用 `const modelValue = defineModel<boolean>()`。
  - 修改 `handleClose` 等方法中的 `emit` 调用，变为对 `modelValue.value` 的直接更新。

## 3. 验证构建与类型检查
- 运行 `npm run typecheck`。
- 确认由于依赖安装和代码重构导致的所有可能报错已被解决。

## 4. DevSecOps 前端安全增强 (新)
- **Token 安全策略调整**：
  - 修改 `nuxt-i18n/app/composables/useAuth.ts`。
  - 移除通过 `useCookie` 或 `localStorage` 手动保存和管理 `auth_token` 的代码。
  - 移除请求头中手动注入 `Authorization: Bearer <token>` 的逻辑（依赖后端 `HttpOnly` Cookie）。
- **Zod 客户端验证**：
  - 在项目内安装 `zod`（如果尚未安装）。
  - 在 `nuxt-i18n/app/components/AuthModal.vue` 中导入 `zod`，并针对邮箱、密码字段构建模式（例如 `z.string().email()` 以及 `z.string().min(8).regex(/[A-Z]/)`）。
  - 拦截表单提交：在发送后端请求前，执行本地验证，如果验证失败，直接在 UI 显示报错（利用原有的 `error` 状态展示），防止触发不必要的 API 请求。
- **最终验证**：
  - 进入 `nuxt-i18n` 目录运行 `npm run typecheck` 或 `npm run build`。

## 5. Level 7 Backend Plan
- **WebSocket 实时支持**:
  - `cd go-backend`
  - 运行 `go get github.com/gorilla/websocket`。
  - 创建 `internal/api/v1/ticket/hub.go`，建立一个并发安全的 WebSockets Hub (`Hub` struct，包含客户端连接的 map、注册和注销通道、广播通道)。
  - 提供 HTTP 到 WebSocket 的 Upgrade 端点（例如 `/api/v1/ws`）。
- **AI 语义搜索存根**:
  - 运行 `go get github.com/sashabaranov/go-openai github.com/pgvector/pgvector-go`。
  - 修改 `internal/repository/product_repository.go`，增加 `SemanticSearchPublic(ctx context.Context, query string) ([]domain.Product, error)` 存根方法，模拟调用 OpenAI 生成嵌入，并使用类似 `ORDER BY embedding <=> ?` 的原生 SQL 语句进行 pgvector 搜索。
- **验证编译**:
  - `cd go-backend`
  - 运行 `go mod tidy`
  - 运行 `go build ./...`，确保项目编译无误。

## 5.5 Level 9 Error & Health Plan
- **统一错误响应**:
  - `cd go-backend`
  - 创建 `internal/api/v1/apierror/error.go`。
  - 定义 `AppError` 结构体，包含 `Code`、`Message` 和 `StatusCode` 字段。
  - 编写 `Send(c *gin.Context, err error)` 辅助函数：判断错误是否为 `*AppError`，是则返回相应的状态码和 JSON 响应；否则返回 500 并屏蔽内部细节。
  - 修改 `internal/api/v1/auth/handler.go`（或类似核心 handler），将其使用 `c.JSON(400, gin.H{"error": err.Error()})` 的地方替换为 `apierror.Send`。
- **深度健康检查**:
  - 更新 `internal/api/v1/router.go`（或定义了 `/health` 接口的地方）。
  - 在 `/health` 接口中添加对数据库（`db.DB().Ping()`）和 Redis（`redis.Client.Ping()`）的检查。如果任何一个失败，则返回 HTTP 503 Service Unavailable。
  - 新增 `/ready` 接口，执行相同的健康和连通性检查。
- **验证编译**:
  - `cd go-backend`
  - 运行 `go mod tidy` 和 `go build ./...` 确保编译通过。

## 6. Level 9 Infrastructure Plan
- **Database Migrations**:
  - `cd go-backend` 并运行 `go get github.com/golang-migrate/migrate/v4`
  - 修改 `internal/pkg/database/migrate.go`：如果环境是 production，则默认禁用 GORM `AutoMigrate`（或仅做 warning 打印）；增加逻辑通过 `migrate.New` 等方式读取 `migrations/` 目录执行 SQL 迁移。
- **Async Job Queue**:
  - `cd go-backend` 并运行 `go get github.com/hibiken/asynq`
  - 创建 `internal/pkg/worker/worker.go`：封装 Asynq Server 的启动和任务处理逻辑。
  - 创建 `internal/pkg/worker/client.go`：封装 Asynq Client，用于生产者投递任务。
  - 修改 `cmd/server/main.go`：初始化 Redis 连接给 Asynq Server，将 worker server 与原有的 HTTP Server 并排启动，监听退出信号时一并优雅关闭 (`server.Shutdown()`)。
- **验证编译**:
  - `cd go-backend`
  - 运行 `go mod tidy` 和 `go build ./...`。
