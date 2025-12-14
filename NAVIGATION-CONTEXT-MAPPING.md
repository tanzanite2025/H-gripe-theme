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

## 当前问题（Observed）
1. 访问 `/products` 时会显示属于 guides 的入口（这是当前设计导致的，因为 `productsNavItems` 本身包含了 `/guides/*`）。
2. Guides 的表现不一致：
   - `/guides`（首页）显示的是“偏 guides 的过滤菜单”。
   - `/guides/*`（子页，例如 `/guides/sizecharts`）会回退到完整的 products 菜单。
3. products 横向菜单里看不到 `wheelsbuild blog`，原因是：
   - `ProductsTopNav.vue` 目前在菜单中存在 `shop` 时，会过滤掉 `wheelsbuild-blog`（以及 `about-tools`）。

## 期望的“导航上下文（Navigation Context）”规则
不要再依赖单一的路由相等判断来推断菜单（例如仅在 `path === '/guides'` 时才认为是 guides）。
而是定义一个明确的 **Nav Context Mapping（导航上下文映射）**，基于当前路由选择使用哪一套菜单。

### 上下文类型（Contexts）
- `company`
- `support`
- `guides`
- `products`

### 映射表（草案）
按需要补充/调整：

| 路由模式 | 期望上下文 | 备注 |
| --- | --- | --- |
| `/company` 与 `/company/*` | `company` | 使用 `companyNavItems` |
| `/support` 与 `/support/*` | `support` | 使用 `supportNavItems` + support layout |
| `/guides` 与 `/guides/*` | `guides` | 在 guides 内**必须**始终保持 guides 菜单，不得切回 products 菜单 |
| `/products` | `products` | Products 聚合页 |
| `/shop` | `products` | 虽是根路由，但必须保持 products 菜单 |
| `/spoke-calculator` | `products` | 虽是根路由，但必须保持 products 菜单 |
| `/wheelsbuild` | `products` | 虽是根路由，但必须保持 products 菜单 |

## Products 菜单内容（草案）
你希望 products 横向菜单主要解决“用户不知道怎么选商品”的问题（选型/指南/工具入口）。
购买辅助的更重内容（售后/支付/物流等）将放在商品详情页的说明与侧边栏快捷入口中。

### 已确认结论（2025-12-14）
- Products 横向菜单：保留选型相关入口（Shop + Tire Size Charts + Technical + Wheelset Guide + Spoke Calculator），不承载 Wheelsbuild blog。
- Products 横向菜单：本期不放 Support 售后入口（Warranty/After sales/Shipping/Payment）。
- Guides 横向菜单：进入 `/guides` 或任意 `/guides/*` 时，必须稳定显示 Guides 菜单，不得切回 Products 菜单。

实现注意点：
- 现有 `ProductsTopNav.vue` 存在“当菜单中存在 `shop` 时过滤 `about-tools` 与 `wheelsbuild-blog`”的规则。
  - 本期规范中这两项都不在 Products 横向菜单里，因此该过滤规则属于冗余逻辑，建议直接删除以保持代码清晰。

请确认 Products 下应该显示哪些菜单项（标记 Keep/Remove/Add）：

### 候选菜单项（现有）
- [yes] Shop（`/shop`）（id: `shop`）
- [no] Tools guide（`/guides/tools`）（id: `about-tools`）【仅在 Guides 横向菜单展示】
- [yes] Tire Size Charts（`/guides/sizecharts`）（id: `tire-size-charts`）
- [yes] Technical（`/guides/technical`）（id: `technical-docs`）
- [yes] Wheelset Guide（`/guides/wheelset-buyers`）（id: `wheelset-buyers`）
- [yes] Spoke Calculator（`/spoke-calculator`）（id: `spoke-calculator`）
- [no] Wheelsbuild blog（`/wheelsbuild`）（id: `wheelsbuild-blog`）[这个单独放一边，把它从里面剔除，这个不是太重要，这个到时候会另外的入口去放，先从里面剔除】

### 候选“购买辅助”的 Support 入口（当前不在 productsNavItems）
（仅当你希望它们显示在 Products 菜单时才勾选。）
- [no] Warranty（`/support/warranty`）  
- [no] After sales（`/support/after-sales`）
- [no] Shipping（`/support/shipping`）
- [no] Payment（`/support/payment`）

需要决策：
- 当用户从 Products 菜单点击 Support 相关链接时：
  - 当前已决定：Products 横向菜单不展示 Support 入口，因此该决策**本期不适用（N/A）**。
  - 若未来需要在 Products 横向菜单加入 Support 入口，可再在此处决策：
    - 我们是否允许进入 Support 上下文（横向菜单切换为 Support 那套）？
    - 或者我们需要一个“products 上下文包装页”（例如 `/products/warranty`）来保持 products 菜单不变？
    - [ ] 允许切换到 Support 上下文
    - [ ] 保持 Products 菜单可见（需要包装页或路由调整）

## Guides 菜单内容（草案）
Guides 菜单需要在所有 `/guides/*` 路由下保持稳定。

请确认 Guides 下应该显示哪些菜单项：
- [yes] Tools（`/guides/tools`）
- [yes] Tire Size Charts（`/guides/sizecharts`）
- [yes] Technical（`/guides/technical`）
- [yes] Wheelset Buyers Guide（`/guides/wheelset-buyers`）

## 实施计划（在本文件确认之后再执行）
1. 引入一个小函数（放在 `ProductsTopNav.vue` 内或抽到 `app/utils/…`），将 `route.path` 映射到 `navContext`。
2. 更新 `ProductsTopNav.vue` 里的 `items` 计算逻辑：
   - `company` 上下文返回 `companyNavItems`。
   - `guides` 上下文返回仅 guides 的菜单项列表。
   - `products` 上下文返回配置好的“购买辅助”菜单列表。
3. 如果你希望显示 blog，则移除或调整“当存在 `shop` 时隐藏 `wheelsbuild-blog`”的规则。
4. 人工回归检查：
   - `/products` 显示 products 购买辅助菜单。
   - `/shop` 保持同一套 products 菜单。
   - `/guides` 显示 guides 菜单。
   - `/guides/sizecharts` 保持 guides 菜单。
   - 在 guides 各页面之间跳转时，横向菜单不再切回 products。

## 回归检查清单（测试时填写）
- [ ] `/products` 横向菜单符合 Products 规范
- [ ] `/shop` 横向菜单符合 Products 规范
- [ ] `/spoke-calculator` 横向菜单符合 Products 规范
- [ ] `/wheelsbuild` 横向菜单符合 Products 规范
- [ ] `/guides` 横向菜单符合 Guides 规范
- [ ] `/guides/tools` 横向菜单符合 Guides 规范
- [ ] `/guides/sizecharts` 横向菜单符合 Guides 规范
- [ ] `/guides/technical` 横向菜单符合 Guides 规范
- [ ] `/guides/wheelset-buyers` 横向菜单符合 Guides 规范
- [ ] Support 与 Company 表现保持不变
