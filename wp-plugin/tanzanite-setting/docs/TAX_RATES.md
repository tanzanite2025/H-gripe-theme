# Tax Rates - 税率管理

**页面路径**: `admin.php?page=tanzanite-settings-tax-rates`  
**权限要求**: `manage_options`  
**REST API**: `/wp-json/tanzanite/v1/tax-rates`

---

## 📋 功能概述

Tax Rates 页面提供灵活的税率管理功能，支持创建多个税率模板，并可关联到商品。

---

## ✨ 主要功能

### 1. 税率列表

#### 显示字段
- **税率名称** - 税率模板名称
- **税率** - 税率百分比
- **适用地区** - 税率适用的地区
- **描述** - 税率说明
- **状态** - 启用/禁用
- **排序** - 显示顺序
- **操作** - 编辑/删除

---

### 2. 创建/编辑税率

#### 表单字段

**基本信息**
- **税率名称** (必填)
  - 例如：增值税、消费税
- **税率** (必填)
  - 百分比，例如：13（表示 13%）
- **适用地区** (可选)
  - 例如：中国大陆、香港、台湾
- **描述** (可选)
  - 税率说明和适用范围

**状态设置**
- **启用/禁用** - 是否激活此税率
- **排序** - 显示顺序（数字越小越靠前）

---

### 3. 税率应用

#### 商品关联

在商品编辑页面：
1. 找到"税率模板"多选框
2. 选择适用的税率
3. 保存商品

#### 税费计算

**计算公式**:
```
税费 = 商品价格 × 税率百分比
总价 = 商品价格 + 税费
```

**示例**:
```
商品价格：¥100
税率：13%
税费：100 × 0.13 = ¥13
总价：100 + 13 = ¥113
```

---

## 🔌 REST API

### 获取税率列表

**端点**: `GET /wp-json/tanzanite/v1/tax-rates`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "name": "增值税",
        "rate": 13.00,
        "region": "中国大陆",
        "description": "一般纳税人增值税",
        "is_active": true,
        "sort_order": 1,
        "meta": null,
        "created_at": "2025-11-01 10:00:00",
        "updated_at": "2025-11-01 10:00:00"
      }
    ]
  }
}
```

### 创建税率

**端点**: `POST /wp-json/tanzanite/v1/tax-rates`

**请求示例**:
```json
{
  "name": "增值税",
  "rate": 13,
  "region": "中国大陆",
  "description": "一般纳税人增值税",
  "is_active": true,
  "sort_order": 1
}
```

### 更新税率

**端点**: `PUT /wp-json/tanzanite/v1/tax-rates/{id}`

### 删除税率

**端点**: `DELETE /wp-json/tanzanite/v1/tax-rates/{id}`

---

## 💻 前端集成

### Nuxt.js 示例

```vue
<template>
  <div class="tax-calculator">
    <h3>价格计算器</h3>
    
    <input 
      v-model.number="price" 
      type="number"
      placeholder="输入商品价格"
    />

    <select v-model="selectedTaxRate">
      <option value="">选择税率</option>
      <option v-for="tax in taxRates" :key="tax.id" :value="tax.rate">
        {{ tax.name }} ({{ tax.rate }}%)
      </option>
    </select>

    <div class="result">
      <p>商品价格: ¥{{ price }}</p>
      <p>税费: ¥{{ taxAmount }}</p>
      <p>总价: ¥{{ totalPrice }}</p>
    </div>
  </div>
</template>

<script setup>
const { $wpApi } = useNuxtApp()

const price = ref(100)
const selectedTaxRate = ref(0)
const taxRates = ref([])

// 计算税费
const taxAmount = computed(() => {
  return (price.value * selectedTaxRate.value / 100).toFixed(2)
})

// 计算总价
const totalPrice = computed(() => {
  return (parseFloat(price.value) + parseFloat(taxAmount.value)).toFixed(2)
})

// 获取税率列表
const fetchTaxRates = async () => {
  const response = await $wpApi('/tax-rates')
  if (response.success) {
    taxRates.value = response.data.items.filter(item => item.is_active)
  }
}

onMounted(() => {
  fetchTaxRates()
})
</script>
```

---

## 🎯 使用场景

### 1. 跨境电商
- 不同国家/地区税率
- 自动计算税费
- 税费明细展示

### 2. 合规要求
- 符合税法规定
- 税费记录
- 税务报表

### 3. 价格透明
- 显示含税价
- 税费明细
- 用户知情权

---

## 📝 注意事项

### 1. 税率设置
- 确保税率准确
- 定期更新税率
- 区分不同地区

### 2. 商品关联
- 正确选择税率
- 检查税费计算
- 测试订单流程

### 3. 法律合规
- 遵守当地税法
- 保留税费记录
- 定期审计

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
