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
