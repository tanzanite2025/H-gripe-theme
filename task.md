# 任务列表

- [x] 在 `nuxt-i18n` 中安装 `pinia` 和 `@pinia/nuxt` 依赖。
- [x] 修改 `nuxt-i18n/nuxt.config.ts`，将 `'@pinia/nuxt'` 添加到 `modules` 数组中。
- [x] 重构 `nuxt-i18n/app/components/AuthModal.vue`，将 `modelValue` 和 `update:modelValue` 事件替换为 Vue 3.4+ 的 `defineModel` 宏。
- [x] 重构 `nuxt-i18n/app/components/WishlistDrawer.vue`，将 `modelValue` 和 `update:modelValue` 事件替换为 Vue 3.4+ 的 `defineModel` 宏。
- [x] 在终端运行 `npm run typecheck` 进行编译和类型检查，确保 `defineModel` 和 Pinia 运行正常。
