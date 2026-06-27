# Site Search Popular Keywords & Chips Design

> 目标：在 **SiteHeader 搜索 Bottom Sheet** 与 **/shop 顶部搜索** 中，统一增加「热门搜索」胶囊 TAB，且所有搜索逻辑仍然只走现有的搜索函数，不产生第二套搜索流程。

本文仅是交互与结构设计说明，不包含实现代码，供之后开发与 Review 对照使用。

---

## 1. 总体设计原则

- **单一搜索逻辑**：
  - 真实的搜索行为（拼 query、提交、跳转到 `/shop`、刷新结果）仍然只通过 **已有的搜索函数** 完成：
    - 在 `ProductSearchPanel.vue` 中：`searchProducts()` → `emit('search', ...)` → `useShopSearchSheet.submit()`。
    - 在 `/shop/index.vue` 中：顶部搜索自己的 `runSearch()`（或类似函数）。
  - 「热门搜索」TAB **不直接发请求，不直接跳路由**，只负责更新搜索组件内部的状态。

- **统一的热门关键词来源**：
  - 新建 `~/utils/popularSearchKeywords.ts`，导出一个字符串数组 `popularSearchKeywords`，两个入口复用：
    - SiteHeader 搜索 Bottom Sheet (`ProductSearchPanel.vue`)。
    - `/shop/index.vue` 顶部搜索区域。

- **统一的胶囊样式**：
  - 搜索组件内部本身如果已有「用于筛选的 TAB / 多选胶囊」样式（例如高级筛选中的 tag 风格），
  - **热门搜索 TAB + 搜索框内部展示的选中 TAB，样式必须与之保持一致**：同一套圆角、背景色、字体大小与间距。
  - 避免做出两套风格不同的胶囊，后期再返工统一样式。

- **方案 B + 视觉增强**：
  - 逻辑上是「回填 + 选中高亮」，不新增搜索流程。
  - 视觉上，将搜索框改造成一个「chips + 输入」的外壳，使多选关键词不会在输入框里混成一团。

---

## 2. 数据结构与状态约定

在需要支持热门搜索的两个地方（`ProductSearchPanel.vue` 和 `/shop/index.vue` 顶部搜索）中，共享以下状态约定：

- `selectedKeywords: string[]`
  - 当前被选中的热门搜索关键词列表。
  - 由热门 TAB 点击、输入框内部的胶囊关闭按钮共同维护。

- `freeTextQuery: string`
  - 用户在输入框中手动输入的自由文本（非热门关键词部分）。

- `productSearchQuery: string`
  - **真正用于搜索提交的 query 字符串**。
  - 每当 `selectedKeywords` 或 `freeTextQuery` 变化时，统一同步：

    ```ts
    productSearchQuery = [
      ...selectedKeywords,
      freeTextQuery.trim(),
    ]
      .filter(Boolean)
      .join(' ')
    ```

  - 对于后端或 `/shop` 页来说，只关心 `productSearchQuery`，不关心 chip 是如何产生的。

> 重要：所有搜索请求（无论是点击 Search 按钮，还是以后选择 TAB 后自动搜索）都只使用 `productSearchQuery`，不在其他地方再单独拼 query。

---

## 3. PopularSearchChips 组件设计

### 3.1 组件职责

`PopularSearchChips.vue` 只负责 **展示** 和 **点击回调**，不直接发搜索、不改路由。

- 渲染一组胶囊形 TAB（热门关键词）。
- 支持单选 / 多选（由父组件控制逻辑）。
- 提供选中态样式（与搜索组件已有的 TAB 胶囊保持一致）。

### 3.2 Props & Emits（概念设计）

- Props：
  - `keywords: string[]` — 所有可选的热门关键词。
  - `modelValue?: string[]` — 当前选中的关键词（用于多选，v-model 形式可选）。
  - `title?: string` — 上方标题文本，默认 `'Popular searches'`，之后可接 i18n。

- Emits：
  - `update:modelValue(selected: string[])` — 当选中集合变化时触发。
  - 或者简化为：`select(keyword: string)`，由父组件在外部维护 `selectedKeywords`。

> 实际实现时可以二选一，推荐使用 `modelValue` + `update:modelValue` 的 v-model 形式，便于在多个父组件中复用。

### 3.3 胶囊样式约定

- PopularSearchChips 中的 TAB **必须与搜索组件内部现有的 TAB / 多选胶囊样式统一**：
  - 同一套：
    - 背景色 / hover 颜色
    - 圆角半径（pill 形）
    - 字体大小 / 行高
    - 内边距 / 间距
  - 如果现在已经有某个 `.filter-tag` / `.chip` / 类似 class，用它作为单一视觉源：
    - PopularSearchChips 只复用该 class，不新造第二套视觉。

---

## 4. 搜索输入框外壳：chips + input

### 4.1 结构

将 `ProductSearchPanel.vue` 中当前的单个 `<input class="search-input-c" />` 替换为一个“外壳 + 内部 input”的结构：

```vue
<div class="search-input-shell">  <!-- 外壳：整体形态仍然看起来像一个输入框 -->
  <!-- 在输入框内部展示已选 TAB（胶囊） -->
  <span
    v-for="keyword in selectedKeywords"
    :key="keyword"
    class="search-chip-in-input"  <!-- 复用与 PopularSearchChips 一致的胶囊样式 -->
  >
    <span class="search-chip-in-input__label">{{ keyword }}</span>
    <button
      type="button"
      class="search-chip-in-input__close"
      @click="removeKeyword(keyword)"  <!-- 调用与 TAB 二次点击相同的逻辑 -->
    >
      ×
    </button>
  </span>

  <!-- 真正可编辑的文本输入区 -->
  <input
    v-model="freeTextQuery"
    class="search-input-inner"  <!-- 无边框 / 透明背景，靠外壳提供视觉 -->
    ...
  />
</div>
```

### 4.2 行为

- `selectedKeywords` 决定：
  - 下方热门区 TAB 的选中状态；
  - 搜索框内部的胶囊显示内容。

- 点击热门 TAB：
  - 调用统一的 `toggleKeyword(keyword)`：

    - 若该词不在 `selectedKeywords` 中：加入数组，TAB 高亮 + 输入框中多一个胶囊。
    - 若该词已在 `selectedKeywords` 中：从数组中删除，TAB 取消高亮 + 输入框中的胶囊消失。

- 点击输入框中胶囊右侧的关闭按钮：
  - 与「再次点击热门 TAB」完全等价：同样调用 `toggleKeyword(keyword)`。

- 任意一次 `selectedKeywords` / `freeTextQuery` 变化时：

  - 重新计算：
    ```ts
    productSearchQuery = [
      ...selectedKeywords,
      freeTextQuery.trim(),
    ].filter(Boolean).join(' ')
    ```

> 这样满足需求：**二次点击 TAB 或点击输入框内的关闭按钮时，同时取消 TAB 选中态和输入框里的文本表示**。

### 4.3 视觉风格要求

- `search-input-shell` 整体外观：
  - 保持与当前 `.search-input-c` 一致的：
    - 背景色、圆角、阴影
    - 聚焦态的边框 / 光晕

- `search-chip-in-input`：
  - 直接复用 PopularSearchChips 胶囊样式（或公用的 chip class），
  - 实现视觉上“在一个大的输入框里排了一串胶囊 + 光标”的效果。

---

## 5. 搜索触发逻辑

### 5.1 统一使用现有搜索函数

- 在 `ProductSearchPanel.vue` 中：
  - 仍然使用已有的 `searchProducts()` 函数：
    - `searchProducts()` 内部读取 `productSearchQuery` 和 `filters`，通过 `emit('search', ...)` 通知父级。
    - 父级 `ShopSearchSheet.vue` 再调用 `useShopSearchSheet.submit()`，负责关闭弹窗和跳转 `/shop`。

- 在 `/shop/index.vue` 中（顶部搜索）：
  - 仍然使用现有的 `runSearch()` 或类似函数，根据当前 `productSearchQuery` 和筛选条件刷新商品列表。

> 不在 PopularSearchChips 或胶囊组件内部写任何 `router.push` 或直接请求的逻辑。

### 5.2 何时触发搜索

当前版本只采用**一种**触发方式：

- **仅在用户点击 Search 按钮时触发搜索**（更稳定）：
  - 热门 TAB / 输入框内关闭按钮只负责更新 `selectedKeywords` 和 `freeTextQuery`，
  - 用户可以多次选择 / 取消 TAB、修改输入内容，然后点击一次 Search 统一提交。

> 如果未来需要，可以在实现层扩展为「选中 / 取消 TAB 时自动触发搜索」，但也必须调用同一个 `searchProducts()` / `runSearch()`，不能在 chips 内部单独实现第二套搜索逻辑。

---

## 6. /shop 顶部搜索的复用方式

- `/shop/index.vue` 顶部搜索区域应复用同一套模式：
  - 使用 `popularSearchKeywords` 数据源。
  - 使用 `PopularSearchChips.vue` 作为热门 TAB 区块。
  - 使用「chips + input」外壳展示已选关键词。
  - 维护自己的 `selectedKeywords / freeTextQuery / productSearchQuery`。
  - 通过本页的 `runSearch()`（或等价函数）执行搜索。

- 这样：
  - 从 SiteHeader 搜索 Bottom Sheet 搜索 → 跳转到 `/shop`；
  - 在 `/shop` 顶部搜索直接操作；
  - 视觉与交互都保持一致，用户不会感觉是两个系统。

---

## 7. i18n & 文案

- 热门关键词数组 `popularSearchKeywords`：
  - 第一版使用固定的英文关键词数组，不走翻译（它们本身就是用户可能输入的英文 query）：
    ```ts
    export const popularSearchKeywords = [
      'Carbon rim',
      'Sapim',
      'Spoke',
      'Carbon Wheels',
    ]
    ```
- 标题 "Popular searches"：
  - 初始可以在组件内部硬编码，后续迁移到 `en.json`：
    - 例如 key：`"search.popularTitle": "Popular searches"`，
  - 由 `PopularSearchChips.vue` 使用 `$t('search.popularTitle')` 渲染。

---

## 8. 实施顺序建议

1. 新建 `~/utils/popularSearchKeywords.ts`，写出一组初始热门关键词。
2. 新建 `PopularSearchChips.vue`，只负责：
   - 展示 chips；
   - 维护选中态；
   - 通过 emit 通知父组件。（样式严格复用现有 TAB 胶囊样式。）
3. 在 `ProductSearchPanel.vue` 中：
   - 增加 `selectedKeywords / freeTextQuery / productSearchQuery` 状态。
   - 改造搜索输入为「chips + input shell」。
   - 接入 `PopularSearchChips` 作为热门搜索区域。
4. 验证 SiteHeader 搜索 Bottom Sheet 的交互：
   - 点击 TAB：选中 / 取消；
   - 输入框内胶囊同步变化，关闭按钮可单独移除关键词；
   - Search 按钮发起搜索，行为与改造前一致。
5. 在 `/shop/index.vue` 顶部搜索中复用相同组件与逻辑。
6. 最后根据需要，将 "Popular searches" 文案接入 i18n。
