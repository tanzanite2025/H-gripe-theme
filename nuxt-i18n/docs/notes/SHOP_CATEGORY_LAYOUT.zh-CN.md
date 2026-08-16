# /shop 商品分类布局现状与维护边界

Last updated: 2026-07-24

本文档是 `/shop` 页面分类导航的当前事实源。旧版 WordPress / site settings 插件方案已不再作为当前实现依据；如果需要追溯历史，请看 `nuxt-i18n/docs/archive/`。

## 1. 当前目标

`/shop` 页面现在采用“商品列表 + 左侧本地分类菜单”的结构：

- 桌面端：分类菜单紧贴页面左侧视觉中心，用本地组件渲染，不引入外部依赖。
- 移动端：使用同一份分类数据渲染 chips，居中换行，避免贴边和单行拥挤。
- 桌面端与移动端共用 `selectedCategory`，保证状态唯一，不维护两套分类逻辑。
- 分类只作为 `/shop` 页面自己的入口能力，不影响顶部 mega menu 的层级归属。

## 2. 当前事实源

### 后端

当前分类来源是 Go backend 的商品类型接口：

- `GET /api/v1/products/specification-templates`
- Go handler：`go-backend/internal/api/v1/product/handler.go` 的 `ListProductSpecificationTemplates`
- 路由注册：`go-backend/internal/api/v1/router.go`

商品列表过滤使用：

- `GET /api/v1/products?...&product_specification_template=<slug>`

也就是说，前端长期应该依赖稳定的 `slug`，而不是把分类 ID 当作跨系统事实源。

### Nuxt 前端

主要文件：

- `nuxt-i18n/app/composables/useShopCategories.ts`
  - 拉取 `/products/specification-templates`
  - 标准化为 `ShopCategory[]`
  - fallback categories 只允许在 `import.meta.dev` 本地开发环境使用；生产环境必须展示真实空分类或接口错误
- `nuxt-i18n/app/pages/shop/index.vue`
  - 维护唯一的 `selectedCategory`
  - 在 `buildProductQueryParams()` 中把 product specification template 分类写入 `product_specification_template`
  - 消费 `useShopSearchSheet()` 传来的 `presetCategorySlug`
- `nuxt-i18n/app/components/shop/ShopCategoryVerticalMenu.vue`
  - 桌面端纵向分类菜单
- `nuxt-i18n/app/components/CategoryChips.vue`
  - 移动端分类 chips

## 3. 当前已收口

- 已有 `useShopCategories()` 统一分类数据加载。
- 桌面端已改成本地纵向分类菜单，不依赖 `21st.dev` 外部组件。
- 移动端已改成居中、可换行的 chips。
- `/shop` 已有 `selectedCategory` 作为单一分类状态。
- 商品请求已在 product specification template 分类选中时带上 `product_specification_template=<slug>`。
- Inner tube / search sheet 入口传来的 `presetCategorySlug` 已能映射到 `/shop` 分类。
- 分类切换不会清空其它搜索关键词和筛选 payload。
- `/shop` 分类 fallback 边界已收口：DEV 可用本地示例防止开发空白，生产不再把示例分类当事实源。
- 桌面端分类菜单和移动端 chips 已有空分类/错误/加载状态；真实空分类不会只显示一个“All”。
- 商品空状态已区分“全站暂无商品”和“当前分类暂无商品”。

## 4. 当前还没完全闭环

这些才是后续真正需要关注的点，不要再按旧 checklist 重做已经完成的组件。

1. 真实商品类型数据
   - 需要确认生产/DEV 的 `/api/v1/products/specification-templates` 返回完整、启用、排序正确的数据。
   - 如果后台商品模板或商品类型重置为单一事实源后字段名变化，需同步此接口和本文档。

2. 真实数据验收
   - 当前代码已按真实接口空值处理，但还需要在 DEV / staging 上用真实 `/api/v1/products/specification-templates` 响应做一次截图或接口验收。
   - 验收重点：只返回启用类型、按 `sort_order ASC, id ASC` 排序、slug 与商品列表过滤参数一致。

3. 与高级筛选共存
   - 分类、关键词、价格、属性筛选现在是合并查询。
   - 后续如果新增更细的 tube / rim / frame 专项筛选，要继续保持 `selectedCategory + filters` 的组合方式，不要把分类逻辑复制进多个组件。

4. i18n
   - 分类名当前来自后端数据。
   - 当前前端只翻译固定 UI 文案，不翻译动态分类名。
   - 如果后端后续需要固定 20 个 storefront locale 的分类名，应由 `/products/specification-templates` 按 locale 返回或返回明确的 localized map；不要在 Nuxt 静态语言包里手写商品类型事实源。

## 5. 维护规则

- 不要恢复旧的 `/shop` Categories 卡片。
- 不要把商品图库、营销菜单或顶部 mega menu 的分类数据混入 `/shop` 商品类型。
- 不要新增外部 UI 依赖来实现这个分类菜单；当前组件应保持本地可控。
- 后续新增三级或更细分类时，优先让后端接口返回层级结构或稳定 slug，再由 Nuxt 组件自动渲染；不要在页面里手写某个分类的特殊按钮。
- 任何涉及 `/shop` 分类来源、查询参数或 preset slug 行为的修改，都要同步更新本文档。
