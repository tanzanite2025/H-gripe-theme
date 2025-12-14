# 主页 FAQ 预览集成

> 创建时间：2024-12-07  
> 完成时间：2024-12-07  
> 状态：✅ 已完成

---

## 一、需求背景

在主页主内容区块下方添加 FAQ 预览区域，让用户快速浏览常见问题。

### 设计要求
- 控制高度，避免内容过长
- 不显示 "All" 标签（会导致 63 条全部展示）
- 每个分类只显示有限数量的问题
- 提供"查看全部"链接跳转到 `/support/faqs`

---

## 二、技术方案

### 选择方案 C：创建轻量级主页 FAQ 组件

**组件名称**：`HomeFaqPreview.vue`

**核心功能**：
1. 使用 `getAllFaqData()` 获取所有 FAQ 数据
2. 按页面分类显示标签（不含 "All"）
3. 每个分类最多显示 N 个问题
4. 手风琴展开/收起
5. 底部"查看全部 FAQ"链接

**Props 设计**：
| Prop | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `maxItemsPerCategory` | `number` | `3` | 每个分类最多显示条目数 |
| `defaultCategory` | `string` | 第一个 | 默认选中的分类 |

---

## 三、实施步骤

### 步骤 1：创建 HomeFaqPreview.vue 组件 ✅
- [x] 创建 `app/components/HomeFaqPreview.vue`
- [x] 实现分类标签切换（按页面分类，无 "All" 标签）
- [x] 实现条目数量限制（`maxItemsPerCategory` prop）
- [x] 实现手风琴展开/收起
- [x] 添加"查看全部"链接（跳转到 `/support/faqs`）

### 步骤 2：集成到主页 ✅
- [x] 找到主页文件：`app/pages/index.vue`
- [x] 在主页内容区块下方添加 `HomeFaqPreview` 组件
- [x] 设置 `max-items-per-category="4"`

### 步骤 3：测试和调整
- [ ] 测试分类切换
- [ ] 测试展开/收起
- [ ] 测试"查看全部"链接
- [ ] 调整样式和间距（如需要）

---

## 四、组件说明

### 与 PageFaq.vue 的区别

| 特性 | PageFaq.vue | HomeFaqPreview.vue |
|------|-------------|-------------------|
| 数据来源 | 单个页面 FAQ | 所有页面 FAQ |
| 分类标签 | 显示页面内分类 | 显示页面名称 |
| "All" 标签 | 无 | 无（设计决策） |
| 条目限制 | maxItems prop | maxItemsPerCategory |
| 使用场景 | 各业务页面 | 仅主页 |

### 数据流

```
getAllFaqData() 
  → 获取所有 8 个页面的 FAQ
  → 按 pageId 分组
  → 每组取前 N 个条目
  → 渲染
```

---

## 五、维护指南

### 修改每个分类显示的条目数
修改组件调用时的 `max-items-per-category` prop：
```vue
<HomeFaqPreview :max-items-per-category="5" />
```

### 修改默认选中的分类
修改组件调用时的 `default-category` prop：
```vue
<HomeFaqPreview default-category="support-payment" />
```

### 注意事项
- 该组件依赖 `getAllFaqData()` 函数
- 新增页面 FAQ 后会自动出现在分类标签中
- 如需隐藏某个页面，需在组件内添加过滤逻辑

---

## 六、相关文件

- `app/components/HomeFaqPreview.vue` - 主页 FAQ 预览组件 ✅ 已创建
- `app/data/faq/index.ts` - FAQ 数据注册表（使用 `getAllFaqData()`）
- `app/pages/index.vue` - 主页文件 ✅ 已集成

---

## 七、维护指南

### 修改显示条目数
在 `app/pages/index.vue` 中修改 prop：
```vue
<HomeFaqPreview :max-items-per-category="5" />
```

### 修改默认选中分类
```vue
<HomeFaqPreview default-category="support-payment" />
```

### 隐藏某个页面的 FAQ
如需在主页隐藏某个页面的 FAQ，需在 `HomeFaqPreview.vue` 中添加过滤逻辑：
```typescript
const allPages = computed(() => 
  getAllFaqData().filter(p => p.pageId !== 'page-to-hide')
)
```

### 注意事项
- 新增页面 FAQ 后会自动出现在主页分类标签中
- 组件依赖 `getAllFaqData()` 函数，确保 FAQ 数据已正确注册
- 每个分类最多显示 `maxItemsPerCategory` 个条目

---

*文档完成 - 最后更新：2024-12-07*
