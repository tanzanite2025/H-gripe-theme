# [Safe Style Refactoring] Unifying Navigation Styles

## 目标

统一导航组件（Top Nav & Pill Tabs）的样式定义到全局 CSS 文件中，以消除重复代码并统一视觉规范。
**采用“温和替换”策略**：先引入新样式，确保生效且无误后，再清理旧样式，避免破坏现有功能。

## User Review Required
>
> [!IMPORTANT]
> 这是一个纯样式重构。我将创建一个新的 CSS 文件 `nav.css` 并将其添加到 Nuxt 配置中。
> **关键策略**：组件文件中的 `<style scoped>` 将暂时保留，但我会通过修改 template 中的 `class` 属性来优先使用全局类。确认无误后，我们再在后续的 Cleanup 阶段删除组件内的 redundant styles。

## Proposed Changes

### 1. Global CSS Definition [NEW]

#### [NEW] [nav.css](file:///c%3A/Users/P16V/Desktop/Wordpress/tanzanite-theme/nuxt-i18n/app/assets/css/components/nav.css)

- 定义 `.nav-top-bar`, `.nav-top-bar__item` (for Products/Support nav)
- 定义 `.nav-pill-tabs`, `.nav-pill-item` (for Membership/Policies tabs)
- 使用 Tailwind `@apply` 确保颜色值与设计系统一致。

### 2. Nuxt Configuration

#### [MODIFY] [nuxt.config.ts](file:///c%3A/Users/P16V/Desktop/Wordpress/tanzanite-theme/nuxt-i18n/nuxt.config.ts)

- 在 `css` 数组中引入 `~/assets/css/components/nav.css`。

### 3. Component Updates (Incremental)

对于以下每个组件，我们将执行：

1. **Template**: 将硬编码的 BEM 类名（如 `.products-top-nav`）替换或追加全局类名（`.nav-top-bar`）。
2. **Script**: 暂时保持不变。
3. **Style**: **暂时保留** 原有代码，防止因 CSS 权重问题导致崩坏，或者作为 fallback。

#### [MODIFY] [ProductsTopNav.vue](file:///c%3A/Users/P16V/Desktop/Wordpress/tanzanite-theme/nuxt-i18n/app/components/ProductsTopNav.vue)

- Apply `.nav-top-bar` & `.nav-top-bar__item` classes.

#### [MODIFY] [SupportTopNav.vue](file:///c%3A/Users/P16V/Desktop/Wordpress/tanzanite-theme/nuxt-i18n/app/components/SupportTopNav.vue)

- Apply `.nav-top-bar` & `.nav-top-bar__item` classes.

#### [MODIFY] [MembershipAndPointsTabs.vue](file:///c%3A/Users/P16V/Desktop/Wordpress/tanzanite-theme/nuxt-i18n/app/components/MembershipAndPointsTabs.vue)

- Apply `.nav-pill-tabs` & `.nav-pill-item` classes.

#### [MODIFY] [PoliciesTabs.vue](file:///c%3A/Users/P16V/Desktop/Wordpress/tanzanite-theme/nuxt-i18n/app/components/PoliciesTabs.vue)

- Apply `.nav-pill-tabs` & `.nav-pill-item` classes.

## Verification Plan

### Manual Verification

- **Visual Check**:
  - 检查 `/products` 顶部导航是否保持半透明背景和 Cyan 激活色。
  - 检查 `/policies/privacy` 胶囊 tab 是否保持圆角和 hover 效果。
  - 检查 `/membershipandpoints` 胶囊 tab 是否正常。
- **Conflict Check**:
  - 确保在此阶段，新添加的全局类名优先级足够（Tailwind `@apply` 生成的 CSS 通常具有合理的 specificity，但我们会手动检查是否需要 `!important` 覆盖原有 scoped 样式，或者通过移除 scoped 样式中的属性来验证）。
  - *Note*: 为了验证新样式确实生效，我可能会在本地微调颜色（极不明显地）或者通过浏览器开发者工具确认 class 来源是 `nav.css`。

### Safe Rollback

- 如果发现任何样式错乱，只需回滚组件 template 的 class 修改，组件即恢复使用内部 scoped 样式。
