# 🛒 购物车数据对接分析报告

## 📊 当前状况

### ✅ 购物车组件现状

**文件位置：** `app/components/CartDrawer.vue`

**数据来源：** `app/composables/useCart.ts`

**存储方式：** 
- ❌ **仅使用 localStorage** - 数据存储在浏览器本地
- ❌ **没有对接后端 API** - 没有与 WordPress 数据库同步
- ❌ **没有使用 tanzanite-setting 插件**

### 🔍 代码分析

#### 1. 购物车数据结构

```typescript
export interface CartItem {
  id: number           // 商品 ID
  title: string        // 商品标题
  slug: string         // 商品 slug
  price: number        // 价格
  quantity: number     // 数量
  thumbnail?: string   // 缩略图
  sku?: string        // SKU
  maxStock?: number   // 最大库存
  weight?: number     // 重量
}
```

#### 2. 数据存储位置

```typescript
// useCart.ts 第 33-42 行
if (import.meta.client) {
  const saved = localStorage.getItem('tanzanite_cart')
  if (saved) {
    try {
      cartItems.value = JSON.parse(saved)
    } catch (e) {
      console.error('Failed to load cart from localStorage', e)
    }
  }
}
```

**问题：**
- ✅ 数据只存在浏览器 localStorage
- ❌ 清除浏览器缓存后数据丢失
- ❌ 不同设备/浏览器无法同步
- ❌ 没有持久化到数据库

---

## ⚠️ 存在的问题

### 1. 数据不持久化
- 购物车数据只存在浏览器本地
- 用户换设备或清除缓存后数据丢失
- 无法跨设备同步

### 2. 没有后端验证
- 价格、库存等信息没有实时验证
- 可能出现价格不一致
- 库存控制不准确

### 3. 缺少订单管理
- 没有订单创建功能
- 没有订单历史记录
- 没有支付集成

### 4. 没有与 WooCommerce 集成
- 如果网站使用 WooCommerce，购物车数据不同步
- 无法使用 WooCommerce 的商品、库存、订单系统

---

## 💡 解决方案

### 方案 1: 创建 tanzanite-setting 购物车 API（推荐）⭐⭐⭐⭐⭐

#### 需要在 tanzanite-setting 插件中添加：

**1. 数据库表设计**

```sql
-- 购物车表
CREATE TABLE wp_tanzanite_cart (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  user_id bigint(20) NOT NULL,
  session_id varchar(255) DEFAULT NULL,
  product_id bigint(20) NOT NULL,
  quantity int(11) NOT NULL DEFAULT 1,
  price decimal(10,2) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id),
  KEY user_id (user_id),
  KEY session_id (session_id)
);

-- 订单表
CREATE TABLE wp_tanzanite_orders (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  user_id bigint(20) NOT NULL,
  order_number varchar(50) NOT NULL,
  status varchar(20) NOT NULL DEFAULT 'pending',
  subtotal decimal(10,2) NOT NULL,
  shipping decimal(10,2) NOT NULL,
  tax decimal(10,2) NOT NULL,
  total decimal(10,2) NOT NULL,
  shipping_address text,
  payment_method varchar(50),
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY order_number (order_number),
  KEY user_id (user_id)
);

-- 订单商品表
CREATE TABLE wp_tanzanite_order_items (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  order_id bigint(20) NOT NULL,
  product_id bigint(20) NOT NULL,
  quantity int(11) NOT NULL,
  price decimal(10,2) NOT NULL,
  subtotal decimal(10,2) NOT NULL,
  PRIMARY KEY (id),
  KEY order_id (order_id)
);
```

**2. REST API 端点**

需要在插件中添加以下 API：

```php
// includes/api/class-cart-api.php

class Tanzanite_Cart_API {
  
  // GET /wp-json/tanzanite/v1/cart
  // 获取购物车
  public function get_cart($request) {
    $user_id = get_current_user_id();
    $session_id = $this->get_session_id();
    
    // 查询购物车
    $cart_items = $this->get_cart_items($user_id, $session_id);
    
    return rest_ensure_response([
      'success' => true,
      'data' => $cart_items
    ]);
  }
  
  // POST /wp-json/tanzanite/v1/cart/add
  // 添加到购物车
  public function add_to_cart($request) {
    $product_id = $request->get_param('product_id');
    $quantity = $request->get_param('quantity');
    
    // 验证商品
    $product = wc_get_product($product_id);
    if (!$product) {
      return new WP_Error('invalid_product', '商品不存在', ['status' => 404]);
    }
    
    // 检查库存
    if (!$product->has_enough_stock($quantity)) {
      return new WP_Error('out_of_stock', '库存不足', ['status' => 400]);
    }
    
    // 添加到购物车
    $this->add_cart_item($product_id, $quantity);
    
    return rest_ensure_response([
      'success' => true,
      'message' => '已添加到购物车'
    ]);
  }
  
  // PUT /wp-json/tanzanite/v1/cart/update
  // 更新购物车数量
  public function update_cart($request) {
    $cart_item_id = $request->get_param('cart_item_id');
    $quantity = $request->get_param('quantity');
    
    $this->update_cart_item($cart_item_id, $quantity);
    
    return rest_ensure_response([
      'success' => true,
      'message' => '购物车已更新'
    ]);
  }
  
  // DELETE /wp-json/tanzanite/v1/cart/remove
  // 从购物车移除
  public function remove_from_cart($request) {
    $cart_item_id = $request->get_param('cart_item_id');
    
    $this->remove_cart_item($cart_item_id);
    
    return rest_ensure_response([
      'success' => true,
      'message' => '已从购物车移除'
    ]);
  }
  
  // POST /wp-json/tanzanite/v1/cart/sync
  // 同步本地购物车到服务器
  public function sync_cart($request) {
    $local_cart = $request->get_param('cart_items');
    
    // 合并本地和服务器购物车
    $merged_cart = $this->merge_carts($local_cart);
    
    return rest_ensure_response([
      'success' => true,
      'data' => $merged_cart
    ]);
  }
  
  // POST /wp-json/tanzanite/v1/orders/create
  // 创建订单
  public function create_order($request) {
    $cart_items = $request->get_param('cart_items');
    $shipping_address = $request->get_param('shipping_address');
    $payment_method = $request->get_param('payment_method');
    
    // 创建订单
    $order = $this->create_order_from_cart($cart_items, $shipping_address, $payment_method);
    
    return rest_ensure_response([
      'success' => true,
      'data' => [
        'order_id' => $order->id,
        'order_number' => $order->order_number
      ]
    ]);
  }
}
```

**3. 管理后台页面**

在 WordPress 后台添加菜单：

```
WordPress 后台
└── Tanzanite Settings
    ├── Dashboard
    ├── Loyalty Settings
    ├── 🛒 Cart & Orders (新增)
    │   ├── Active Carts (活跃购物车)
    │   ├── Orders (订单列表)
    │   └── Settings (购物车设置)
    └── General Settings
```

**管理页面功能：**
- 查看所有用户的购物车
- 查看订单列表
- 订单状态管理
- 购物车设置（税率、运费规则等）

---

### 方案 2: 集成 WooCommerce（如果已安装）⭐⭐⭐⭐

如果网站已经使用 WooCommerce：

**优势：**
- ✅ 无需重复开发
- ✅ 完整的电商功能
- ✅ 成熟的支付集成
- ✅ 丰富的插件生态

**需要修改：**

```typescript
// useCart.ts 修改为调用 WooCommerce API

const addToCart = async (product: Omit<CartItem, 'quantity'>) => {
  try {
    const response = await $fetch('/wp-json/wc/store/v1/cart/add-item', {
      method: 'POST',
      body: {
        id: product.id,
        quantity: 1
      }
    })
    
    // 更新本地状态
    cartItems.value = response.items
    return { success: true, message: 'Added to cart' }
  } catch (error) {
    return { success: false, message: 'Failed to add to cart' }
  }
}
```

---

### 方案 3: 混合方案（推荐用于过渡）⭐⭐⭐

**保留 localStorage 作为缓存，同时同步到服务器：**

```typescript
// useCart.ts 修改

const addToCart = async (product: Omit<CartItem, 'quantity'>) => {
  // 1. 先添加到本地（快速响应）
  const existingItem = cartItems.value.find(item => item.id === product.id)
  if (existingItem) {
    existingItem.quantity++
  } else {
    cartItems.value.push({ ...product, quantity: 1 })
  }
  saveCart() // 保存到 localStorage
  
  // 2. 同步到服务器（后台）
  try {
    await $fetch('/wp-json/tanzanite/v1/cart/sync', {
      method: 'POST',
      body: {
        cart_items: cartItems.value
      }
    })
  } catch (error) {
    console.error('Failed to sync cart to server', error)
  }
  
  return { success: true, message: 'Added to cart' }
}

// 页面加载时从服务器恢复购物车
const loadCartFromServer = async () => {
  try {
    const response = await $fetch('/wp-json/tanzanite/v1/cart')
    if (response.success && response.data) {
      cartItems.value = response.data
      saveCart() // 同步到 localStorage
    }
  } catch (error) {
    // 如果服务器失败，使用本地缓存
    console.error('Failed to load cart from server', error)
  }
}
```

---

## 📋 实施建议

### 阶段 1: 立即修复（1-2天）

1. **添加服务器同步**
   - 修改 `useCart.ts` 添加 API 调用
   - 实现购物车同步功能

2. **创建基础 API**
   - 在 tanzanite-setting 插件中添加购物车 API
   - 实现基本的增删改查

### 阶段 2: 完善功能（3-5天）

1. **添加订单系统**
   - 创建订单表
   - 实现订单创建 API
   - 添加订单管理页面

2. **添加管理后台**
   - 购物车列表页面
   - 订单管理页面
   - 设置页面

### 阶段 3: 优化体验（2-3天）

1. **添加支付集成**
   - 集成支付网关
   - 添加支付回调

2. **优化性能**
   - 添加缓存
   - 优化数据库查询

---

## 🎯 推荐方案

**推荐使用方案 1 + 方案 3 的组合：**

1. ✅ 在 tanzanite-setting 插件中创建购物车 API
2. ✅ 保留 localStorage 作为本地缓存（快速响应）
3. ✅ 自动同步到服务器（数据持久化）
4. ✅ 添加管理后台页面

**优势：**
- 用户体验好（本地缓存快速响应）
- 数据安全（服务器持久化）
- 跨设备同步
- 便于管理和统计

---

## 📁 需要创建的文件

### tanzanite-setting 插件

```
tanzanite-setting/
├── includes/
│   ├── api/
│   │   ├── class-cart-api.php          (新增)
│   │   └── class-order-api.php         (新增)
│   ├── admin/
│   │   ├── class-cart-admin.php        (新增)
│   │   └── class-order-admin.php       (新增)
│   └── database/
│       └── class-cart-schema.php       (新增)
```

### 前端修改

```
tanzanite-theme/
└── app/
    └── composables/
        └── useCart.ts                   (修改)
```

---

## ✅ 结论

**当前购物车系统的问题：**
- ❌ 只使用 localStorage，数据不持久化
- ❌ 没有对接后端 API
- ❌ 没有使用 tanzanite-setting 插件
- ❌ 缺少订单管理功能

**建议：**
1. 在 tanzanite-setting 插件中创建购物车和订单管理系统
2. 添加 REST API 端点
3. 修改前端代码对接 API
4. 添加管理后台页面

**预计工作量：**
- 后端开发：5-7 天
- 前端对接：2-3 天
- 测试优化：2-3 天
- **总计：9-13 天**

需要我帮你开始实现吗？我可以先创建 API 端点的代码框架。
