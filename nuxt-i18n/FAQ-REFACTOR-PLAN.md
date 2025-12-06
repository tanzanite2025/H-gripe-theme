# FAQ 重构计划：从弹窗模式迁移到页面内嵌模式

> 创建时间：2024-12-06  
> 状态：待实施

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

### 第二阶段：创建新 FAQ 基础设施 ⬜

- [ ] **步骤 2.1**：创建类型定义文件 `app/data/faq/types.ts`
- [ ] **步骤 2.2**：创建通用组件 `app/components/PageFaq.vue`
  - 手风琴样式（点击展开/收起）
  - 响应式设计，移动端友好
  - 支持浅色/深色主题
  - 平滑动画过渡
- [ ] **步骤 2.3**：创建数据导出入口 `app/data/faq/index.ts`

### 第三阶段：试点页面 ⬜

选择 1-2 个页面进行试点：

- [ ] **步骤 3.1**：创建试点页面的 FAQ 数据文件（如 `support-payment.ts`）
- [ ] **步骤 3.2**：在试点页面集成 `PageFaq` 组件
- [ ] **步骤 3.3**：测试效果，调整样式

### 第四阶段：推广到其他页面 ⬜

- [ ] **步骤 4.1**：整理所有需要 FAQ 的页面清单
- [ ] **步骤 4.2**：为每个页面创建对应的 FAQ 数据文件
- [ ] **步骤 4.3**：逐个页面集成组件

### 第五阶段：汇总页面 ⬜

- [ ] **步骤 5.1**：修改 `/support/faqs` 页面
- [ ] **步骤 5.2**：导入所有 FAQ 数据
- [ ] **步骤 5.3**：按分类/来源页面分组显示
- [ ] **步骤 5.4**：添加搜索/筛选功能（可选）

### 第六阶段：收尾工作 ⬜

- [ ] **步骤 6.1**：（可选）移除 WordPress 插件中的 FAQ JSON 生成逻辑
- [ ] **步骤 6.2**：更新相关文档
- [ ] **步骤 6.3**：最终测试所有页面

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

> ⏸️ **暂缓处理**：多语言翻译待网站全部完成后统一处理，避免碎片化翻译和重复劳动。

### 当前阶段策略

1. **FAQ 数据直接使用英文**，不使用 i18n 翻译键
2. **功能优先**，先完成 FAQ 重构的核心功能
3. **翻译文件暂不修改**，保持现有 34 个语言文件不变

### 数据文件格式（简化版）

```typescript
// app/data/faq/support-payment.ts
export const supportPaymentFaq = {
  pageKey: 'support-payment',
  pageTitle: 'Payment FAQ',           // 直接英文，不用翻译键
  pagePath: '/support/payment',
  items: [
    { 
      id: '1', 
      question: 'What payment methods do you accept?',  // 直接英文
      answer: 'We accept PayPal, Visa, Mastercard...'   // 直接英文
    },
  ]
}
```

### 网站完成后的翻译计划

1. **全局扫描**：统一扫描所有需要翻译的文本
2. **提取翻译键**：批量提取并规范化命名
3. **批量翻译**：一次性完成 34 个语言文件
4. **质量保证**：统一术语，保持一致性

### 好处

- ✅ 避免碎片化翻译，减少遗漏
- ✅ 避免重复劳动，提高效率
- ✅ 避免数据混乱，保持一致性
- ✅ 专注当前任务，功能优先


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
| 第二阶段：基础设施 | ⬜ 待开始 | - | - | - |
| 第三阶段：试点页面 | ⬜ 待开始 | - | - | - |
| 第四阶段：推广 | ⬜ 待开始 | - | - | - |
| 第五阶段：汇总页面 | ⬜ 待开始 | - | - | - |
| 第六阶段：收尾工作 | ⬜ 待开始 | - | - | - |

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

*文档结束 - 请在实施过程中更新进度状态*
