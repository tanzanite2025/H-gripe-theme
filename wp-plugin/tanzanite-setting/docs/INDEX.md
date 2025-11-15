# Tanzanite Settings 文档索引

**插件版本**: 0.2.1  
**最后更新**: 2025-11-11

---

## 📚 文档导航

### 🎯 快速开始
- [README](../README.md) - 插件总览和安装指南
- [REST API 完整文档](./REST_API.md) - API 接口文档

---

### 🛍️ 商品管理

#### 核心功能
- [All Products](./ALL_PRODUCTS.md) - 商品列表管理
  - 商品查询和筛选
  - 批量操作
  - 分类和标签管理
  
- [Add New Product](./ADD_PRODUCT.md) - 添加新商品
  - 商品基本信息
  - SKU 规格管理
  - 价格和库存设置
  
- [Attributes](./ATTRIBUTES.md) - 商品属性管理
  - 属性创建和编辑
  - 属性值管理
  - 商品关联

- [Reviews](./REVIEWS.md) - 商品评论管理
  - 评论审核
  - 评分管理
  - 评论回复

---

### 📦 订单管理

#### 订单处理
- [All Orders](./ALL_ORDERS.md) - 订单列表
  - 订单查询
  - 状态管理
  - 批量操作
  
- [Order Detail](./ORDER_DETAIL.md) - 订单详情
  - 订单信息查看
  - 订单编辑
  - 物流信息

- [Order Bulk](./ORDER_BULK.md) - 批量操作
  - 批量发货
  - 批量导出
  - 批量状态更新

---

### 💳 支付与税费

#### 支付配置
- [Payment Method](./PAYMENT_METHOD.md) - 支付方式管理
  - 支付方式配置
  - 图标上传
  - 多货币支持

#### 税费管理
- [Tax Rates](./TAX_RATES.md) - 税率管理
  - 税率创建
  - 区域设置
  - 商品关联

---

### 🚚 物流管理

#### 物流配置
- [Shipping Templates](./SHIPPING_TEMPLATES.md) - 运费模板
  - 模板创建
  - 区域运费
  - 免邮设置

- [Carriers & Tracking](./CARRIERS.md) - 物流公司
  - 物流公司管理
  - 物流代码
  - API 配置

- [Tracking Providers](./TRACKING_PROVIDERS.md) - 物流追踪
  - 追踪设置
  - 查询接口
  - 通知配置

---

### 🎁 营销与奖励

#### 积分系统
- [Loyalty & Points](./LOYALTY_SETTINGS.md) - 积分系统
  - 积分获取规则
  - 积分消费规则
  - 会员等级
  - 推荐奖励
  - 每日签到

#### 优惠券和礼品卡
- [Gift Cards & Coupons](./GIFTCARDS_COUPONS.md) - 礼品卡和优惠券
  - 优惠券管理
  - 礼品卡发行
  - 积分兑换
  - 使用规则

---

### 👥 会员管理

#### 会员系统
- [Member Profiles](./MEMBER_PROFILES.md) - 会员管理
  - 会员列表
  - 会员等级
  - 积分查询
  - 会员资料

---

### 📊 系统功能

#### 数据管理
- [SKU Importer](./SKU_IMPORTER.md) - SKU 批量导入
  - CSV 导入
  - 数据映射
  - 错误处理

- [Audit Logs](./AUDIT_LOGS.md) - 审计日志
  - 操作记录
  - 日志查询
  - 安全审计

#### URL 管理
- [URLLink](./URLLINK.md) - URL 管理与重写
  - URL 目录树
  - 自定义 URL 路径
  - 批量操作
  - 301 重定向

---

## 🔌 API 文档

### REST API 端点

#### 商品 API
```
GET    /tanzanite/v1/products
POST   /tanzanite/v1/products
PUT    /tanzanite/v1/products/{id}
DELETE /tanzanite/v1/products/{id}
GET    /tanzanite/v1/categories
GET    /tanzanite/v1/tags
```

#### 订单 API
```
GET    /tanzanite/v1/orders
POST   /tanzanite/v1/orders
PUT    /tanzanite/v1/orders/{id}
DELETE /tanzanite/v1/orders/{id}
```

#### 支付 API
```
GET    /tanzanite/v1/payment-methods
POST   /tanzanite/v1/payment-methods
PUT    /tanzanite/v1/payment-methods/{id}
```

#### 税率 API
```
GET    /tanzanite/v1/tax-rates
POST   /tanzanite/v1/tax-rates
PUT    /tanzanite/v1/tax-rates/{id}
```

#### 积分 API
```
GET    /tanzanite/v1/loyalty/points
POST   /tanzanite/v1/loyalty/checkin
POST   /tanzanite/v1/loyalty/referral/generate
POST   /tanzanite/v1/loyalty/referral/apply
GET    /tanzanite/v1/loyalty/referral/stats
```

#### 优惠券 API
```
GET    /tanzanite/v1/coupons
POST   /tanzanite/v1/coupons/validate
POST   /tanzanite/v1/coupons/apply
```

#### 礼品卡 API
```
GET    /tanzanite/v1/giftcards
POST   /tanzanite/v1/giftcards/validate
POST   /tanzanite/v1/giftcards/apply
POST   /tanzanite/v1/redeem/exchange
```

**完整 API 文档**: [REST_API.md](./REST_API.md)

---

## 📖 使用指南

### 新手入门

1. **安装插件**
   - 上传插件到 WordPress
   - 激活插件
   - 自动创建数据库表

2. **基础配置**
   - 配置支付方式
   - 设置税率
   - 创建运费模板
   - 配置积分系统

3. **添加商品**
   - 创建商品分类
   - 添加商品属性
   - 创建商品
   - 设置 SKU 和价格

4. **处理订单**
   - 查看新订单
   - 更新订单状态
   - 安排发货
   - 更新物流信息

### 高级功能

1. **积分系统**
   - 配置积分获取规则
   - 设置会员等级
   - 开启推荐奖励
   - 积分兑换礼品卡

2. **营销活动**
   - 创建优惠券
   - 发行礼品卡
   - 设置促销活动
   - 会员专享优惠

3. **数据分析**
   - 查看销售报表
   - 分析用户行为
   - 导出数据
   - 审计日志

---

## 🔧 开发文档

### 钩子和过滤器

#### 商品钩子
```php
// 商品创建后
do_action('tanzanite_product_created', $product_id, $product_data);

// 商品更新后
do_action('tanzanite_product_updated', $product_id, $product_data);

// 修改商品查询
apply_filters('tanzanite_products_query_args', $args);
```

#### 订单钩子
```php
// 订单创建后
do_action('tanzanite_order_created', $order_id, $order_data);

// 订单状态变更
do_action('tanzanite_order_status_changed', $order_id, $old_status, $new_status);

// 订单完成后
do_action('tanzanite_order_completed', $order_id, $order_data);
```

#### 积分钩子
```php
// 积分增加后
do_action('tanzanite_points_earned', $user_id, $points, $reason);

// 积分消费后
do_action('tanzanite_points_spent', $user_id, $points, $reason);
```

### 自定义开发

#### 添加自定义 REST API
```php
add_action('rest_api_init', function() {
    register_rest_route('tanzanite/v1', '/custom-endpoint', [
        'methods' => 'GET',
        'callback' => 'my_custom_callback',
        'permission_callback' => 'is_user_logged_in'
    ]);
});
```

#### 扩展商品字段
```php
add_filter('tanzanite_product_meta_fields', function($fields) {
    $fields['custom_field'] = [
        'type' => 'string',
        'label' => '自定义字段',
        'default' => ''
    ];
    return $fields;
});
```

---

## 🐛 故障排除

### 常见问题

1. **分类法无效错误**
   - [解决方案](../README.md#故障排除)

2. **REST API 403 错误**
   - [解决方案](../README.md#故障排除)

3. **数据库表未创建**
   - [解决方案](../README.md#故障排除)

### 调试技巧

1. **启用 WordPress 调试**
```php
define('WP_DEBUG', true);
define('WP_DEBUG_LOG', true);
```

2. **查看审计日志**
   - 进入 Audit Logs 页面
   - 查看操作记录

3. **检查 REST API**
```bash
# 测试 API 端点
curl -X GET "https://example.com/wp-json/tanzanite/v1/products" \
  -H "X-WP-Nonce: YOUR_NONCE"
```

---

## 📞 获取帮助

### 技术支持

- **文档**: 本文档库
- **GitHub**: [Issues](https://github.com/tanzanite/issues)
- **邮箱**: support@tanzanite.com

### 社区资源

- **论坛**: [Tanzanite 论坛](https://forum.tanzanite.com)
- **教程**: [视频教程](https://tutorials.tanzanite.com)
- **示例**: [代码示例](https://examples.tanzanite.com)

---

## 📝 贡献文档

如果你发现文档有错误或需要改进，欢迎贡献：

1. Fork 项目
2. 修改文档
3. 提交 Pull Request

---

## 📄 许可证

本文档采用 CC BY-SA 4.0 许可证。

---

**文档维护**: Tanzanite Team  
**最后更新**: 2025-11-11
