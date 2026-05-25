# Nuxt 3 前端集成指南

本目录包含了将博客多语言功能集成到 Nuxt 3 前端的完整示例代码。

---

## 📁 文件说明

| 文件 | 说明 |
|------|------|
| `useI18n.ts` | i18n 组合式函数，提供语言管理功能 |
| `LanguageSwitcher.vue` | 语言切换组件（3种样式） |
| `PostTranslations.vue` | 文章翻译链接组件 |
| `blog-post-page.vue` | 博客文章页面完整示例 |
| `nuxt.config.example.ts` | Nuxt 配置示例 |

---

## 🚀 快速开始

### 1. 复制文件到 Nuxt 项目

```bash
# 复制组合式函数
cp useI18n.ts your-nuxt-project/composables/

# 复制组件
cp LanguageSwitcher.vue your-nuxt-project/components/
cp PostTranslations.vue your-nuxt-project/components/

# 复制页面示例（可选）
cp blog-post-page.vue your-nuxt-project/pages/blog/[slug].vue
```

### 2. 配置环境变量

创建 `.env` 文件：

```env
NUXT_PUBLIC_API_BASE=http://localhost:9000
NUXT_PUBLIC_SITE_URL=https://tanzanite.site
```

### 3. 更新 Nuxt 配置

将 `nuxt.config.example.ts` 中的配置合并到你的 `nuxt.config.ts`。

### 4. 使用组件

```vue
<template>
  <div>
    <!-- 语言切换器 -->
    <LanguageSwitcher display-mode="dropdown" />
    
    <!-- 文章翻译链接 -->
    <PostTranslations :post-id="123" />
  </div>
</template>
```

---

## 🎨 组件使用

### LanguageSwitcher 组件

语言切换组件支持 3 种显示模式：

#### 1. Select 模式（下拉选择器）

```vue
<LanguageSwitcher display-mode="select" />
```

#### 2. Buttons 模式（按钮列表）

```vue
<LanguageSwitcher display-mode="buttons" />
```

#### 3. Dropdown 模式（下拉菜单）

```vue
<LanguageSwitcher display-mode="dropdown" />
```

**Props**:
- `displayMode`: `'select' | 'buttons' | 'dropdown'` - 显示模式（默认: `'select'`）
- `showAllLanguages`: `boolean` - 是否显示所有语言（默认: `false`，只显示启用的语言）

---

### PostTranslations 组件

显示文章的所有翻译版本链接。

```vue
<PostTranslations 
  :post-id="123" 
  title="Read this article in other languages"
  :show-current-locale="true"
/>
```

**Props**:
- `postId`: `number` - 文章ID（必需）
- `title`: `string` - 标题文本（可选）
- `showCurrentLocale`: `boolean` - 是否显示当前语言（默认: `true`）

---

## 🔧 useI18n 组合式函数

### 基本用法

```vue
<script setup>
const { 
  locale,              // 当前语言
  getLanguages,        // 获取语言列表
  setLanguage,         // 设置语言
  switchLanguage,      // 切换语言并刷新
  getPostTranslations, // 获取文章翻译
  localizeUrl          // 本地化 URL
} = useI18n()
</script>
```

### API 方法

#### getLanguages()

获取支持的语言列表。

```typescript
const languages = await getLanguages()
// 返回: Language[]
```

#### getPostTranslations(postId)

获取文章的所有翻译版本。

```typescript
const translations = await getPostTranslations(123)
// 返回: Record<string, PostTranslation>
```

#### setLanguage(locale)

设置用户语言偏好（保存到 Cookie）。

```typescript
const success = await setLanguage('zh')
// 返回: boolean
```

#### switchLanguage(locale)

切换语言并刷新页面。

```typescript
await switchLanguage('zh')
// 自动导航到对应语言的 URL 并刷新
```

#### detectLanguage()

检测用户语言偏好。

```typescript
const detectedLocale = await detectLanguage()
// 返回: string
```

#### localizeUrl(path, locale?)

构建本地化 URL。

```typescript
const url = localizeUrl('/blog/post', 'zh')
// 返回: '/zh/blog/post'

const url = localizeUrl('/blog/post', 'en')
// 返回: '/blog/post' (英文不加前缀)
```

---

## 📄 页面集成示例

### 博客文章页面

```vue
<template>
  <div>
    <article>
      <h1>{{ post?.title }}</h1>
      <div v-html="post?.content"></div>
    </article>
    
    <!-- 显示翻译链接 -->
    <PostTranslations :post-id="post.id" />
  </div>
</template>

<script setup>
const route = useRoute()
const { locale } = useI18n()

// 获取文章数据
const { data: post } = await useFetch(
  `/api/v1/content/posts/${route.params.slug}`,
  {
    headers: {
      'Accept-Language': locale.value
    }
  }
)

// 设置 SEO 元标签
useHead({
  title: post.value?.title,
  meta: [
    { name: 'description', content: post.value?.excerpt }
  ]
})
</script>
```

### 博客列表页面

```vue
<template>
  <div>
    <h1>Blog</h1>
    
    <div v-for="post in posts" :key="post.id">
      <NuxtLink :to="localizeUrl(`/blog/${post.slug}`)">
        <h2>{{ post.title }}</h2>
        <p>{{ post.excerpt }}</p>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup>
const { locale, localizeUrl } = useI18n()

// 获取文章列表
const { data: posts } = await useFetch('/api/v1/content/posts', {
  query: {
    locale: locale.value,
    status: 'published',
    page: 1,
    page_size: 10
  }
})
</script>
```

---

## 🌐 SEO 优化

### Hreflang 标签

在文章页面自动添加 Hreflang 标签：

```vue
<script setup>
const { getPostTranslations } = useI18n()
const translations = await getPostTranslations(post.value.id)

useHead({
  link: [
    // Canonical
    { rel: 'canonical', href: post.value.canonical_url },
    
    // Hreflang
    ...Object.entries(translations).map(([locale, trans]) => ({
      rel: 'alternate',
      hreflang: locale,
      href: `https://tanzanite.site${trans.url}`
    })),
    
    // x-default
    { 
      rel: 'alternate', 
      hreflang: 'x-default', 
      href: translations.en?.url 
    }
  ]
})
</script>
```

### 结构化数据（JSON-LD）

```vue
<script setup>
useHead({
  script: [
    {
      type: 'application/ld+json',
      children: JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'BlogPosting',
        headline: post.value.title,
        inLanguage: locale.value,
        translationOfWork: Object.entries(translations)
          .map(([loc, trans]) => ({
            '@type': 'BlogPosting',
            url: `https://tanzanite.site${trans.url}`,
            inLanguage: loc
          }))
      })
    }
  ]
})
</script>
```

---

## 🎨 样式定制

所有组件都使用 scoped CSS，可以通过以下方式定制样式：

### 1. 覆盖 CSS 变量

```css
/* 在全局样式中 */
:root {
  --language-switcher-bg: #ffffff;
  --language-switcher-border: #e5e7eb;
  --language-switcher-hover: #3b82f6;
}
```

### 2. 使用深度选择器

```vue
<style>
.my-component :deep(.language-select) {
  border-color: red;
}
</style>
```

### 3. 完全自定义

复制组件代码并修改样式部分。

---

## 🔍 常见问题

### Q: 如何在服务端渲染（SSR）中使用？

A: 所有组件和组合式函数都支持 SSR。确保在 `onMounted` 钩子中调用客户端专用的代码。

### Q: 如何处理语言切换时的页面刷新？

A: `switchLanguage()` 方法会自动处理 URL 导航和页面刷新。如果不想刷新，使用 `setLanguage()` 并手动重新加载数据。

### Q: 如何添加新语言？

A: 在 Go 后端的 `internal/api/v1/i18n/handler.go` 中添加新语言到 `SupportedLanguages` 数组。

### Q: 如何自定义 URL 结构？

A: 修改 `useI18n.ts` 中的 `localizeUrl()` 函数。

### Q: 如何处理 404 页面的多语言？

A: 创建 `error.vue` 页面并使用 `useI18n()` 获取当前语言。

---

## 📚 相关文档

- [Go 后端 API 文档](../../API.md)
- [博客多语言迁移指南](../../BLOG_I18N_MIGRATION_GUIDE.md)
- [快速参考](../../I18N_QUICK_REFERENCE.md)

---

## 🤝 贡献

如果你有改进建议或发现问题，欢迎提交 Issue 或 Pull Request。

---

**版本**: 1.0.0  
**更新日期**: 2026-05-25
