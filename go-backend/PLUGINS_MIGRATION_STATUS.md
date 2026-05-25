# WordPress 插件迁移到 Go 后端 - 状态分析

## 📊 当前状态总览

### ✅ 已完成的基础架构
- Go 项目结构完整
- 基础领域模型（User, Post, Product, Cart等）
- Repository 和 Service 层
- REST API 框架
- 中间件系统
- 数据库迁移工具

### ⚠️ 待迁移的 WordPress 插件

目前有 **7个WordPress插件** 仍然是PHP代码，需要转换到Go后端：

| # | 插件名称 | 功能描述 | 优先级 | 状态 |
|---|---------|---------|--------|------|
| 1 | **tanzanite-blog-i18n** | 博客多语言管理 | P1 | ⚠️ 部分完成 |
| 2 | **tanzanite-setting** | 商城核心（商品/订单/物流） | P0 | ❌ 未开始 |
| 3 | **tanzanite-customer-service** | 客服系统 | P2 | ❌ 未开始 |
| 4 | **tanzanite-faq-content** | FAQ内容管理 | P1 | ⚠️ 部分完成 |
| 5 | **tanzanite-photo-gallery** | 图片库管理 | P2 | ❌ 未开始 |
| 6 | **tanzanite-product-registry** | 产品注册系统 | P2 | ❌ 未开始 |
| 7 | **tanzanite-subscription** | 订阅与通知系统 | P2 | ⚠️ 部分完成 |

---

## 📋 详细插件分析

### 1. tanzanite-blog-i18n（博客国际化）

#### 功能范围
- ✅ 文章多语言管理（34种语言）
- ✅ 翻译组关联
- ✅ 语言分类法（tz_lang）
- ✅ REST API：列表、详情、翻译映射
- ✅ 自动创建分类（news/wheelsbuild）

#### Go后端对应实现
- ✅ Post 模型已创建（支持 locale 和 parent_id）
- ✅ PostRepository 已实现
- ✅ PostService 已实现
- ✅ Content API 已创建
- ⚠️ **缺失**：翻译组UUID管理
- ⚠️ **缺失**：后台创建翻译功能

#### 需要补充
```go
// 1. 添加翻译组字段到 Post 模型
type Post struct {
    // ... 现有字段
    TranslationGroup string `gorm:"index" json:"translation_group"` // UUID
}

// 2. 添加翻译相关方法到 PostRepository
func (r *PostRepository) FindTranslations(groupID string) ([]post.Post, error)
func (r *PostRepository) CreateTranslation(sourceID uint, targetLocale string) (*post.Post, error)

// 3. 添加翻译API端点
GET /api/v1/content/posts/:id/translations
POST /api/v1/content/posts/:id/translations
```

---

### 2. tanzanite-setting（商城核心）⭐ **最重要**

#### 功能范围（非常庞大）

##### 商品管理
- ❌ 商品CRUD（已有基础，需扩展）
- ❌ 商品属性管理
- ❌ SKU管理（多规格）
- ❌ 商品评价系统
- ❌ 批量操作
- ❌ 阶梯价格
- ❌ 会员价格

##### 订单管理
- ❌ 订单CRUD
- ❌ 订单状态流转
- ❌ 订单商品项
- ❌ 批量发货
- ❌ 订单导出

##### 支付系统
- ❌ 支付方式管理
- ❌ 支付图标上传
- ❌ 多货币支持
- ❌ 税率管理
- ❌ 税费计算

##### 物流系统
- ❌ 运费模板
- ❌ 物流公司管理
- ❌ 物流追踪（17TRACK API）
- ❌ 批量发货

##### 营销系统
- ❌ 积分系统
- ❌ 优惠券管理
- ❌ 礼品卡系统
- ❌ 推荐奖励
- ❌ 每日签到

##### 会员系统
- ❌ 会员等级
- ❌ 会员资料
- ❌ 积分交易记录

##### 系统功能
- ❌ 审计日志
- ❌ SKU批量导入
- ❌ URL管理
- ❌ SEO管理

#### Go后端对应实现
- ⚠️ Product 模型（基础版已有）
- ⚠️ Cart 模型（基础版已有）
- ❌ Order 模型（**缺失**）
- ❌ Payment 模型（**缺失**）
- ❌ Shipping 模型（**缺失**）
- ❌ Coupon 模型（**缺失**）
- ❌ GiftCard 模型（**缺失**）
- ❌ Loyalty 模型（**缺失**）
- ❌ Review 模型（**缺失**）

#### 需要新增的模型和功能

```go
// 1. 订单系统
type Order struct {
    ID              uint
    OrderNumber     string
    UserID          uint
    Status          string // pending, paid, shipped, completed, cancelled
    PaymentMethod   string
    PaymentStatus   string
    ShippingMethod  string
    ShippingStatus  string
    TotalAmount     float64
    ShippingFee     float64
    TaxAmount       float64
    DiscountAmount  float64
    Items           []OrderItem
    ShippingAddress Address
    BillingAddress  Address
    CreatedAt       time.Time
    PaidAt          *time.Time
    ShippedAt       *time.Time
    CompletedAt     *time.Time
}

type OrderItem struct {
    ID        uint
    OrderID   uint
    ProductID uint
    SKU       string
    Quantity  int
    Price     float64
    Subtotal  float64
}

// 2. 支付系统
type PaymentMethod struct {
    ID          uint
    Name        string
    Code        string
    Icon        string
    FeeType     string // fixed, percentage
    FeeValue    float64
    Enabled     bool
    SortOrder   int
}

type TaxRate struct {
    ID          uint
    Country     string
    State       string
    Rate        float64
    Name        string
    Enabled     bool
}

// 3. 物流系统
type ShippingTemplate struct {
    ID          uint
    Name        string
    Type        string // weight, quantity, price
    FreeShipping bool
    FreeThreshold float64
    Rules       []ShippingRule
}

type ShippingRule struct {
    ID         uint
    TemplateID uint
    MinValue   float64
    MaxValue   float64
    Fee        float64
    Region     string
}

type Carrier struct {
    ID          uint
    Name        string
    Code        string
    TrackingURL string
    APIKey      string
    Enabled     bool
}

// 4. 营销系统
type Coupon struct {
    ID          uint
    Code        string
    Type        string // fixed, percentage
    Value       float64
    MinAmount   float64
    MaxDiscount float64
    StartDate   time.Time
    EndDate     time.Time
    UsageLimit  int
    UsedCount   int
    Enabled     bool
}

type GiftCard struct {
    ID          uint
    Code        string
    Balance     float64
    InitialValue float64
    Status      string
    ExpiresAt   *time.Time
}

type LoyaltyTransaction struct {
    ID          uint
    UserID      uint
    Type        string // earn, spend, expire
    Points      int
    Balance     int
    Source      string // order, checkin, referral
    SourceID    uint
    Description string
    CreatedAt   time.Time
}

// 5. 评价系统
type Review struct {
    ID          uint
    ProductID   uint
    UserID      uint
    Rating      int
    Title       string
    Content     string
    Images      []string
    Status      string // pending, approved, rejected
    Featured    bool
    CreatedAt   time.Time
}
```

---

### 3. tanzanite-customer-service（客服系统）

#### 功能范围
- ❌ 客服对话管理
- ❌ 工单系统
- ❌ 实时聊天
- ❌ 客服分配

#### Go后端需要实现
```go
type Ticket struct {
    ID          uint
    UserID      uint
    Subject     string
    Status      string // open, in_progress, resolved, closed
    Priority    string // low, medium, high, urgent
    Category    string
    AssignedTo  uint
    Messages    []TicketMessage
    CreatedAt   time.Time
    UpdatedAt   time.Time
    ResolvedAt  *time.Time
}

type TicketMessage struct {
    ID          uint
    TicketID    uint
    UserID      uint
    IsStaff     bool
    Content     string
    Attachments []string
    CreatedAt   time.Time
}
```

---

### 4. tanzanite-faq-content（FAQ管理）

#### 功能范围
- ⚠️ FAQ CRUD（基础已有）
- ⚠️ 分类管理（基础已有）
- ❌ 排序功能
- ❌ 浏览统计

#### Go后端对应实现
- ✅ FAQ 模型已创建
- ✅ FAQRepository 已实现
- ✅ FAQService 已实现
- ✅ FAQ API 已创建
- ⚠️ **需要增强**：排序、统计功能

---

### 5. tanzanite-photo-gallery（图片库）

#### 功能范围
- ❌ 图片上传
- ❌ 图片分类
- ❌ 图片标签
- ❌ 图片搜索
- ❌ 图片压缩

#### Go后端需要实现
```go
type Gallery struct {
    ID          uint
    Name        string
    Description string
    CoverImage  string
    Images      []GalleryImage
    CreatedAt   time.Time
}

type GalleryImage struct {
    ID          uint
    GalleryID   uint
    URL         string
    Thumbnail   string
    Title       string
    Description string
    Tags        []string
    Order       int
    CreatedAt   time.Time
}
```

---

### 6. tanzanite-product-registry（产品注册）

#### 功能范围
- ❌ 产品注册表单
- ❌ 注册记录管理
- ❌ 保修信息
- ❌ 序列号验证

#### Go后端需要实现
```go
type ProductRegistration struct {
    ID              uint
    UserID          uint
    ProductID       uint
    SerialNumber    string
    PurchaseDate    time.Time
    PurchaseProof   string
    WarrantyExpires time.Time
    Status          string
    CreatedAt       time.Time
}
```

---

### 7. tanzanite-subscription（订阅系统）

#### 功能范围
- ⚠️ 邮箱订阅（基础已有）
- ❌ 订阅确认
- ❌ 退订管理
- ❌ 邮件通知
- ❌ 群发功能

#### Go后端对应实现
- ✅ Subscription 模型已创建
- ❌ 邮件发送服务（**缺失**）
- ❌ 订阅确认流程（**缺失**）
- ❌ 通知触发器（**缺失**）

#### 需要补充
```go
type EmailTemplate struct {
    ID          uint
    Name        string
    Subject     string
    Body        string
    Type        string // subscription_confirm, new_post, new_product
}

type EmailLog struct {
    ID          uint
    To          string
    Subject     string
    Status      string // sent, failed, pending
    Error       string
    SentAt      time.Time
}

// 邮件服务
type EmailService struct {
    smtpHost string
    smtpPort int
    username string
    password string
}

func (s *EmailService) SendSubscriptionConfirm(email, token string) error
func (s *EmailService) SendNewPostNotification(subscribers []string, post *Post) error
func (s *EmailService) SendNewProductNotification(subscribers []string, product *Product) error
```

---

## 🎯 迁移优先级和时间估算

### 阶段一：核心商城功能（4-6周）⭐ **最高优先级**

#### Week 1-2: 订单系统
- [ ] Order 模型和数据库表
- [ ] OrderRepository 和 OrderService
- [ ] 订单 CRUD API
- [ ] 订单状态流转
- [ ] 订单列表和详情

#### Week 3-4: 支付和物流
- [ ] PaymentMethod 模型和管理
- [ ] TaxRate 模型和计算
- [ ] ShippingTemplate 模型
- [ ] Carrier 模型和追踪
- [ ] 支付和物流 API

#### Week 5-6: 营销系统
- [ ] Coupon 模型和验证
- [ ] GiftCard 模型和使用
- [ ] LoyaltyTransaction 模型
- [ ] 积分获取和消费
- [ ] 营销 API

### 阶段二：内容和评价（2-3周）

#### Week 7-8: 评价和博客增强
- [ ] Review 模型和管理
- [ ] 评价 CRUD API
- [ ] 博客翻译组完善
- [ ] 评价审核流程

#### Week 9: FAQ和图片库
- [ ] FAQ 排序和统计
- [ ] Gallery 模型
- [ ] 图片上传和管理
- [ ] 图片 API

### 阶段三：客服和订阅（2-3周）

#### Week 10-11: 客服系统
- [ ] Ticket 模型
- [ ] TicketMessage 模型
- [ ] 工单 CRUD API
- [ ] 客服分配逻辑

#### Week 12: 订阅和通知
- [ ] 邮件服务实现
- [ ] 订阅确认流程
- [ ] 通知触发器
- [ ] 群发功能

### 阶段四：产品注册（1周）

#### Week 13: 产品注册
- [ ] ProductRegistration 模型
- [ ] 注册 API
- [ ] 保修管理

---

## 📝 迁移策略建议

### 1. 渐进式迁移
- 保持 WordPress 插件运行
- 逐个功能迁移到 Go
- 通过 Nginx 路由分流
- 验证后切换流量

### 2. 数据迁移
- 为每个插件编写数据导出脚本
- 创建 Go 导入工具
- 验证数据完整性
- 保留 WordPress 备份

### 3. API 兼容性
- 保持相同的 REST 端点
- 相同的请求/响应格式
- 相同的错误码
- 前端无需修改

### 4. 测试策略
- 单元测试每个新功能
- 集成测试 API 端点
- 性能测试关键路径
- 用户验收测试

---

## 🔧 立即可做的事情

### 1. 补充现有模型
```bash
# 在 go-backend/internal/domain/ 下创建新模型
- order/
- payment/
- shipping/
- coupon/
- loyalty/
- review/
- ticket/
- gallery/
- registration/
```

### 2. 扩展 Repository 层
```bash
# 为新模型创建 Repository
- order_repository.go
- payment_repository.go
- shipping_repository.go
- coupon_repository.go
- loyalty_repository.go
- review_repository.go
```

### 3. 实现 Service 层
```bash
# 为新模型创建 Service
- order_service.go
- payment_service.go
- shipping_service.go
- coupon_service.go
- loyalty_service.go
- review_service.go
```

### 4. 创建 API 处理器
```bash
# 在 internal/api/v1/ 下创建新处理器
- order/
- payment/
- shipping/
- marketing/
- review/
- customer-service/
```

---

## 📊 完成度统计

### 当前完成度
```
基础架构:        100% ✅
用户系统:        100% ✅
文章系统:         80% ⚠️
产品系统:         60% ⚠️
购物车系统:       70% ⚠️
FAQ系统:          80% ⚠️
订阅系统:         40% ⚠️

订单系统:          0% ❌
支付系统:          0% ❌
物流系统:          0% ❌
营销系统:          0% ❌
评价系统:          0% ❌
客服系统:          0% ❌
图片库:            0% ❌
产品注册:          0% ❌

总体完成度:       ~35%
```

### 需要完成的工作量
```
已完成:    ~3,500 行代码
待完成:    ~6,500 行代码
总计:      ~10,000 行代码
```

---

## 🎯 建议的下一步行动

### 立即开始（本周）
1. **创建订单系统模型和表结构**
2. **实现订单 CRUD 基础功能**
3. **创建支付方式管理**

### 短期目标（2周内）
1. **完成订单系统核心功能**
2. **实现支付和税率管理**
3. **创建物流模板系统**

### 中期目标（1个月内）
1. **完成营销系统（优惠券、积分）**
2. **实现评价系统**
3. **完善博客翻译功能**

### 长期目标（3个月内）
1. **完成所有7个插件的迁移**
2. **全面测试和优化**
3. **生产环境部署**

---

**文档版本**: v1.0  
**创建日期**: 2026-05-25  
**状态**: 需要继续开发

**关键结论**: 
- ✅ 基础架构已完成
- ⚠️ 核心商城功能需要大量开发
- ❌ 7个WordPress插件功能大部分未实现
- 📊 总体完成度约35%，还需要约6,500行代码
