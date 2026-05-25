# 博客多语言功能快速参考

## 🚀 API 端点速查

### 语言管理
```bash
# 获取支持的语言列表（34种）
GET /api/v1/i18n/languages

# 检测用户语言偏好
GET /api/v1/i18n/detect

# 设置用户语言偏好
POST /api/v1/i18n/set-language
Body: {"locale": "zh"}
```

### 翻译查询
```bash
# 获取文章的所有翻译版本
GET /api/v1/i18n/translations/:post_id
```

### Sitemap
```bash
# Sitemap 索引
GET /sitemap.xml

# Hreflang Sitemap（包含所有语言版本）
GET /sitemap-hreflang.xml

# 单语言 Sitemap
GET /sitemap-en.xml
GET /sitemap-zh.xml
GET /sitemap-fr.xml
```

---

## 📦 数据模型

### Post 模型新增字段
```go
type Post struct {
    // 翻译关联
    TranslationGroupID *uint  `json:"translation_group_id"`
    
    // SEO 元数据
    MetaKeywords    string `json:"meta_keywords"`
    CanonicalURL    string `json:"canonical_url"`
    
    // 关联（不存储）
    Translations []Post `json:"translations,omitempty"`
}
```

---

## 🔧 配置

### config.yaml
```yaml
server:
  base_url: "https://tanzanite.site"  # 用于生成 Sitemap
```

---

## 🌍 支持的语言（34种）

### 欧洲语言
en, fr, de, es, it, pt, ru, nl, pl, tr, sv, no, da, fi, cs, hu, ro

### 亚洲语言
zh, zh-TW, ja, ko, vi, th, id, ms, hi, bn, ta, te, mr, ur

### 中东语言
ar, fa, he

---

## 📝 使用示例

### 前端集成（Nuxt 3）

#### 1. 语言切换组件
```vue
<template>
  <select v-model="locale" @change="setLanguage">
    <option v-for="lang in languages" :value="lang.code">
      {{ lang.native_name }}
    </option>
  </select>
</template>

<script setup>
const locale = useCookie('locale')
const languages = await $fetch('/api/v1/i18n/languages')

const setLanguage = async () => {
  await $fetch('/api/v1/i18n/set-language', {
    method: 'POST',
    body: { locale: locale.value }
  })
  window.location.reload()
}
</script>
```

#### 2. 文章翻译链接
```vue
<template>
  <div v-if="translations.length > 1">
    <h3>Available in:</h3>
    <a v-for="trans in translations" :href="trans.url">
      {{ trans.locale }}
    </a>
  </div>
</template>

<script setup>
const props = defineProps(['postId'])
const { translations } = await $fetch(`/api/v1/i18n/translations/${props.postId}`)
</script>
```

#### 3. SEO 元标签（Hreflang）
```vue
<script setup>
const post = await $fetch('/api/v1/content/posts/123')
const translations = await $fetch(`/api/v1/i18n/translations/${post.id}`)

useHead({
  link: [
    // Canonical
    { rel: 'canonical', href: post.canonical_url },
    
    // Hreflang
    ...Object.entries(translations).map(([locale, trans]) => ({
      rel: 'alternate',
      hreflang: locale,
      href: `https://tanzanite.site${trans.url}`
    })),
    
    // x-default
    { rel: 'alternate', hreflang: 'x-default', href: translations.en?.url }
  ]
})
</script>
```

---

## 🗄️ 数据库

### 添加翻译组ID列
```sql
ALTER TABLE posts ADD COLUMN translation_group_id INTEGER;
CREATE INDEX idx_posts_translation_group_id ON posts(translation_group_id);
```

### 添加 SEO 字段
```sql
ALTER TABLE posts ADD COLUMN meta_keywords VARCHAR(255);
ALTER TABLE posts ADD COLUMN canonical_url VARCHAR(500);
```

---

## 🧪 测试命令

### 快速测试
```bash
# Linux/Mac
./test-i18n-api.sh

# Windows
.\test-i18n-api.ps1
```

### 手动测试
```bash
# 语言列表
curl http://localhost:9000/api/v1/i18n/languages | jq

# 文章翻译
curl http://localhost:9000/api/v1/i18n/translations/1 | jq

# Sitemap
curl http://localhost:9000/sitemap-hreflang.xml
```

---

## 📚 相关文档

- **完整迁移指南**: `BLOG_I18N_MIGRATION_GUIDE.md`
- **实现完成报告**: `BLOG_I18N_IMPLEMENTATION_COMPLETE.md`
- **API 文档**: `API.md`
- **变更日志**: `CHANGELOG.md`

---

## 🔍 常见问题

### Q: 如何创建翻译文章？
A: 创建文章时，设置相同的 `translation_group_id`，不同的 `locale`。

### Q: 如何查询某篇文章的所有翻译？
A: 使用 `GET /api/v1/i18n/translations/:post_id`。

### Q: Sitemap 多久更新一次？
A: 实时生成，建议添加缓存（如 Redis，TTL 1小时）。

### Q: 如何添加新语言？
A: 在 `internal/api/v1/i18n/handler.go` 的 `SupportedLanguages` 数组中添加。

### Q: URL 结构是什么？
A: 
- 英文: `/blog/post-slug`
- 其他: `/{locale}/blog/post-slug`

---

**版本**: 1.4.0  
**更新日期**: 2026-05-25
