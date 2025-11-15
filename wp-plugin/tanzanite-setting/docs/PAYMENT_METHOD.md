# Payment Method - 支付方式管理

**页面路径**: `admin.php?page=tanzanite-settings-payment`  
**权限要求**: `manage_options`  
**REST API**: `/wp-json/tanzanite/v1/payment-methods`

---

## 📋 功能概述

Payment Method 页面提供支付方式的完整管理，支持多种支付方式配置、图标上传和多货币支持。

---

## ✨ 主要功能

### 1. 创建支付方式

**字段**:
- **名称** (name) - 支付方式名称
- **代码** (code) - 唯一标识符
- **图标** (icon_url) - 支付图标 URL
- **支持货币** (currencies) - 货币列表
- **默认货币** (default_currency) - 默认货币
- **状态** (is_active) - 启用/禁用
- **排序** (sort_order) - 显示顺序

**示例**:
```json
{
  "name": "支付宝",
  "code": "alipay",
  "icon_url": "https://example.com/alipay.png",
  "currencies": ["CNY", "USD"],
  "default_currency": "CNY",
  "is_active": true,
  "sort_order": 1
}
```

---

### 2. 图标上传

**步骤**:
1. 点击"选择图片"按钮
2. 从媒体库选择或上传图片
3. 实时预览图标
4. 保存支付方式

**图标规范**:
- 尺寸：120x60px
- 格式：PNG、JPG、SVG
- 大小：< 100KB
- 背景：透明或白色

---

### 3. 多货币配置

**支持货币**:
```
CNY - 人民币
USD - 美元
EUR - 欧元
GBP - 英镑
JPY - 日元
HKD - 港币
```

**配置方式**:
- 输入货币代码（逗号分隔）
- 自动转换为大写
- 验证货币代码有效性

---

## 🔌 REST API

### 获取支付方式列表

```
GET /wp-json/tanzanite/v1/payment-methods
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
        "is_active": true,
        "sort_order": 1
      }
    ]
  }
}
```

---

## 💻 前端集成

```vue
<template>
  <div class="payment-methods">
    <div 
      v-for="method in paymentMethods" 
      :key="method.id"
      class="payment-option"
      :class="{ active: selected === method.code }"
      @click="selectPayment(method.code)"
    >
      <img :src="method.icon_url" :alt="method.name" />
      <span>{{ method.name }}</span>
    </div>
  </div>
</template>

<script setup>
const { $wpApi } = useNuxtApp()

const paymentMethods = ref([])
const selected = ref('')

const fetchPaymentMethods = async () => {
  const response = await $wpApi('/payment-methods')
  if (response.success) {
    paymentMethods.value = response.data.items.filter(m => m.is_active)
  }
}

const selectPayment = (code) => {
  selected.value = code
  emit('update:modelValue', code)
}

onMounted(() => {
  fetchPaymentMethods()
})
</script>
```

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
