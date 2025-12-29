# 🎨 Navigation Style Analysis & Unification Plan

## 🚨 现状分析 (Current Status)

经过对关键组件的代码审查，确认存在以下 **样式重复定义** 和 **硬编码颜色** 问题，这直接导致了维护困难和潜在的视觉不一致。

### 1. 横向顶部导航 (Top Nav Bar)

涉及组件：

- `ProductsTopNav.vue`
- `SupportTopNav.vue`

**问题**：

- 两者完全复制了一套 CSS 代码。
- **Hardcoded Colors**:
  - 背景：`rgba(15, 23, 42, 0.92)`
  - 边框：`rgba(148, 163, 184, 0.3)`
  - 激活色 (Cyan): `#38bdf8` (对应 Tailwind `sky-400`)
  - 悬停色：`#e5f2ff` (对应 Tailwind `sky-50`)
- **风险**：修改其中一个（例如调整透明度），很容易忘记修改另一个，导致 Support 和 Products 区域顶栏不一致。

### 2. 胶囊式选项卡 (Pill Tabs)

涉及组件：

- `MembershipAndPointsTabs.vue`
- `PoliciesTabs.vue`

**问题**：

- class 名 `company-tabs__item` 在两个文件中被重复定义。
- 样式完全一致：圆角 `9999px`，背景 `rgba(31, 41, 55, 0.9)`，激活变白。
- 若未来设计要求改变胶囊圆角或颜色，需多处修改。

### 3. 全局按钮与工具类

- `MembershipAndPointsTabs.vue` 内部定义了 `.btn-gradient` (`linear-gradient(to right, #40ffaa, #6b73ff)`)。
- 这种通用按钮样式应该提取为全局组件或工具类。

---

## 🛠️ 优化方案 (Proposed Solution)

建议分两步走，不直接重写所有逻辑，而是抽取样式。

### 方案 A：提取全局 CSS 组件类 (推荐)

在 `app/assets/css` 下新建 `components/nav.css`，定义标准类：

```css
/* Top Nav Bar (Products / Support) */
.nav-top-bar {
  @apply w-full border-b border-slate-400/30 bg-slate-900/90 backdrop-blur-md;
}
.nav-top-bar__item {
  @apply text-white border-b-[3px] border-transparent hover:text-sky-50;
}
.nav-top-bar__item--active {
  @apply text-white font-semibold border-sky-400;
}

/* Pill Tabs (Membership / Policies) */
.nav-pill-tabs {
  @apply flex gap-3 overflow-x-auto;
}
.nav-pill-item {
  @apply rounded-full px-5 py-2 text-sm font-medium text-white bg-gray-800/90 backdrop-blur-sm transition-all;
}
.nav-pill-item--active {
  @apply bg-white text-slate-900 font-semibold shadow-lg;
}
```

### 方案 B：封装 Vue 组件

创建 `BaseTopNav.vue` 和 `BasePillTabs.vue` 组件，但这涉及 HTML 结构的变动，风险略高。**建议先执行方案 A（样式统一）**。

## ✅ 执行计划 (Next Steps)

1. **创建文件**: 新建 `app/assets/css/components.css` 并引入 `nuxt.config.ts`。
2. **定义样式**: 将上述 CSS 提取进去，使用 Tailwind `@apply` 语法以保持与设计系统的颜色一致性（使用 `sky-400` 而不是 `#38bdf8`）。
3. **重构组件**:
    - 修改 `ProductsTopNav` / `SupportTopNav` 引用新类名。
    - 修改 `MembershipAndPointsTabs` / `PoliciesTabs` 引用新类名。
4. **验证**: 确保视觉效果 1:1 还原，但代码通过 Tailwind 统一了色值。
