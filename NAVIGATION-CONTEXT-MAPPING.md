# 导航上下文映射规范

## 目标
防止主导航下方的横向二级导航（“top tabs”）在用户浏览过程中**意外切换到另一套菜单**。

目标体验（Target UX）：
- 在购物 / 浏览商品的上下文中，始终显示同一套 **购买辅助（purchase-help）** 菜单。
- 在指南（Guides）上下文中，始终显示同一套 **Guides** 菜单。
- Support 与 Company 保持各自的菜单体系不变。

## 当前实现位置（代码分布）
- **Products / Guides / Company** 的横向菜单组件：
  - `nuxt-i18n/app/components/ProductsTopNav.vue`
- **Products 菜单项定义**：
  - `nuxt-i18n/app/utils/productsNav.ts`（`productsNavItems`）
- **Company 菜单项定义**：
  - `nuxt-i18n/app/utils/companyNav.ts`（`companyNavItems`）
- **Support 菜单**：
  - `nuxt-i18n/app/components/SupportTopNav.vue`
  - `nuxt-i18n/app/utils/supportNav.ts`（`supportNavItems`）
- 渲染横向菜单的布局（layouts）：
  - `nuxt-i18n/app/layouts/products.vue`（渲染 `ProductsTopNav`）
  - `nuxt-i18n/app/layouts/support.vue`（渲染 `SupportTopNav`）
 - 导航上下文映射 helper（唯一数据源）：
   - `nuxt-i18n/app/utils/navContext.ts`（`getNavContextFromRoute`）

## 期望的“导航上下文（Navigation Context）”规则
不要再依赖单一的路由相等判断来推断菜单（例如仅在 `path === '/guides'` 时才认为是 guides）。
而是定义一个明确的 **Nav Context Mapping（导航上下文映射）**，基于当前路由选择使用哪一套菜单。

### 上下文类型（Contexts）
- `company`
- `support`
- `guides`
- `blog`
- `products`

### 映射表（草案）
按需要补充/调整：

| 路由模式 | 期望上下文 | 备注 |
| --- | --- | --- |
| `/company` 与 `/company/*` | `company` | 使用 `companyNavItems` |
| `/support` 与 `/support/*` | `support` | 默认使用 `supportNavItems` + `support` layout；特殊情况见下文 Test Report 规则 |
| `/guides` 与 `/guides/*` | `guides` | 默认在 guides 内保持 Guides 菜单；从 Products 带 `?nav=products` 进入时视为 `products` 上下文 |
| `/products` | `products` | Products 聚合页 |
| `/shop` | `products` | 虽是根路由，但必须保持 products 菜单 |
| `/spoke-calculator` | `products` | 虽是根路由，但必须保持 products 菜单 |
| `/membershipandpoints` | `products` | 根路由，但必须保持 products 菜单 |
| `/picture-warehouse` | `products` | 根路由，但必须保持 products 菜单 |
| `/blog` 与 `/blog/*` | `blog` | Blog 区域，必须稳定显示 Blog 菜单，不得显示 Products 菜单 |

## Products 菜单内容（草案）
你希望 products 横向菜单主要解决“用户不知道怎么选商品”的问题（选型/指南/工具入口）。
购买辅助的更重内容（售后/支付/物流等）将放在商品详情页的说明与侧边栏快捷入口中。

### 已确认结论（2025-12-14）
- Products 横向菜单：保留选型相关入口（Shop + Tire Size Charts + Technical + Wheelset Guide + Spoke Calculator + Test Report + Membership and Points + Picture warehouse），不承载 Wheelsbuild blog。
- Products 横向菜单：本期不放 Support 售后入口（Warranty/After sales/Shipping/Payment）。
- Guides 横向菜单：进入 `/guides` 或任意 `/guides/*` 时，必须稳定显示 Guides 菜单，不得切回 Products 菜单。

请确认 Products 下应该显示哪些菜单项（标记 Keep/Remove/Add）：

### 候选菜单项（现有）
- [yes] Shop（`/shop`）（id: `shop`）
- [no] Tools guide（`/guides/tools`）（id: `about-tools`）【仅在 Guides 横向菜单展示】
- [yes] Tire Size Charts（`/guides/sizecharts`）（id: `tire-size-charts`）
- [yes] Technical（`/guides/technical`）（id: `technical-docs`）
- [yes] Wheelset Guide（`/guides/wheelset-buyers`）（id: `wheelset-buyers`）
- [yes] Spoke Calculator（`/spoke-calculator`）（id: `spoke-calculator`）
- [yes] Test Report（`/support/test-report`）（id: `test-report`）【从 Products 进入需保持 Products 横向菜单】
- [yes] Membership and Points（`/membershipandpoints`）（id: `membership-and-points`）
- [yes] Picture warehouse（`/picture-warehouse`）（id: `picture-warehouse`）



需要决策：
- 当用户从 Products 菜单点击 Support 相关链接时：
  - 当前已决定：Products 横向菜单加入 `Test Report`，但从 Products 进入 `/support/test-report` 时必须保持 Products 横向菜单。
    - 实现方式：沿用统一 query 标记 `?nav=products`（避免出现多套实现方式）。
  - 其它 Support 售后入口（Warranty/After sales/Shipping/Payment）本期仍不放进 Products 横向菜单。

## Guides 菜单内容（草案）
Guides 菜单需要在所有 `/guides/*` 路由下保持稳定。

请确认 Guides 下应该显示哪些菜单项：
- [yes] Tools（`/guides/tools`）
- [yes] Tire Size Charts（`/guides/sizecharts`）
- [yes] Technical（`/guides/technical`）
- [yes] Wheelset Buyers Guide（`/guides/wheelset-buyers`）

## Blog 菜单内容（草案）
Blog 菜单需要在所有 `/blog/*` 路由下保持稳定。

请确认 Blog 下应该显示哪些菜单项：
- [yes] All / Blog home（`/blog`）
- [yes] News（`/blog/news`）
- [yes] Wheelsbuild（`/blog/wheelsbuild`）

## 特殊规则：Products -> Guides 页面仍保持 Products 横向菜单
需求背景：
- `/guides/*` 默认属于 Guides 上下文，因此直接访问 `/guides/sizecharts` 时应显示 Guides 横向菜单。
- 但当用户从 Products 横向菜单点击进入这些指南页时（它们被当作“选型/怎么选”的入口），希望仍保持 Products 横向菜单，避免中断。

实现方式（已落地）：
- 当 `ProductsTopNav` 处于 Products 上下文且菜单项 `to` 以 `/guides/` 开头时，跳转链接会自动拼接 query：`?nav=products`。
- `ProductsTopNav` 会读取 `route.query.nav`：
  - 如果 `nav=products`，则即使当前路径是 `/guides/*`，也强制使用 Products 菜单。
  - 如果没有该 query，则 `/guides/*` 仍按 Guides 上下文显示 Guides 菜单。

示例：
- 从 Products 菜单点击 Tire Size Charts：
  - `/guides/sizecharts?nav=products`（保持 Products 横向菜单）
- 直接访问或从顶部主导航进入 Guides：
  - `/guides/sizecharts`（显示 Guides 横向菜单）

维护注意事项（以后往 Products 横向菜单新增入口时）：
- 如果新增入口指向 `/guides/*`，并且你希望它在“从 Products 进入”时保持 Products 横向菜单：
  - 确保该入口是通过 `ProductsTopNav` 渲染（会自动加 `?nav=products`）。

## 特殊规则：Products -> Support Test Report 仍保持 Products 横向菜单
需求背景：
- `Test Report` 页面在 `/support/test-report`，默认使用 support layout，并展示 Support 横向菜单。
- 但该入口同时出现在 Products 横向菜单中，且要求“从 Products 进入时横向菜单仍保持 Products”。

实现方式（已落地，沿用同一个 query 机制）：
- Products 横向菜单里的 `Test Report` 指向 `/support/test-report?nav=products`。
- 在 support layout 中：
  - 当 `route.query.nav=products` 时，顶部渲染 `ProductsTopNav`；
  - 否则仍渲染 `SupportTopNav`（Support 分区内正常浏览不受影响）。

示例：
- 从 Products 横向菜单点击 Test Report：
  - `/support/test-report?nav=products`（保持 Products 横向菜单）
- 从 Support 横向菜单点击 Test Report：
  - `/support/test-report`（保持 Support 横向菜单）

维护注意事项：
- 仅当你明确希望“从 Products 进入某个 Support 页面也保持 Products 横向菜单”时，才使用 `?nav=products`。
- Support 分区内部链接保持不带该 query，以避免 Support 用户路径被 Products 横向菜单覆盖。

## 实现概览（navContext helper 已落地）
1. 在 `app/utils/navContext.ts` 中引入 `getNavContextFromRoute`，统一根据 `route.path` 与 `route.query.nav` 计算 `navContext`（`company/support/guides/products/blog`）。
2. 在 `ProductsTopNav.vue` 中使用 `getNavContextFromRoute`，根据 `navContext` 在 `companyNavItems` / Guides 菜单 / Blog 菜单 / Products 购买辅助菜单之间切换。
3. 在 `layouts/support.vue` 中同样使用 `getNavContextFromRoute`：
   - 当 `navContext === 'products'` 时强制渲染 `ProductsTopNav`；
   - 其余情况渲染 `SupportTopNav`，保证从 SUPPORT 顶部进入时始终看到 Support 横向菜单。

## 回归检查清单（测试时填写）
- [ ] `/products` 横向菜单符合 Products 规范
- [ ] `/shop` 横向菜单符合 Products 规范
- [ ] `/spoke-calculator` 横向菜单符合 Products 规范
- [ ] `/guides` 横向菜单符合 Guides 规范
- [ ] `/guides/tools` 横向菜单符合 Guides 规范
- [ ] `/guides/sizecharts` 横向菜单符合 Guides 规范
- [ ] `/guides/technical` 横向菜单符合 Guides 规范
- [ ] `/guides/wheelset-buyers` 横向菜单符合 Guides 规范
- [ ] Support 与 Company 表现保持不变
