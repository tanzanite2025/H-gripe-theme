# All Orders - 订单列表管理

**页面路径**: `admin.php?page=tanzanite-settings-orders`  
**权限要求**: `tanz_view_orders`  
**REST API**: `/wp-json/tanzanite/v1/orders`

---

## 📋 功能概述

All Orders 页面提供完整的订单管理功能，包括订单查看、筛选、状态更新和批量操作。

---

## ✨ 主要功能

### 1. 订单列表展示

#### 显示字段
- **订单号** - 唯一订单编号
- **用户 ID** - 下单用户
- **订单状态** - 当前状态
- **总金额** - 订单总价
- **支付方式** - 使用的支付方式
- **创建时间** - 下单时间
- **更新时间** - 最后更新时间
- **操作** - 查看详情/编辑/删除

#### 订单状态
- **pending** - 待支付
- **paid** - 已支付
- **shipped** - 已发货
- **completed** - 已完成
- **cancelled** - 已取消
- **refunded** - 已退款

---

### 2. 订单筛选

#### 基础筛选

**关键词搜索**
- 订单号
- 用户 ID
- 收货人姓名
- 收货人电话

**状态筛选**
- 全部状态
- 待支付
- 已支付
- 已发货
- 已完成
- 已取消
- 已退款

**时间筛选**
- 今天
- 最近 7 天
- 最近 30 天
- 自定义日期范围

**排序方式**
- 创建时间（默认）
- 更新时间
- 订单金额

---

### 3. 批量操作

#### 批量更新状态
- 批量标记为已支付
- 批量标记为已发货
- 批量标记为已完成
- 批量取消订单

#### 批量导出
- 导出选中订单
- CSV 格式
- 包含订单详情和商品信息

---

## 🔌 REST API

### 获取订单列表

**端点**: `GET /wp-json/tanzanite/v1/orders`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | integer | 否 | 页码，默认 1 |
| per_page | integer | 否 | 每页数量，默认 20 |
| keyword | string | 否 | 搜索关键词 |
| status | string | 否 | 订单状态 |
| user_id | integer | 否 | 用户 ID |
| date_from | string | 否 | 开始日期（YYYY-MM-DD） |
| date_to | string | 否 | 结束日期（YYYY-MM-DD） |
| sort | string | 否 | 排序字段 |
| order | string | 否 | 排序方向 |

**请求示例**:

```javascript
// 获取待支付订单
const response = await fetch('/wp-json/tanzanite/v1/orders?status=pending', {
  headers: {
    'X-WP-Nonce': wpNonce
  }
})

// 搜索订单
const searchResponse = await fetch('/wp-json/tanzanite/v1/orders?keyword=ORD20251111001', {
  headers: {
    'X-WP-Nonce': wpNonce
  }
})

// 按日期筛选
const dateResponse = await fetch('/wp-json/tanzanite/v1/orders?date_from=2025-11-01&date_to=2025-11-11', {
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
        "id": 1001,
        "order_number": "ORD20251111001",
        "user_id": 123,
        "status": "paid",
        "total": 7999.00,
        "subtotal": 7999.00,
        "tax": 0.00,
        "shipping_fee": 0.00,
        "discount": 0.00,
        "coupon_code": null,
        "coupon_discount": 0.00,
        "giftcard_discount": 0.00,
        "points_used": 0,
        "points_value": 0.00,
        "payment_method": "alipay",
        "shipping_name": "张三",
        "shipping_phone": "13800138000",
        "shipping_address": "北京市朝阳区...",
        "tracking_number": null,
        "carrier_code": null,
        "notes": "",
        "created_at": "2025-11-11 10:00:00",
        "updated_at": "2025-11-11 10:30:00"
      }
    ],
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total_pages": 10,
      "total": 195
    }
  }
}
```

### 更新订单状态

**端点**: `PUT /wp-json/tanzanite/v1/orders/{id}`

**请求参数**:

```json
{
  "status": "shipped",
  "tracking_number": "SF1234567890",
  "carrier_code": "shunfeng"
}
```

**响应示例**:

```json
{
  "success": true,
  "data": {
    "message": "订单状态已更新"
  }
}
```

---

## 💻 前端集成

### Nuxt.js 示例

```vue
<template>
  <div class="orders-page">
    <!-- 筛选栏 -->
    <div class="filters">
      <input 
        v-model="filters.keyword" 
        @input="searchOrders"
        placeholder="搜索订单号或用户"
      />

      <select v-model="filters.status" @change="fetchOrders">
        <option value="">全部状态</option>
        <option value="pending">待支付</option>
        <option value="paid">已支付</option>
        <option value="shipped">已发货</option>
        <option value="completed">已完成</option>
        <option value="cancelled">已取消</option>
      </select>

      <input 
        v-model="filters.date_from" 
        type="date"
        @change="fetchOrders"
      />
      <input 
        v-model="filters.date_to" 
        type="date"
        @change="fetchOrders"
      />
    </div>

    <!-- 订单列表 -->
    <table class="orders-table">
      <thead>
        <tr>
          <th><input type="checkbox" @change="selectAll" /></th>
          <th>订单号</th>
          <th>用户</th>
          <th>状态</th>
          <th>金额</th>
          <th>支付方式</th>
          <th>创建时间</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="order in orders" :key="order.id">
          <td><input type="checkbox" v-model="selected" :value="order.id" /></td>
          <td>{{ order.order_number }}</td>
          <td>{{ order.user_id }}</td>
          <td>
            <span :class="`status-${order.status}`">
              {{ getStatusText(order.status) }}
            </span>
          </td>
          <td>¥{{ order.total }}</td>
          <td>{{ order.payment_method }}</td>
          <td>{{ formatDate(order.created_at) }}</td>
          <td>
            <button @click="viewOrder(order.id)">查看</button>
            <button @click="updateStatus(order.id)">更新状态</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- 批量操作 -->
    <div v-if="selected.length > 0" class="bulk-actions">
      <select v-model="bulkAction">
        <option value="">选择操作</option>
        <option value="paid">标记为已支付</option>
        <option value="shipped">标记为已发货</option>
        <option value="completed">标记为已完成</option>
        <option value="cancelled">取消订单</option>
      </select>
      <button @click="applyBulkAction">应用</button>
    </div>

    <!-- 分页 -->
    <div class="pagination">
      <button @click="prevPage" :disabled="pagination.page === 1">
        上一页
      </button>
      <span>第 {{ pagination.page }} / {{ pagination.total_pages }} 页</span>
      <button @click="nextPage" :disabled="pagination.page === pagination.total_pages">
        下一页
      </button>
    </div>
  </div>
</template>

<script setup>
const { $wpApi } = useNuxtApp()

const orders = ref([])
const selected = ref([])
const bulkAction = ref('')
const pagination = ref({
  page: 1,
  per_page: 20,
  total_pages: 1,
  total: 0
})

const filters = reactive({
  keyword: '',
  status: '',
  date_from: '',
  date_to: ''
})

// 获取订单列表
const fetchOrders = async () => {
  const params = {
    page: pagination.value.page,
    per_page: pagination.value.per_page,
    ...filters
  }

  const response = await $wpApi('/orders', { params })
  
  if (response.success) {
    orders.value = response.data.items
    pagination.value = response.data.pagination
  }
}

// 更新订单状态
const updateStatus = async (orderId) => {
  const newStatus = prompt('输入新状态（paid/shipped/completed/cancelled）:')
  
  if (!newStatus) return

  const response = await $wpApi(`/orders/${orderId}`, {
    method: 'PUT',
    body: { status: newStatus }
  })

  if (response.success) {
    alert('状态已更新')
    fetchOrders()
  }
}

// 批量操作
const applyBulkAction = async () => {
  if (!bulkAction.value || selected.value.length === 0) return

  const confirmed = confirm(`确定要对 ${selected.value.length} 个订单执行此操作吗？`)
  
  if (!confirmed) return

  for (const orderId of selected.value) {
    await $wpApi(`/orders/${orderId}`, {
      method: 'PUT',
      body: { status: bulkAction.value }
    })
  }

  alert('批量操作完成')
  selected.value = []
  bulkAction.value = ''
  fetchOrders()
}

// 状态文本
const getStatusText = (status) => {
  const statusMap = {
    pending: '待支付',
    paid: '已支付',
    shipped: '已发货',
    completed: '已完成',
    cancelled: '已取消',
    refunded: '已退款'
  }
  return statusMap[status] || status
}

// 初始加载
onMounted(() => {
  fetchOrders()
})
</script>

<style scoped>
.status-pending { color: #f59e0b; }
.status-paid { color: #10b981; }
.status-shipped { color: #3b82f6; }
.status-completed { color: #22c55e; }
.status-cancelled { color: #ef4444; }
.status-refunded { color: #6b7280; }
</style>
```

---

## 🎯 使用场景

### 1. 订单处理
- 查看新订单
- 确认支付状态
- 安排发货
- 更新物流信息

### 2. 客户服务
- 查询订单状态
- 处理退款申请
- 修改订单信息

### 3. 数据分析
- 统计订单数量
- 分析销售趋势
- 导出订单数据

---

## 📝 注意事项

### 1. 订单状态流转
```
pending → paid → shipped → completed
         ↓
    cancelled / refunded
```

### 2. 权限控制
- 查看订单：`tanz_view_orders`
- 编辑订单：`tanz_edit_orders`
- 删除订单：管理员

### 3. 数据安全
- 敏感信息脱敏
- 操作日志记录
- 定期备份数据

---

## 🔗 相关页面

- [Order Detail](./ORDER_DETAIL.md) - 订单详情
- [Order Bulk](./ORDER_BULK.md) - 批量操作

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
