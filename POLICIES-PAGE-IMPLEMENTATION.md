# Policies 页面实施指南（样式全复用）

## 目标

在 `tanzanite-theme/nuxt-i18n`（Nuxt）中实现一组 Policies 页面：

- **总入口页**：`/policies`
  - 以“块/卡片”形式展示 4 个子页面入口
- **子页面（带横向 TAB）**：
  - `/policies/cookie`
  - `/policies/privacy`
  - `/policies/refund-return`
  - `/policies/terms`

## 核心要求（重要）

- **样式全复用**：完全沿用现有页面的视觉风格（`products` layout + pills tab 样式 + 卡片网格样式）。
- **仅允许变化**：
  - 路由（slug）
  - TAB 名称（cookie / privacy / refund-return / terms）
  - 每个页面正文内容

---

## 路由与文件对照

| TAB 名称 | slug | 路由 | 目标文件 |
|---|---|---|---|
| cookie | `cookie` | `/policies/cookie` | `nuxt-i18n/app/pages/policies/cookie.vue` |
| privacy | `privacy` | `/policies/privacy` | `nuxt-i18n/app/pages/policies/privacy.vue` |
| refund-return | `refund-return` | `/policies/refund-return` | `nuxt-i18n/app/pages/policies/refund-return.vue` |
| terms | `terms` | `/policies/terms` | `nuxt-i18n/app/pages/policies/terms.vue` |

---

## 仓库现状（你现在已经有了大部分骨架）

已存在：

- `nuxt-i18n/app/pages/policies/` 下的 `cookie/privacy/refund-return/terms/index` 页面文件
- `nuxt-i18n/app/components/PoliciesTabs.vue` **存在但为空文件**（当前 TAB 实际渲染不出来）
- `nuxt-i18n/app/components/AppFooter.vue` 已有 `/policies/*` 链接入口

因此，本任务主要是：

1. **补齐 `PoliciesTabs.vue`（TAB 横向菜单）**
2. **把 `/policies` 改成“总页卡片入口”**（而不是重定向到某个 policy）
3. **确保 4 个子页面都统一使用 `layout: 'products'` 并引用 `PoliciesTabs`**

---

## 样式复用说明（明确写给未来维护者）

- **Layout 复用**：所有 policies 页面使用 `definePageMeta({ layout: 'products' })`，与 `company/*`、`guides/*` 完全一致。
- **TAB 样式复用**：TAB 的 pills 样式复用现有的 `.company-tabs` / `.company-tabs__item` / `.company-tabs__item--active`。
  - 可直接从 `nuxt-i18n/app/pages/company/about.vue` 或 `nuxt-i18n/app/pages/policies/index.vue` 复制对应 CSS（保持数值不变）。
- **卡片入口样式复用**：`/policies` 总页的卡片网格复用 `nuxt-i18n/app/pages/company/index.vue` 的 `company-grid/company-card` 视觉（同样保持 CSS 数值不变）。

---

## 实施步骤

### Step 1：实现 `PoliciesTabs.vue`（TAB 横向菜单组件）

文件：`nuxt-i18n/app/components/PoliciesTabs.vue`

建议实现要点：

- 固定 tabs 配置：
  - `cookie/privacy/refund-return/terms`
- 用 `NuxtLink` 跳转：`/policies/${slug}`
- 当前高亮：
  - 方案 A：每个子页传 `current-slug`（当前项目已这么写）
  - 方案 B：组件内部 `useRoute()` 自动推导（可选）
- 组件内写一份 scoped CSS：**原样复制** `.company-tabs` 相关样式，保证“完全一致”。

完成后，你不需要在每个子页手写 TAB UI，只需要 `<PoliciesTabs current-slug="cookie" />`。

### Step 2：将 `/policies` 变为“总页入口（卡片块）”

编辑文件：`nuxt-i18n/app/pages/policies/index.vue`

目标：展示 4 个块（卡片），分别链接到：

- `/policies/cookie`
- `/policies/privacy`
- `/policies/refund-return`
- `/policies/terms`

实现方式（最省事、样式最一致）：

- 直接参考并复用 `nuxt-i18n/app/pages/company/index.vue` 的结构：
  - `sections` 数组 + `v-for` 渲染 `NuxtLink`
  - 复制其 `company-grid/company-card` CSS（保持不改）
- `definePageMeta({ layout: 'products' })`
- `useHead({ title: 'Policies' })`

> 注意：当前 `pages/policies/index.vue` 里有 `<NuxtPage />` 和 `router.replace()` 逻辑，它更像“嵌套路由容器/重定向页”。按你的需求（/policies 是总页），这些逻辑应该移除。

### Step 3：统一 4 个子页面（只保留 TAB + 正文内容）

逐个检查以下文件：

- `pages/policies/cookie.vue`
- `pages/policies/privacy.vue`
- `pages/policies/refund-return.vue`
- `pages/policies/terms.vue`

要求：

- 顶部都有：`<PoliciesTabs current-slug="..." />`
- 都有：`definePageMeta({ layout: 'products' })`
- 都有：`useHead({ title: '...' })`
- **正文内容随你替换**，但外层容器结构（如 `.company-page` / `.policies-content`）保持一致，以复用现有布局与间距。

> 当前仓库里 `terms.vue` 没有 `definePageMeta({ layout: 'products' })`，建议补齐以保证与其它 policies 页一致。

---

## 本地验证（Smoke Test）

- 运行 `pnpm dev`
- 逐个访问确认：
  - `/policies`（应显示 4 个卡片块）
  - `/policies/cookie`（TAB 存在，cookie 高亮）
  - `/policies/privacy`（TAB 存在，privacy 高亮）
  - `/policies/refund-return`（TAB 存在，refund-return 高亮）
  - `/policies/terms`（TAB 存在，terms 高亮）

---

## 验收清单

- `/policies` 是总页入口（卡片快捷入口），不是自动跳转到某个子页
- 四个子页的 TAB 样式完全一致（pills），并且激活态正确
- 所有 policies 页面都使用 `products` layout（样式与 `company/guides` 完全一致）
- 仅路由、TAB 文案、正文内容发生变化（无新增视觉风格）
