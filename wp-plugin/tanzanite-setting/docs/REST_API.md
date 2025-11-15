# Tanzanite Settings REST API 完整文档

**API 版本**: v1  
**命名空间**: `/wp-json/tanzanite/v1/`  
**认证方式**: WordPress Nonce / JWT Token

---

## 📋 API 概述

Tanzanite Settings 提供完整的 REST API 接口，支持所有后台功能的前端调用。

---

## 🔐 认证

### Nonce 认证（推荐用于同域请求）

```javascript
// 获取 Nonce（在 WordPress 页面中）
const nonce = wpApiSettings.nonce

// 发送请求
fetch('/wp-json/tanzanite/v1/products', {
  headers: {
    'X-WP-Nonce': nonce
  }
})
```

### JWT Token 认证（推荐用于跨域/移动端）

```javascript
// 登录获取 Token
const loginResponse = await fetch('/wp-json/jwt-auth/v1/token', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'user',
    password: 'pass'
  })
})

const { token } = await loginResponse.json()

// 使用 Token
fetch('/wp-json/tanzanite/v1/products', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
})
```

---

## 🛍️ 商品 API

### 获取商品列表

```
GET /tanzanite/v1/products
```

**参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| per_page | int | 每页数量 |
| keyword | string | 搜索关键词 |
| status | string | 状态筛选 |
| category | int | 分类 ID |
| tags | array | 标签 slug |
| sort | string | 排序字段 |
| order | string | ASC/DESC |

**示例**:
```javascript
GET /wp-json/tanzanite/v1/products?page=1&per_page=20&category=5
```

---

### 获取单个商品

```
GET /tanzanite/v1/products/{id}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "id": 123,
    "title": "iPhone 15 Pro",
    "price_regular": 7999.00,
    "price_sale": 7499.00,
    "stock_qty": 100,
    "categories": [...],
    "tags": [...]
  }
}
```

---

### 创建商品

```
POST /tanzanite/v1/products
```

**权限**: `tanz_edit_products`

**请求体**:
```json
{
  "title": "新商品",
  "content": "商品描述",
  "status": "publish",
  "price_regular": 999.00,
  "price_sale": 899.00,
  "stock_qty": 50,
  "category_ids": [5, 10],
  "tag_ids": [1, 2, 3]
}
```

---

### 更新商品

```
PUT /tanzanite/v1/products/{id}
```

**权限**: `tanz_edit_products`

---

### 删除商品

```
DELETE /tanzanite/v1/products/{id}
```

**权限**: `tanz_edit_products`

---

## 📦 订单 API

### 获取订单列表

```
GET /tanzanite/v1/orders
```

**参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| per_page | int | 每页数量 |
| keyword | string | 搜索关键词 |
| status | string | 订单状态 |
| user_id | int | 用户 ID |
| date_from | string | 开始日期 |
| date_to | string | 结束日期 |

---

### 获取订单详情

```
GET /tanzanite/v1/orders/{id}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "id": 1001,
    "order_number": "ORD20251111001",
    "user_id": 123,
    "status": "paid",
    "total": 7999.00,
    "items": [
      {
        "product_id": 456,
        "product_title": "iPhone 15 Pro",
        "quantity": 1,
        "price": 7999.00
      }
    ],
    "shipping": {
      "name": "张三",
      "phone": "13800138000",
      "address": "北京市..."
    }
  }
}
```

---

### 创建订单

```
POST /tanzanite/v1/orders
```

**请求体**:
```json
{
  "user_id": 123,
  "items": [
    {
      "product_id": 456,
      "quantity": 1,
      "price": 7999.00
    }
  ],
  "shipping": {
    "name": "张三",
    "phone": "13800138000",
    "address": "北京市朝阳区..."
  },
  "payment_method": "alipay"
}
```

---

### 更新订单状态

```
PUT /tanzanite/v1/orders/{id}
```

**请求体**:
```json
{
  "status": "shipped",
  "tracking_number": "SF1234567890",
  "carrier_code": "shunfeng"
}
```

---

## 💳 支付方式 API

### 获取支付方式列表

```
GET /tanzanite/v1/payment-methods
```

**响应**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "name": "支付宝",
        "code": "alipay",
        "icon_url": "https://...",
        "currencies": ["CNY", "USD"],
        "default_currency": "CNY",
        "is_active": true
      }
    ]
  }
}
```

---

### 创建支付方式

```
POST /tanzanite/v1/payment-methods
```

**权限**: `manage_options`

---

## 💰 税率 API

### 获取税率列表

```
GET /tanzanite/v1/tax-rates
```

---

### 创建税率

```
POST /tanzanite/v1/tax-rates
```

**请求体**:
```json
{
  "name": "增值税",
  "rate": 13,
  "region": "中国大陆",
  "is_active": true
}
```

---

## 🎁 积分 API

### 获取用户积分

```
GET /tanzanite/v1/loyalty/points
```

**权限**: 已登录用户

**响应**:
```json
{
  "success": true,
  "data": {
    "points": 1000,
    "level": "Gold",
    "next_level": "Platinum",
    "next_level_points": 10000
  }
}
```

---

### 每日签到

```
POST /tanzanite/v1/loyalty/checkin
```

**权限**: 已登录用户

**响应**:
```json
{
  "success": true,
  "data": {
    "message": "签到成功",
    "points_earned": 10,
    "total_points": 1010,
    "checkin_streak": 5
  }
}
```

---

### 生成推荐码

```
POST /tanzanite/v1/loyalty/referral/generate
```

**权限**: 已登录用户

**响应**:
```json
{
  "success": true,
  "data": {
    "code": "REF12345678",
    "url": "https://example.com/?ref=REF12345678"
  }
}
```

---

### 应用推荐码

```
POST /tanzanite/v1/loyalty/referral/apply
```

**权限**: 已登录用户

**请求体**:
```json
{
  "code": "REF12345678"
}
```

---

### 获取推荐统计

```
GET /tanzanite/v1/loyalty/referral/stats
```

**权限**: 已登录用户

**响应**:
```json
{
  "success": true,
  "data": {
    "referral_code": "REF12345678",
    "referral_url": "https://...",
    "referral_count": 5,
    "referred_by": null
  }
}
```

---

## 🎫 优惠券 API

### 获取优惠券列表

```
GET /tanzanite/v1/coupons
```

---

### 验证优惠券

```
POST /tanzanite/v1/coupons/validate
```

**请求体**:
```json
{
  "code": "SUMMER2025",
  "total": 100.00
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "valid": true,
    "discount": 10.00,
    "discount_type": "fixed_amount",
    "message": "优惠券可用"
  }
}
```

---

### 应用优惠券

```
POST /tanzanite/v1/coupons/apply
```

**请求体**:
```json
{
  "code": "SUMMER2025",
  "order_id": 1001
}
```

---

## 🎁 礼品卡 API

### 获取礼品卡列表

```
GET /tanzanite/v1/giftcards
```

---

### 验证礼品卡

```
POST /tanzanite/v1/giftcards/validate
```

**请求体**:
```json
{
  "card_code": "GC20251111001",
  "amount": 50.00
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "valid": true,
    "balance": 100.00,
    "can_use": 50.00
  }
}
```

---

### 应用礼品卡

```
POST /tanzanite/v1/giftcards/apply
```

**请求体**:
```json
{
  "card_code": "GC20251111001",
  "amount": 50.00,
  "order_id": 1001
}
```

---

### 积分兑换礼品卡

```
POST /tanzanite/v1/redeem/exchange
```

**权限**: 已登录用户

**请求体**:
```json
{
  "points": 1000,
  "value": 10
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "giftcard": {
      "id": 123,
      "card_code": "GC20251111002",
      "balance": 10.00,
      "expires_at": "2026-11-11"
    }
  }
}
```

---

## 📊 分类和标签 API

### 获取商品分类

```
GET /tanzanite/v1/categories
```

**响应**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 5,
        "name": "手机",
        "slug": "phones",
        "parent": 0,
        "count": 25
      }
    ]
  }
}
```

---

### 获取商品标签

```
GET /tanzanite/v1/tags
```

---

## 🔍 搜索 API

### 全局搜索

```
GET /tanzanite/v1/search
```

**参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| q | string | 搜索关键词 |
| type | string | 搜索类型（product/order/user） |

---

## 📝 审计日志 API

### 获取审计日志

```
GET /tanzanite/v1/audit-logs
```

**权限**: `manage_options`

**参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| per_page | int | 每页数量 |
| action | string | 操作类型 |
| user_id | int | 用户 ID |
| date_from | string | 开始日期 |
| date_to | string | 结束日期 |

---

## 🚨 错误处理

### 错误响应格式

```json
{
  "success": false,
  "data": {
    "code": "invalid_parameter",
    "message": "参数无效",
    "status": 400
  }
}
```

### 常见错误码

| 错误码 | HTTP 状态 | 说明 |
|--------|----------|------|
| `unauthorized` | 401 | 未授权 |
| `forbidden` | 403 | 无权限 |
| `not_found` | 404 | 资源不存在 |
| `invalid_parameter` | 400 | 参数无效 |
| `server_error` | 500 | 服务器错误 |

---

## 📊 响应格式

### 成功响应

```json
{
  "success": true,
  "data": {
    // 响应数据
  }
}
```

### 列表响应

```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total_pages": 5,
      "total": 95
    }
  }
}
```

---

## 🔧 开发工具

### Postman 集合

下载 Postman 集合：[tanzanite-api.postman_collection.json](./tanzanite-api.postman_collection.json)

### cURL 示例

```bash
# 获取商品列表
curl -X GET "https://example.com/wp-json/tanzanite/v1/products" \
  -H "X-WP-Nonce: YOUR_NONCE"

# 创建商品
curl -X POST "https://example.com/wp-json/tanzanite/v1/products" \
  -H "Content-Type: application/json" \
  -H "X-WP-Nonce: YOUR_NONCE" \
  -d '{"title":"新商品","price_regular":999}'
```

---

## 📚 相关文档

- [插件总览](../README.md)
- [功能文档索引](./INDEX.md)
- [前端集成指南](./FRONTEND_INTEGRATION.md)

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
