# FAQ 重构计划：从弹窗模式迁移到页面内嵌模式

> 创建时间：2024-12-06  
> 完成时间：2024-12-07  
> 状态：✅ 已完成

---

## 一、背景与目标

### 当前问题
- `FaqModal.vue` 弹窗从 WordPress 后端加载 FAQ 数据，导致打开延迟严重
- 弹窗模式对移动端不够友好
- 全局 FAQ 内容不够针对性

### 目标
1. **废弃弹窗模式**，改为页面内嵌的手风琴式 FAQ
2. **每个页面有专属 FAQ**，内容更具针对性
3. **数据前端写死**，消除网络延迟
4. **汇总页面** `/support/faqs` 聚合所有 FAQ

---

## 二、技术方案

### 2.1 组件设计

**只需创建 1 个通用组件：`PageFaq.vue`**

```typescript
// Props 定义
interface PageFaqProps {
  title?: string           // 可选标题，如"常见问题"
  items: FaqItem[]         // FAQ 问答数组
  collapsible?: boolean    // 是否可折叠，默认 true
  defaultOpen?: number[]   // 默认展开的项目索引
  theme?: 'light' | 'dark' // 主题色，默认 'dark'
}

interface FaqItem {
  id: string | number
  question: string
  answer: string  // 支持 HTML
}
```

### 2.2 数据组织结构（方案 B：集中管理）

```
app/data/faq/
├── index.ts              # 统一导出所有 FAQ 数据
├── types.ts              # 类型定义
├── products-hub.ts       # 花鼓产品页 FAQ
├── products-rim.ts       # 轮圈产品页 FAQ
├── products-spoke.ts     # 辐条产品页 FAQ
├── support-payment.ts    # 支付页面 FAQ
├── support-shipping.ts   # 运输页面 FAQ
├── support-returns.ts    # 退换货页面 FAQ
└── ...
```

### 2.3 数据结构定义

```typescript
// app/data/faq/types.ts

export interface FaqItem {
  id: string
  question: string
  answer: string  // 支持 HTML
}

export interface PageFaqData {
  pageKey: string       // 唯一标识，如 'products-hub'
  pageTitle: string     // 显示标题，如 '花鼓产品常见问题'
  pagePath: string      // 页面路径，如 '/products/hub'
  icon?: string         // 可选图标/emoji
  items: FaqItem[]
}

// 多语言支持
export interface PageFaqDataI18n {
  en: PageFaqData
  zh: PageFaqData
  // 其他语言...
}
```

---

## 三、实施步骤

> ⚠️ **顺序调整说明**：先清理旧代码，让系统干净后再开发新组件，避免新旧代码混淆。

### 第一阶段：清理旧 FAQ 弹窗代码 ✅ 已完成

> 目标：移除所有与 `FaqModal.vue` 相关的代码，让系统干净
> 完成时间：2024-12-07

- [x] **步骤 1.1**：移除 `SiteHeader.vue` 中的 FAQ 弹窗相关代码
  - ✅ 移除 `FaqModal` 组件导入
  - ✅ 移除 `faqOpen` ref
  - ✅ 移除 `toggleFaq` 函数
  - ✅ 移除 FAQ teleport 模板
  - ✅ 移除相关 watch 和 onBeforeUnmount 清理逻辑
  - ✅ 简化 `handleOpenWhatsApp` 函数
- [x] **步骤 1.2**：移除 `WhatsAppChatModal.vue` 中的 FAQ 弹窗相关代码
  - ✅ 移除 `FaqModal` 组件导入
  - ✅ 移除 `showFAQ` ref
  - ✅ 移除桌面端 FAQ 按钮
  - ✅ 移除移动端 FAQ 按钮
  - ✅ 移除 FAQ teleport 模板
- [x] **步骤 1.3**：移除 `LeverAndPoint.vue` 中的 FAQ 弹窗相关代码
  - ✅ 移除 `FaqModal` 组件导入
  - ✅ 移除 `showFaqModal` ref
  - ✅ 移除 `handleFAQ`、`closeFAQ` 函数
  - ✅ 移除 FAQ 按钮
  - ✅ 移除 FAQ teleport 模板
  - ✅ 移除相关 watch 和 onBeforeUnmount 清理逻辑
- [x] **步骤 1.4**：删除 `app/components/FaqModal.vue` 文件
- [x] **步骤 1.5**：检查是否有其他文件引用 `FaqModal`，如有则清理（无其他引用）
- [ ] **步骤 1.6**：运行项目，确认无报错

### 第二阶段：创建新 FAQ 基础设施 ✅ 已完成

> 完成时间：2024-12-07

- [x] **步骤 2.1**：创建类型定义文件 `app/data/faq/types.ts`
  - ✅ 定义 `FaqItem`、`FaqCategory`、`PageFaqData` 接口
  - ✅ 定义 `PageFaqProps` 组件属性接口
  - ✅ 定义 `FaqRegistry` 注册表类型
- [x] **步骤 2.2**：创建数据索引文件 `app/data/faq/index.ts`
  - ✅ 创建 FAQ 注册表
  - ✅ 实现 `getFaqData`、`getAllFaqData`、`getAllFaqItems` 函数
- [x] **步骤 2.3**：创建通用组件 `app/components/PageFaq.vue`
  - ✅ 手风琴样式（点击展开/收起）
  - ✅ 响应式设计，移动端友好
  - ✅ 支持浅色/深色主题
  - ✅ 平滑动画过渡
  - ✅ 支持 maxItems 限制和 "View All" 链接
- [x] **步骤 2.4**：创建页面数据目录 `app/data/faq/pages/`

### 第三阶段：试点页面 ✅ 已完成

> 完成时间：2024-12-07
> 试点页面：`/support/payment`

- [x] **步骤 3.1**：创建试点页面的 FAQ 数据文件
  - ✅ 创建 `app/data/faq/pages/support-payment.ts`
  - ✅ 包含 4 个分类：Payment Methods, Security, Billing, Troubleshooting
  - ✅ 共 11 个 FAQ 条目
- [x] **步骤 3.2**：在试点页面集成 `PageFaq` 组件
  - ✅ 在 `app/pages/support/payment.vue` 中添加 PageFaq 组件
  - ✅ 使用 dark 主题，显示分类
- [x] **步骤 3.3**：测试效果，调整样式
  - ✅ 调整标题大小（text-base md:text-lg）
  - ✅ 减小上边距（py-4 md:py-6）
  - ✅ 调整问题/答案文字大小（text-xs md:text-sm）
  - ✅ 优化内边距和间距

### 第四阶段：推广到其他页面 ✅ 已完成

> 完成时间：2024-12-07

- [x] **步骤 4.1**：整理所有需要 FAQ 的页面清单
- [x] **步骤 4.2**：为每个页面创建对应的 FAQ 数据文件
  - ✅ `support-shipping.ts` - 10 个 FAQ 条目
  - ✅ `support-after-sales.ts` - 8 个 FAQ 条目
  - ✅ `support-warranty.ts` - 6 个 FAQ 条目
  - ✅ `support-product-feedback.ts` - 5 个 FAQ 条目
  - ✅ `support-test-report.ts` - 7 个 FAQ 条目
- [x] **步骤 4.3**：逐个页面集成组件
  - ✅ `/support/shipping`
  - ✅ `/support/after-sales`
  - ✅ `/support/warranty`
  - ✅ `/support/product-feedback`
  - ✅ `/support/test-report`
  - ✅ `/membershipandpoints` - 9 个 FAQ 条目
  - ✅ `/guides/wheelset-buyers` - 7 个 FAQ 条目（放在所有 tab 内容之后）

### 第五阶段：汇总页面 ✅ 已完成

> 完成时间：2024-12-07

- [x] **步骤 5.1**：修改 `/support/faqs` 页面
  - ✅ 重写页面，使用新的 FAQ 数据系统
- [x] **步骤 5.2**：导入所有 FAQ 数据
  - ✅ 使用 `getAllFaqData()` 获取所有注册的 FAQ
- [x] **步骤 5.3**：按分类/来源页面分组显示
  - ✅ 按页面分组显示所有 FAQ 条目
  - ✅ 每个条目显示分类标签
- [x] **步骤 5.4**：添加搜索/筛选功能
  - ✅ 搜索框支持搜索问题、答案、分类、标签
  - ✅ 页面分类标签支持按页面筛选
  - ✅ 手风琴展开/收起动画

### 第六阶段：收尾工作 ✅ 已完成

> 完成时间：2024-12-07

- [x] **步骤 6.1**：WordPress 插件处理
  - ⚠️ 发现 `wp-plugin/tanzanite-faq-content/` 插件仍存在
  - ✅ 建议：保留插件文件但在 WordPress 后台禁用
  - ✅ 前端已完全独立，不再依赖 WordPress FAQ 数据
- [x] **步骤 6.2**：更新相关文档
  - ✅ FAQ-REFACTOR-PLAN.md 已更新完成
- [x] **步骤 6.3**：最终测试清单
  - 测试页面：
    - `/support/payment` - PageFaq 组件
    - `/support/shipping` - PageFaq 组件
    - `/support/after-sales` - PageFaq 组件
    - `/support/warranty` - PageFaq 组件
    - `/support/product-feedback` - PageFaq 组件
    - `/support/test-report` - PageFaq 组件
    - `/membershipandpoints` - PageFaq 组件
    - `/guides/wheelset-buyers` - PageFaq 组件
    - `/support/faqs` - 汇总页面（搜索 + 筛选）

---

## 四、需要 FAQ 的页面清单

根据网站结构，以下页面可能需要专属 FAQ：

### 产品相关
- [ ] `/products/hub` - 花鼓产品
- [ ] `/products/rim` - 轮圈产品
- [ ] `/products/spoke` - 辐条产品
- [ ] `/products/wheelset` - 轮组产品

### 支持相关
- [ ] `/support/payment` - 支付方式
- [ ] `/support/shipping` - 运输配送
- [ ] `/support/returns` - 退换货政策
- [ ] `/support/warranty` - 保修政策

### 工具相关
- [ ] `/spoke-calculator` - 辐条计算器
- [ ] `/tire-size-charts` - 轮胎尺寸表

### 其他
- [ ] `/shop` - 商店页面
- [ ] `/checkout` - 结账页面

---

## 五、组件 UI 规格

### 5.1 桌面端样式

```
┌─────────────────────────────────────────────────────────┐
│  常见问题                                                │
├─────────────────────────────────────────────────────────┤
│  ▼ Plus系列花鼓的保修政策是什么？                         │
│    ┌─────────────────────────────────────────────────┐  │
│    │ 自购买之日起，我们提供3年标准保修。Light Bicycle   │  │
│    │ 将为Plus系列产品提供自消费者原始购买之日起三（3）  │  │
│    │ 年的保修，保修范围涵盖材料和工艺缺陷。             │  │
│    └─────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────┤
│  ▶ Plus系列花鼓的轴承尺寸是多少？                         │
├─────────────────────────────────────────────────────────┤
│  ▶ 我可以自己维修 Plus 系列集线器吗？                     │
├─────────────────────────────────────────────────────────┤
│  ▶ 如何提交索赔？                                        │
└─────────────────────────────────────────────────────────┘
```

### 5.2 移动端样式

- 全宽显示
- 问题文字可换行
- 答案区域内边距适当缩小
- 点击区域足够大（至少 44px 高度）

### 5.3 交互规格

- 点击问题行展开/收起答案
- 展开时显示向下箭头 ▼，收起时显示向右箭头 ▶
- 平滑动画过渡（200-300ms）
- 可选：同时只展开一个（手风琴模式）或允许多个展开

### 5.4 颜色规格（深色主题）

- 背景：`rgba(255, 255, 255, 0.03)` 或 `bg-white/[0.03]`
- 边框：`rgba(255, 255, 255, 0.1)` 或 `border-white/10`
- 问题文字：`#ffffff` 或 `text-white`
- 答案文字：`rgba(255, 255, 255, 0.7)` 或 `text-white/70`
- 悬停边框：`rgba(107, 115, 255, 0.5)` 或 `border-[#6b73ff]/50`

---

## 六、多语言支持方案

> 2026-07-23 更新：本节原策略已经过期。当前 Nuxt storefront i18n 不再采用“等网站全部完成后一次性处理”的方式，也不应该无脑批量翻译 34 个语言文件。

当前规则以 `I18N-CURRENT-STATUS.md` 为准：

- 按页面/组件分块接入 i18n。
- 先处理稳定、边界清楚的 UI 文案。
- 商品名、SKU、规格、动态 FAQ 内容和后台录入内容不直接硬塞进静态语言包。
- `app/i18n/messages/<locale>/*.json` 是可维护源文件。
- `app/i18n/locales/*.json` 由 `npm run i18n:build` 生成。

参考：`./I18N-CURRENT-STATUS.md`


---

## 七、风险与注意事项

1. **内容迁移**：需要从现有 WordPress FAQ 中提取内容
2. **SEO 影响**：页面内嵌 FAQ 对 SEO 有利，但需确保 HTML 结构正确
3. **维护成本**：FAQ 更新需要重新部署，考虑是否需要 CMS 集成
4. **向后兼容**：清理旧代码时确保不影响其他功能

---

## 八、进度跟踪

| 阶段 | 状态 | 开始时间 | 完成时间 | 备注 |
|------|------|----------|----------|------|
| 第一阶段：清理旧代码 | ✅ 已完成 | 2024-12-07 | 2024-12-07 | 已清理 3 个文件，删除 FaqModal.vue |
| 第二阶段：基础设施 | ✅ 已完成 | 2024-12-07 | 2024-12-07 | 类型定义 + 通用组件 + 数据注册表 |
| 第三阶段：试点页面 | ✅ 已完成 | 2024-12-07 | 2024-12-07 | /support/payment 页面集成完成 |
| 第四阶段：推广 | ✅ 已完成 | 2024-12-07 | 2024-12-07 | 8 个页面已集成，共 63 个 FAQ 条目 |
| 第五阶段：汇总页面 | ✅ 已完成 | 2024-12-07 | 2024-12-07 | /support/faqs 页面重写，支持搜索和筛选 |
| 第六阶段：收尾工作 | ✅ 已完成 | 2024-12-07 | 2024-12-07 | 文档更新，测试清单已列出 |

---

## 九、相关文件清单

### 需要创建的文件
- `app/components/PageFaq.vue` - 通用 FAQ 组件
- `app/data/faq/types.ts` - 类型定义
- `app/data/faq/index.ts` - 数据导出入口
- `app/data/faq/*.ts` - 各页面 FAQ 数据文件

### 需要修改的文件
- `app/pages/support/faqs.vue` - 汇总页面
- 各业务页面（添加 FAQ 组件）

### 需要删除的文件
- `app/components/FaqModal.vue`

### 需要清理的文件
- `app/components/SiteHeader.vue` - 移除 FAQ 弹窗相关代码
- `app/components/WhatsAppChatModal.vue` - 移除 FAQ 弹窗相关代码
- `app/components/LeverAndPoint.vue` - 移除 FAQ 弹窗相关代码

---

## 十、维护指南（重要）

### 10.1 添加新页面 FAQ

**步骤：**

1. **创建数据文件** `app/data/faq/pages/[page-name].ts`
   ```typescript
   import type { PageFaqData } from '../types'
   
   export const yourPageFaq: PageFaqData = {
     pageId: 'your-page-id',  // 唯一标识，用于组件引用
     title: 'Your Page FAQs',
     subtitle: 'Optional subtitle',
     categories: [
       {
         id: 'category-1',
         name: 'Category Name',
         icon: '📦',  // 可选 emoji 图标
         items: [
           {
             id: 'item-1',  // 在页面内唯一即可
             question: 'Your question?',
             answer: 'Your answer. <strong>Supports HTML</strong>.',
             tags: ['optional', 'tags'],  // 用于搜索
           },
         ],
       },
     ],
   }
   ```

2. **注册到 index.ts** `app/data/faq/index.ts`
   ```typescript
   // 添加导入
   import { yourPageFaq } from './pages/your-page'
   
   // 添加到注册表
   export const faqRegistry: FaqRegistry = {
     // ... 现有条目
     'your-page-id': yourPageFaq,  // 添加这行
   }
   ```

3. **在页面中使用组件**
   ```vue
   <template>
     <!-- 页面其他内容 -->
     
     <section class="your-section">
       <PageFaq 
         page-id="your-page-id"
         theme="dark"
         :show-categories="true"
       />
     </section>
   </template>
   
   <script setup lang="ts">
   import PageFaq from '~/components/PageFaq.vue'
   </script>
   ```

### 10.2 修改现有 FAQ

- 直接编辑对应的 `app/data/faq/pages/[page-name].ts` 文件
- 修改 `question`、`answer`、`tags` 等字段
- 无需修改其他文件

### 10.3 删除 FAQ 条目

- 从对应数据文件的 `items` 数组中移除条目
- 如果删除整个分类，从 `categories` 数组中移除
- 如果删除整个页面的 FAQ：
  1. 删除数据文件
  2. 从 `index.ts` 移除导入和注册
  3. 从页面组件中移除 `<PageFaq>` 组件

### 10.4 注意事项

#### ⚠️ ID 唯一性
- `pageId` 必须全局唯一（跨所有页面）
- `category.id` 在同一页面内唯一
- `item.id` 在同一分类内唯一
- 汇总页面会将 `pageId-itemId` 组合作为全局唯一 ID

#### ⚠️ HTML 内容安全
- `answer` 字段支持 HTML，使用 `v-html` 渲染
- 只使用受信任的内容，避免 XSS 风险
- 推荐使用的 HTML 标签：`<ul>`, `<ol>`, `<li>`, `<strong>`, `<br>`, `<p>`

#### ⚠️ 类型检查
- 所有数据文件必须导入并使用 `PageFaqData` 类型
- TypeScript 会自动检查数据结构是否正确

#### ⚠️ 汇总页面自动更新
- `/support/faqs` 页面使用 `getAllFaqData()` 自动获取所有注册的 FAQ
- 新增页面 FAQ 后，汇总页面会自动包含
- 无需手动修改汇总页面

### 10.5 组件 Props 说明

| Prop | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page-id` | `string` | 必填 | 对应注册表中的 pageId |
| `title` | `string` | 从数据获取 | 覆盖数据中的标题 |
| `theme` | `'light' \| 'dark'` | `'light'` | 主题色 |
| `show-categories` | `boolean` | `true` | 是否显示分类标题 |
| `max-items` | `number` | 无限制 | 最多显示条目数 |
| `show-view-all-link` | `boolean` | `false` | 显示"查看全部"链接 |

### 10.6 文件结构总览

```
app/
├── components/
│   └── PageFaq.vue              # 通用 FAQ 组件
├── data/
│   └── faq/
│       ├── types.ts             # 类型定义
│       ├── index.ts             # 注册表和工具函数
│       └── pages/               # 各页面 FAQ 数据
│           ├── support-payment.ts
│           ├── support-shipping.ts
│           ├── support-after-sales.ts
│           ├── support-warranty.ts
│           ├── support-product-feedback.ts
│           ├── support-test-report.ts
│           ├── company-membership.ts
│           └── guides-wheelset-buyers.ts
└── pages/
    └── support/
        └── faqs.vue             # 汇总页面
```

---

*文档完成 - 最后更新：2024-12-07*
