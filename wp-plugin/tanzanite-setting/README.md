# Tanzanite Settings Plugin

**版本**: 0.2.1  
**作者**: Tanzanite Team  
**WordPress 版本**: 5.0+  
**PHP 版本**: 7.4+

---

## 📋 插件简介

Tanzanite Settings 是一个功能强大的 WordPress 电商管理插件，提供完整的商品管理、订单处理、会员系统、积分奖励等功能。专为现代电商网站设计，支持 REST API 对接，适合与前端框架（如 Nuxt.js、React）配合使用。

---

## ✨ 核心功能

### 🛍️ 商品管理
- **商品列表** - 完整的商品 CRUD 操作
- **商品属性** - 自定义商品属性和规格
- **商品分类** - 层级分类管理
- **商品标签** - 灵活的标签系统
- **商品评论** - 评分和评论管理
- **SKU 管理** - 多规格商品支持
- **库存管理** - 实时库存跟踪

### 📦 订单管理
- **订单列表** - 订单查询和筛选
- **订单详情** - 完整的订单信息展示
- **订单状态** - 多状态流转管理
- **批量操作** - 订单批量处理
- **订单导出** - 数据导出功能

### 💳 支付与税费
- **支付方式** - 多种支付方式配置
- **支付图标** - 自定义支付方式图标
- **多货币支持** - 支持多种货币
- **税率管理** - 灵活的税率配置
- **税费计算** - 自动税费计算

### 🚚 物流管理
- **物流模板** - 运费模板配置
- **物流公司** - 物流公司管理
- **物流追踪** - 订单追踪功能
- **批量发货** - 订单批量发货

### 🎁 营销与奖励
- **积分系统** - 完整的积分获取和消费
- **优惠券** - 优惠券创建和管理
- **礼品卡** - 礼品卡发行和使用
- **积分兑换** - 积分兑换礼品卡
- **推荐奖励** - 推荐好友获得积分
- **每日签到** - 签到获得积分

### 👥 会员管理
- **会员列表** - 会员信息管理
- **会员等级** - 基于积分的等级系统
- **会员资料** - 完整的会员资料

### 📊 系统功能
- **审计日志** - 完整的操作日志
- **数据导入** - SKU 批量导入
- **REST API** - 完整的 API 接口
- **权限管理** - 细粒度权限控制
- **URL 管理** - 自定义 URL 结构和重定向
- **SEO 管理** - 多语言 SEO、结构化数据、站点地图

---

## 📁 功能页面列表

### 商品相关
1. [All Products](./docs/ALL_PRODUCTS.md) - 商品列表管理
2. [Add New Product](./docs/ADD_PRODUCT.md) - 添加新商品
3. [Attributes](./docs/ATTRIBUTES.md) - 商品属性管理
4. [Reviews](./docs/REVIEWS.md) - 商品评论管理

### 订单相关
5. [All Orders](./docs/ALL_ORDERS.md) - 订单列表
6. [Order Detail](./docs/ORDER_DETAIL.md) - 订单详情
7. [Order Bulk](./docs/ORDER_BULK.md) - 订单批量操作

### 支付与税费
8. [Payment Method](./docs/PAYMENT_METHOD.md) - 支付方式管理
9. [Tax Rates](./docs/TAX_RATES.md) - 税率管理

### 物流相关
10. [Shipping Templates](./docs/SHIPPING_TEMPLATES.md) - 运费模板
11. [Carriers & Tracking](./docs/CARRIERS.md) - 物流公司管理
12. [Tracking Providers](./docs/TRACKING_PROVIDERS.md) - 物流追踪设置

### 营销与会员
13. [Loyalty & Points](./docs/LOYALTY_SETTINGS.md) - 积分系统设置
14. [Gift Cards & Coupons](./docs/GIFTCARDS_COUPONS.md) - 礼品卡和优惠券
15. [Member Profiles](./docs/MEMBER_PROFILES.md) - 会员管理

### 系统功能
16. [SKU Importer](./docs/SKU_IMPORTER.md) - SKU 批量导入
17. [Audit Logs](./docs/AUDIT_LOGS.md) - 审计日志
18. [URLLink](./docs/URLLINK.md) - URL 管理与重写
19. [SEO Settings](./docs/SEO_SETTINGS.md) - SEO 设置与优化

---

## 🔌 REST API

插件提供完整的 REST API 接口，所有功能都可以通过 API 访问。

### API 命名空间
```
/wp-json/tanzanite/v1/
```

### 主要 API 端点

#### 商品 API
```
GET    /products           - 获取商品列表
GET    /products/{id}      - 获取单个商品
POST   /products           - 创建商品
PUT    /products/{id}      - 更新商品
DELETE /products/{id}      - 删除商品
```

#### 订单 API
```
GET    /orders             - 获取订单列表
GET    /orders/{id}        - 获取订单详情
POST   /orders             - 创建订单
PUT    /orders/{id}        - 更新订单
```

#### 支付方式 API
```
GET    /payment-methods    - 获取支付方式列表
POST   /payment-methods    - 创建支付方式
PUT    /payment-methods/{id} - 更新支付方式
```

#### 税率 API
```
GET    /tax-rates          - 获取税率列表
POST   /tax-rates          - 创建税率
PUT    /tax-rates/{id}     - 更新税率
```

#### 积分 API
```
GET    /loyalty/points     - 获取用户积分
POST   /loyalty/checkin    - 每日签到
POST   /loyalty/referral/generate - 生成推荐码
POST   /loyalty/referral/apply - 应用推荐码
```

#### 优惠券 API
```
GET    /coupons            - 获取优惠券列表
POST   /coupons/validate   - 验证优惠券
POST   /coupons/apply      - 应用优惠券
```

#### 礼品卡 API
```
GET    /giftcards          - 获取礼品卡列表
POST   /giftcards/validate - 验证礼品卡
POST   /giftcards/apply    - 应用礼品卡
```

**详细 API 文档**: [REST API 完整文档](./docs/REST_API.md)

---

## 🗄️ 数据库表

插件创建以下自定义数据库表：

### 商品相关
- `wp_tanz_products` - 商品主表（使用 WordPress 自定义文章类型）
- `wp_tanz_attributes` - 商品属性
- `wp_tanz_attribute_terms` - 属性值

### 订单相关
- `wp_tanz_orders` - 订单主表
- `wp_tanz_order_items` - 订单商品项
- `wp_tanz_order_skus` - 订单 SKU 详情

### 支付与税费
- `wp_tanz_payment_methods` - 支付方式
- `wp_tanz_tax_rates` - 税率

### 物流相关
- `wp_tanz_shipping_templates` - 运费模板
- `wp_tanz_carriers` - 物流公司

### 营销与奖励
- `wp_tanz_coupons` - 优惠券
- `wp_tanz_giftcards` - 礼品卡
- `wp_tanz_rewards_transactions` - 积分交易记录

### 系统表
- `wp_tanz_audit_logs` - 审计日志

---

## 🚀 安装与激活

### 安装步骤

1. **上传插件**
   ```bash
   上传到 /wp-content/plugins/tanzanite-setting/
   ```

2. **激活插件**
   - 进入 WordPress 后台
   - 插件 → 已安装的插件
   - 找到 "Tanzanite Settings"
   - 点击"激活"

3. **自动初始化**
   - 插件会自动创建数据库表
   - 注册自定义文章类型和分类法
   - 初始化默认配置

### 系统要求

- WordPress 5.0+
- PHP 7.4+
- MySQL 5.7+ 或 MariaDB 10.2+
- 推荐内存：256MB+

---

## ⚙️ 配置

### 基础配置

1. **积分系统**
   - 进入 Loyalty & Points 页面
   - 配置积分获取规则
   - 设置积分兑换比例

2. **支付方式**
   - 进入 Payment Method 页面
   - 添加支付方式
   - 上传支付图标

3. **税率设置**
   - 进入 Tax Rates 页面
   - 添加税率模板
   - 关联到商品

4. **物流设置**
   - 进入 Shipping Templates 页面
   - 创建运费模板
   - 配置物流公司

### 高级配置

详见各功能页面的文档。

---

## 🔐 权限系统

插件使用自定义权限控制访问：

### 权限列表

| 权限 | 说明 | 默认角色 |
|------|------|---------|
| `tanz_view_products` | 查看商品 | Administrator, Shop Manager |
| `tanz_edit_products` | 编辑商品 | Administrator, Shop Manager |
| `tanz_view_orders` | 查看订单 | Administrator, Shop Manager |
| `tanz_edit_orders` | 编辑订单 | Administrator, Shop Manager |
| `manage_options` | 管理设置 | Administrator |

### 权限配置

可以使用插件如 "User Role Editor" 来自定义权限分配。

---

## 🎨 前端集成

### Nuxt.js 集成示例

```javascript
// composables/useProducts.ts
export const useProducts = () => {
  const { $wpApi } = useNuxtApp()

  const getProducts = async (params = {}) => {
    return await $wpApi('/products', { params })
  }

  const getProduct = async (id) => {
    return await $wpApi(`/products/${id}`)
  }

  return {
    getProducts,
    getProduct
  }
}
```

### React 集成示例

```javascript
// hooks/useProducts.js
import { useState, useEffect } from 'react'

export const useProducts = () => {
  const [products, setProducts] = useState([])
  const [loading, setLoading] = useState(false)

  const fetchProducts = async () => {
    setLoading(true)
    const response = await fetch('/wp-json/tanzanite/v1/products')
    const data = await response.json()
    setProducts(data.items)
    setLoading(false)
  }

  useEffect(() => {
    fetchProducts()
  }, [])

  return { products, loading, fetchProducts }
}
```

---

## 📝 开发指南

### 添加自定义功能

```php
// 添加自定义 REST API 端点
add_action('rest_api_init', function() {
    register_rest_route('tanzanite/v1', '/custom-endpoint', [
        'methods' => 'GET',
        'callback' => 'my_custom_callback',
        'permission_callback' => 'is_user_logged_in'
    ]);
});
```

### 钩子和过滤器

```php
// 修改商品查询
add_filter('tanzanite_products_query_args', function($args) {
    // 自定义查询参数
    return $args;
});

// 订单创建后
add_action('tanzanite_order_created', function($order_id, $order_data) {
    // 自定义逻辑
}, 10, 2);
```

---

## 🐛 故障排除

### 常见问题

#### 1. 分类法无效错误

**问题**: 商品筛选时提示 "无效的分类法"

**解决方案**:
- 确保已激活插件
- 刷新 WordPress 重写规则（设置 → 固定链接 → 保存更改）
- 检查分类法是否正确注册

#### 2. REST API 403 错误

**问题**: API 请求返回 403 Forbidden

**解决方案**:
- 检查 Nonce 是否正确
- 确认用户有相应权限
- 检查服务器 .htaccess 配置

#### 3. 数据库表未创建

**问题**: 插件激活后数据库表未创建

**解决方案**:
- 停用并重新激活插件
- 检查数据库用户权限
- 查看 WordPress 调试日志

### 调试模式

启用 WordPress 调试：

```php
// wp-config.php
define('WP_DEBUG', true);
define('WP_DEBUG_LOG', true);
define('WP_DEBUG_DISPLAY', false);
```

---

## 📊 性能优化

### 数据库优化

```sql
-- 为常用查询添加索引
ALTER TABLE wp_tanz_orders ADD INDEX idx_status (status);
ALTER TABLE wp_tanz_orders ADD INDEX idx_user (user_id);
ALTER TABLE wp_tanz_products ADD INDEX idx_status (post_status);
```

### 缓存建议

- 使用对象缓存（Redis/Memcached）
- 启用 WordPress 页面缓存
- 使用 CDN 加速静态资源

---

## 🔄 更新日志

### 0.2.1 (2025-11-11)
- ✅ 修复商品分类法名称不一致问题
- ✅ 新增商品标签分类法
- ✅ 修复积分系统严重错误
- ✅ 添加礼品卡封面图片功能
- ✅ 完善积分获取功能（订单、签到、推荐）

### 0.2.0 (2025-11-10)
- ✅ 重构积分系统
- ✅ 添加优惠券和礼品卡功能
- ✅ 完善 REST API

### 0.1.7 (2025-11-09)
- ✅ 添加税率管理功能
- ✅ 优化商品编辑页面

### 0.1.6 (2025-11-08)
- ✅ 添加支付方式图标上传
- ✅ 支持多货币配置

### 0.1.5 (2025-11-07)
- ✅ 初始版本发布
- ✅ 基础商品和订单管理

---

## 📞 支持与反馈

### 技术支持

- **文档**: [完整文档](./docs/)
- **问题反馈**: GitHub Issues
- **邮箱**: support@tanzanite.com

### 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

---

## 📄 许可证

本插件采用 GPL v2 或更高版本许可证。

---

## 🙏 致谢

感谢所有贡献者和用户的支持！

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
