# FAQ 系统实现指南（最终方案）

## 📋 项目概述

**目标**：创建一个支持 34 种语言的 FAQ 系统

**核心架构**：
- **SEO 设置**：扩展现有 `tanzanite-setting` 插件（FAQ Schema、Sitemap）
- **内容管理**：新建独立 `tanzanite-faq-content` 插件（编辑界面、JSON 生成）
- **前端展示**：Nuxt 3 + i18n（读取 JSON 文件）
- **数据传输**：静态 JSON 文件（无 REST API）

---

## ⚠️ 重要说明：语言代码与文件命名

### **语言代码来源**
FAQ 插件的语言列表**直接从 Nuxt i18n 配置读取**：
- **源文件目录**：`nuxt-i18n/i18n/locales/*.json`
- **读取方式**：扫描目录中的所有 `.json` 文件
- **语言代码**：使用文件名作为语言代码（如 `zh_cn.json` → `zh_cn`）

### **语言代码一致性（关键！）**
```
Nuxt i18n locale     →  FAQ 数据库 locale  →  生成的 JSON 文件
─────────────────────────────────────────────────────────────
zh_cn                →  zh_cn              →  zh_cn.json
en                   →  en                 →  en.json
fr                   →  fr                 →  fr.json
ja                   →  ja                 →  ja.json
```

**⚠️ 绝对不要修改语言代码格式！** 必须与 Nuxt i18n 的 locale 完全一致。

### **前端语言切换流程**
```javascript
// faq.vue
const { locale } = useI18n()  // 从 Nuxt i18n 获取当前语言

// 自动跟随语言切换
const { data: faqData } = await useFetch(
  () => `${config.public.wordpressUrl}/wp-content/uploads/faq/${locale.value}.json`
)
```

**工作原理**：
1. 用户切换语言 → Nuxt i18n 的 `locale` 改变（如 `zh_cn` → `en`）
2. FAQ 页面响应 → `useFetch` 的 URL 自动更新
3. 请求新文件 → 从 `zh_cn.json` 切换到 `en.json`
4. 页面重新渲染 → 显示新语言的 FAQ

### **为什么使用独立 JSON 文件？**

**✅ 优点：**
- 内容与 UI 翻译分离，职责清晰
- FAQ 内容可以动态更新，无需重新构建 Nuxt
- 可以按需加载（只在 FAQ 页面加载）
- WordPress 管理员可以随时修改，立即生效
- 与 Nuxt i18n 系统完美集成，自动跟随语言切换

**❌ 不要合并到 Nuxt 翻译文件：**
- 每次修改 FAQ 需要重新构建和部署 Nuxt
- 翻译文件会变得很大
- 所有页面都会加载 FAQ 数据（即使不需要）
- 内容管理不灵活

### **文件位置对比**

| 文件类型 | 位置 | 用途 |
|---------|------|------|
| Nuxt UI 翻译 | `nuxt-i18n/i18n/locales/zh_cn.json` | 页面 UI 文本翻译 |
| FAQ 数据 | `/wp-content/uploads/faq/zh_cn.json` | FAQ 问答内容数据 |

**✅ 两者不冲突，各司其职！**

---

## 🎯 系统架构（两个插件分离）

```
┌─────────────────────────────────────────────────────────────┐
│                    WordPress 后台                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────┐    ┌──────────────────────┐      │
│  │  SEO 插件（扩展）    │    │  FAQ 内容插件（新建）│      │
│  │  tanzanite-setting   │    │  tanzanite-faq-content│     │
│  ├──────────────────────┤    ├──────────────────────┤      │
│  │ • 34 语言配置        │◄───│ • 读取语言列表       │      │
│  │ • FAQ Schema 开关    │    │ • FAQ 编辑界面       │      │
│  │ • FAQ Sitemap        │    │ • 分类管理           │      │
│  │ • Meta 模板          │    │ • 34 语言编辑        │      │
│  └──────────────────────┘    │ • 直接生成 JSON ✨   │      │
│                               └──────────────────────┘      │
│                                        │                     │
│                                        ↓                     │
│                              直接写入 JSON 文件              │
│                              (无 REST API)                   │
└────────────────────────────────────────┼─────────────────────┘
                                         │
                                         ↓
                    /wp-content/uploads/faq/
                    ├── en.json
                    ├── zh.json
                    └── ... (34 个文件)
                                         │
                                         ↓ (HTTP 请求)
                    ┌────────────────────┴─────────────────────┐
                    │         Nuxt 前端 (nuxt-i18n)            │
                    ├──────────────────────────────────────────┤
                    │  pages/faq.vue                           │
                    │  • useFetch() 读取 JSON                  │
                    │  • 显示 FAQ 内容                         │
                    │  • 使用 SEO 插件的 Schema 设置           │
                    └──────────────────────────────────────────┘
```

---

## 📂 项目结构（最终版）

```
wordpress/
├── wp-content/
│   ├── plugins/
│   │   ├── tanzanite-setting/                    # 现有插件（扩展）
│   │   │   └── includes/
│   │   │       └── class-mytheme-seo.php         # 添加 FAQ SEO 设置
│   │   │
│   │   └── tanzanite-faq-content/                # 新建插件 ✨
│   │       ├── tanzanite-faq-content.php         # 主文件
│   │       ├── includes/
│   │       │   ├── class-faq-database.php        # 数据库操作
│   │       │   ├── class-faq-editor.php          # 编辑界面
│   │       │   └── class-faq-json-generator.php  # JSON 生成器
│   │       └── assets/
│   │           ├── css/admin.css
│   │           └── js/admin.js
│   │
│   └── uploads/
│       └── faq/                                  # JSON 文件存储
│           ├── en.json
│           ├── zh.json
│           └── ... (34个语言文件)

nuxt-i18n/
└── app/
    └── pages/
        └── faq.vue                         # FAQ 页面
```

---

## 🚀 实施步骤（最终方案）

### **阶段 1：扩展 SEO 插件（FAQ SEO 设置）**

#### 步骤 1.1：修改 SEO 插件配置
**文件**：`wp-content/plugins/tanzanite-setting/includes/class-mytheme-seo.php`

**任务**：
- [ ] 在 `default_settings()` 中添加 FAQ Schema 配置
- [ ] 在 `filter_wp_sitemaps_index()` 中添加 FAQ 页面
- [ ] 添加 FAQ meta 模板

**预计时间**：30 分钟

---

### **阶段 2：创建 FAQ 内容管理插件**

#### 步骤 2.1：创建插件文件夹
```bash
cd wp-content/plugins/
mkdir tanzanite-faq-content
cd tanzanite-faq-content
```

#### 步骤 2.2：创建数据库表
**表结构**：
- `wp_mytheme_faq` - 主表（id, category, order_num, created_at, updated_at）
- `wp_mytheme_faq_i18n` - 多语言表（faq_id, locale, question, answer）

**预计时间**：30 分钟

#### 步骤 2.3：实现 JSON 生成器
**文件**：`includes/class-faq-json-generator.php`

**功能**：
- 从 SEO 插件读取 34 种语言列表
- 从数据库读取 FAQ 数据
- 按语言生成 JSON 文件到 `/wp-content/uploads/faq/`
- 保存时自动触发生成

**预计时间**：1-2 小时

#### 步骤 2.4：创建后台编辑界面
**文件**：`includes/class-faq-editor.php`

**功能**：
- 34 个语言标签页编辑
- 分类管理（产品、配送、退货、支付）
- 富文本编辑器
- 拖拽排序
- 保存后自动生成 JSON

**预计时间**：2-3 小时

---

### **阶段 3：Nuxt 前端页面**

#### 步骤 3.1：创建 FAQ 页面
**文件**：`app/pages/faq.vue`

**功能**：
- 使用 `useFetch()` 读取 JSON 文件
- 根据 `locale` 加载对应语言
- 显示分类和 FAQ 列表
- 添加 FAQ Schema.org 结构化数据

**预计时间**：1-2 小时

#### 步骤 3.2：配置路由（自动）
Nuxt i18n 自动生成 34 个路由：
- `/en/faq`, `/zh/faq`, `/ja/faq` ... 

**无需手动配置**

---

### **阶段 4：集成与测试**

#### 步骤 4.1：SEO 插件测试
- [ ] 验证 FAQ Schema 配置已添加
- [ ] 检查 Sitemap 包含 FAQ 页面
- [ ] 确认语言列表可被 FAQ 插件读取

#### 步骤 4.2：FAQ 插件测试
- [ ] 创建测试 FAQ（多语言）
- [ ] 验证 JSON 文件生成到 `/wp-content/uploads/faq/`
- [ ] 检查文件权限（644）
- [ ] 验证 JSON 结构正确

#### 步骤 4.3：Nuxt 前端测试
- [ ] 访问 `/en/faq` 验证英文显示
- [ ] 访问 `/zh/faq` 验证中文显示
- [ ] 测试语言切换功能
- [ ] 验证 Schema.org 结构化数据
- [ ] 检查 SEO meta 标签

#### 步骤 4.4：性能测试
- [ ] 检查 JSON 文件大小（应 < 50KB）
- [ ] 测试首次加载速度
- [ ] 验证浏览器缓存生效
- [ ] 测试 CDN 加速（如果配置）

---

## 💻 核心代码实现

### 1. SEO 插件扩展（FAQ 设置）

**文件**：`tanzanite-setting/includes/class-mytheme-seo.php`

**修改点 1：添加 FAQ Schema 配置**
```php
// 在 default_settings() 方法中
'schema' => [
    'product' => [...],
    'faq' => [              // 新增
        'enabled' => true,
        'organizationName' => 'Tanzanite',
        'organizationUrl' => home_url(),
    ],
],
```

**修改点 2：添加 FAQ 到 Sitemap**
```php
// 在 filter_wp_sitemaps_index() 方法中
$languages = $this->get_languages();
foreach ($languages as $lang) {
    $sitemaps[] = [
        'loc' => home_url("/{$lang}/faq"),
        'lastmod' => current_time('c'),
    ];
}
```

---

### 2. FAQ 内容插件主文件

**文件**：`tanzanite-faq-content/tanzanite-faq-content.php`

**文件路径**：`app/pages/faq.vue`

**功能**：
- 根据当前语言加载对应 JSON
- 显示 FAQ 列表
- 搜索和过滤功能
- SEO 优化

---

## 🌍 支持的 34 种语言

```javascript
const languages = [
  'en', 'zh', 'ja', 'ko', 'es', 'fr', 'de', 'it', 'pt', 'ru',
  'ar', 'hi', 'bn', 'pa', 'jv', 'ms', 'te', 'vi', 'mr', 'ta',
  'tr', 'ur', 'gu', 'pl', 'uk', 'fa', 'ml', 'kn', 'or', 'my',
  'th', 'nl', 'sv', 'ro'
]
```

---

## 📊 JSON 文件结构

```json
{
  "categories": [
    {
      "id": "product",
      "name": "Product Questions",
      "icon": "📦",
      "items": [
        {
          "id": 1,
          "question": "How do I track my order?",
          "answer": "You can track your order by...",
          "order": 1
        }
      ]
    },
    {
      "id": "shipping",
      "name": "Shipping & Delivery",
      "icon": "🚚",
      "items": [...]
    }
  ],
  "meta": {
    "last_updated": "2025-11-16T22:00:00Z",
    "total_items": 50
  }
}
```

---

## ⚙️ 配置说明

### WordPress 配置

1. **文件权限**
```bash
chmod 755 /wp-content/uploads/faq/
chmod 644 /wp-content/uploads/faq/*.json
```

2. **Nginx 配置**（可选，用于缓存）
```nginx
location /wp-content/uploads/faq/ {
    expires 1h;
    add_header Cache-Control "public, immutable";
    add_header Access-Control-Allow-Origin "*";
}
```

### Nuxt 配置

1. **环境变量**
```env
# .env
WORDPRESS_URL=https://yoursite.com
FAQ_JSON_BASE_URL=https://yoursite.com/wp-content/uploads/faq
```

2. **Nuxt 配置**
```javascript
// nuxt.config.ts
export default defineNuxtConfig({
  runtimeConfig: {
    public: {
      wordpressUrl: process.env.WORDPRESS_URL,
      faqJsonBaseUrl: process.env.FAQ_JSON_BASE_URL
    }
  }
})
```

---

## 🔧 故障排查

### 问题 1：JSON 文件未生成
**检查**：
- WordPress 文件夹权限
- PHP 错误日志
- 插件是否激活

**解决**：
```bash
sudo chown -R www-data:www-data /wp-content/uploads/faq/
sudo chmod -R 755 /wp-content/uploads/faq/
```

### 问题 2：Nuxt 无法加载 JSON
**检查**：
- JSON 文件 URL 是否正确
- CORS 设置
- 网络请求（浏览器开发者工具）

**解决**：
在 WordPress `.htaccess` 添加：
```apache
<IfModule mod_headers.c>
    <FilesMatch "\.(json)$">
        Header set Access-Control-Allow-Origin "*"
    </FilesMatch>
</IfModule>
```

### 问题 3：语言切换不生效
**检查**：
- Nuxt i18n 配置
- 路由是否正确生成
- locale 变量是否正确

---

## 📈 性能优化

### 1. 浏览器缓存
```javascript
// Nuxt 中
const { data } = await useFetch(url, {
  headers: {
    'Cache-Control': 'public, max-age=3600'
  }
})
```

### 2. CDN 加速
将 `/wp-content/uploads/faq/` 文件夹同步到 CDN

### 3. 文件压缩
在 WordPress 插件中启用 JSON 压缩：
```php
$json = json_encode($data, JSON_UNESCAPED_UNICODE);
file_put_contents($file . '.gz', gzencode($json, 9));
```

---

## 🎨 UI/UX 建议

### WordPress 后台
- 使用拖拽排序（Sortable.js）
- 富文本编辑器（TinyMCE）
- 实时预览功能
- 批量操作（导入/导出）

### Nuxt 前端
- 搜索功能（实时过滤）
- 分类导航
- 折叠/展开动画
- 移动端优化

---

## 🔐 安全考虑

1. **输入验证**
```php
// WordPress 插件中
$question = sanitize_text_field($_POST['question']);
$answer = wp_kses_post($_POST['answer']);
```

2. **权限检查**
```php
if (!current_user_can('manage_options')) {
    wp_die('Unauthorized');
}
```

3. **Nonce 验证**
```php
check_ajax_referer('faq_nonce', 'nonce');
```

---

## 📝 开发检查清单

### WordPress 插件
- [ ] 插件激活/停用钩子
- [ ] 数据库表创建
- [ ] 管理页面 UI
- [ ] AJAX 保存功能
- [ ] JSON 文件生成
- [ ] 错误处理
- [ ] 权限检查

### Nuxt 前端
- [ ] FAQ 页面创建
- [ ] JSON 数据获取
- [ ] 多语言路由
- [ ] SEO 优化
- [ ] 响应式设计
- [ ] 加载状态
- [ ] 错误处理

### 测试
- [ ] 创建 FAQ 测试
- [ ] 编辑 FAQ 测试
- [ ] 删除 FAQ 测试
- [ ] 多语言切换测试
- [ ] 性能测试
- [ ] 移动端测试
- [ ] SEO 测试

---

## 🎯 扩展功能：FAQ 快捷弹窗（可选）

### **功能概述**

在 SiteHeader 添加 FAQ 按钮，点击后弹出快捷 FAQ 弹窗，显示精选的常见问题，方便用户快速查看。

### **架构设计**

```
SiteHeader
├── 分享按钮
├── 翻译转换器
└── FAQ 按钮（新增）✨
         ↓ 点击
    FAQQuickModal 组件
    ├── 显示精选问题（5-10 条）
    ├── 可折叠/展开
    └── "查看完整 FAQ" 按钮 → 跳转到 /faq 页面
```

### **实施步骤**

#### **步骤 1：在 SiteHeader 添加 FAQ 按钮**

**文件**：`app/components/SiteHeader.vue`

**位置**：在分享按钮和翻译转换器之间

```vue
<!-- 桌面端 -->
<button 
  class="pointer-events-auto text-white shadow-[0_2px_8px_#2aa3ff40] hover:shadow-[0_4px_12px_#2aa3ff40] transition-all duration-200 w-[115px] h-12 rounded-full inline-flex items-center justify-center bg-black border-2 border-[#6b73ff]" 
  @click.stop="toggleFAQQuick()" 
  :aria-expanded="faqQuickOpen" 
  aria-haspopup="dialog" 
  aria-label="Open FAQ quick view"
>
  <svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
    <circle cx="12" cy="12" r="10"/>
    <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
    <line x1="12" y1="17" x2="12.01" y2="17"/>
  </svg>
</button>

<!-- 移动端类似 -->
```

**预计时间**：15 分钟

---

#### **步骤 2：创建 FAQQuickModal 组件**

**文件**：`app/components/FAQQuickModal.vue`

**功能**：
- 读取 JSON 数据（与 faq.vue 相同）
- 只显示前 5-10 条精选问题
- 折叠/展开交互
- 黑色主题 + 蓝色边框
- "查看完整 FAQ" 按钮

**核心代码结构**：

```vue
<template>
  <Teleport to="body">
    <Transition name="fade">
      <div v-if="isFAQQuickOpen" class="fixed inset-0 z-50 flex items-center justify-center">
        <!-- 遮罩层 -->
        <div class="absolute inset-0 bg-black/80" @click="closeFAQQuick"></div>
        
        <!-- 弹窗内容 -->
        <div class="relative bg-black border-2 border-[#6b73ff] rounded-2xl w-full max-w-2xl max-h-[80vh] overflow-hidden">
          <!-- 头部 -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-white/10">
            <h2 class="text-2xl font-bold text-white">❓ Quick FAQ</h2>
            <button @click="closeFAQQuick" class="...">×</button>
          </div>
          
          <!-- FAQ 列表 -->
          <div class="overflow-y-auto p-6 max-h-[60vh]">
            <div v-for="item in featuredFAQs" :key="item.id" class="...">
              <!-- FAQ 项目 -->
            </div>
          </div>
          
          <!-- 底部按钮 -->
          <div class="px-6 py-4 border-t border-white/10">
            <NuxtLink to="/faq" @click="closeFAQQuick" class="...">
              查看完整 FAQ →
            </NuxtLink>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
const { locale } = useI18n()
const config = useRuntimeConfig()

// 状态管理
const isFAQQuickOpen = ref(false)
const openItems = ref([])

// 获取 FAQ 数据
const { data: faqData } = await useFetch(
  () => `${config.public.wordpressUrl}/wp-content/uploads/faq/${locale.value}.json`,
  { 
    key: `faq-quick-${locale.value}`,
    lazy: true // 按需加载
  }
)

// 精选问题（前 10 条）
const featuredFAQs = computed(() => {
  if (!faqData.value || !faqData.value.categories) return []
  
  return faqData.value.categories
    .flatMap(cat => cat.items)
    .slice(0, 10) // 只取前 10 条
})

// 打开/关闭弹窗
const closeFAQQuick = () => {
  isFAQQuickOpen.value = false
  openItems.value = []
}

// 全局事件监听（从 SiteHeader 触发）
if (import.meta.client) {
  window.addEventListener('open-faq-quick', () => {
    isFAQQuickOpen.value = true
  })
}
</script>
```

**预计时间**：1-1.5 小时

---

#### **步骤 3：添加状态管理（可选）**

**文件**：`app/composables/useFAQ.ts`

```typescript
export const useFAQ = () => {
  const isFAQQuickOpen = useState('faq-quick-open', () => false)
  
  const openFAQQuick = () => {
    isFAQQuickOpen.value = true
  }
  
  const closeFAQQuick = () => {
    isFAQQuickOpen.value = false
  }
  
  return {
    isFAQQuickOpen,
    openFAQQuick,
    closeFAQQuick
  }
}
```

**预计时间**：15 分钟

---

### **精选问题选择策略**

#### **方案 A：硬编码（推荐 - 简单）**

```javascript
// 在组件中直接指定
const featuredFAQs = computed(() => {
  if (!faqData.value) return []
  
  // 只取前 10 条
  return faqData.value.categories
    .flatMap(cat => cat.items)
    .slice(0, 10)
})
```

#### **方案 B：按分类选择（推荐 - 平衡）**

```javascript
// 每个分类取前 2-3 条
const featuredFAQs = computed(() => {
  if (!faqData.value) return []
  
  return faqData.value.categories.flatMap(cat => 
    cat.items.slice(0, 2) // 每个分类 2 条
  )
})
```

#### **方案 C：后台标记（高级 - 灵活）**

需要修改：
1. 数据库添加 `featured` 字段
2. 后台添加"精选"复选框
3. JSON 生成器包含 `featured` 标记
4. 前端过滤 `featured: true` 的问题

**预计额外时间**：1-2 小时

---

### **UI/UX 设计**

#### **弹窗样式**

```
┌─────────────────────────────────────────────┐
│  ❓ Quick FAQ                          [X]  │
├─────────────────────────────────────────────┤
│                                             │
│  📦 Product Questions                       │
│  ▼ How do I track my order?                │
│     You can track your order by logging... │
│                                             │
│  ▶ What is your return policy?             │
│  ▶ How long does shipping take?            │
│                                             │
│  � Shipping & Delivery                     │
│  ▶ Do you ship internationally?            │
│  ▶ What are the shipping costs?            │
│                                             │
│  💳 Payment Methods                         │
│  ▶ What payment methods do you accept?     │
│                                             │
├─────────────────────────────────────────────┤
│  [查看完整 FAQ →]                           │
└─────────────────────────────────────────────┘
```

#### **样式要点**

- 背景：`bg-black`
- 边框：`border-[#6b73ff]`
- 文字：`text-white`
- 最大高度：`max-h-[80vh]`
- 滚动：`overflow-y-auto`
- 动画：`fade` 过渡效果

---

### **性能优化**

#### **1. 按需加载**

```javascript
// 使用 lazy: true，弹窗打开时才加载
const { data } = await useFetch(url, { lazy: true })
```

#### **2. 数据缓存**

```javascript
// 缓存 1 小时
const { data } = await useFetch(url, {
  key: `faq-quick-${locale.value}`,
  getCachedData: (key) => {
    // 缓存逻辑
  }
})
```

#### **3. 限制显示数量**

```javascript
// 只显示前 10 条，减少渲染负担
.slice(0, 10)
```

---

### **开发检查清单**

- [ ] SiteHeader 添加 FAQ 按钮（桌面端）
- [ ] SiteHeader 添加 FAQ 按钮（移动端）
- [ ] 创建 FAQQuickModal 组件
- [ ] 实现数据加载逻辑
- [ ] 实现折叠/展开交互
- [ ] 添加"查看完整 FAQ"按钮
- [ ] 测试多语言切换
- [ ] 测试响应式布局
- [ ] 优化动画效果
- [ ] 测试性能

---

### **预估时间**

| 任务 | 时间 | 难度 |
|------|------|------|
| 添加 FAQ 按钮 | 15 分钟 | ⭐ 简单 |
| 创建弹窗组件 | 1-1.5 小时 | ⭐⭐ 中等 |
| 数据加载逻辑 | 30 分钟 | ⭐⭐ 中等 |
| 跳转功能 | 15 分钟 | ⭐ 简单 |
| 样式优化 | 30 分钟 | ⭐⭐ 中等 |
| **总计** | **2.5-3 小时** | ⭐⭐ 中等 |

---

### **注意事项**

1. **数据依赖**：需要先在 WordPress 后台创建 FAQ 内容并生成 JSON 文件
2. **语言同步**：确保 FAQ 内容已翻译成所有需要的语言
3. **精选选择**：建议先用方案 A 或 B，后续根据需求升级到方案 C
4. **性能监控**：注意 JSON 文件大小，避免加载过慢

---

## �� 部署流程

### 开发环境
1. 在本地 WordPress 开发插件
2. 在本地 Nuxt 开发页面
3. 测试完整流程
4. （可选）开发 FAQ 快捷弹窗

### 生产环境
1. 上传插件到生产服务器
2. 激活插件
3. 创建测试 FAQ
4. 验证 JSON 文件生成
5. 部署 Nuxt 前端
6. 验证完整功能
7. （可选）部署 FAQ 快捷弹窗

---

## 📞 支持与维护

### 日常维护
- 定期备份 WordPress 数据库
- 监控 JSON 文件大小
- 检查错误日志
- 更新内容

### 扩展功能
- 添加搜索统计
- FAQ 点击统计
- 用户反馈功能
- AI 智能推荐

---

## 📚 相关资源

- [WordPress Plugin Handbook](https://developer.wordpress.org/plugins/)
- [Nuxt 3 Documentation](https://nuxt.com/)
- [Nuxt i18n Module](https://i18n.nuxtjs.org/)
- [Schema.org FAQ](https://schema.org/FAQPage)

---

## 🎯 下一步行动

1. **立即开始**：创建 WordPress 插件基础文件
2. **第一个里程碑**：实现单语言 FAQ 管理
3. **第二个里程碑**：扩展到 34 种语言
4. **第三个里程碑**：Nuxt 前端集成
5. **最终目标**：完整系统上线

---

## 📅 预估时间

| 阶段 | 时间 | 说明 |
|------|------|------|
| WordPress 插件基础 | 2-3 小时 | 数据库、基础 UI |
| 多语言管理界面 | 3-4 小时 | 34 语言编辑器 |
| JSON 生成器 | 1-2 小时 | 文件生成逻辑 |
| Nuxt 前端页面 | 2-3 小时 | 页面 + SEO |
| 测试与优化 | 2-3 小时 | 完整测试 |
| **总计** | **10-15 小时** | 完整实现 |

---

**准备好开始了吗？我们从创建 WordPress 插件主文件开始！** 🚀
