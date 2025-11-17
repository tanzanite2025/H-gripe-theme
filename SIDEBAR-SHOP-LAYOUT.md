# Sidebar 商店布局架构文档

**项目**: Tanzanite Theme  
**更新日期**: 2025-11-17  
**功能**: Sidebar 双栏布局 + 商品筛选 + 搜索结果展示

---

## 📐 整体架构

### 桌面端布局

```
┌─────────────────────────────────────────────────────────────────┐
│  SiteHeader                                                      │
├──────────────────┬──────────────────────────────────────────────┤
│                  │                                               │
│  Sidebar 第一栏  │  Sidebar 第二栏                               │
│  分类导航        │  高级筛选 + 固定信息                          │
│  (280px)         │  (280px)                                      │
│                  │                                               │
│  📁 静态页面     │  🔍 高级筛选器                                │
│  ├─ 🏠 首页      │  ┌─────────────────────────────┐            │
│  ├─ 🛍️ 商店     │  │ 价格: $0 - $5000            │            │
│  ├─ ❓ FAQ      │  │ 库存: ☑ 有货 ☐ 预售        │            │
│  └─ 📞 联系     │  │ 排序: 价格 ↓                │            │
│                  │  └─────────────────────────────┘            │
│  ──────────────  │                                               │
│                  │  ─────────────────                            │
│  📦 商品分类     │                                               │
│  └─ 💍 戒指 ▼   │  📞 联系信息                                  │
│      ├─ 钻石    │  ┌─────────────────────────────┐            │
│      ├─ 黄金    │  │ 📞 +1 234 567 8900          │            │
│      └─ 铂金    │  │ 📧 info@tanzanite.site      │            │
│  └─ 📿 项链 ▼   │  │ 🕐 Mon-Fri 9:00-18:00       │            │
│      ├─ 金项链  │  └─────────────────────────────┘            │
│      └─ 银项链  │                                               │
│  └─ 💎 手镯 ▶   │  🌐 社交图标                                  │
│                  │  [FB] [TW] [IG] [WA]                        │
│                  │                                               │
└──────────────────┴──────────────────────────────────────────────┘
                                    │
                                    │ 点击分类触发
                                    ↓
                    ┌───────────────────────────────────┐
                    │  右侧弹窗（SidePanel）             │
                    │  [← 返回] 戒指 - 钻石              │
                    ├───────────────────────────────────┤
                    │  📦 商品搜索结果                   │
                    │  ┌─────────────────────────────┐  │
                    │  │ 💍 Diamond Ring             │  │
                    │  │ $1,299 | 库存: 50           │  │
                    │  │ ⭐⭐⭐⭐⭐ (128)             │  │
                    │  │         [查看详情 →]        │  │
                    │  └─────────────────────────────┘  │
                    │                                    │
                    │  [加载更多]                        │
                    └───────────────────────────────────┘
```

### 移动端布局（分页结构）

```
第 1 页：分类导航
┌─────────────────────────────────┐
│  ┌─ 第 1 页 ─┬─ 第 2 页 ─┐     │
│  │   ●       │     ○      │     │
│  └───────────┴────────────┘     │
├─────────────────────────────────┤
│  📁 静态页面                     │
│  ├─ 🏠 首页                      │
│  ├─ 🛍️ 商店                     │
│  └─ ❓ FAQ                       │
│                                  │
│  📦 商品分类（默认折叠）         │
│  └─ 💍 戒指 ▶                   │
│  └─ 📿 项链 ▶                   │
│  └─ 💎 手镯 ▶                   │
└─────────────────────────────────┘

第 2 页：筛选器 + 固定信息
┌─────────────────────────────────┐
│  ┌─ 第 1 页 ─┬─ 第 2 页 ─┐     │
│  │   ○       │     ●      │     │
│  └───────────┴────────────┘     │
├─────────────────────────────────┤
│  🔍 高级筛选器                   │
│  ┌─────────────────────────┐    │
│  │ 价格: $0 - $5000        │    │
│  │ 库存: ☑ 有货            │    │
│  │ 排序: 价格 ↓            │    │
│  └─────────────────────────┘    │
│                                  │
│  📞 联系信息                     │
│  📧 info@tanzanite.site         │
│                                  │
│  🌐 社交图标                     │
│  [FB] [TW] [IG] [WA]            │
└─────────────────────────────────┘
```

---

## 🎯 核心功能

### 1. Sidebar 双栏布局

#### 第一栏：分类导航
- **静态页面列表**
  - 首页、商店、FAQ、联系我们等
  - 固定显示，不可折叠
  
- **商品分类树**
  - 一级分类：戒指、项链、手镯等
  - 二级分类：钻石戒指、黄金戒指等
  - 桌面端：二级分类默认展开
  - 移动端：二级分类默认折叠

#### 第二栏：筛选器 + 固定信息
- **高级筛选器组件**（可复用）
  - 价格范围滑块
  - 库存状态复选框
  - 排序下拉菜单
  - 其他自定义筛选条件
  
- **联系信息**（固定显示）
  - 电话号码
  - 邮箱地址
  - 营业时间
  
- **社交图标**（固定显示）
  - Facebook、Twitter、Instagram、WhatsApp

### 2. 右侧弹窗（SidePanel）

- **商品搜索结果展示**
  - 商品缩略图
  - 商品标题
  - 价格和库存
  - 评分（可选）
  - [查看详情] 按钮
  
- **交互逻辑**
  - 点击第一栏的分类 → 弹出 SidePanel
  - 应用第二栏的筛选条件
  - 实时更新商品列表
  - 点击 [查看详情] → 跳转到商品详情页

### 3. 移动端分页

- **第 1 页：分类导航**
  - 静态页面列表
  - 商品分类树（默认折叠）
  
- **第 2 页：筛选器 + 固定信息**
  - 高级筛选器
  - 联系信息
  - 社交图标
  
- **切换方式**
  - 顶部标签切换
  - 底部按钮切换
  - 滑动手势（可选）

---

## 📦 组件结构

### 组件层级

```
pages/shop.vue
├─ SiteHeader.vue
├─ Sidebar.vue (主容器)
│  ├─ SidebarColumn1.vue (第一栏)
│  │  ├─ StaticPagesList.vue (静态页面)
│  │  └─ CategoryTree.vue (商品分类树)
│  │
│  └─ SidebarColumn2.vue (第二栏)
│     ├─ AdvancedFilter.vue (高级筛选器) ⭐ 可复用组件
│     ├─ ContactInfo.vue (联系信息)
│     └─ SocialIcons.vue (社交图标)
│
└─ SidePanel.vue (右侧弹窗)
   ├─ ProductSearchResults.vue (商品搜索结果)
   └─ ProductCardSimple.vue (商品简要卡片)
```

### 核心组件说明

#### 1. AdvancedFilter.vue（高级筛选器）⭐

**功能**：
- 价格范围筛选
- 库存状态筛选
- 排序方式选择
- 其他自定义筛选条件

**特点**：
- ✅ 可复用组件
- ✅ 可在其他页面调用
- ✅ 支持自定义配置
- ✅ 响应式设计

**Props**：
```typescript
interface FilterProps {
  // 初始筛选条件
  initialFilters?: {
    priceRange?: [number, number]
    inStock?: boolean
    sortBy?: string
    // 其他筛选条件
  }
  
  // 可用的筛选选项
  options?: {
    showPriceRange?: boolean
    showStockFilter?: boolean
    showSortBy?: boolean
    priceMin?: number
    priceMax?: number
    sortOptions?: Array<{label: string, value: string}>
  }
  
  // 样式配置
  compact?: boolean // 紧凑模式
  theme?: 'dark' | 'light'
}
```

**Emits**：
```typescript
interface FilterEmits {
  // 筛选条件变化
  'update:filters': (filters: FilterState) => void
  
  // 重置筛选
  'reset': () => void
}
```

**使用示例**：
```vue
<!-- 在 Sidebar 中使用 -->
<AdvancedFilter
  v-model:filters="filters"
  :options="{
    showPriceRange: true,
    showStockFilter: true,
    priceMin: 0,
    priceMax: 10000
  }"
  @update:filters="handleFilterChange"
/>

<!-- 在其他页面使用 -->
<AdvancedFilter
  v-model:filters="searchFilters"
  :compact="true"
  theme="light"
/>
```

#### 2. CategoryTree.vue（商品分类树）

**功能**：
- 显示商品分类层级
- 支持展开/折叠
- 响应式行为（桌面/移动端）

**Props**：
```typescript
interface CategoryTreeProps {
  // 分类数据
  categories: Array<Category>
  
  // 桌面端是否默认展开
  desktopExpanded?: boolean
  
  // 移动端是否默认展开
  mobileExpanded?: boolean
  
  // 当前选中的分类
  selectedCategory?: string
}
```

**Emits**：
```typescript
interface CategoryTreeEmits {
  // 分类点击事件
  'category-click': (category: Category) => void
}
```

#### 3. SidePanel.vue（右侧弹窗）

**功能**：
- 显示商品搜索结果
- 支持分页/无限滚动
- 响应式弹出动画

**Props**：
```typescript
interface SidePanelProps {
  // 是否打开
  open: boolean
  
  // 当前分类
  category?: Category
  
  // 筛选条件
  filters?: FilterState
  
  // 弹出方向
  direction?: 'left' | 'right'
}
```

**Emits**：
```typescript
interface SidePanelEmits {
  // 关闭事件
  'close': () => void
  
  // 商品点击事件
  'product-click': (product: Product) => void
}
```

---

## 🔄 数据流程

### 1. 用户筛选流程

```
用户在第二栏调整筛选器
    ↓
AdvancedFilter 组件 emit 'update:filters'
    ↓
Sidebar 组件更新 filters 状态
    ↓
传递给 SidePanel 组件
    ↓
SidePanel 根据 filters 重新请求商品数据
    ↓
显示更新后的商品列表
```

### 2. 分类点击流程

```
用户点击第一栏的分类
    ↓
CategoryTree 组件 emit 'category-click'
    ↓
Sidebar 组件接收事件
    ↓
打开 SidePanel 组件
    ↓
传递 category 和 filters 参数
    ↓
SidePanel 请求该分类的商品数据
    ↓
显示商品列表
```

### 3. 商品详情跳转流程

```
用户点击商品卡片的 [查看详情]
    ↓
NuxtLink 跳转到 /products/[slug]
    ↓
加载商品详情页（SSG 静态页面）
    ↓
显示完整商品信息
```

---

## 🛠️ 实施步骤

### 阶段 1：创建可复用组件

#### 1.1 创建 AdvancedFilter 组件

**文件路径**: `nuxt-i18n/app/components/AdvancedFilter.vue`

**功能清单**：
- [ ] 价格范围滑块（使用 Vue3 Slider 或自定义）
- [ ] 库存状态复选框
- [ ] 排序下拉菜单
- [ ] 响应式布局（桌面/移动端）
- [ ] 主题支持（dark/light）
- [ ] 紧凑模式支持

**依赖**：
```json
{
  "@vueuse/core": "^10.x",
  "vue3-slider": "^1.x" // 或其他滑块库
}
```

#### 1.2 创建 CategoryTree 组件

**文件路径**: `nuxt-i18n/app/components/CategoryTree.vue`

**功能清单**：
- [ ] 递归渲染分类树
- [ ] 展开/折叠动画
- [ ] 响应式行为（桌面/移动端默认状态）
- [ ] 高亮当前选中分类
- [ ] 图标支持（Emoji 或 SVG）

#### 1.3 创建 SidePanel 组件

**文件路径**: `nuxt-i18n/app/components/SidePanel.vue`

**功能清单**：
- [ ] 弹出动画（从右侧滑入）
- [ ] 背景遮罩
- [ ] 点击遮罩关闭
- [ ] 商品列表渲染
- [ ] 加载状态
- [ ] 分页/无限滚动

#### 1.4 创建辅助组件

**ContactInfo.vue**:
- [ ] 联系电话
- [ ] 邮箱地址
- [ ] 营业时间
- [ ] 可配置显示项

**SocialIcons.vue**:
- [ ] 社交平台图标
- [ ] 链接配置
- [ ] 悬停效果

**ProductCardSimple.vue**:
- [ ] 商品缩略图
- [ ] 标题、价格、库存
- [ ] 评分显示（可选）
- [ ] [查看详情] 按钮

---

### 阶段 2：构建 Sidebar 主容器

#### 2.1 创建 Sidebar 组件

**文件路径**: `nuxt-i18n/app/components/Sidebar.vue`

**功能清单**：
- [ ] 桌面端双栏布局
- [ ] 移动端分页布局
- [ ] 响应式切换
- [ ] 状态管理（当前页、选中分类、筛选条件）

#### 2.2 创建子组件

**SidebarColumn1.vue**:
- [ ] 静态页面列表
- [ ] 商品分类树
- [ ] 滚动容器

**SidebarColumn2.vue**:
- [ ] 高级筛选器
- [ ] 联系信息
- [ ] 社交图标
- [ ] 固定布局

---

### 阶段 3：集成到页面

#### 3.1 修改 shop.vue 页面

**文件路径**: `nuxt-i18n/app/pages/shop.vue`

**功能清单**：
- [ ] 引入 Sidebar 组件
- [ ] 引入 SidePanel 组件
- [ ] 状态管理（筛选条件、选中分类）
- [ ] API 请求逻辑

#### 3.2 WordPress API 集成

**端点需求**：
```
GET /wp-json/tanzanite/v1/categories
  - 获取商品分类树

GET /wp-json/tanzanite/v1/products
  - 参数: category, priceMin, priceMax, inStock, sortBy
  - 返回: 商品列表（分页）
```

---

### 阶段 4：样式和动画

#### 4.1 Tailwind 配置

**功能清单**：
- [ ] 自定义颜色（品牌色）
- [ ] 自定义断点（如需要）
- [ ] 动画类

#### 4.2 过渡动画

**功能清单**：
- [ ] SidePanel 滑入/滑出动画
- [ ] 分类展开/折叠动画
- [ ] 筛选器交互动画
- [ ] 加载状态动画

---

### 阶段 5：测试和优化

#### 5.1 功能测试

**测试清单**：
- [ ] 桌面端双栏布局显示正常
- [ ] 移动端分页切换正常
- [ ] 分类点击触发 SidePanel
- [ ] 筛选器更新商品列表
- [ ] 商品详情跳转正常
- [ ] 响应式布局适配

#### 5.2 性能优化

**优化清单**：
- [ ] 组件懒加载
- [ ] 图片懒加载
- [ ] API 请求缓存
- [ ] 虚拟滚动（如商品很多）
- [ ] 防抖/节流（筛选器）

---

## 📊 数据结构

### Category（分类）

```typescript
interface Category {
  id: number
  name: string
  slug: string
  icon?: string
  parent?: number
  children?: Category[]
  count?: number // 商品数量
}
```

### FilterState（筛选状态）

```typescript
interface FilterState {
  priceRange: [number, number]
  inStock: boolean
  sortBy: 'price_asc' | 'price_desc' | 'newest' | 'popular'
  // 其他筛选条件
  [key: string]: any
}
```

### Product（商品）

```typescript
interface Product {
  id: number
  title: string
  slug: string
  price: number
  stock: number
  image: string
  rating?: number
  reviews?: number
  category: string
  // 其他字段
}
```

---

## 🎨 样式规范

### 颜色

```css
/* 主色调 */
--color-primary-green: #40ffaa
--color-primary-blue: #6b73ff

/* 背景 */
--color-bg-dark: #000000
--color-bg-card: rgba(255, 255, 255, 0.05)

/* 边框 */
--color-border: rgba(255, 255, 255, 0.1)

/* 文字 */
--color-text-primary: #ffffff
--color-text-secondary: rgba(255, 255, 255, 0.7)
--color-text-tertiary: rgba(255, 255, 255, 0.4)
```

### 间距

```css
/* Sidebar 宽度 */
--sidebar-column-width: 280px

/* 内边距 */
--spacing-sm: 0.5rem
--spacing-md: 1rem
--spacing-lg: 1.5rem

/* 圆角 */
--radius-sm: 0.5rem
--radius-md: 0.75rem
--radius-lg: 1rem
```

---

## 🔗 相关文档

- [Nuxt SSG 架构文档](./NUXT-SSG-ARCHITECTURE.md)
- [FAQ 实施指南](./FAQ-IMPLEMENTATION-GUIDE.md)
- [URLLink 使用文档](./wp-plugin/tanzanite-setting/docs/URLLINK.md)

---

## 📝 注意事项

### 1. AdvancedFilter 组件设计

- ⚠️ 必须设计为**完全独立**的组件
- ⚠️ 不依赖特定的父组件或页面
- ⚠️ 通过 props 和 emits 通信
- ⚠️ 支持多种配置和主题

### 2. 响应式设计

- ⚠️ 桌面端和移动端行为差异较大
- ⚠️ 使用 `@vueuse/core` 的 `useBreakpoints` 检测屏幕尺寸
- ⚠️ 确保所有交互在移动端可用

### 3. 性能考虑

- ⚠️ 商品列表可能很长，考虑虚拟滚动
- ⚠️ 筛选器变化频繁，使用防抖
- ⚠️ 图片使用懒加载
- ⚠️ API 请求使用缓存

### 4. 可访问性

- ⚠️ 键盘导航支持
- ⚠️ ARIA 标签
- ⚠️ 焦点管理
- ⚠️ 屏幕阅读器支持

---

## 🚀 实施优先级

### P0（必须完成）
1. ✅ AdvancedFilter 组件（可复用）
2. ✅ CategoryTree 组件
3. ✅ SidePanel 组件
4. ✅ Sidebar 主容器
5. ✅ 基础样式和布局

### P1（重要）
1. ⭐ 移动端分页功能
2. ⭐ 商品搜索结果展示
3. ⭐ WordPress API 集成
4. ⭐ 响应式适配

### P2（优化）
1. 🔧 动画和过渡效果
2. 🔧 虚拟滚动
3. 🔧 性能优化
4. 🔧 可访问性增强

---

**最后更新**: 2025-11-17  
**维护者**: Tanzanite Team  
**状态**: 待实施
