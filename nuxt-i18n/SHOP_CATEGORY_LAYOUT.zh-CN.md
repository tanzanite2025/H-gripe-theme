# /shop 商品分类布局设计

## 一、目标

希望在 `/shop` 页面上，以可维护、可复用的方式展示「商品分类」，满足：

- **桌面端**：商品网格左侧有清晰的分类导航，方便快速按类别浏览（例如 Rims、Hubs、Spokes 等）。
- **移动端**：提供轻量级的分类选择入口，不明显增加页面高度。
- **桌面 & 移动端**：共用同一份分类数据和选中状态，避免两套逻辑各自维护、容易出错。

本文档作为实现与后续维护的设计说明，方便以后改 UI 时不误删、误改分类功能。

---

## 二、数据模型（Data Model）

- 数据来源：
  - **核心来源是自建商城插件 Tanzanite Setting**：插件在工作区 `tanzanite-theme/wp-plugin/tanzanite-setting` 中，定义了自定义商品类型 `tanz_product`，并通过 REST 命名空间 `tanzanite/v1` 暴露所有自建商城的数据（例如 `GET /wp-json/tanzanite/v1/products`，已支持整型 `category` 参数按分类 ID 过滤）。
  - 如果需要更细的商品类型分类（如 `Rims`、`Hubs`、`Spokes`），可以：
    - 在插件里继续使用 / 扩展商品分类（taxonomy），并单独提供一个「商品分类列表」接口；
    - 或者复用插件现有的 Attributes 体系（`tanzanite/v1/attributes`、`tanzanite/v1/attributes/filterable`），通过某个专用 Attribute Group 来表示「商品类型」。

- 前端数据结构（示例 TypeScript 结构）：

  ```ts
  interface ShopCategory {
    id: number
    slug: string
    name: string
    count?: number
  }
  ```

- 在 `/shop` 页面维护的状态：

  ```ts
  const selectedCategory = ref<ShopCategory | null>(null) // null 表示“全部”
  const categories = ref<ShopCategory[]>([])
  ```

- 与现有搜索参数的集成：
  - 在 `ProductSearchFiltersPayload` 或请求参数中，增加 `category`（或 `categories`）字段。
  - 点击某个分类时，复用当前的 `handleSearch` / `loadProducts` 流程，只是在 payload 里带上选中的分类信息。

---

## 三、桌面端布局（Desktop Layout）

### 3.1 结构

在 `pages/shop.vue` 中，围绕当前商品列表区域外再包一层「两列布局」容器：

- 左列：`CategorySidebar`，用于展示商品分类导航；
- 右列：现有的商品网格区域。

示意结构：

```vue
<section class="mt-6 flex gap-4">
  <aside class="hidden md:block w-56">
    <CategorySidebar
      :categories="categories"
      :selected="selectedCategory"
      @select="onCategorySelect"
    />
  </aside>

  <div class="flex-1">
    <!-- 现有商品列表模块 -->
  </div>
</section>
```

这里 `w-56` 只是一个参考宽度，实际可以根据 UI 效果在 220–260px 左右微调。

### 3.2 CategorySidebar 行为

- 展示一个**纵向分类列表**：
  - 第一项：`All`（全部），表示 `selectedCategory = null`；
  - 后续项：每个真实分类 `ShopCategory`。
- 点击某一项时：
  1. 更新 `selectedCategory`；
  2. 基于当前 `currentSearch` 构造一个新的 `ProductSearchPayload`，附带分类过滤条件；
  3. 调用 `handleSearch(payload)`，触发 `loadProducts` 重新加载商品。
- 当前选中的分类需要有明显的视觉高亮，例如：
  - 背景色加深；
  - 左侧彩色边框；
  - 加粗文字等。
- 可选：展示数量，例如 `Rims (12)`，前提是后端返回了每个分类下商品数量。
- 可选：给整个分类栏加 `position: sticky`，保证在长页面滚动时分类一直可见。

---

## 四、移动端布局（Mobile Layout）

桌面端有独立的左侧分类栏，而在移动端不适合再增加一整列，因此采用更轻量的方案。

### 4.1 推荐方式：横向滚动标签（chips）

- 在商品网格上方增加一行**横向可滚动**的分类标签：

```vue
<div class="md:hidden mb-3 overflow-x-auto">
  <CategoryChips
    :categories="categories"
    :selected="selectedCategory"
    @select="onCategorySelect"
  />
</div>
```

- `CategoryChips` 渲染标签，例如：
  - `[All] [Rims] [Hubs] [Spokes] [Accessories] ...`
- 行为：
  - 使用与桌面端相同的 `onCategorySelect` 处理函数，共享逻辑；
  - 标签样式可以使用 Tailwind 工具类，例如：
    - `whitespace-nowrap`、`space-x-2`、`px-3 py-1`、`rounded-full` 等。
- 优点：
  - 只占一行高度，横向滑动即可查看更多分类；
  - 不会把移动端页面拉得过长，但仍然可以快速切换分类。

### 4.2 备选方式（如分类非常多时）

- 使用一个 `Categories` 按钮，点击后弹出一个小面板 / bottom sheet 展示全部分类列表；
- 选择分类后自动关闭面板，同时触发同样的 `onCategorySelect` 逻辑。

目前设计优先采用 **横向 chips** 方案，仅当分类数量非常多或文字过长时，再考虑切换到折叠面板形式。

---

## 五、共用行为与逻辑（Shared Behavior & Logic）

为避免重复实现逻辑，建议抽象一个 composable：`useShopCategories`。

```ts
export const useShopCategories = () => {
  const categories = ref<ShopCategory[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadCategories = async () => {
    // 从 Tanzanite Setting 插件提供的 REST 接口获取分类数据。
    // 推荐由后端在 tanzanite/v1 命名空间下提供一个明确的「商品分类列表」接口，
    // 便于前端一次性拉取所有可选分类用于 Sidebar / chips 展示。
    // 当前已经有商品列表接口 /wp-json/tanzanite/v1/products 可接收 category 参数做分类筛选，
    // 分类本身的列表可以通过：
    // - 新增 categories Controller，或
    // - 直接暴露自定义 taxonomy 对应的 REST 端点（如 tanz_product 的分类）。
    // DEV 环境依然可以先用本地 mock 数据兜底。
  }

  return { categories, loading, error, loadCategories }
}
```

在 `/shop` 页中：

- 在 `onMounted` 中调用 `loadCategories()`；
- 在页面级别维护唯一的 `selectedCategory`；
- 将 `categories` 与 `selectedCategory` 同时传入：
  - 桌面端的 `CategorySidebar`；
  - 移动端的 `CategoryChips`。

统一的分类选择处理函数示例：

```ts
const onCategorySelect = (category: ShopCategory | null) => {
  selectedCategory.value = category

  const base = currentSearch.value || { query: '', filters: { priceRange: [0, 5000], attributes: {} } }

  const next: ProductSearchPayload = {
    ...base,
    filters: {
      ...base.filters,
      // 根据后端约定挂载分类过滤字段
      // 例如: category: category?.slug ?? undefined
    },
  }

  handleSearch(next)
}
```

---

## 六、后端请求参数（Request Parameters）

需要与后端确认分类在商品接口中的最终表现形式，目前已知与规划如下：

- **方案 A：单独的分类字段（当前 Products 接口已经支持）**
  - 在 `Tanzanite_REST_Products_Controller` 中，`GET /wp-json/tanzanite/v1/products` 已读取整型的 `category` 参数：
    - `$category = (int) ( $request->get_param( 'category' ) ?: 0 );`
  - 前端在构造查询参数时，可以直接传入选中分类的 ID，例如：
    - `params.category = selectedCategory.id`
  - 分类 ID 的含义由插件内部定义的商品分类 / taxonomy 决定。

- **方案 B：通过 attributes 传递（可选扩展）**
  - 如果后续选择用 Attribute Group 来表达「商品类型」，则可以将分类信息放入 `attributes` 字段；
  - 示例：`params.attributes = { ...payload.filters.attributes, product_type: [selectedCategory.slug] }`；
  - 这种方式在结构上更接近目前 Color / Diameter / Brake 等属性筛选的实现，利于后续统一维护。

本文档只描述前端结构与组合方式，并**不强绑定**具体接口路径或字段名，在实际落地时可按后端最终约定来调整映射。

---

## 七、体验与交互注意事项（UX Notes）

- 分类选中状态在桌面与移动端之间必须**同步**：
  - 在桌面点击 Sidebar 分类时，移动端 chips 中对应标签也要处于选中态；
  - 在移动端切换分类时，Sidebar 中高亮项也要同步更新（因为两者共用 `selectedCategory`）。
- 切换分类时，**不要自动重置其他筛选条件**（价格区间、颜色、Diameter、Brake 等），除非明确要这么设计：
  - 用户通常希望先选一个大类，再继续通过高级筛选细化结果。
- 对 `All`（全部）状态要有清晰的视觉区分：
  - 例如 chips 中的 `All` 有特殊边框 / 背景，Sidebar 里第一项 `All` 高亮等；
  - 方便用户理解“当前是在看所有商品”还是“只看某个分类”。

---

## 八、实现检查清单（Implementation Checklist）

- [ ] 新增 `useShopCategories` composable，并确认真实 API endpoint 或准备 DEV mock 数据。
- [ ] 在 `/shop` 页面中增加 `selectedCategory` 状态，并在 `buildProductQueryParams` 中接入分类过滤逻辑。
- [ ] 实现桌面端的 `CategorySidebar`（仅 `md+` 显示），并接好 `onCategorySelect` 事件。
- [ ] 实现移动端的 `CategoryChips`（仅 `< md` 显示），支持横向滚动与选中高亮。
- [ ] 确认分类切换时商品网格会重新加载，并与现有 `AdvancedFilter` 共存良好，不互相覆盖状态。
- [ ] 手动测试：
  - [ ] 桌面端：侧边分类 + 商品网格的布局与交互；
  - [ ] 移动端：chips 选择分类 + 单列商品网格的交互表现。
