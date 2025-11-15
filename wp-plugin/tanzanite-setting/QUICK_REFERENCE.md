# Tanzanite Settings 快速参考指南

**版本**: 0.2.1  
**快速查找**: Ctrl+F 搜索关键词

---

## 🚀 快速开始

### 安装
```bash
1. 上传插件到 /wp-content/plugins/tanzanite-setting/
2. 激活插件
3. 自动创建数据库表
```

### 基础配置
```
1. 配置支付方式 → Payment Method
2. 设置税率 → Tax Rates
3. 配置积分系统 → Loyalty & Points
4. 添加商品 → Add New Product
```

---

## 📋 功能页面速查

| 页面 | 路径 | 权限 | 功能 |
|------|------|------|------|
| All Products | `tanzanite-settings` | `tanz_view_products` | 商品列表 |
| Add Product | `tanzanite-settings-add-product` | `tanz_edit_products` | 添加商品 |
| Attributes | `tanzanite-settings-attributes` | `manage_options` | 商品属性 |
| Reviews | `tanzanite-settings-reviews` | `tanz_view_products` | 商品评论 |
| All Orders | `tanzanite-settings-orders` | `tanz_view_orders` | 订单列表 |
| Payment Method | `tanzanite-settings-payment` | `manage_options` | 支付方式 |
| Tax Rates | `tanzanite-settings-tax-rates` | `manage_options` | 税率管理 |
| Shipping Templates | `tanzanite-settings-shipping` | `manage_options` | 运费模板 |
| Carriers | `tanzanite-settings-carriers` | `manage_options` | 物流公司 |
| Loyalty & Points | `tanzanite-settings-rewards` | `manage_options` | 积分系统 |
| Gift Cards & Coupons | `tanzanite-settings-giftcards` | `manage_options` | 礼品卡优惠券 |
| Member Profiles | `tanzanite-settings-members` | `manage_options` | 会员管理 |
| Audit Logs | `tanzanite-settings-audit` | `manage_options` | 审计日志 |

---

## 🔌 API 端点速查

### 商品 API
```
GET    /tanzanite/v1/products              # 商品列表
GET    /tanzanite/v1/products/{id}         # 商品详情
POST   /tanzanite/v1/products              # 创建商品
PUT    /tanzanite/v1/products/{id}         # 更新商品
DELETE /tanzanite/v1/products/{id}         # 删除商品
GET    /tanzanite/v1/categories            # 分类列表
GET    /tanzanite/v1/tags                  # 标签列表
```

### 订单 API
```
GET    /tanzanite/v1/orders                # 订单列表
GET    /tanzanite/v1/orders/{id}           # 订单详情
POST   /tanzanite/v1/orders                # 创建订单
PUT    /tanzanite/v1/orders/{id}           # 更新订单
```

### 积分 API
```
GET    /tanzanite/v1/loyalty/points        # 用户积分
POST   /tanzanite/v1/loyalty/checkin       # 每日签到
POST   /tanzanite/v1/loyalty/referral/generate  # 生成推荐码
POST   /tanzanite/v1/loyalty/referral/apply     # 应用推荐码
GET    /tanzanite/v1/loyalty/referral/stats     # 推荐统计
```

### 优惠券 API
```
GET    /tanzanite/v1/coupons               # 优惠券列表
POST   /tanzanite/v1/coupons/validate      # 验证优惠券
POST   /tanzanite/v1/coupons/apply         # 应用优惠券
```

### 礼品卡 API
```
GET    /tanzanite/v1/giftcards             # 礼品卡列表
POST   /tanzanite/v1/giftcards/validate    # 验证礼品卡
POST   /tanzanite/v1/giftcards/apply       # 应用礼品卡
POST   /tanzanite/v1/redeem/exchange       # 积分兑换
```

---

## 🗄️ 数据库表速查

| 表名 | 用途 | 主要字段 |
|------|------|---------|
| `wp_tanz_orders` | 订单 | id, order_number, user_id, status, total |
| `wp_tanz_order_items` | 订单商品 | order_id, product_id, quantity, price |
| `wp_tanz_payment_methods` | 支付方式 | id, name, code, icon_url, currencies |
| `wp_tanz_tax_rates` | 税率 | id, name, rate, region, is_active |
| `wp_tanz_coupons` | 优惠券 | id, code, discount_type, discount_value |
| `wp_tanz_giftcards` | 礼品卡 | id, card_code, balance, cover_image |
| `wp_tanz_rewards_transactions` | 积分交易 | user_id, action, points_delta, notes |
| `wp_tanz_carriers` | 物流公司 | id, name, code, tracking_url |
| `wp_tanz_audit_logs` | 审计日志 | action, resource_type, user_id |

---

## 🎯 常用代码片段

### 前端 - 获取商品列表
```javascript
const { $wpApi } = useNuxtApp()

const products = await $wpApi('/products', {
  params: {
    page: 1,
    per_page: 20,
    category: 5
  }
})
```

### 前端 - 创建订单
```javascript
const order = await $wpApi('/orders', {
  method: 'POST',
  body: {
    user_id: 123,
    items: [
      { product_id: 456, quantity: 1, price: 999 }
    ],
    payment_method: 'alipay'
  }
})
```

### 前端 - 每日签到
```javascript
const result = await $wpApi('/loyalty/checkin', {
  method: 'POST'
})

if (result.success) {
  alert(`签到成功！获得 ${result.data.points_earned} 积分`)
}
```

### 后端 - 添加钩子
```php
// 订单创建后
add_action('tanzanite_order_created', function($order_id, $order_data) {
    // 自定义逻辑
}, 10, 2);

// 积分增加后
add_action('tanzanite_points_earned', function($user_id, $points, $reason) {
    // 发送通知
}, 10, 3);
```

---

## 🔧 常用配置

### 积分系统配置
```php
// 配置存储在 option: tanzanite_loyalty_config
{
  "enabled": true,
  "points_per_unit": 10,           // 1元=10积分
  "daily_checkin_points": 10,      // 签到积分
  "referral": {
    "enabled": true,
    "bonus_inviter": 50,           // 邀请者奖励
    "bonus_invitee": 30            // 被邀请者奖励
  }
}
```

### 积分兑换配置
```php
// 配置存储在 option: tz_redeem_*
tz_redeem_enabled: '1'
tz_redeem_exchange_rate: 100        // 100积分=1元
tz_redeem_min_points: 1000          // 最少1000积分
tz_redeem_max_value_per_day: 500    // 每天最多500元
tz_redeem_card_expiry_days: 365     // 有效期365天
```

---

## 📊 订单状态流转

```
pending (待支付)
    ↓
paid (已支付)
    ↓
shipped (已发货)
    ↓
completed (已完成)

可随时转为:
cancelled (已取消)
refunded (已退款)
```

---

## 🎁 积分获取方式

| 方式 | 配置项 | 默认值 | API |
|------|--------|--------|-----|
| 订单完成 | `points_per_unit` | 1元=1积分 | 自动 |
| 每日签到 | `daily_checkin_points` | 10积分 | `POST /loyalty/checkin` |
| 推荐好友 | `referral.bonus_inviter` | 50积分 | `POST /loyalty/referral/apply` |
| 被推荐 | `referral.bonus_invitee` | 30积分 | `POST /loyalty/referral/apply` |
| 商品评论 | 待实现 | - | - |
| 社交分享 | 待实现 | - | - |

---

## 🔐 权限列表

| 权限 | 说明 | 默认角色 |
|------|------|---------|
| `tanz_view_products` | 查看商品 | Administrator, Shop Manager |
| `tanz_edit_products` | 编辑商品 | Administrator, Shop Manager |
| `tanz_view_orders` | 查看订单 | Administrator, Shop Manager |
| `tanz_edit_orders` | 编辑订单 | Administrator, Shop Manager |
| `manage_options` | 管理设置 | Administrator |

---

## 🐛 故障排除速查

### 分类法无效
```
问题: 商品筛选提示"无效的分类法"
解决: 
1. 确认分类法已注册 (tanz_product_category, tanz_product_tag)
2. 刷新固定链接 (设置 → 固定链接 → 保存)
3. 检查代码中分类法名称一致性
```

### REST API 403
```
问题: API 请求返回 403 Forbidden
解决:
1. 检查 Nonce 是否正确
2. 确认用户有相应权限
3. 检查 .htaccess 配置
```

### 数据库表未创建
```
问题: 插件激活后表未创建
解决:
1. 停用并重新激活插件
2. 检查数据库用户权限
3. 查看 WordPress 调试日志
```

---

## 📞 获取帮助

### 文档
- [完整文档](./docs/INDEX.md)
- [REST API 文档](./docs/REST_API.md)
- [功能页面文档](./docs/)

### 支持
- GitHub Issues
- 邮箱: support@tanzanite.com
- 论坛: forum.tanzanite.com

---

## 🔗 相关链接

- [插件主页](../README.md)
- [更新日志](../README.md#更新日志)
- [开发指南](./docs/INDEX.md#开发文档)

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
