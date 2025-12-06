# 产品编码与保修查询系统

> 创建时间：2024-12-07  
> 状态：� 实施中

---

## 一、系统概述

### 1.1 目标

建立产品编码管理系统，实现：
1. **后台管理**：录入产品编码、客户信息、订单关联
2. **用户查询**：登录用户可查询产品保修状态
3. **记录追踪**：维修、延保、换货等记录管理

### 1.2 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     前端 (Nuxt + i18n)                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  /support/warranty-check                             │   │
│  │  - 需要用户登录                                      │   │
│  │  - 多语言支持 (i18n)                                 │   │
│  │  - 输入编码 → 显示保修状态                           │   │
│  └─────────────────────────────────────────────────────┘   │
│                            ↓ API 请求（需认证）             │
└─────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────┐
│                   后端 (WordPress 插件)                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  tanzanite-product-registry                          │   │
│  │  - 后台界面（中文）                                  │   │
│  │  - REST API（需认证）                                │   │
│  │  - 数据库管理                                        │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 二、数据模型

### 2.1 产品类型表 (tz_product_types)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INT AUTO_INCREMENT | 主键 |
| type_code | VARCHAR(50) | 类型代码（唯一） |
| type_name | VARCHAR(100) | 显示名称（中文） |
| type_name_en | VARCHAR(100) | 显示名称（英文，用于前端） |
| default_warranty_months | INT | 默认保修月数 |
| sort_order | INT | 排序 |
| is_active | TINYINT(1) | 是否启用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**初始数据：**

| type_code | type_name | type_name_en | default_warranty_months |
|-----------|-----------|--------------|------------------------|
| hub | 花鼓 | Hub | 36 |
| rim | 轮圈 | Rim | 36 |
| wheelset | 轮组 | Wheelset | 36 |
| spoke | 辐条 | Spoke | 36 |
| other | 其他 | Other | 36 |

### 2.2 产品登记表 (tz_product_registry)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INT AUTO_INCREMENT | 主键 |
| product_code | VARCHAR(100) | 产品编码（唯一索引） |
| product_type_id | INT | 关联产品类型 |
| product_name | VARCHAR(200) | 产品名称/型号 |
| ship_date | DATE | 出货日期（精确到月，存为当月1日） |
| warranty_months | INT | 保修月数（可覆盖类型默认值） |
| order_id | VARCHAR(100) | 订单号 |
| customer_name | VARCHAR(100) | 客户姓名 |
| customer_email | VARCHAR(100) | 客户邮箱 |
| customer_phone | VARCHAR(50) | 客户电话 |
| notes | TEXT | 备注 |
| created_by | INT | 创建人（WP用户ID） |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**索引：**
- UNIQUE INDEX on `product_code`
- INDEX on `order_id`
- INDEX on `customer_email`
- INDEX on `ship_date`

### 2.3 保修记录表 (tz_warranty_records)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INT AUTO_INCREMENT | 主键 |
| product_id | INT | 关联产品ID |
| record_type | ENUM('repair', 'extend', 'replace') | 记录类型 |
| record_date | DATE | 记录日期 |
| description | TEXT | 描述 |
| extend_months | INT | 延保月数（仅 extend 类型使用） |
| operator | VARCHAR(100) | 操作人 |
| created_at | DATETIME | 创建时间 |

**记录类型说明：**

| record_type | 中文 | 英文 | 说明 |
|-------------|------|------|------|
| repair | 维修 | Repair | 产品维修记录 |
| extend | 延保 | Warranty Extension | 延长保修期 |
| replace | 换货 | Replacement | 产品更换 |

---

## 三、后台功能（WordPress 插件）

### 3.1 菜单结构

```
Tanzanite 产品管理
├── 产品登记
│   ├── 所有产品
│   ├── 添加产品
│   └── 批量导入
├── 产品类型
│   └── 类型管理
├── 保修记录
│   └── 所有记录
└── 系统设置
    └── 基本设置
```

### 3.2 产品登记功能

#### 添加产品表单

```
┌─────────────────────────────────────────────────────────┐
│  添加产品                                                │
├─────────────────────────────────────────────────────────┤
│  产品编码 *    [________________]                        │
│  产品类型 *    [花鼓 ▼]                                  │
│  产品名称      [________________]                        │
│  出货日期 *    [2024] 年 [12] 月                         │
│  保修期限      [36] 个月  □ 使用类型默认值               │
│  ─────────────────────────────────────────────────────  │
│  订单号        [________________]                        │
│  客户姓名      [________________]                        │
│  客户邮箱      [________________]                        │
│  客户电话      [________________]                        │
│  备注          [________________]                        │
│  ─────────────────────────────────────────────────────  │
│                              [取消]  [保存]              │
└─────────────────────────────────────────────────────────┘
```

#### 产品列表

```
┌─────────────────────────────────────────────────────────────────────────┐
│  所有产品                                    [搜索: ________] [筛选 ▼]  │
├─────────────────────────────────────────────────────────────────────────┤
│  编码        │ 类型  │ 名称      │ 出货日期  │ 保修至    │ 客户    │ 操作 │
│─────────────────────────────────────────────────────────────────────────│
│  ABC123     │ 花鼓  │ TZ-H01   │ 2024-12  │ 2027-12  │ 张三    │ 编辑 │
│  XYZ789     │ 轮组  │ TZ-W02   │ 2024-11  │ 2027-11  │ 李四    │ 编辑 │
│  ...        │ ...   │ ...      │ ...      │ ...      │ ...     │ ...  │
├─────────────────────────────────────────────────────────────────────────┤
│  共 156 条记录                              [< 1 2 3 4 5 >]  [导出 Excel] │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 批量导入

支持 Excel/CSV 格式：

| 产品编码 | 产品类型 | 产品名称 | 出货年月 | 保修月数 | 订单号 | 客户姓名 | 客户邮箱 | 客户电话 | 备注 |
|----------|----------|----------|----------|----------|--------|----------|----------|----------|------|
| ABC123 | hub | TZ-H01 | 2024-12 | 36 | ORD001 | 张三 | ... | ... | ... |

### 3.3 产品类型管理

```
┌─────────────────────────────────────────────────────────────────┐
│  产品类型管理                                      [添加类型]   │
├─────────────────────────────────────────────────────────────────┤
│  代码      │ 中文名称  │ 英文名称   │ 默认保修  │ 状态  │ 操作  │
│─────────────────────────────────────────────────────────────────│
│  hub      │ 花鼓     │ Hub       │ 36个月   │ 启用  │ 编辑  │
│  rim      │ 轮圈     │ Rim       │ 36个月   │ 启用  │ 编辑  │
│  wheelset │ 轮组     │ Wheelset  │ 36个月   │ 启用  │ 编辑  │
│  spoke    │ 辐条     │ Spoke     │ 36个月   │ 启用  │ 编辑  │
│  other    │ 其他     │ Other     │ 36个月   │ 启用  │ 编辑  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.4 保修记录管理

#### 添加记录（在产品详情页）

```
┌─────────────────────────────────────────────────────────┐
│  添加保修记录                                            │
├─────────────────────────────────────────────────────────┤
│  产品编码    ABC123 (TZ-H01 花鼓)                        │
│  记录类型 *  ○ 维修  ○ 延保  ○ 换货                      │
│  记录日期 *  [2024-12-07]                                │
│  延保月数    [__] 个月  (仅延保时填写)                    │
│  描述        [________________________________]          │
│                              [取消]  [保存]              │
└─────────────────────────────────────────────────────────┘
```

---

## 四、前端功能（Nuxt）

### 4.1 页面路由

```
/support/warranty-check  - 保修查询页面（需登录）
```

### 4.2 用户界面

#### 未登录状态

```
┌─────────────────────────────────────────────────────────┐
│           🔒 Warranty Check                             │
├─────────────────────────────────────────────────────────┤
│                                                         │
│     Please log in to check your product warranty.       │
│                                                         │
│                    [Log In]                             │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

#### 已登录 - 查询界面

```
┌─────────────────────────────────────────────────────────┐
│           🔍 Warranty Check                             │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Enter your product code:                               │
│  ┌─────────────────────────────┐ ┌──────────┐          │
│  │ ABC123                      │ │  Check   │          │
│  └─────────────────────────────┘ └──────────┘          │
│                                                         │
│  Your product code can be found on the product          │
│  packaging or warranty card.                            │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

#### 查询结果 - 保修有效

```
┌─────────────────────────────────────────────────────────┐
│           ✅ Warranty Valid                             │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Product Code:    ABC123                                │
│  Product Type:    Hub                                   │
│  Product Name:    TZ-H01 Pro Hub                        │
│  Ship Date:       December 2024                         │
│  Warranty Period: 36 months                             │
│  Warranty Until:  December 2027                         │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │  ⏱️ Remaining: 35 months 25 days                │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
│  Service Records:                                       │
│  └─ No records                                          │
│                                                         │
│                    [Check Another]                      │
└─────────────────────────────────────────────────────────┘
```

#### 查询结果 - 保修过期

```
┌─────────────────────────────────────────────────────────┐
│           ❌ Warranty Expired                           │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Product Code:    XYZ789                                │
│  Product Type:    Wheelset                              │
│  Warranty Until:  March 2024                            │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │  ⚠️ Expired 9 months ago                        │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
│  Need help? Contact our support team.                   │
│                    [Contact Support]                    │
└─────────────────────────────────────────────────────────┘
```

#### 查询结果 - 未找到

```
┌─────────────────────────────────────────────────────────┐
│           ❓ Product Not Found                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  We couldn't find a product with code: INVALID123       │
│                                                         │
│  Please check:                                          │
│  • The code is entered correctly                        │
│  • The code is from a Tanzanite product                 │
│                                                         │
│  If you believe this is an error, please contact us.    │
│                    [Contact Support]                    │
└─────────────────────────────────────────────────────────┘
```

### 4.3 i18n 翻译键

```yaml
warranty:
  title: "Warranty Check"
  login_required: "Please log in to check your product warranty."
  login_button: "Log In"
  input_label: "Enter your product code:"
  input_placeholder: "e.g. ABC123"
  check_button: "Check"
  help_text: "Your product code can be found on the product packaging or warranty card."
  
  result:
    valid: "Warranty Valid"
    expired: "Warranty Expired"
    not_found: "Product Not Found"
    
  fields:
    product_code: "Product Code"
    product_type: "Product Type"
    product_name: "Product Name"
    ship_date: "Ship Date"
    warranty_period: "Warranty Period"
    warranty_until: "Warranty Until"
    remaining: "Remaining"
    expired_ago: "Expired {time} ago"
    
  records:
    title: "Service Records"
    no_records: "No records"
    repair: "Repair"
    extend: "Warranty Extension"
    replace: "Replacement"
    
  actions:
    check_another: "Check Another"
    contact_support: "Contact Support"
    
  errors:
    not_found_message: "We couldn't find a product with code: {code}"
    check_tips:
      - "The code is entered correctly"
      - "The code is from a Tanzanite product"
    error_contact: "If you believe this is an error, please contact us."
```

---

## 五、API 设计

### 5.1 认证方式

**复用现有 tanzanite-setting 插件的会员系统：**

- 继承 `Tanzanite_REST_Controller` 基类
- 使用 `is_user_logged_in` 作为 `permission_callback`（与现有 Members API 一致）
- API 命名空间：`tanzanite/v1`（与现有 API 保持一致）
- 前端通过 Cookie + X-WP-Nonce 认证（与现有系统一致）
- 未登录用户返回 401

### 5.2 API 端点

#### 查询保修状态

```
GET /wp-json/tanzanite/v1/warranty/{product_code}

Headers:
  Authorization: Bearer {token}  或 Cookie 认证

Response 200:
{
  "success": true,
  "data": {
    "product_code": "ABC123",
    "product_type": {
      "code": "hub",
      "name": "Hub",
      "name_zh": "花鼓"
    },
    "product_name": "TZ-H01 Pro Hub",
    "ship_date": "2024-12",
    "warranty_months": 36,
    "warranty_end": "2027-12",
    "status": "valid",  // valid | expired
    "remaining": {
      "months": 35,
      "days": 25,
      "total_days": 1120
    },
    "records": [
      {
        "type": "repair",
        "type_name": "Repair",
        "date": "2025-03-15",
        "description": "Bearing replacement"
      }
    ]
  }
}

Response 401:
{
  "success": false,
  "error": "unauthorized",
  "message": "Please log in to access this feature."
}

Response 404:
{
  "success": false,
  "error": "not_found",
  "message": "Product not found."
}
```

### 5.3 安全措施

1. **认证要求**：必须登录才能查询（复用现有会员系统）
2. **限流**：每用户每分钟最多 5 次查询（防止暴力枚举）
3. **日志**：记录所有查询请求（可复用现有审计日志系统）
4. **敏感信息**：API 不返回客户联系方式（仅返回产品和保修信息）

---

## 六、实施计划

### 第一阶段：后端基础（WordPress 插件）✅ 已完成

- [x] 创建插件基础结构
- [x] 创建数据库表（3张表）
- [x] 产品类型管理（CRUD）
- [x] 产品登记管理（CRUD）
- [x] 保修记录管理（CRUD）
- [x] REST API 接口（含限流）

### 第二阶段：后端扩展 ✅ 已完成

- [x] 批量导入功能（CSV 格式，支持中英文表头）
- [x] 数据导出功能（CSV 格式，含 BOM 支持 Excel）
- [x] 下载导入模板功能
- [x] 拖拽上传界面
- [x] 导入结果反馈（成功/失败统计）

### 第三阶段：前端页面 ✅ 已完成

- [x] 创建 `/support/warranty-check` 页面
- [x] 登录状态检测
- [x] 查询表单和结果展示
- [x] i18n 多语言支持（en.json, zh_cn.json）
- [x] 响应式设计

### 第四阶段：测试和优化

- [ ] 功能测试
- [ ] 安全测试
- [ ] 性能优化
- [ ] 文档完善

---

## 七、文件清单

### WordPress 插件

```
wp-plugin/tanzanite-product-registry/
├── tanzanite-product-registry.php    # 插件主文件
├── includes/
│   ├── class-database.php            # 数据库操作
│   ├── class-product-types.php       # 产品类型管理
│   ├── class-product-registry.php    # 产品登记管理
│   ├── class-warranty-records.php    # 保修记录管理
│   ├── class-api.php                 # REST API
│   └── class-import-export.php       # 导入导出
├── admin/
│   ├── class-admin-menu.php          # 后台菜单
│   ├── views/                        # 后台页面模板
│   └── assets/                       # 后台资源文件
└── languages/                        # 翻译文件（如需要）
```

### Nuxt 前端

```
app/
├── pages/
│   └── support/
│       └── warranty-check.vue        # 保修查询页面
├── composables/
│   └── useWarrantyCheck.ts           # 查询逻辑
└── locales/
    ├── en.json                       # 英文翻译（添加 warranty 部分）
    └── zh.json                       # 中文翻译（添加 warranty 部分）
```

---

## 八、注意事项

### 8.1 编码格式

- 现有编码格式保持不变（字母+数字组合）
- 编码仅作为唯一标识，不包含业务逻辑
- 所有业务信息（日期、类型、保修期）存储在数据库中

### 8.2 保修计算逻辑

```
保修截止日期 = 出货日期 + 保修月数 + 延保月数之和

示例：
- 出货日期：2024-12-01
- 保修月数：36
- 延保记录：+6个月
- 保修截止：2028-06-01
```

### 8.3 数据安全

**后台权限：**
- 插件内不做额外权限检查（你登录即代表最高权限）
- 后台页面仅在 WordPress 管理后台显示
- 安全性依赖 WordPress 本身的登录保护

**建议的 WordPress 账户安全措施：**
- 使用强密码
- 启用两步验证（2FA）插件，如 Wordfence、Google Authenticator
- 限制登录尝试次数
- 定期更换密码
- 使用 HTTPS

**API 安全：**
- 用户查询 API 需要登录（复用现有会员系统）
- 不向用户暴露其他客户信息
- 限流防止暴力枚举

**数据备份：**
- 定期备份数据库（建议使用 UpdraftPlus 等插件）

### 8.4 扩展性

- 产品类型可动态添加
- 保修记录类型可扩展（修改 ENUM）
- API 可扩展更多端点

---

*文档创建 - 待确认后开始实施*
