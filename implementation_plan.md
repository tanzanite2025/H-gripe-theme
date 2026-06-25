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
