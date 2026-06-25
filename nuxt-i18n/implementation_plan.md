# 技术实现方案 (Tanzanite 优化计划阶段 2 和 3)

## 涉及修改的文件列表
1. `app/app.vue`
2. `app/components/WhatsAppChatModal.vue`
3. `app/components/feedback/SuggestionForm.vue`
4. `app/components/MembershipAndPointsTabs.vue`
5. `app/pages/shop.vue`

## 详细技术方案

### 1. 全局组件的懒加载 (`app/app.vue`)
- **操作**：由于全局引入组件可能导致首次加载包体积过大，Nuxt 允许通过 `Lazy` 前缀实现按需加载。在 `app.vue` 中查找相关组件 (`CartDrawer`, `CheckoutModal`, `WhatsAppChatModal`, `ShopSearchSheet`) 的标签。
- **修改**：将这些标签统一加上 `Lazy` 前缀（例如：`<LazyWhatsAppChatModal />`）。

### 2. `AuthModal` 组件按需加载改造
- **操作**：移除对 `AuthModal` 的强制静态引入。由于 `AuthModal` 通常是在用户交互后才渲染，通过模板懒加载可以显著减少初始加载耗时。
- **修改**：在涉及的文件中（`WhatsAppChatModal.vue`, `SuggestionForm.vue`, `MembershipAndPointsTabs.vue`），在 `<script>` 内移除 `import AuthModal from '...'`。在 `<template>` 内将 `<AuthModal>` 全部替换为 `<LazyAuthModal>`。

### 3. 重构页面数据获取以支持 SSR (`app/pages/shop.vue`)
- **当前问题**：`shop.vue` 中数据的获取使用了纯客户端的 `$fetch` 并包裹在 `onMounted` 钩子内，这导致服务器端渲染 (SSR) 时不会返回页面产品数据，对 SEO 和首屏渲染非常不利。
- **修改**：
  - 将 `$fetch` 移出 `onMounted`，放置到 `<script setup>` 的顶层。
  - 使用 Nuxt 3 的原生 SSR 数据获取组合式 API：`const { data, refresh } = await useAsyncData('shop-products', () => $fetch(...))`。
  - 由于路由查询参数（例如分类或分页）可能会发生变更，为了保持状态同步，使用 `watch(() => route.query, () => { refresh() })` 或直接在 `useAsyncData` 的 `watch` 选项中配置以便在 query 更新时自动重取数据。

## 潜在风险和破坏性变更
- **按需加载组件未挂载时的引用丢失**：若涉及 `ref` 引用这些被改为懒加载的组件，可能会在初始化时报 `undefined` 错误。由于本次变更主要是弹窗组件（如 Modal 和 Drawer），这类组件一般由全局状态控制展示/隐藏，风险较低。
- **SSR 数据获取的依赖冲突**：转移至 `useAsyncData` 后，如果有依赖仅限客户端的变量（如 `window`, `localStorage`），会在服务端渲染时报错。需要确保 API 的参数均可从服务端（如 `route.query`）中安全获取。
- **页面响应式表现变化**：`useAsyncData` 会阻止页面导航直到数据获取完成，这可能会影响路由跳转时的用户感知（可以通过 `pending` 状态处理加载动画）。

## 测试验证
所有更改完成后，将在根目录执行 `npm run build`，以验证代码的编译是否通过，并确认无 TS 类型不匹配问题。
