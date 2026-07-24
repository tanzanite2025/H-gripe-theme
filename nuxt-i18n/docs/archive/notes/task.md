# 任务拆解 (Tanzanite 优化计划阶段 2 和 3)

## 任务目标
执行前端性能与 SSR 修复优化，具体包括延迟加载全局重组件、移除静态引入、修复 SSR 数据拉取等。

## Check-list 检查项

- [ ] **1. 全局组件按需延迟加载 (`app.vue`)**
  - [ ] 检查并确认目标组件的位置。
  - [ ] 将 `CartDrawer` 修改为 `<LazyCartDrawer />`。
  - [ ] 将 `CheckoutModal` 修改为 `<LazyCheckoutModal />`。
  - [ ] 将 `WhatsAppChatModal` 修改为 `<LazyWhatsAppChatModal />`。
  - [ ] 将 `ShopSearchSheet` 修改为 `<LazyShopSearchSheet />`。

- [ ] **2. 移除 AuthModal 的静态引入并使用懒加载**
  - [ ] 修改 `app/components/WhatsAppChatModal.vue`：移除 `import AuthModal...`，替换模板中对应的标签为 `<LazyAuthModal />`。
  - [ ] 修改 `app/components/feedback/SuggestionForm.vue`：移除静态引入，替换为 `<LazyAuthModal />`。
  - [ ] 修改 `app/components/MembershipAndPointsTabs.vue`：移除静态引入，替换为 `<LazyAuthModal />`。

- [ ] **3. 重构 SSR 数据获取逻辑 (`app/pages/shop.vue`)**
  - [ ] 移除 `onMounted` 中的 `$fetch` 数据获取。
  - [ ] 引入并使用 `useAsyncData('shop-products', () => $fetch(...))` 在 setup 根层级实现 SSR 数据拉取。
  - [ ] 添加 `watch` 监听路由的 query 变化，并调用 `refresh()` 以保证参数变化时数据的响应式更新。

- [ ] **4. 编译与测试**
  - [ ] 执行 `npm run build` 命令。
  - [ ] 确保无编译及 TypeScript 类型报错。
