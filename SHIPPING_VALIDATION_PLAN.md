# 配送地区验证功能实现计划

## 实施进度

| 阶段 | 状态 | 完成时间 | 备注 |
|------|------|----------|------|
| 阶段 0：准备工作 | ✅ 完成 | 2024-12-07 | 分析现有系统结构 |
| 阶段 1：运费模板扩展 | ✅ 完成 | 2024-12-07 | 添加 zip_ranges + eta_days + service 字段 |
| 阶段 2：包装规则系统 | ✅ 完成 | 2024-12-07 | 数据库 + API + 后台 |
| 阶段 3：前端基础设施 | ✅ 完成 | 2024-12-07 | countries.ts + composables |
| 阶段 4：结账流程改造 | ✅ 完成 | 2024-12-07 | 国家选择 + 验证 |
| **里程碑 1** | ✅ 达成 | 2024-12-07 | 核心功能完成，可上线 |
| 阶段 5：运费计算集成 | ✅ 完成 | 2024-12-07 | 包装计算 + 地区匹配 |
| 阶段 6：运费透明度 | ✅ 完成 | 2024-12-07 | 明细展示 |
| **里程碑 2** | ✅ 达成 | 2024-12-07 | 运费计算完善 |
| 阶段 7-10：体验优化 | ✅ 完成 | 2024-12-07 | 购物车UX + 替代方案 |
| **里程碑 3** | ✅ 达成 | 2024-12-07 | 全部完成 |

### 阶段 1 完成详情

**改动文件**：
- `wp-plugin/tanzanite-setting/includes/rest-api/class-rest-shippingtemplates-controller.php`
  - `sanitize_shipping_rules()` 添加字段：`zip_ranges`, `eta_min_days`, `eta_max_days`, `service`, `service_label`
  - `regions` 字段自动转大写
- `wp-plugin/tanzanite-setting/includes/admin/class-shipping-admin.php`
  - 后台 UI 添加「邮编范围」输入框
- `wp-plugin/tanzanite-setting/assets/js/shipping-templates.js`
  - JS 支持邮编范围的读取、编辑、保存、显示

**验证方法**：
```bash
# 调用 API 验证返回数据包含新字段
curl /wp-json/tanzanite/v1/shipping-templates
```

---

### 阶段 2 实施计划：包装规则系统

> **目标**：创建独立的包装规则管理功能，与运费模板代码完全分离

#### 2.1 新建文件清单

| 文件路径 | 说明 | 预计大小 | 状态 |
|----------|------|----------|------|
| `includes/rest-api/class-rest-packaging-controller.php` | 包装规则 REST API | ~15KB | ✅ 已创建 |
| `includes/admin/class-packaging-admin.php` | 包装规则后台页面 | ~6KB | ✅ 已创建 |
| `assets/js/packaging-rules.js` | 包装规则后台 JS | ~10KB | ✅ 已创建 |

#### 2.2 需要修改的文件

| 文件路径 | 改动内容 | 改动量 | 状态 |
|----------|----------|--------|------|
| `includes/class-plugin.php` | 注册新菜单项 + 加载新控制器 | ~20 行 | ✅ 已修改 |
| `tanzanite-setting.php` | 插件激活时创建数据库表 | - | ✅ (通过 API 安装) |

#### 2.3 数据库表

**表 1：`wp_tanz_packaging_rules`**
```sql
CREATE TABLE `wp_tanz_packaging_rules` (
  `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `rule_name` VARCHAR(100) NOT NULL,
  `description` TEXT,
  `box_weight` DECIMAL(10,3) NOT NULL DEFAULT 0,
  `box_length` DECIMAL(10,2) DEFAULT NULL,
  `box_width` DECIMAL(10,2) DEFAULT NULL,
  `box_height` DECIMAL(10,2) DEFAULT NULL,
  `max_items` INT DEFAULT NULL,
  `max_weight` DECIMAL(10,3) DEFAULT NULL,
  `priority` INT DEFAULT 0,
  `is_active` TINYINT(1) DEFAULT 1,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**表 2：`wp_tanz_packaging_rule_applies`**
```sql
CREATE TABLE `wp_tanz_packaging_rule_applies` (
  `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `rule_id` INT UNSIGNED NOT NULL,
  `apply_type` ENUM('category', 'tag', 'product', 'all') NOT NULL,
  `apply_value` VARCHAR(100) DEFAULT NULL,
  FOREIGN KEY (`rule_id`) REFERENCES `wp_tanz_packaging_rules`(`id`) ON DELETE CASCADE
);
```

#### 2.4 实施步骤

| 步骤 | 任务 | 状态 |
|------|------|------|
| 2.4.1 | 在 `tanzanite-setting.php` 添加数据库表创建代码 | ✅ (通过 API 安装) |
| 2.4.2 | 创建 `class-rest-packaging-controller.php` (REST API) | ✅ |
| 2.4.3 | 创建 `class-packaging-admin.php` (后台页面) | ✅ |
| 2.4.4 | 创建 `packaging-rules.js` (后台 JS) | ✅ |
| 2.4.5 | 在 `class-plugin.php` 注册菜单和控制器 | ✅ |
| 2.4.6 | 测试验证 | ⏳ |

#### 2.5 文件结构预览

```
tanzanite-setting/
├── includes/
│   ├── admin/
│   │   ├── class-shipping-admin.php      (现有)
│   │   └── class-packaging-admin.php     (新建) ← 包装规则后台页面
│   └── rest-api/
│       ├── class-rest-shippingtemplates-controller.php (现有)
│       └── class-rest-packaging-controller.php         (新建) ← 包装规则 API
└── assets/
    └── js/
        ├── shipping-templates.js         (现有)
        └── packaging-rules.js            (新建) ← 包装规则后台 JS
```

#### 2.6 设计原则

1. **完全独立**：包装规则的代码与运费模板代码完全分离
2. **遵循现有模式**：参考 `class-shipping-admin.php` 和 `class-rest-shippingtemplates-controller.php` 的结构
3. **最小侵入**：只在 `class-plugin.php` 中添加菜单注册，不修改其他现有文件
4. **可独立禁用**：如果出问题，可以快速注释掉相关代码

---

## 背景

当前结账流程存在以下问题：
- 用户可以在任何国家/地区下单，即使该地区没有配送服务
- 结账表单缺少国家/地区选择字段
- 运费计算没有根据用户所在地区匹配配送规则
- 没有阻止不可配送地区的用户付款

## 目标

确保只有在后台 Shipping Templates 中设置了配送规则的地区才能进行商品下单和付款。

---

## 现有系统分析

### 后端（WordPress 插件）

**文件位置**：`wp-plugin/tanzanite-setting/includes/rest-api/class-rest-shippingtemplates-controller.php`

**配送模板数据结构（当前）**：
```json
{
  "id": 1,
  "template_name": "默认模板",
  "is_active": true,
  "rules": [
    {
      "type": "weight",
      "min": 0,
      "max": 10,
      "fee": 50,
      "free_over": 500,
      "regions": ["JP", "US", "DE"]  // ISO 国家代码，空数组表示全球配送
    }
  ],
  "meta": {
    "carrier": "sf_express",
    "currency": "USD"
  }
}
```

**配送模板数据结构（扩展后，支持邮编范围）**：
```json
{
  "id": 1,
  "template_name": "美国分区运费",
  "is_active": true,
  "rules": [
    {
      "type": "weight",
      "min": 0,
      "max": 10,
      "fee": 40,
      "regions": ["US"],
      "zip_ranges": ["10001-10999", "11001-11999"]  // 纽约地区
    },
    {
      "type": "weight",
      "min": 0,
      "max": 10,
      "fee": 80,
      "regions": ["US"],
      "zip_ranges": ["99501-99950"]  // 阿拉斯加（偏远地区）
    },
    {
      "type": "weight",
      "min": 0,
      "max": 10,
      "fee": 50,
      "regions": ["US"],
      "zip_ranges": []  // 空数组 = 美国其他地区（兜底规则）
    },
    {
      "type": "weight",
      "min": 0,
      "max": 10,
      "fee": 60,
      "regions": ["DE"],
      "zip_ranges": []  // 德国全境
    }
  ],
  "meta": {
    "carrier": "sf_express",
    "currency": "USD"
  }
}
```

**邮编范围匹配逻辑**：
1. 先匹配 `regions`（国家）— **regions 为空或不包含用户国家 → 不可配送**
2. 再匹配 `zip_ranges`（邮编范围）
3. `zip_ranges` 为空表示该国家的兜底规则（国家已匹配的前提下）
4. 优先匹配更具体的规则（有邮编范围的优先于兜底规则）
5. 邮编范围格式：`"起始邮编-结束邮编"`，支持纯数字和字母数字混合
6. **没有任何规则匹配 → 不可配送**（安全优先原则）

**API 端点**：
- `GET /wp-json/tanzanite/v1/shipping-templates` - 获取所有配送模板

### 前端（Nuxt）

**相关文件**：
- `app/components/CheckoutModal.vue` - 结账弹窗
- `app/composables/useCartCalculation.ts` - 购物车计算逻辑
- `app/composables/useCart.ts` - 购物车状态管理

**当前表单字段**：
- Recipient (name)
- Phone
- Address
- City
- Zip Code
- Payment Method
- Notes

**缺失**：Country（国家/地区）选择

---

## 实现步骤

### 阶段 1：基础设施

#### 步骤 1.1：创建国家列表数据
- [ ] 创建 `app/data/countries.ts`
- [ ] 包含常用国家的 ISO 代码和名称
- [ ] 按字母顺序排列，常用国家置顶

```ts
// app/data/countries.ts
export interface Country {
  code: string  // ISO 3166-1 alpha-2
  name: string
}

export const COUNTRIES: Country[] = [
  // 常用国家置顶
  { code: 'US', name: 'United States' },
  { code: 'GB', name: 'United Kingdom' },
  { code: 'DE', name: 'Germany' },
  { code: 'JP', name: 'Japan' },
  { code: 'CN', name: 'China' },
  // ... 其他国家按字母顺序
]
```

#### 步骤 1.2：创建配送验证 Composable
- [ ] 创建 `app/composables/useShippingValidation.ts`
- [ ] 封装配送可用性检查逻辑

```ts
// app/composables/useShippingValidation.ts
export const useShippingValidation = () => {
  // 获取所有可配送的国家代码
  const getAvailableCountries = (templates: ShippingTemplate[]): string[] => { ... }
  
  // 检查指定国家是否可配送
  const isCountryShippable = (countryCode: string, templates: ShippingTemplate[]): boolean => { ... }
  
  // 获取指定国家的配送规则
  const getShippingRulesForCountry = (countryCode: string, templates: ShippingTemplate[]): ShippingRule[] => { ... }
  
  return { getAvailableCountries, isCountryShippable, getShippingRulesForCountry }
}
```

---

### 阶段 2：结账表单改造

#### 步骤 2.1：添加国家选择字段
- [ ] 在 `CheckoutModal.vue` 的 Shipping Address 区域添加 Country 下拉框
- [ ] 放在 City 和 Zip Code 之前
- [ ] 使用搜索过滤功能（国家列表较长）

#### 步骤 2.2：更新表单数据结构
- [ ] 在 `form` ref 中添加 `country` 字段
- [ ] 更新 `isFormValid` 验证逻辑

```ts
const form = ref({
  name: '',
  phone: '',
  address: '',
  country: '',  // 新增
  city: '',
  zip: '',
  paymentMethod: 'credit_card',
  notes: '',
})
```

---

### 阶段 3：配送可用性验证

#### 步骤 3.1：在结账弹窗中集成验证
- [ ] 在 `CheckoutModal.vue` 中引入 `useShippingValidation`
- [ ] 添加 `isShippingAvailable` computed 属性
- [ ] 监听国家选择变化，实时验证

```ts
const isShippingAvailable = computed(() => {
  if (!form.value.country) return false
  return isCountryShippable(form.value.country, shippingTemplates.value)
})
```

#### 步骤 3.2：显示不可配送提示
- [ ] 当 `isShippingAvailable` 为 false 时，显示警告信息
- [ ] 提示文案：「We currently do not ship to [Country Name]. Please contact support for assistance.」
- [ ] 样式：红色/橙色警告框

#### 步骤 3.3：禁用付款按钮
- [ ] 更新 `isFormValid` 逻辑，加入 `isShippingAvailable` 条件
- [ ] 付款按钮 disabled 时显示原因

```ts
const isFormValid = computed(() => {
  return (
    form.value.name.trim() !== '' &&
    form.value.phone.trim() !== '' &&
    form.value.address.trim() !== '' &&
    form.value.country !== '' &&
    form.value.city.trim() !== '' &&
    form.value.paymentMethod !== '' &&
    isShippingAvailable.value  // 新增
  )
})
```

---

### 阶段 4：运费计算更新

#### 步骤 4.1：更新 calculateShipping 函数
- [ ] 在 `useCartCalculation.ts` 中修改 `calculateShipping`
- [ ] 添加 `countryCode` 参数
- [ ] 根据 `regions` 过滤匹配的规则

```ts
const calculateShipping = (
  items: Array<{ weight?: number; quantity: number; price: number }>,
  subtotal: number,
  countryCode: string  // 新增
): number => {
  // 1. 过滤出适用于该国家的规则
  // 2. 在匹配的规则中计算运费
  // 3. 如果没有匹配规则，返回 -1 表示不可配送
}
```

#### 步骤 4.2：更新 CheckoutModal 中的运费显示
- [ ] 传入用户选择的国家
- [ ] 处理不可配送的情况（运费显示为 N/A）

---

### 阶段 5：用户体验优化

#### 步骤 5.1：国家选择智能排序
- [ ] 根据后台配送模板中的 regions，将可配送国家置顶
- [ ] 不可配送国家显示为灰色或带标记

#### 步骤 5.2：购物车阶段预警
- [ ] 在购物车弹窗中也显示配送提示
- [ ] 「Shipping available to: US, UK, DE, JP...」

#### 步骤 5.3：地址自动填充
- [ ] 如果用户已登录且有保存的地址，自动填充
- [ ] 根据浏览器语言/IP 预选国家（可选）

---

## 文件改动清单

### 前端（Nuxt）

| 文件 | 改动类型 | 说明 |
|------|----------|------|
| `app/data/countries.ts` | 新建 | 国家列表数据 |
| `app/composables/useShippingValidation.ts` | 新建 | 配送验证逻辑（含邮编范围匹配） |
| `app/composables/usePackagingCalculation.ts` | 新建 | 包装规则计算逻辑 |
| `app/components/CheckoutModal.vue` | 修改 | 添加国家选择、验证逻辑、禁用付款 |
| `app/composables/useCartCalculation.ts` | 修改 | 更新运费计算，集成包装计算 |
| `app/composables/useCart.ts` | 修改 | 添加 country 到 shippingAddress |

### 后端（WordPress 插件）

| 文件 | 改动类型 | 说明 |
|------|----------|------|
| `wp-plugin/tanzanite-setting/includes/rest-api/class-rest-shippingtemplates-controller.php` | 修改 | 添加 `zip_ranges` 字段支持 |
| `wp-plugin/tanzanite-setting/includes/rest-api/class-rest-packaging-controller.php` | 新建 | 包装规则 REST API |
| `wp-plugin/tanzanite-setting/includes/admin/class-packaging-admin.php` | 新建 | 包装规则后台管理页面 |
| `wp-plugin/tanzanite-setting/assets/js/packaging-rules.js` | 新建 | 包装规则后台 JS |
| `wp-plugin/tanzanite-setting/assets/js/shipping-templates.js` | 修改 | 添加邮编范围输入框 |
| `wp-plugin/tanzanite-setting/tanzanite-setting.php` | 修改 | 注册新表和新页面 |

### 数据库

| 表名 | 改动类型 | 说明 |
|------|----------|------|
| `wp_tanz_packaging_rules` | 新建 | 包装规则表 |
| `wp_tanz_packaging_rule_applies` | 新建 | 包装规则适用范围表 |
| `wp_tanz_products` | 修改 | 添加 `packaging_rule_id` 和 `packaging_override` 字段 |

---

## 测试用例

### 正常流程
1. 用户选择可配送国家（如 US）
2. 运费正常计算并显示
3. 可以正常提交订单

### 不可配送流程
1. 用户选择不可配送国家（如 XX）
2. 显示「不可配送」警告
3. 付款按钮禁用
4. 无法提交订单

### 边界情况
1. 配送模板 regions 为空 → **不可配送**（必须明确配置地区才允许配送）
2. 没有激活的配送模板 → 所有国家不可配送
3. 用户未选择国家 → 付款按钮禁用
4. 用户所在国家没有匹配的运费规则 → 不可配送

---

## 实施计划（重新整理）

### 设计原则

1. **先后端后前端**：确保 API 稳定后再开发前端
2. **先核心后优化**：先实现阻止不可配送下单，再优化体验
3. **每阶段可独立测试**：每个阶段完成后都能验证功能
4. **渐进式上线**：可以分阶段部署，降低风险

---

### 阶段 0：准备工作（0.5 天）

> **目标**：确保现有系统稳定，准备测试数据

| 序号 | 任务 | 说明 | 验收标准 |
|------|------|------|----------|
| 0.1 | 备份现有数据库 | 防止改动出错 | 备份文件存在 |
| 0.2 | 检查现有运费模板数据 | 确认 `regions` 字段格式 | 数据格式正确 |
| 0.3 | 准备测试用例数据 | 各国家、各重量段的测试订单 | 测试数据就绪 |

**风险点**：无
**依赖**：无

---

### 阶段 1：后端 - 运费模板扩展（1 天）

> **目标**：运费模板支持邮编范围，为前端提供完整数据

| 序号 | 任务 | 文件 | 验收标准 |
|------|------|------|----------|
| 1.1 | 运费规则添加 `zip_ranges` 字段 | `class-rest-shippingtemplates-controller.php` | API 返回包含 zip_ranges |
| 1.2 | 运费规则添加 `estimated_days_min/max` 字段 | 同上 | API 返回包含时效字段 |
| 1.3 | 后台 UI 添加邮编范围输入 | `shipping-templates.js` | 后台可编辑邮编范围 |
| 1.4 | 后台 UI 添加时效输入 | 同上 | 后台可编辑预计天数 |

**测试方法**：
```bash
# 调用 API 验证返回数据
curl /wp-json/tanzanite/v1/shipping-templates
```

**风险点**：现有数据兼容性（zip_ranges 默认为空数组）
**依赖**：无

---

### 阶段 2：后端 - 包装规则系统（2 天）

> **目标**：完成包装规则的后端 CRUD 功能

| 序号 | 任务 | 文件 | 验收标准 |
|------|------|------|----------|
| 2.1 | 创建数据库表 | `tanzanite-setting.php` (activation hook) | 表创建成功 |
| 2.2 | 包装规则 REST API | `class-rest-packaging-controller.php` | CRUD 接口可用 |
| 2.3 | 后台管理页面 - 列表 | `class-packaging-admin.php` | 可查看规则列表 |
| 2.4 | 后台管理页面 - 编辑 | 同上 + `packaging-rules.js` | 可创建/编辑规则 |
| 2.5 | 配置默认包装规则 | 后台操作 | 至少有 1 个默认规则 |

**测试方法**：
```bash
# 创建包装规则
curl -X POST /wp-json/tanzanite/v1/packaging-rules -d '{...}'
# 获取包装规则
curl /wp-json/tanzanite/v1/packaging-rules
```

**风险点**：数据库表创建失败（需要检查权限）
**依赖**：阶段 1 完成

---

### 阶段 3：前端 - 基础设施（1 天）

> **目标**：创建前端所需的数据和工具函数，不改动现有 UI

| 序号 | 任务 | 文件 | 验收标准 |
|------|------|------|----------|
| 3.1 | 国家列表数据 | `app/data/countries.ts` | 导出 COUNTRIES 数组 |
| 3.2 | 邮编格式配置 | `app/data/zipFormats.ts` | 导出 ZIP_FORMAT_HINTS |
| 3.3 | 配送验证 composable | `app/composables/useShippingValidation.ts` | 函数可调用 |
| 3.4 | 包装计算 composable | `app/composables/usePackagingCalculation.ts` | 函数可调用 |

**测试方法**：
```ts
// 在控制台测试
const { isCountryShippable } = useShippingValidation()
console.log(isCountryShippable('US', templates)) // true/false
```

**风险点**：无（纯新增文件，不影响现有功能）
**依赖**：阶段 2 完成（需要 API 数据）

---

### 阶段 4：前端 - 结账流程核心改造（1.5 天）

> **目标**：实现核心功能 - 阻止不可配送地区下单

| 序号 | 任务 | 文件 | 验收标准 |
|------|------|------|----------|
| 4.1 | 添加国家选择下拉框 | `CheckoutModal.vue` | UI 显示正常 |
| 4.2 | 表单数据添加 country 字段 | 同上 | 数据绑定正确 |
| 4.3 | 集成配送可用性验证 | 同上 | 选择国家后实时验证 |
| 4.4 | 不可配送警告提示 | 同上 | 显示警告信息 |
| 4.5 | 禁用付款按钮 | 同上 | 不可配送时按钮禁用 |
| 4.6 | 邮编格式提示 | 同上 | 根据国家显示格式提示 |

**测试方法**：
1. 选择可配送国家 → 可以付款
2. 选择不可配送国家 → 按钮禁用 + 警告显示
3. 不选择国家 → 按钮禁用

**风险点**：UI 布局变化可能影响现有样式
**依赖**：阶段 3 完成

**⚠️ 里程碑 1：核心功能完成**
> 此时系统已经可以阻止不可配送地区下单，可以先部署上线

---

### 阶段 5：前端 - 运费计算集成（1 天）

> **目标**：运费计算考虑包装重量和地区匹配

| 序号 | 任务 | 文件 | 验收标准 |
|------|------|------|----------|
| 5.1 | 更新 calculateShipping 函数 | `useCartCalculation.ts` | 支持 countryCode 参数 |
| 5.2 | 集成包装计算 | 同上 | 运费包含包装重量 |
| 5.3 | 集成邮编范围匹配 | 同上 | 根据邮编匹配运费规则 |
| 5.4 | CheckoutModal 传入国家/邮编 | `CheckoutModal.vue` | 运费实时更新 |

**测试方法**：
1. 选择不同国家 → 运费变化
2. 输入不同邮编 → 运费变化（如美国不同州）
3. 添加多件商品 → 包装计算正确

**风险点**：运费计算逻辑变化可能导致金额不一致
**依赖**：阶段 4 完成

---

### 阶段 6：前端 - 运费透明度（0.5 天）

> **目标**：让用户理解运费是怎么算出来的

| 序号 | 任务 | 文件 | 验收标准 |
|------|------|------|----------|
| 6.1 | 运费明细展开面板 | `CheckoutModal.vue` | 可展开查看明细 |
| 6.2 | 显示包裹数量 | 同上 | 显示「X 个包裹」 |
| 6.3 | 显示总重量 | 同上 | 显示「总重 X.Xkg」 |
| 6.4 | 显示预计送达时间 | 同上 | 显示「预计 X-X 天」 |

**测试方法**：点击运费展开，查看明细是否正确

**风险点**：无
**依赖**：阶段 5 完成

**⚠️ 里程碑 2：运费计算完善**
> 此时运费计算已经完全准确，可以部署上线

---

### 阶段 7：用户体验优化 - 购物车（0.5 天）

> **目标**：在购物车阶段就让用户知道运费情况

| 序号 | 任务 | 文件 | 验收标准 |
|------|------|------|----------|
| 7.1 | 购物车添加国家选择 | `CartModal.vue` 或相关组件 | UI 显示正常 |
| 7.2 | 购物车显示运费预估 | 同上 | 显示预估运费 |
| 7.3 | 免运费进度条 | 同上 | 显示进度和差额 |
| 7.4 | 记住用户国家选择 | localStorage | 下次自动填充 |

**测试方法**：
1. 选择国家后运费预估显示
2. 关闭购物车再打开，国家保持选中

**风险点**：无
**依赖**：阶段 6 完成

---

### 阶段 8：用户体验优化 - 不可配送处理（0.5 天）

> **目标**：给不可配送用户提供出路

| 序号 | 任务 | 文件 | 验收标准 |
|------|------|------|----------|
| 8.1 | 不可配送时显示联系客服按钮 | `CheckoutModal.vue` | 按钮可点击 |
| 8.2 | 点击跳转聊天窗口 | 同上 | 打开聊天并预填话题 |
| 8.3 | 显示可配送国家列表 | 同上 | 列出可配送国家 |

**测试方法**：选择不可配送国家，点击联系客服

**风险点**：无
**依赖**：阶段 4 完成（可与阶段 5-7 并行）

---

### 阶段 9：用户体验优化 - 多地址管理（1 天）

> **目标**：让用户可以保存和快速选择地址

| 序号 | 任务 | 文件 | 验收标准 |
|------|------|------|----------|
| 9.1 | 地址数据结构定义 | `app/types/address.ts` | 类型定义完成 |
| 9.2 | 地址存储 composable | `app/composables/useAddresses.ts` | 可保存/读取地址 |
| 9.3 | 结账时地址选择 UI | `CheckoutModal.vue` | 可选择已保存地址 |
| 9.4 | 保存新地址功能 | 同上 | 可保存当前地址 |

**测试方法**：
1. 填写地址后保存
2. 下次结账时可选择已保存地址

**风险点**：无
**依赖**：阶段 4 完成

---

### 阶段 10：后端 - 商品包装规则关联（0.5 天）

> **目标**：允许单个商品指定包装规则

| 序号 | 任务 | 文件 | 验收标准 |
|------|------|------|----------|
| 10.1 | 商品表添加 packaging_rule_id 字段 | 数据库迁移 | 字段存在 |
| 10.2 | 商品 API 返回包装规则 | 商品 REST API | API 返回包含字段 |
| 10.3 | 商品编辑页添加包装规则选择 | 商品编辑页 | 可选择包装规则 |

**测试方法**：编辑商品，选择包装规则，保存后验证

**风险点**：商品数据结构变化
**依赖**：阶段 2 完成

---

## 实施时间线

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ 周一        周二        周三        周四        周五        周六        周日  │
├─────────────────────────────────────────────────────────────────────────────┤
│ [阶段0]     [阶段1]     [阶段2]     [阶段2]     [阶段3]     [阶段4]     [阶段4]│
│ 准备工作    运费扩展    包装规则    包装规则    前端基础    结账改造    结账改造│
│ 0.5天       1天         (续)        完成        1天         (续)        完成  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                            ⚠️ 里程碑1：可上线 │
├─────────────────────────────────────────────────────────────────────────────┤
│ [阶段5]     [阶段6]     [阶段7]     [阶段8]     [阶段9]     [阶段10]    缓冲   │
│ 运费集成    运费透明    购物车UX    不可配送    多地址      商品关联    测试   │
│ 1天         0.5天       0.5天       0.5天       1天         0.5天       修复   │
├─────────────────────────────────────────────────────────────────────────────┤
│                         ⚠️ 里程碑2：运费完善                 ⚠️ 里程碑3：全部完成│
└─────────────────────────────────────────────────────────────────────────────┘

预计总工期：10-12 天（含测试和修复时间）
```

---

## 里程碑检查点

### 里程碑 1：核心功能（阶段 0-4 完成后）
- ✅ 后台可配置运费模板（含邮编范围）
- ✅ 后台可配置包装规则
- ✅ 前端结账时必须选择国家
- ✅ 不可配送地区无法付款
- **可以上线**：阻止了不可配送地区下单的风险

### 里程碑 2：运费完善（阶段 5-6 完成后）
- ✅ 运费计算包含包装重量
- ✅ 运费根据国家/邮编匹配规则
- ✅ 用户可查看运费明细
- **可以上线**：运费计算完全准确

### 里程碑 3：体验优化（阶段 7-10 完成后）
- ✅ 购物车阶段可预估运费
- ✅ 免运费进度提示
- ✅ 不可配送时有替代方案
- ✅ 多地址管理
- ✅ 商品可指定包装规则
- **完整上线**：用户体验最佳

---

## 回滚方案

### 如果阶段 4 出问题（结账流程）
```ts
// 临时禁用国家验证
const isShippingAvailable = computed(() => true) // 临时返回 true
```

### 如果阶段 5 出问题（运费计算）
```ts
// 回退到原有运费计算逻辑
const calculateShipping = (items, subtotal) => {
  // 使用原有逻辑，不传 countryCode
}
```

### 如果后端 API 出问题
- 前端使用缓存数据
- 或显示「运费待确认」，允许下单后人工核算

---

## 风险与注意事项

1. **后台配送模板为空**：需要处理没有配送模板的情况，建议默认不可配送
2. **国家代码一致性**：确保前后端使用相同的 ISO 3166-1 alpha-2 代码
3. **缓存问题**：配送模板可能被缓存，后台修改后前端需要刷新
4. **多模板匹配**：如果多个模板都匹配同一国家，需要确定优先级规则

---

## 后续扩展

- 多物流渠道选择
- 根据 IP 自动检测国家

---

## 用户体验优化建议

### UX1. 运费预估时机优化

**问题**：用户需要到结账页面才能看到运费，可能加了一堆商品后发现运费很贵或不可配送。

**解决方案**：

| 位置 | 功能 | 优先级 |
|------|------|--------|
| 商品详情页 | 显示「预估运费」（基于默认国家或用户上次选择的国家） | P2 |
| 购物车弹窗 | 国家选择 + 实时运费预估 | P1 |
| 页脚/侧边栏 | 「运费计算器」小工具，用户可提前查询任意地区运费 | P3 |

**实现思路**：
```ts
// 记住用户上次选择的国家
const lastCountry = localStorage.getItem('tz_shipping_country') || 'US'

// 商品详情页显示预估运费
const estimatedShipping = computed(() => {
  return calculateShippingForCountry(lastCountry, productWeight)
})
```

---

### UX2. 不可配送时提供替代方案

**问题**：用户选择不可配送地区后只看到「不可配送」，没有出路。

**解决方案**：

```
┌─────────────────────────────────────────────────────────────────┐
│ ⚠️ Sorry, we currently don't ship to [Alaska]                   │
│                                                                 │
│ Here are some options:                                          │
│                                                                 │
│ 💬 [Contact Support] - Ask about special shipping arrangements  │
│                                                                 │
│ 📍 Ship to a different address?                                 │
│    We ship to the continental US for $50                        │
│                                                                 │
│ 🔔 [Notify me] when shipping to Alaska becomes available        │
└─────────────────────────────────────────────────────────────────┘
```

**功能点**：
- **联系客服按钮**：跳转到聊天窗口，预填「配送咨询」话题
- **显示最近可配送地区**：如用户在阿拉斯加，提示可以发到美国本土地址
- **到货通知**：当该地区开通配送时邮件通知用户（需要邮件订阅功能）

---

### UX3. 运费透明度 - 显示计算明细

**问题**：用户只看到最终运费数字，不知道怎么算出来的。

**解决方案**：在结账页面添加运费明细展开面板

```
┌─────────────────────────────────────────────────────────────────┐
│ Shipping                                              $45.00    │
│ ├─ 📦 Package 1: Carbon Wheelset                               │
│ │   └─ Product: 1.5kg + Box: 0.8kg = 2.3kg                     │
│ ├─ 📦 Package 2: Spokes & Accessories                          │
│ │   └─ Product: 0.5kg + Box: 0.3kg = 0.8kg                     │
│ ├─────────────────────────────────────────────────────────────  │
│ │ Total Weight: 3.1kg                                          │
│ │ Shipping Zone: US - New York (10001)                         │
│ │ Rate: $15/kg × 3.1kg = $46.50 → $45.00 (rounded)             │
│ └─ 🚚 Estimated Delivery: 7-15 business days                   │
└─────────────────────────────────────────────────────────────────┘
```

**数据结构扩展**：
```ts
interface ShippingBreakdown {
  packages: Array<{
    name: string
    productWeight: number
    boxWeight: number
    totalWeight: number
  }>
  totalWeight: number
  zone: string
  rate: string
  fee: number
  estimatedDays: string
}
```

---

### UX4. 邮编输入体验优化

**问题**：用户可能输错邮编格式，导致匹配失败。

**解决方案**：

#### 4.1 根据国家显示格式提示
```ts
const ZIP_FORMAT_HINTS: Record<string, { placeholder: string; pattern: string; hint: string }> = {
  US: { placeholder: '10001', pattern: '^\\d{5}(-\\d{4})?$', hint: '5 digits (e.g., 10001)' },
  GB: { placeholder: 'SW1A 1AA', pattern: '^[A-Z]{1,2}\\d[A-Z\\d]? ?\\d[A-Z]{2}$', hint: 'e.g., SW1A 1AA' },
  DE: { placeholder: '10115', pattern: '^\\d{5}$', hint: '5 digits (e.g., 10115)' },
  JP: { placeholder: '100-0001', pattern: '^\\d{3}-?\\d{4}$', hint: '7 digits (e.g., 100-0001)' },
  CA: { placeholder: 'K1A 0B1', pattern: '^[A-Z]\\d[A-Z] ?\\d[A-Z]\\d$', hint: 'e.g., K1A 0B1' },
  AU: { placeholder: '2000', pattern: '^\\d{4}$', hint: '4 digits (e.g., 2000)' },
  CN: { placeholder: '100000', pattern: '^\\d{6}$', hint: '6 digits (e.g., 100000)' },
}
```

#### 4.2 实时格式校验
```vue
<input
  v-model="form.zip"
  :placeholder="zipFormatHint.placeholder"
  :pattern="zipFormatHint.pattern"
  @blur="validateZipFormat"
/>
<span v-if="zipFormatError" class="text-red-500 text-sm">
  {{ zipFormatError }}
</span>
<span v-else class="text-gray-400 text-sm">
  {{ zipFormatHint.hint }}
</span>
```

---

### UX5. 免运费门槛提示

**问题**：用户不知道还差多少可以免运费。

**解决方案**：

#### 5.1 购物车显示免运费进度
```
┌─────────────────────────────────────────────────────────────────┐
│ 🚚 Free shipping on orders over $200                            │
│ ████████████████████░░░░░░░░░░  $156 / $200                     │
│ Add $44 more to get FREE SHIPPING!                              │
└─────────────────────────────────────────────────────────────────┘
```

#### 5.2 接近免运费时弹出提示
```ts
// 当用户准备结账时，如果差额小于 20%
const freeShippingGap = freeShippingThreshold - subtotal
if (freeShippingGap > 0 && freeShippingGap < freeShippingThreshold * 0.2) {
  showToast(`Add just $${freeShippingGap.toFixed(2)} more for FREE shipping!`)
}
```

#### 5.3 推荐凑单商品
```
┌─────────────────────────────────────────────────────────────────┐
│ 💡 Add one of these to get FREE shipping:                       │
│                                                                 │
│ [Inner Tube - $12]  [Tire Lever - $8]  [Patch Kit - $6]        │
└─────────────────────────────────────────────────────────────────┘
```

---

### UX6. 预计送达时间显示

**问题**：用户不知道什么时候能收到货。

**解决方案**：

#### 6.1 运费模板添加时效字段
```sql
-- 在 shipping rules 中添加
ALTER TABLE shipping_templates ADD COLUMN estimated_days_min INT DEFAULT NULL;
ALTER TABLE shipping_templates ADD COLUMN estimated_days_max INT DEFAULT NULL;
```

```json
{
  "type": "weight",
  "min": 0,
  "max": 10,
  "fee": 50,
  "regions": ["US"],
  "estimated_days_min": 7,
  "estimated_days_max": 15
}
```

#### 6.2 结账页面显示
```
Shipping: $45.00
📅 Estimated delivery: Dec 14 - Dec 22, 2025
```

#### 6.3 偏远地区额外提示
```
⚠️ Shipping to Alaska may take longer than usual (15-25 business days)
```

---

### UX7. 多地址管理

**问题**：用户可能想把商品寄到不同地址（如礼物），或有多个常用地址。

**解决方案**：

#### 7.1 保存多个收货地址
```ts
interface SavedAddress {
  id: string
  label: string        // "Home", "Office", "Mom's House"
  name: string
  phone: string
  country: string
  city: string
  address: string
  zip: string
  isDefault: boolean
}
```

#### 7.2 结账时快速选择
```
┌─────────────────────────────────────────────────────────────────┐
│ Shipping Address                                                │
│                                                                 │
│ ○ 🏠 Home (Default)                                             │
│   John Doe, 123 Main St, New York, NY 10001                    │
│                                                                 │
│ ○ 🏢 Office                                                     │
│   John Doe, 456 Business Ave, New York, NY 10002               │
│                                                                 │
│ ○ ➕ Add new address                                            │
└─────────────────────────────────────────────────────────────────┘
```

#### 7.3 数据存储
- **游客**：localStorage 存储（最多 3 个地址）
- **登录用户**：后端数据库存储（无限制）

---

### UX8. 货币一致性

**问题**：运费货币可能与商品价格货币不一致，造成用户困惑。

**解决方案**：

- 运费模板中的 `meta.currency` 必须与商品价格货币一致
- 如果支持多货币，运费也要自动转换
- 显示时统一使用用户选择的货币

```ts
// 确保运费和商品使用相同货币
const displayShipping = computed(() => {
  const fee = shippingFee.value
  const templateCurrency = shippingTemplate.value?.meta?.currency || 'USD'
  
  if (templateCurrency !== userCurrency.value) {
    return convertCurrency(fee, templateCurrency, userCurrency.value)
  }
  return fee
})
```

---

## 用户体验优化实现优先级

| 优先级 | 功能 | 说明 | 阶段 |
|--------|------|------|------|
| **P1** | 购物车运费预估 | 用户选择国家后立即显示预估运费 | 第三阶段 |
| **P1** | 免运费进度提示 | 显示「再买 $XX 免运费」+ 进度条 | 第三阶段 |
| **P2** | 运费明细展示 | 显示重量、包装、计算过程 | 第四阶段 |
| **P2** | 预计送达时间 | 在运费模板中添加时效字段 | 第四阶段 |
| **P2** | 邮编格式校验 | 根据国家校验邮编格式 + 提示 | 第三阶段 |
| **P2** | 不可配送替代方案 | 联系客服 / 最近可配送地区 / 到货通知 | 第五阶段 |
| **P3** | 商品详情页运费预估 | 显示预估运费到该商品 | 第五阶段 |
| **P3** | 多地址管理 | 保存多个收货地址 | 第五阶段 |
| **P3** | 运费计算器小工具 | 独立的运费查询功能 | 第五阶段 |
| **P3** | 凑单推荐 | 接近免运费时推荐商品 | 第五阶段 |

---

## 附录：包装规则系统设计（方案 B）

### B1. 问题背景

运费计算时需要考虑包装重量，但存在以下复杂场景：

| 场景 | 示例 | 问题 |
|------|------|------|
| 大件商品 | 轮组（1件=1包裹） | 简单，商品重量+包装重量 |
| 小件商品 | 辐条（50根=1包裹） | 多件共用包装，不能按件数乘包装重量 |
| 混合订单 | 轮组+辐条+内胎 | 部分商品可以塞进大件包装 |

### B2. 解决方案：包装规则表

创建独立的「包装规则」配置，与商品分类/标签关联，实现灵活的包装计算。

---

### B3. 数据结构设计

#### B3.1 包装规则表 `tanz_packaging_rules`

```sql
CREATE TABLE `wp_tanz_packaging_rules` (
  `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `rule_name` VARCHAR(100) NOT NULL COMMENT '规则名称',
  `description` TEXT COMMENT '规则描述',
  `box_weight` DECIMAL(10,3) NOT NULL DEFAULT 0 COMMENT '包装重量(kg)',
  `box_length` DECIMAL(10,2) DEFAULT NULL COMMENT '包装长度(cm)',
  `box_width` DECIMAL(10,2) DEFAULT NULL COMMENT '包装宽度(cm)',
  `box_height` DECIMAL(10,2) DEFAULT NULL COMMENT '包装高度(cm)',
  `max_items` INT DEFAULT NULL COMMENT '每包装最大件数',
  `max_weight` DECIMAL(10,3) DEFAULT NULL COMMENT '每包装最大承重(kg)',
  `max_volume` DECIMAL(10,6) DEFAULT NULL COMMENT '每包装最大体积(m³)',
  `priority` INT DEFAULT 0 COMMENT '优先级(数字越大优先级越高)',
  `is_active` TINYINT(1) DEFAULT 1,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX `idx_active_priority` (`is_active`, `priority` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### B3.2 包装规则适用范围表 `tanz_packaging_rule_applies`

```sql
CREATE TABLE `wp_tanz_packaging_rule_applies` (
  `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `rule_id` INT UNSIGNED NOT NULL,
  `apply_type` ENUM('category', 'tag', 'product', 'all') NOT NULL COMMENT '适用类型',
  `apply_value` VARCHAR(100) DEFAULT NULL COMMENT '适用值(分类ID/标签ID/商品ID)',
  FOREIGN KEY (`rule_id`) REFERENCES `wp_tanz_packaging_rules`(`id`) ON DELETE CASCADE,
  INDEX `idx_rule_type` (`rule_id`, `apply_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### B3.3 商品表扩展（可选覆盖）

```sql
-- 在商品表中添加字段，允许单个商品覆盖默认包装规则
ALTER TABLE `wp_tanz_products` ADD COLUMN `packaging_rule_id` INT UNSIGNED DEFAULT NULL COMMENT '指定包装规则ID';
ALTER TABLE `wp_tanz_products` ADD COLUMN `packaging_override` JSON DEFAULT NULL COMMENT '包装规则覆盖(JSON)';
```

---

### B4. 包装规则配置示例

#### B4.1 规则数据示例

```json
[
  {
    "id": 1,
    "rule_name": "轮组专用包装",
    "description": "碳纤维轮组独立包装，每个包装只能放1个轮组",
    "box_weight": 0.8,
    "box_length": 75,
    "box_width": 75,
    "box_height": 15,
    "max_items": 1,
    "max_weight": null,
    "priority": 100,
    "applies_to": [
      { "type": "category", "value": "wheelsets" },
      { "type": "category", "value": "carbon-wheels" }
    ]
  },
  {
    "id": 2,
    "rule_name": "小件通用包装",
    "description": "辐条、气嘴、螺丝等小件，可合并包装",
    "box_weight": 0.3,
    "box_length": 30,
    "box_width": 20,
    "box_height": 15,
    "max_items": 200,
    "max_weight": 5,
    "priority": 50,
    "applies_to": [
      { "type": "category", "value": "spokes" },
      { "type": "category", "value": "accessories" },
      { "type": "tag", "value": "small-parts" }
    ]
  },
  {
    "id": 3,
    "rule_name": "内胎/外胎包装",
    "description": "内胎外胎可以多个合并，但有体积限制",
    "box_weight": 0.4,
    "box_length": 40,
    "box_width": 40,
    "box_height": 20,
    "max_items": 10,
    "max_weight": 8,
    "priority": 60,
    "applies_to": [
      { "type": "category", "value": "tubes" },
      { "type": "category", "value": "tires" }
    ]
  },
  {
    "id": 4,
    "rule_name": "默认包装",
    "description": "未匹配其他规则的商品使用此规则",
    "box_weight": 0.5,
    "box_length": 50,
    "box_width": 40,
    "box_height": 30,
    "max_items": 5,
    "max_weight": 10,
    "priority": 0,
    "applies_to": [
      { "type": "all", "value": null }
    ]
  }
]
```

---

### B5. 包装计算逻辑

#### B5.1 计算流程

```
购物车商品列表
    ↓
1. 为每个商品匹配包装规则（按优先级）
    ↓
2. 按包装规则分组商品
    ↓
3. 计算每组需要多少个包装
    ↓
4. 累加所有包装的重量/体积
    ↓
5. 返回总发货重量/体积
```

#### B5.2 核心计算函数

```ts
// app/composables/usePackagingCalculation.ts

interface PackagingRule {
  id: number
  rule_name: string
  box_weight: number      // kg
  box_length?: number     // cm
  box_width?: number      // cm
  box_height?: number     // cm
  max_items?: number      // 每包装最大件数
  max_weight?: number     // 每包装最大承重 kg
  max_volume?: number     // 每包装最大体积 m³
  priority: number
  applies_to: Array<{ type: 'category' | 'tag' | 'product' | 'all'; value: string | null }>
}

interface CartItemWithPackaging {
  product_id: number
  quantity: number
  weight: number          // 单件商品重量 kg
  volume?: number         // 单件商品体积 m³
  category_ids: string[]
  tag_ids: string[]
  packaging_rule_id?: number  // 商品指定的包装规则
}

interface PackagingResult {
  total_weight: number           // 总发货重量
  total_volume: number           // 总发货体积
  package_count: number          // 包裹数量
  packages: Array<{
    rule_id: number
    rule_name: string
    items: Array<{ product_id: number; quantity: number }>
    weight: number
    volume: number
  }>
}

/**
 * 为商品匹配包装规则
 */
const matchPackagingRule = (
  item: CartItemWithPackaging,
  rules: PackagingRule[]
): PackagingRule | null => {
  // 1. 如果商品指定了包装规则，优先使用
  if (item.packaging_rule_id) {
    const specified = rules.find(r => r.id === item.packaging_rule_id)
    if (specified) return specified
  }

  // 2. 按优先级排序
  const sortedRules = [...rules].sort((a, b) => b.priority - a.priority)

  // 3. 查找匹配的规则
  for (const rule of sortedRules) {
    for (const apply of rule.applies_to) {
      switch (apply.type) {
        case 'product':
          if (String(item.product_id) === apply.value) return rule
          break
        case 'category':
          if (item.category_ids.includes(apply.value!)) return rule
          break
        case 'tag':
          if (item.tag_ids.includes(apply.value!)) return rule
          break
        case 'all':
          return rule
      }
    }
  }

  return null
}

/**
 * 计算一组商品需要多少个包装
 */
const calculatePackageCount = (
  items: Array<{ quantity: number; weight: number; volume?: number }>,
  rule: PackagingRule
): number => {
  const totalQty = items.reduce((sum, i) => sum + i.quantity, 0)
  const totalWeight = items.reduce((sum, i) => sum + i.weight * i.quantity, 0)
  const totalVolume = items.reduce((sum, i) => sum + (i.volume || 0) * i.quantity, 0)

  let packageCount = 1

  // 按件数限制
  if (rule.max_items) {
    packageCount = Math.max(packageCount, Math.ceil(totalQty / rule.max_items))
  }

  // 按重量限制
  if (rule.max_weight) {
    packageCount = Math.max(packageCount, Math.ceil(totalWeight / rule.max_weight))
  }

  // 按体积限制
  if (rule.max_volume && totalVolume > 0) {
    packageCount = Math.max(packageCount, Math.ceil(totalVolume / rule.max_volume))
  }

  return packageCount
}

/**
 * 计算购物车的包装结果
 */
const calculatePackaging = (
  items: CartItemWithPackaging[],
  rules: PackagingRule[]
): PackagingResult => {
  // 1. 为每个商品匹配规则并分组
  const groups = new Map<number, {
    rule: PackagingRule
    items: CartItemWithPackaging[]
  }>()

  for (const item of items) {
    const rule = matchPackagingRule(item, rules)
    if (!rule) continue

    if (!groups.has(rule.id)) {
      groups.set(rule.id, { rule, items: [] })
    }
    groups.get(rule.id)!.items.push(item)
  }

  // 2. 计算每组的包装
  const packages: PackagingResult['packages'] = []
  let totalWeight = 0
  let totalVolume = 0
  let packageCount = 0

  for (const [ruleId, group] of groups) {
    const { rule, items: groupItems } = group

    // 计算商品总重量
    const productWeight = groupItems.reduce(
      (sum, i) => sum + i.weight * i.quantity, 0
    )
    const productVolume = groupItems.reduce(
      (sum, i) => sum + (i.volume || 0) * i.quantity, 0
    )

    // 计算需要多少个包装
    const count = calculatePackageCount(groupItems, rule)

    // 计算包装重量
    const boxWeight = rule.box_weight * count
    const boxVolume = rule.box_length && rule.box_width && rule.box_height
      ? (rule.box_length * rule.box_width * rule.box_height / 1000000) * count  // cm³ → m³
      : 0

    // 累加
    totalWeight += productWeight + boxWeight
    totalVolume += productVolume + boxVolume
    packageCount += count

    packages.push({
      rule_id: ruleId,
      rule_name: rule.rule_name,
      items: groupItems.map(i => ({ product_id: i.product_id, quantity: i.quantity })),
      weight: productWeight + boxWeight,
      volume: productVolume + boxVolume,
    })
  }

  return {
    total_weight: Math.round(totalWeight * 1000) / 1000,  // 保留3位小数
    total_volume: Math.round(totalVolume * 1000000) / 1000000,
    package_count: packageCount,
    packages,
  }
}

export const usePackagingCalculation = () => {
  const packagingRules = ref<PackagingRule[]>([])

  const loadPackagingRules = async () => {
    try {
      const response = await $fetch<{ items: PackagingRule[] }>(
        '/wp-json/tanzanite/v1/packaging-rules'
      )
      packagingRules.value = response.items || []
    } catch (error) {
      console.error('Failed to load packaging rules:', error)
    }
  }

  return {
    packagingRules,
    loadPackagingRules,
    matchPackagingRule,
    calculatePackageCount,
    calculatePackaging,
  }
}
```

---

### B6. 后端 API 设计

#### B6.1 REST API 端点

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/wp-json/tanzanite/v1/packaging-rules` | 获取所有包装规则 |
| GET | `/wp-json/tanzanite/v1/packaging-rules/{id}` | 获取单个规则 |
| POST | `/wp-json/tanzanite/v1/packaging-rules` | 创建规则 |
| PUT | `/wp-json/tanzanite/v1/packaging-rules/{id}` | 更新规则 |
| DELETE | `/wp-json/tanzanite/v1/packaging-rules/{id}` | 删除规则 |
| POST | `/wp-json/tanzanite/v1/packaging/calculate` | 计算包装（可选，后端计算） |

#### B6.2 后端文件

| 文件 | 说明 |
|------|------|
| `includes/rest-api/class-rest-packaging-controller.php` | 包装规则 REST API |
| `includes/admin/class-packaging-admin.php` | 后台管理页面 |
| `assets/js/packaging-rules.js` | 后台 JS |

---

### B7. 后台管理界面

#### B7.1 包装规则列表页

```
┌─────────────────────────────────────────────────────────────────┐
│ Packaging Rules                                    [+ Add Rule] │
├─────────────────────────────────────────────────────────────────┤
│ Name              │ Box Weight │ Max Items │ Priority │ Status  │
├───────────────────┼────────────┼───────────┼──────────┼─────────┤
│ 轮组专用包装       │ 0.8 kg     │ 1         │ 100      │ ✓ Active│
│ 小件通用包装       │ 0.3 kg     │ 200       │ 50       │ ✓ Active│
│ 内胎/外胎包装      │ 0.4 kg     │ 10        │ 60       │ ✓ Active│
│ 默认包装          │ 0.5 kg     │ 5         │ 0        │ ✓ Active│
└─────────────────────────────────────────────────────────────────┘
```

#### B7.2 包装规则编辑页

```
┌─────────────────────────────────────────────────────────────────┐
│ Edit Packaging Rule                                             │
├─────────────────────────────────────────────────────────────────┤
│ Rule Name:        [轮组专用包装                              ]  │
│ Description:      [碳纤维轮组独立包装，每个包装只能放1个轮组  ]  │
│                                                                 │
│ ─── Box Dimensions ───                                          │
│ Weight (kg):      [0.8    ]                                     │
│ Length (cm):      [75     ]  Width: [75    ]  Height: [15    ]  │
│                                                                 │
│ ─── Limits ───                                                  │
│ Max Items:        [1      ]  (leave empty for no limit)         │
│ Max Weight (kg):  [       ]  (leave empty for no limit)         │
│                                                                 │
│ ─── Applies To ───                                              │
│ [x] Categories:   [wheelsets] [carbon-wheels] [+ Add]           │
│ [ ] Tags:         [+ Add]                                       │
│ [ ] Products:     [+ Add]                                       │
│ [ ] All Products                                                │
│                                                                 │
│ Priority:         [100    ]  (higher = matched first)           │
│ Status:           [x] Active                                    │
│                                                                 │
│                              [Cancel]  [Save Rule]              │
└─────────────────────────────────────────────────────────────────┘
```

---

### B8. 与运费计算的集成

#### B8.1 更新后的运费计算流程

```
购物车商品
    ↓
1. 调用 calculatePackaging() 获取包装结果
    ↓
2. 获取 total_weight（含包装重量）
    ↓
3. 根据 country + zip + total_weight 匹配运费规则
    ↓
4. 计算运费
```

#### B8.2 更新 useCartCalculation.ts

```ts
// 在 calculateShipping 中使用包装计算结果
const calculateShipping = (
  items: CartItemWithPackaging[],
  countryCode: string,
  zipCode: string,
  packagingRules: PackagingRule[],
  shippingTemplates: ShippingTemplate[]
): { fee: number; weight: number; packages: number } => {
  
  // 1. 计算包装
  const packaging = calculatePackaging(items, packagingRules)
  
  // 2. 查找匹配的运费规则
  const rule = findMatchingShippingRule(
    countryCode,
    zipCode,
    packaging.total_weight,
    shippingTemplates
  )
  
  if (!rule) {
    return { fee: -1, weight: packaging.total_weight, packages: packaging.package_count }
  }
  
  // 3. 计算运费
  const fee = rule.fee
  
  return {
    fee,
    weight: packaging.total_weight,
    packages: packaging.package_count,
  }
}
```

---

### B9. 实现优先级

| 优先级 | 任务 | 说明 |
|--------|------|------|
| P0 | 数据库表创建 | `tanz_packaging_rules` + `tanz_packaging_rule_applies` |
| P0 | REST API | CRUD 接口 |
| P0 | 后台管理页面 | 规则列表 + 编辑页 |
| P1 | 前端 composable | `usePackagingCalculation.ts` |
| P1 | 集成到运费计算 | 更新 `calculateShipping()` |
| P2 | 商品编辑页集成 | 允许单个商品指定/覆盖包装规则 |

---

### B10. 测试用例

#### 场景 1：单个大件商品
```
购物车：1 × 轮组 (1.5kg)
包装规则：轮组专用包装 (box_weight: 0.8kg, max_items: 1)
预期结果：
  - 包裹数：1
  - 总重量：1.5 + 0.8 = 2.3kg
```

#### 场景 2：多个小件商品
```
购物车：50 × 辐条 (0.01kg each)
包装规则：小件通用包装 (box_weight: 0.3kg, max_items: 200)
预期结果：
  - 包裹数：1
  - 总重量：0.5 + 0.3 = 0.8kg
```

#### 场景 3：小件超过包装限制
```
购物车：300 × 辐条 (0.01kg each)
包装规则：小件通用包装 (box_weight: 0.3kg, max_items: 200)
预期结果：
  - 包裹数：2 (ceil(300/200))
  - 总重量：3.0 + 0.6 = 3.6kg
```

#### 场景 4：混合订单
```
购物车：
  - 1 × 轮组 (1.5kg) → 轮组专用包装
  - 50 × 辐条 (0.01kg) → 小件通用包装
  - 2 × 内胎 (0.2kg) → 内胎/外胎包装
预期结果：
  - 包裹数：3
  - 总重量：(1.5+0.8) + (0.5+0.3) + (0.4+0.4) = 3.9kg
```

---

## 附录：邮编范围功能详细设计

### A1. 后端改动

#### A1.1 数据库表结构
当前 `rules` 字段已经是 JSON 格式存储，无需修改表结构，只需在 JSON 中添加 `zip_ranges` 字段。

#### A1.2 后端 sanitize 逻辑更新
**文件**：`wp-plugin/tanzanite-setting/includes/rest-api/class-rest-shippingtemplates-controller.php`

在 `sanitize_shipping_rules()` 函数中添加 `zip_ranges` 字段处理：

```php
// 在 sanitize_shipping_rules() 中添加
$zip_ranges = array();
if ( isset( $rule['zip_ranges'] ) && is_array( $rule['zip_ranges'] ) ) {
    foreach ( $rule['zip_ranges'] as $range ) {
        $zip_ranges[] = sanitize_text_field( $range );
    }
}

$sanitized[] = array_filter(
    array(
        'type'       => $type,
        'min'        => $min,
        'max'        => $max,
        'fee'        => $fee,
        'priority'   => $priority,
        'free_over'  => $free_over,
        'regions'    => $regions,
        'zip_ranges' => $zip_ranges,  // 新增
    ),
    // ...
);
```

#### A1.3 后台管理界面更新
**文件**：`wp-plugin/tanzanite-setting/assets/js/shipping-templates.js`

在规则编辑表单中添加邮编范围输入框：
- 输入格式：`10001-10999, 11001-11999`（逗号分隔多个范围）
- 留空表示该国家的兜底规则

---

### A2. 前端改动

#### A2.1 邮编范围匹配函数

```ts
// app/composables/useShippingValidation.ts

/**
 * 检查邮编是否在指定范围内
 * @param zipCode 用户输入的邮编
 * @param zipRanges 邮编范围数组，如 ["10001-10999", "11001-11999"]
 */
const isZipInRanges = (zipCode: string, zipRanges: string[]): boolean => {
  if (!zipRanges || zipRanges.length === 0) {
    return true // 空数组表示该国家的兜底规则（国家已匹配的前提下）
  }

  const normalizedZip = zipCode.replace(/\s/g, '').toUpperCase()

  for (const range of zipRanges) {
    if (range.includes('-')) {
      const [start, end] = range.split('-').map(s => s.trim().toUpperCase())
      // 数字邮编比较
      if (/^\d+$/.test(normalizedZip) && /^\d+$/.test(start) && /^\d+$/.test(end)) {
        const zipNum = parseInt(normalizedZip)
        const startNum = parseInt(start)
        const endNum = parseInt(end)
        if (zipNum >= startNum && zipNum <= endNum) {
          return true
        }
      }
      // 字母数字邮编比较（如英国 SW1A 1AA）
      else if (normalizedZip >= start && normalizedZip <= end) {
        return true
      }
    } else {
      // 单个邮编精确匹配
      if (normalizedZip === range.trim().toUpperCase()) {
        return true
      }
    }
  }

  return false
}

/**
 * 查找匹配的配送规则（考虑邮编范围）
 */
const findMatchingRule = (
  countryCode: string,
  zipCode: string,
  templates: ShippingTemplate[]
): ShippingRule | null => {
  const normalizedCountry = countryCode.toUpperCase()
  
  // 收集所有匹配国家的规则
  const countryRules: ShippingRule[] = []
  
  for (const template of templates) {
    if (!template.is_active) continue
    
    for (const rule of template.rules || []) {
      // 检查国家匹配 — regions 为空或不包含用户国家则跳过
      const regions = rule.regions || []
      if (regions.length === 0 || !regions.map(r => r.toUpperCase()).includes(normalizedCountry)) {
        continue
      }
      
      countryRules.push(rule)
    }
  }
  
  // 优先匹配有邮编范围的规则
  for (const rule of countryRules) {
    if (rule.zip_ranges && rule.zip_ranges.length > 0) {
      if (isZipInRanges(zipCode, rule.zip_ranges)) {
        return rule
      }
    }
  }
  
  // 其次匹配兜底规则（zip_ranges 为空）
  for (const rule of countryRules) {
    if (!rule.zip_ranges || rule.zip_ranges.length === 0) {
      return rule
    }
  }
  
  return null
}
```

#### A2.2 验证流程更新

```
用户选择国家
    ↓
用户输入邮编
    ↓
实时调用 findMatchingRule(country, zip, templates)
    ↓
┌─────────────────────────────────────┐
│ 返回 null？                          │
└─────────────────────────────────────┘
    ↓ 是                    ↓ 否
显示「该地区暂不支持配送」    显示运费
禁用付款按钮                 允许付款
```

---

### A3. 各国邮编格式参考

| 国家 | 格式 | 示例 |
|------|------|------|
| 美国 (US) | 5位数字 或 5+4位 | 10001, 10001-1234 |
| 英国 (GB) | 字母数字混合 | SW1A 1AA |
| 德国 (DE) | 5位数字 | 10115 |
| 日本 (JP) | 7位数字（含连字符） | 100-0001 |
| 加拿大 (CA) | 字母数字交替 | K1A 0B1 |
| 澳大利亚 (AU) | 4位数字 | 2000 |
| 中国 (CN) | 6位数字 | 100000 |

---

### A4. 后台配置示例

**美国分区运费配置**：

| 地区 | 邮编范围 | 运费 |
|------|----------|------|
| 纽约 | 10001-10999, 11001-11999 | $40 |
| 洛杉矶 | 90001-90899 | $45 |
| 阿拉斯加 | 99501-99950 | $80 |
| 夏威夷 | 96701-96898 | $75 |
| 美国其他地区 | （留空） | $50 |

**德国分区运费配置**：

| 地区 | 邮编范围 | 运费 |
|------|----------|------|
| 柏林 | 10115-14199 | €15 |
| 慕尼黑 | 80331-81929 | €18 |
| 德国其他地区 | （留空） | €20 |

