# All Products - 商品列表管理

**页面路径**: `admin.php?page=tanzanite-settings`  
**权限要求**: `tanz_view_products`  
**REST API**: `/wp-json/tanzanite/v1/products`

---

## 📋 功能概述

All Products 页面是商品管理的核心页面，提供完整的商品列表查看、筛选、搜索和批量操作功能。

---

## ✨ 主要功能

### 1. 商品列表展示

#### 显示字段
- **商品名称** - 商品标题（可点击编辑）
- **SKU 编号** - 商品 SKU（支持多规格）
- **状态** - 发布状态（全部/已发布/草稿/已删除）
- **分类** - 商品所属分类
- **作者 ID** - 创建者 ID
- **操作** - 编辑/删除按钮

#### 分页功能
- 每页显示数量：20 条（可配置）
- 总页数显示
- 页码跳转
- 上一页/下一页

---

### 2. 基础筛选

#### 筛选条件

**关键词搜索**
- 搜索范围：商品名称、SKU、描述
- 支持模糊匹配
- 实时搜索

**状态筛选**
- 全部
- 已发布（publish）
- 草稿（draft）
- 已删除（trash）

**分类筛选**
- 下拉选择分类
- 支持层级分类
- 显示分类名称

**排序方式**
- 更新时间（默认）
- 常规价格
- 库存数量
- 积分奖励

**排序方向**
- 降序（DESC，默认）
- 升序（ASC）

**每页数量**
- 20 条（默认）
- 50 条
- 100 条
- 200 条

---

### 3. 高级筛选

点击"高级筛选"展开更多筛选选项：

#### 标签筛选
- 搜索标签名称
- 支持多标签筛选
- OR 关系（包含任一标签）

#### 作者筛选
- 按作者 ID 筛选
- 查看特定用户创建的商品

#### 库存筛选
- **最小库存** - 库存数量下限
- **最大库存** - 库存数量上限
- 用于查找低库存或高库存商品

#### 价格筛选
- **最小价格** - 价格下限
- **最大价格** - 价格上限
- 按价格区间筛选

#### 积分筛选
- **最小积分** - 积分奖励下限
- **最大积分** - 积分奖励上限

---

### 4. 批量操作

选择多个商品后可执行批量操作：

#### 批量修改状态
- 批量发布
- 批量设为草稿
- 批量删除

#### 批量导出
- 导出选中商品
- CSV 格式
- 包含完整商品信息

---

## 🔌 REST API

### 获取商品列表

**端点**: `GET /wp-json/tanzanite/v1/products`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | integer | 否 | 页码，默认 1 |
| per_page | integer | 否 | 每页数量，默认 20 |
| keyword | string | 否 | 搜索关键词 |
| status | string | 否 | 状态筛选 |
| category | integer | 否 | 分类 ID |
| tags | array | 否 | 标签 slug 数组 |
| author | integer | 否 | 作者 ID |
| inventory_min | integer | 否 | 最小库存 |
| inventory_max | integer | 否 | 最大库存 |
| price_min | float | 否 | 最小价格 |
| price_max | float | 否 | 最大价格 |
| sort | string | 否 | 排序字段 |
| order | string | 否 | 排序方向（ASC/DESC） |

**请求示例**:

```javascript
// 获取第一页商品
const response = await fetch('/wp-json/tanzanite/v1/products?page=1&per_page=20', {
  headers: {
    'X-WP-Nonce': wpNonce
  }
})

// 搜索商品
const searchResponse = await fetch('/wp-json/tanzanite/v1/products?keyword=手机&category=5', {
  headers: {
    'X-WP-Nonce': wpNonce
  }
})

// 高级筛选
const advancedResponse = await fetch('/wp-json/tanzanite/v1/products?tags[]=summer&tags[]=sale&inventory_min=10', {
  headers: {
    'X-WP-Nonce': wpNonce
  }
})
```

**响应示例**:

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 123,
        "title": "iPhone 15 Pro",
        "status": "publish",
        "excerpt": "最新款 iPhone",
        "content": "详细描述...",
        "featured_image_id": 456,
        "thumbnail": "https://example.com/wp-content/uploads/2025/11/iphone-150x150.jpg",
        "categories": [
          {
            "id": 5,
            "name": "手机",
            "slug": "phones"
          }
        ],
        "tags": [
          {
            "id": 10,
            "name": "新品",
            "slug": "new"
          }
        ],
        "price_regular": 7999.00,
        "price_sale": 7499.00,
        "stock_qty": 100,
        "points_reward": 799,
        "is_sticky": false,
        "author_id": 1,
        "created_at": "2025-11-01 10:00:00",
        "updated_at": "2025-11-11 11:00:00"
      }
    ],
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total_pages": 5,
      "total": 95
    },
    "filters": ["keyword", "status", "category", "tags", "author", "inventory_min", "inventory_max"],
    "sorting": ["updated_at", "price_regular", "stock_qty", "points_reward"]
  }
}
```

---

## 💻 前端集成

### Nuxt.js 示例

```vue
<template>
  <div class="products-page">
    <!-- 搜索栏 -->
    <div class="search-bar">
      <input 
        v-model="filters.keyword" 
        @input="searchProducts"
        placeholder="搜索商品名称或 SKU"
      />
    </div>

    <!-- 基础筛选 -->
    <div class="basic-filters">
      <select v-model="filters.status" @change="fetchProducts">
        <option value="">全部状态</option>
        <option value="publish">已发布</option>
        <option value="draft">草稿</option>
      </select>

      <select v-model="filters.category" @change="fetchProducts">
        <option value="">全部分类</option>
        <option v-for="cat in categories" :key="cat.id" :value="cat.id">
          {{ cat.name }}
        </option>
      </select>

      <select v-model="filters.sort" @change="fetchProducts">
        <option value="updated_at">更新时间</option>
        <option value="price_regular">价格</option>
        <option value="stock_qty">库存</option>
      </select>

      <select v-model="filters.order" @change="fetchProducts">
        <option value="DESC">降序</option>
        <option value="ASC">升序</option>
      </select>
    </div>

    <!-- 高级筛选 -->
    <div v-if="showAdvanced" class="advanced-filters">
      <input 
        v-model="filters.tags" 
        placeholder="搜索标签"
      />
      <input 
        v-model.number="filters.inventory_min" 
        type="number"
        placeholder="最小库存"
      />
      <input 
        v-model.number="filters.inventory_max" 
        type="number"
        placeholder="最大库存"
      />
    </div>

    <!-- 商品列表 -->
    <div class="products-list">
      <div 
        v-for="product in products" 
        :key="product.id"
        class="product-item"
      >
        <img :src="product.thumbnail" :alt="product.title" />
        <h3>{{ product.title }}</h3>
        <p>价格: ¥{{ product.price_sale || product.price_regular }}</p>
        <p>库存: {{ product.stock_qty }}</p>
        <div class="categories">
          <span v-for="cat in product.categories" :key="cat.id">
            {{ cat.name }}
          </span>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination">
      <button 
        @click="prevPage" 
        :disabled="pagination.page === 1"
      >
        上一页
      </button>
      <span>第 {{ pagination.page }} / {{ pagination.total_pages }} 页</span>
      <button 
        @click="nextPage" 
        :disabled="pagination.page === pagination.total_pages"
      >
        下一页
      </button>
    </div>
  </div>
</template>

<script setup>
const { $wpApi } = useNuxtApp()

const products = ref([])
const categories = ref([])
const pagination = ref({
  page: 1,
  per_page: 20,
  total_pages: 1,
  total: 0
})

const filters = reactive({
  keyword: '',
  status: '',
  category: '',
  tags: '',
  sort: 'updated_at',
  order: 'DESC',
  inventory_min: null,
  inventory_max: null
})

const showAdvanced = ref(false)

// 获取商品列表
const fetchProducts = async () => {
  const params = {
    page: pagination.value.page,
    per_page: pagination.value.per_page,
    ...filters
  }

  const response = await $wpApi('/products', { params })
  
  if (response.success) {
    products.value = response.data.items
    pagination.value = response.data.pagination
  }
}

// 搜索防抖
const searchProducts = useDebounceFn(() => {
  pagination.value.page = 1
  fetchProducts()
}, 500)

// 分页
const prevPage = () => {
  if (pagination.value.page > 1) {
    pagination.value.page--
    fetchProducts()
  }
}

const nextPage = () => {
  if (pagination.value.page < pagination.value.total_pages) {
    pagination.value.page++
    fetchProducts()
  }
}

// 初始加载
onMounted(() => {
  fetchProducts()
  fetchCategories()
})
</script>
```

---

## 🎯 使用场景

### 1. 日常商品管理
- 查看所有商品
- 快速搜索商品
- 编辑商品信息
- 删除过期商品

### 2. 库存管理
- 查找低库存商品
- 批量补货
- 库存盘点

### 3. 促销活动
- 筛选特定分类商品
- 批量修改价格
- 添加促销标签

### 4. 数据分析
- 按分类统计商品数量
- 查看热门商品
- 导出商品数据

---

## 📝 注意事项

### 1. 性能优化
- 大量商品时建议使用筛选
- 避免一次加载过多数据
- 使用分页减少服务器压力

### 2. 权限控制
- 普通用户只能查看
- 编辑需要 `tanz_edit_products` 权限
- 删除需要管理员权限

### 3. 数据一致性
- 删除商品前检查订单关联
- 批量操作前确认选择
- 重要操作建议备份

---

## 🐛 常见问题

### Q: 为什么筛选没有结果？

**A**: 检查以下几点：
1. 筛选条件是否过于严格
2. 分类法是否正确注册
3. 数据库中是否有对应数据

### Q: 如何批量修改商品？

**A**: 
1. 勾选需要修改的商品
2. 选择批量操作类型
3. 点击"应用"按钮

### Q: 搜索速度慢怎么办？

**A**:
1. 为数据库添加索引
2. 使用更精确的搜索词
3. 启用对象缓存

---

## 🔗 相关页面

- [Add New Product](./ADD_PRODUCT.md) - 添加新商品
- [Attributes](./ATTRIBUTES.md) - 商品属性管理
- [Reviews](./REVIEWS.md) - 商品评论

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
