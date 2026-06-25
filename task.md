# 任务列表

- [x] 在 `nuxt-i18n` 中安装 `pinia` 和 `@pinia/nuxt` 依赖。
- [x] 修改 `nuxt-i18n/nuxt.config.ts`，将 `'@pinia/nuxt'` 添加到 `modules` 数组中。
- [x] 重构 `nuxt-i18n/app/components/AuthModal.vue`，将 `modelValue` 和 `update:modelValue` 事件替换为 Vue 3.4+ 的 `defineModel` 宏。
- [x] 重构 `nuxt-i18n/app/components/WishlistDrawer.vue`，将 `modelValue` 和 `update:modelValue` 事件替换为 Vue 3.4+ 的 `defineModel` 宏。
- [x] 在终端运行 `npm run typecheck` 进行编译和类型检查，确保 `defineModel` 和 Pinia 运行正常。
- [x] 移除 `useAuth.ts` 中手动保存 JWT 到 cookie/localStorage 的逻辑，改为依赖后端设置的 `HttpOnly` cookie。
- [x] 在 `AuthModal.vue` 中引入 `zod`，为邮箱和密码添加严格的客户端验证。
- [x] 运行类型检查或构建，确保无编译错误。
