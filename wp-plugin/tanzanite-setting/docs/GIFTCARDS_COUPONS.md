# Gift Cards & Coupons - 礼品卡和优惠券

**页面路径**: `admin.php?page=tanzanite-settings-giftcards`  
**权限要求**: `manage_options`  
**REST API**: `/wp-json/tanzanite/v1/giftcards`, `/wp-json/tanzanite/v1/coupons`

---

## 📋 功能概述

Gift Cards & Coupons 页面提供完整的礼品卡和优惠券管理功能，支持创建、验证、应用和积分兑换。

---

## 🎁 礼品卡管理

### 1. 创建礼品卡

**字段**:
- **卡号** (card_code) - 自动生成或手动输入
- **面额** (balance) - 礼品卡金额
- **原始金额** (original_value) - 初始金额
- **货币** (currency) - 默认 CNY
- **所有者** (owner_user_id) - 用户 ID
- **封面图片** (cover_image) - 卡片封面 URL
- **状态** (status) - active/used/expired
- **有效期** (expires_at) - 过期时间

**示例**:
```json
{
  "card_code": "GC20251111001",
  "balance": 100.00,
  "original_value": 100.00,
  "currency": "CNY",
  "cover_image": "https://example.com/card-cover.jpg",
  "status": "active",
  "expires_at": "2026-11-11"
}
```

---

### 2. 礼品卡验证

**API**: `POST /tanzanite/v1/giftcards/validate`

**请求**:
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
    "can_use": 50.00,
    "message": "礼品卡可用"
  }
}
```

---

### 3. 礼品卡应用

**API**: `POST /tanzanite/v1/giftcards/apply`

**请求**:
```json
{
  "card_code": "GC20251111001",
  "amount": 50.00,
  "order_id": 1001
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "remaining_balance": 50.00,
    "message": "礼品卡已应用"
  }
}
```

---

### 4. 积分兑换礼品卡

**API**: `POST /tanzanite/v1/redeem/exchange`

**配置**:
- 兑换比例：100 积分 = 1 元
- 最低兑换：1000 积分
- 每日限额：500 元
- 有效期：365 天

**请求**:
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

## 🎫 优惠券管理

### 1. 创建优惠券

**字段**:
- **优惠券代码** (code) - 唯一代码
- **折扣类型** (discount_type) - fixed_amount/percentage
- **折扣值** (discount_value) - 金额或百分比
- **最低消费** (min_purchase) - 使用门槛
- **使用次数限制** (usage_limit) - 总次数
- **每用户限制** (usage_limit_per_user) - 单用户次数
- **有效期** (valid_from, valid_to) - 时间范围
- **状态** (is_active) - 启用/禁用

**示例**:
```json
{
  "code": "SUMMER2025",
  "discount_type": "fixed_amount",
  "discount_value": 10.00,
  "min_purchase": 50.00,
  "usage_limit": 100,
  "usage_limit_per_user": 1,
  "valid_from": "2025-06-01",
  "valid_to": "2025-08-31",
  "is_active": true
}
```

---

### 2. 优惠券验证

**API**: `POST /tanzanite/v1/coupons/validate`

**请求**:
```json
{
  "code": "SUMMER2025",
  "total": 100.00,
  "user_id": 123
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

**验证规则**:
- ✅ 优惠券存在且启用
- ✅ 在有效期内
- ✅ 未超过使用次数
- ✅ 用户未超过限制
- ✅ 订单金额满足最低消费

---

### 3. 优惠券应用

**API**: `POST /tanzanite/v1/coupons/apply`

**请求**:
```json
{
  "code": "SUMMER2025",
  "order_id": 1001,
  "user_id": 123
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "discount": 10.00,
    "message": "优惠券已应用"
  }
}
```

---

## 💻 前端集成

### Nuxt.js 示例

```vue
<template>
  <div class="checkout-page">
    <!-- 优惠券输入 -->
    <div class="coupon-section">
      <input 
        v-model="couponCode" 
        placeholder="输入优惠券代码"
      />
      <button @click="applyCoupon">应用</button>
      <p v-if="couponDiscount > 0" class="success">
        已优惠 ¥{{ couponDiscount }}
      </p>
    </div>

    <!-- 礼品卡输入 -->
    <div class="giftcard-section">
      <input 
        v-model="giftcardCode" 
        placeholder="输入礼品卡号"
      />
      <button @click="applyGiftcard">使用</button>
      <p v-if="giftcardDiscount > 0" class="success">
        已抵扣 ¥{{ giftcardDiscount }}
      </p>
    </div>

    <!-- 订单总计 -->
    <div class="order-summary">
      <p>商品总额: ¥{{ subtotal }}</p>
      <p v-if="couponDiscount > 0">优惠券: -¥{{ couponDiscount }}</p>
      <p v-if="giftcardDiscount > 0">礼品卡: -¥{{ giftcardDiscount }}</p>
      <p class="total">应付金额: ¥{{ finalTotal }}</p>
    </div>
  </div>
</template>

<script setup>
const { $wpApi } = useNuxtApp()

const couponCode = ref('')
const giftcardCode = ref('')
const couponDiscount = ref(0)
const giftcardDiscount = ref(0)
const subtotal = ref(100)

const finalTotal = computed(() => {
  return Math.max(0, subtotal.value - couponDiscount.value - giftcardDiscount.value)
})

// 应用优惠券
const applyCoupon = async () => {
  const response = await $wpApi('/coupons/validate', {
    method: 'POST',
    body: {
      code: couponCode.value,
      total: subtotal.value
    }
  })

  if (response.success && response.data.valid) {
    couponDiscount.value = response.data.discount
    alert('优惠券已应用')
  } else {
    alert(response.data.message || '优惠券无效')
  }
}

// 应用礼品卡
const applyGiftcard = async () => {
  const response = await $wpApi('/giftcards/validate', {
    method: 'POST',
    body: {
      card_code: giftcardCode.value,
      amount: finalTotal.value
    }
  })

  if (response.success && response.data.valid) {
    giftcardDiscount.value = response.data.can_use
    alert('礼品卡已应用')
  } else {
    alert(response.data.message || '礼品卡无效')
  }
}
</script>
```

---

## 🎯 使用场景

### 1. 促销活动
- 节日优惠券
- 满减活动
- 新用户优惠

### 2. 会员福利
- 生日礼品卡
- 会员专属优惠券
- 积分兑换

### 3. 营销推广
- 推荐奖励
- 社交分享
- 邮件营销

---

## 📝 注意事项

### 1. 安全性
- 优惠券代码唯一性
- 防止重复使用
- 验证用户权限

### 2. 性能
- 缓存优惠券规则
- 异步验证
- 批量查询优化

### 3. 用户体验
- 清晰的错误提示
- 实时验证反馈
- 优惠明细展示

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
