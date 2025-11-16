# 🕐 浏览历史组件使用指南

## ✅ 已完成的功能

### 1. 核心文件
- ✅ `composables/useBrowsingHistory.ts` - 数据管理
- ✅ `components/BrowsingHistory.vue` - 显示组件
- ✅ `pages/test-browsing.vue` - 测试页面

### 2. 功能特性
- ✅ 自动记录浏览商品
- ✅ 横向滚动显示
- ✅ 单个删除/批量清空
- ✅ localStorage 持久化
- ✅ 最多保存 20 个商品
- ✅ 自动去重（同一商品只保留最新）
- ✅ 响应式设计（桌面端/移动端）

---

## 🧪 测试步骤

### 1. 访问测试页面
```
http://localhost:3000/test-browsing
```

### 2. 测试功能
1. **添加商品** - 点击"添加商品 1/2/3"按钮
2. **批量添加** - 点击"批量添加 10 个"按钮
3. **查看显示** - 观察浏览历史组件的显示效果
4. **删除单个** - 鼠标悬停在商品上，点击右上角的 ❌ 按钮
5. **清空历史** - 点击"清空历史"按钮
6. **滚动测试** - 桌面端点击左右箭头，移动端左右滑动
7. **刷新测试** - 刷新页面，检查数据是否保留

---

## 📦 集成到现有页面

### 方案 A：在购物车中显示

**文件：** `components/CartDrawer.vue`

```vue
<template>
  <div>
    <!-- 现有购物车内容 -->
    <div class="cart-items">
      <!-- ... -->
    </div>

    <!-- 添加浏览历史组件 -->
    <div class="mt-6">
      <BrowsingHistory />
    </div>
  </div>
</template>

<script setup>
// 无需额外导入，组件会自动注册
</script>
```

---

### 方案 B：在商品详情页记录浏览

**文件：** `pages/product/[id].vue` 或商品详情组件

```vue
<script setup>
import { useBrowsingHistory } from '~/composables/useBrowsingHistory'

const { addToHistory } = useBrowsingHistory()

// 商品数据
const product = ref({
  id: 123,
  title: 'iPhone 15 Pro',
  thumbnail: 'https://example.com/image.jpg',
  price: '$999',
  url: '/product/123'
})

// 页面加载时记录浏览
onMounted(() => {
  if (product.value) {
    addToHistory({
      id: product.value.id,
      title: product.value.title,
      thumbnail: product.value.thumbnail,
      price: product.value.price,
      url: product.value.url
    })
  }
})
</script>
```

---

### 方案 C：在首页显示

**文件：** `pages/index.vue`

```vue
<template>
  <div>
    <!-- 其他首页内容 -->
    
    <!-- 浏览历史区块 -->
    <section class="my-8">
      <BrowsingHistory />
    </section>
  </div>
</template>
```

---

## 🎨 样式定制

### 修改颜色主题

**文件：** `components/BrowsingHistory.vue`

```vue
<!-- 修改标题栏背景色 -->
<div class="bg-gray-50">  <!-- 改为 bg-blue-50 -->

<!-- 修改价格颜色 -->
<p class="text-red-600">  <!-- 改为 text-blue-600 -->

<!-- 修改悬停效果 -->
<div class="hover:border-blue-300">  <!-- 改为其他颜色 -->
```

---

### 调整卡片尺寸

```vue
<!-- 修改卡片宽度 -->
<div class="w-32">  <!-- 改为 w-40 或 w-48 -->

<!-- 修改图片高度 -->
<div class="h-32">  <!-- 改为 h-40 或 h-48 -->
```

---

## 🔧 高级配置

### 修改最大保存数量

**文件：** `composables/useBrowsingHistory.ts`

```typescript
const MAX_ITEMS = 20  // 改为 30 或 50
```

---

### 修改存储键名

```typescript
const STORAGE_KEY = 'tz_browsing_history'  // 改为自定义名称
```

---

### 添加过期时间

```typescript
// 在 addToHistory 函数中添加
const addToHistory = (item) => {
  history.value.unshift({
    ...item,
    viewedAt: new Date().toISOString(),
    expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString() // 7天后过期
  })
  
  // 过滤过期商品
  history.value = history.value.filter(h => new Date(h.expiresAt) > new Date())
}
```

---

## 📊 数据结构

### localStorage 存储格式

```json
{
  "id": 123,
  "title": "iPhone 15 Pro Max",
  "thumbnail": "https://example.com/image.jpg",
  "price": "$1,199.00",
  "url": "/product/123",
  "viewedAt": "2025-11-15T09:30:00.000Z"
}
```

---

## ⚠️ 注意事项

### 1. 商品数据要求
- `id` - 必须唯一
- `title` - 不能为空
- `thumbnail` - 可选，但建议提供
- `price` - 建议格式化为货币格式
- `url` - 必须是有效的商品链接

### 2. 性能优化
- 组件使用 `v-if` 控制显示，无数据时不渲染
- 图片使用 `loading="lazy"` 懒加载
- 限制最大数量为 20 个，避免占用过多存储空间

### 3. 浏览器兼容性
- 需要支持 localStorage
- 需要支持 ES6+
- 移动端需要支持触摸滚动

---

## 🚀 下一步

### 可选增强功能

1. **后端同步**
   - 登录用户可以跨设备同步浏览历史
   - 需要添加后端 API

2. **智能推荐**
   - 基于浏览历史推荐相似商品
   - 需要推荐算法支持

3. **数据分析**
   - 统计用户浏览偏好
   - 用于精准营销

4. **分类筛选**
   - 按商品分类筛选浏览历史
   - 需要商品分类数据

---

## 📝 完成清单

- [x] 创建 Composable
- [x] 创建显示组件（浅色版）
- [x] 创建显示组件（深色版）
- [x] 创建测试页面
- [x] 添加使用文档
- [x] 集成到购物车 ✅
- [ ] 集成到商品详情页（待定）
- [ ] 集成到首页（待定）

---

## ✅ 已集成页面

### 1. 购物车页面
**文件：** `components/CartDrawer.vue`
**组件：** `BrowsingHistoryDark` （深色主题版本）
**位置：** 购物车商品列表下方
**状态：** ✅ 已完成

**特点：**
- 适配购物车的黑色背景
- 在有商品和空购物车时都显示
- 使用半透明白色边框和背景
- 价格显示为绿色（#40ffaa）

---

## 🎉 总结

浏览历史组件已经完成开发、测试并集成到购物车页面。

**测试地址：** 
- 测试页面：http://localhost:3000/test-browsing
- 购物车：打开购物车查看底部的浏览历史

**下一步：**
- 在商品详情页添加浏览记录功能
- 可选：在首页或其他页面显示浏览历史

**需要帮助？** 随时联系开发团队！
