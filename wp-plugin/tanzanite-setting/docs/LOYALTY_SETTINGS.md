# Loyalty & Points - 积分系统设置

**页面路径**: `admin.php?page=tanzanite-settings-rewards`  
**权限要求**: `manage_options`  
**REST API**: `/wp-json/tanzanite/v1/loyalty/*`

---

## 📋 功能概述

Loyalty & Points 页面提供完整的积分系统配置，包括积分获取规则、积分兑换设置、推荐奖励等功能。

---

## ✨ 主要功能

### 1. 积分系统配置

#### 基础设置

**启用/禁用积分系统**
- 全局开关
- 关闭后所有积分功能停用

**积分名称**
- 自定义积分显示名称
- 例如：积分、金币、点数

**积分有效期**
- 永久有效
- 固定期限（天数）
- 滚动期限

---

### 2. 积分获取规则

#### 订单完成获得积分

**配置项**:
- **启用/禁用** - 订单完成是否奖励积分
- **积分比例** - 每消费 1 元获得的积分数
  - 例如：1 元 = 1 积分
  - 例如：1 元 = 10 积分
- **最小订单金额** - 低于此金额不奖励积分
- **排除商品** - 某些商品不参与积分奖励

**计算公式**:
```
积分 = 订单金额 × 积分比例
```

**示例**:
```
订单金额：¥100
积分比例：1 元 = 10 积分
获得积分：100 × 10 = 1000 积分
```

---

#### 每日签到获得积分

**配置项**:
- **启用/禁用** - 是否开启签到功能
- **每日积分** - 每次签到获得的积分数
- **连续签到奖励** - 连续签到额外奖励
  - 连续 7 天：额外 50 积分
  - 连续 30 天：额外 200 积分

**API 端点**:
```
POST /wp-json/tanzanite/v1/loyalty/checkin
```

**请求示例**:
```javascript
const response = await fetch('/wp-json/tanzanite/v1/loyalty/checkin', {
  method: 'POST',
  headers: {
    'X-WP-Nonce': wpNonce
  }
})

// 响应
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

#### 推荐好友获得积分

**配置项**:
- **启用/禁用** - 是否开启推荐功能
- **邀请者奖励** - 邀请人获得的积分
- **被邀请者奖励** - 新用户获得的积分
- **奖励条件** - 新用户完成首单后发放

**推荐流程**:

1. **生成推荐码**
```javascript
POST /wp-json/tanzanite/v1/loyalty/referral/generate

// 响应
{
  "success": true,
  "data": {
    "code": "REF12345678",
    "url": "https://example.com/?ref=REF12345678",
    "message": "推荐码生成成功"
  }
}
```

2. **应用推荐码**
```javascript
POST /wp-json/tanzanite/v1/loyalty/referral/apply
Body: { "code": "REF12345678" }

// 响应
{
  "success": true,
  "data": {
    "message": "推荐码应用成功",
    "points_earned": 30,
    "total_points": 30
  }
}
```

3. **查看推荐统计**
```javascript
GET /wp-json/tanzanite/v1/loyalty/referral/stats

// 响应
{
  "success": true,
  "data": {
    "referral_code": "REF12345678",
    "referral_url": "https://example.com/?ref=REF12345678",
    "referral_count": 5,
    "referred_by": null
  }
}
```

---

#### 商品评论获得积分

**配置项**:
- **启用/禁用** - 评论是否奖励积分
- **文字评论积分** - 纯文字评论奖励
- **图片评论积分** - 带图评论额外奖励
- **精华评论积分** - 被标记为精华的评论奖励
- **每个商品限制** - 每个商品只能评论一次获得积分

---

#### 社交分享获得积分

**配置项**:
- **启用/禁用** - 分享是否奖励积分
- **每日分享次数** - 每天可获得积分的分享次数
- **每次分享积分** - 每次分享获得的积分
- **分享平台** - 支持的分享平台
  - 微信
  - 微博
  - QQ
  - 朋友圈

---

### 3. 积分消费规则

#### 积分抵扣现金

**配置项**:
- **启用/禁用** - 是否允许积分抵扣
- **兑换比例** - 积分与现金的兑换比例
  - 例如：100 积分 = 1 元
- **最低使用积分** - 最少使用多少积分
- **最高抵扣比例** - 订单最多可用积分抵扣的比例
  - 例如：最多抵扣订单金额的 50%
- **排除商品** - 某些商品不可使用积分

**计算公式**:
```
抵扣金额 = 使用积分 ÷ 兑换比例
```

**示例**:
```
使用积分：1000 积分
兑换比例：100 积分 = 1 元
抵扣金额：1000 ÷ 100 = 10 元
```

---

#### 积分兑换礼品卡

**配置项**:
- **启用/禁用** - 是否允许积分兑换礼品卡
- **兑换比例** - 积分与礼品卡的兑换比例
  - 例如：1000 积分 = 10 元礼品卡
- **最低兑换积分** - 最少兑换多少积分
- **每日兑换限额** - 每天最多兑换金额
- **礼品卡有效期** - 兑换的礼品卡有效天数
- **预设面额** - 可选的礼品卡面额
  - 10 元、50 元、100 元、200 元、500 元

**API 端点**:
```
POST /wp-json/tanzanite/v1/redeem/exchange
```

**请求示例**:
```javascript
const response = await fetch('/wp-json/tanzanite/v1/redeem/exchange', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-WP-Nonce': wpNonce
  },
  body: JSON.stringify({
    points: 1000,
    value: 10
  })
})

// 响应
{
  "success": true,
  "data": {
    "giftcard": {
      "id": 123,
      "card_code": "GC20251111001",
      "balance": 10.00,
      "expires_at": "2026-11-11 00:00:00"
    },
    "message": "兑换成功"
  }
}
```

---

### 4. 会员等级系统

#### 等级配置

**等级划分**:
- **Ordinary** - 0-499 积分
- **Bronze** - 500-1999 积分
- **Silver** - 2000-4999 积分
- **Gold** - 5000-9999 积分
- **Platinum** - 10000+ 积分

**等级权益**:
- **折扣优惠** - 不同等级享受不同折扣
- **积分倍率** - 高等级会员获得更多积分
- **专属活动** - 高等级会员专属活动
- **优先服务** - 优先发货、优先客服

---

### 5. 积分交易记录

#### 记录内容
- **用户 ID** - 积分所属用户
- **交易类型** - 获得/消费
- **积分变动** - 增加或减少的积分数
- **关联订单** - 相关订单 ID
- **操作说明** - 积分变动原因
- **创建时间** - 交易时间

#### 查询功能
- 按用户查询
- 按时间范围查询
- 按交易类型查询
- 导出交易记录

---

## 💻 前端集成

### Nuxt.js 示例

#### 1. 用户积分展示

```vue
<template>
  <div class="user-points">
    <div class="points-balance">
      <h3>我的积分</h3>
      <p class="points">{{ userPoints }}</p>
      <p class="level">{{ userLevel }}</p>
    </div>

    <div class="points-actions">
      <button @click="checkin" :disabled="checkedInToday">
        {{ checkedInToday ? '今日已签到' : '每日签到' }}
      </button>
      <button @click="showReferral">
        邀请好友
      </button>
      <button @click="exchangeGiftcard">
        兑换礼品卡
      </button>
    </div>

    <div class="points-history">
      <h4>积分记录</h4>
      <div v-for="record in pointsHistory" :key="record.id" class="record-item">
        <span>{{ record.notes }}</span>
        <span :class="record.action === 'earn' ? 'earn' : 'spend'">
          {{ record.action === 'earn' ? '+' : '-' }}{{ record.points_delta }}
        </span>
        <span>{{ formatDate(record.created_at) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
const { $wpApi } = useNuxtApp()

const userPoints = ref(0)
const userLevel = ref('')
const checkedInToday = ref(false)
const pointsHistory = ref([])

// 获取用户积分
const fetchUserPoints = async () => {
  const response = await $wpApi('/loyalty/points')
  if (response.success) {
    userPoints.value = response.data.points
    userLevel.value = response.data.level
  }
}

// 每日签到
const checkin = async () => {
  const response = await $wpApi('/loyalty/checkin', {
    method: 'POST'
  })

  if (response.success) {
    alert(`签到成功！获得 ${response.data.points_earned} 积分`)
    userPoints.value = response.data.total_points
    checkedInToday.value = true
  } else {
    alert(response.data.message || '签到失败')
  }
}

// 显示推荐码
const showReferral = async () => {
  const response = await $wpApi('/loyalty/referral/stats')
  
  if (response.success) {
    const { referral_code, referral_url } = response.data
    
    if (referral_code) {
      alert(`你的推荐码：${referral_code}\n推荐链接：${referral_url}`)
    } else {
      // 生成推荐码
      const generateResponse = await $wpApi('/loyalty/referral/generate', {
        method: 'POST'
      })
      
      if (generateResponse.success) {
        alert(`推荐码已生成：${generateResponse.data.code}`)
      }
    }
  }
}

// 兑换礼品卡
const exchangeGiftcard = async () => {
  const points = prompt('输入要兑换的积分数（1000积分=10元）:')
  
  if (!points) return

  const value = parseInt(points) / 100

  const response = await $wpApi('/redeem/exchange', {
    method: 'POST',
    body: { points: parseInt(points), value }
  })

  if (response.success) {
    alert(`兑换成功！礼品卡号：${response.data.giftcard.card_code}`)
    fetchUserPoints()
  } else {
    alert(response.data.message || '兑换失败')
  }
}

// 获取积分记录
const fetchPointsHistory = async () => {
  const response = await $wpApi('/loyalty/transactions')
  if (response.success) {
    pointsHistory.value = response.data.items
  }
}

onMounted(() => {
  fetchUserPoints()
  fetchPointsHistory()
})
</script>

<style scoped>
.points-balance {
  text-align: center;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 12px;
}

.points {
  font-size: 48px;
  font-weight: bold;
  margin: 10px 0;
}

.earn { color: #22c55e; }
.spend { color: #ef4444; }
</style>
```

---

## 🎯 使用场景

### 1. 用户激励
- 鼓励用户下单
- 提高用户活跃度
- 增加用户粘性

### 2. 营销活动
- 新用户注册送积分
- 节日双倍积分
- 会员日积分加倍

### 3. 用户留存
- 积分有效期促进消费
- 会员等级体系
- 推荐奖励拉新

---

## 📝 注意事项

### 1. 积分安全
- 防止积分刷取
- 异常交易监控
- 积分回滚机制

### 2. 性能优化
- 积分计算异步处理
- 交易记录定期归档
- 使用缓存减少查询

### 3. 用户体验
- 积分变动及时通知
- 清晰的积分规则说明
- 简单的兑换流程

---

## 🔗 相关页面

- [Gift Cards & Coupons](./GIFTCARDS_COUPONS.md) - 礼品卡和优惠券
- [Member Profiles](./MEMBER_PROFILES.md) - 会员管理

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
