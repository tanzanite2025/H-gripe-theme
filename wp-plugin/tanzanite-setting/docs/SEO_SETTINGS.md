# SEO Settings - SEO 设置与优化

**页面路径**: `admin.php?page=tanzanite-settings-seo`  
**权限要求**: `manage_options`  
**REST API**: `/wp-json/mytheme/v1/seo/*`

---

## 📋 功能概述

SEO Settings 页面集成了 MyTheme SEO Bridge 功能，提供完整的多语言 SEO 元数据管理、结构化数据（Schema.org）、站点地图生成和 404 监控功能。

---

## ✨ 主要功能

### 1. 多语言 SEO

**功能**:
- 多语言配置管理
- 与 Nuxt i18n 集成
- 每种语言独立 SEO 设置
- Hreflang 标签自动生成

**配置方式**:
- 语言列表管理
- 默认语言设置
- 从 Nuxt i18n 导入语言

---

### 2. 元数据管理

**支持类型**:
- 文章 SEO（标题、描述、关键词）
- 页面 SEO（自定义页面元数据）
- 商品 SEO（WooCommerce 商品）
- 分类 SEO（分类和标签）
- 首页 SEO（首页专属设置）

**字段**:
- Meta Title
- Meta Description
- Meta Keywords
- Open Graph 标签
- Twitter Card 标签

---

### 3. 结构化数据（Schema.org）

**Product Schema**:
- 商品名称、品牌、SKU
- 价格和货币
- 库存状态
- 商品图片
- 评分和评论
- GTIN、MPN 支持

**配置选项**:
- 启用/禁用 Product Schema
- 全局品牌设置
- 多语言品牌配置
- 价格来源选择（常规价/促销价）
- 自定义字段映射

---

### 4. 站点地图

**功能**:
- 自动生成 XML 站点地图
- 多语言站点地图支持
- 实时更新
- 外部站点地图集成

**配置**:
- 启用/禁用站点地图
- 按语言分割
- 包含的内容类型
- 更新时自动 Ping 搜索引擎

---

### 5. 404 监控

**功能**:
- 记录所有 404 错误
- 访问统计
- 重定向管理
- 日志清理

**数据**:
- 404 URL
- 访问次数
- 最后访问时间
- 来源页面

---

### 6. URL 管理集成

**功能**:
- 集成 URLLink 功能
- 统一入口管理 URL
- URL 目录树
- 301 重定向

---

## 🔌 REST API

### 元数据 API

```
GET    /mytheme/v1/seo/meta/{id}           # 获取文章元数据
POST   /mytheme/v1/seo/meta/{id}           # 更新文章元数据
```

### 首页 SEO

```
GET    /mytheme/v1/seo/homepage            # 获取首页 SEO
POST   /mytheme/v1/seo/homepage            # 更新首页 SEO
```

### 分类 SEO

```
GET    /mytheme/v1/seo/category/{id}       # 获取分类 SEO
POST   /mytheme/v1/seo/category/{id}       # 更新分类 SEO
```

### 404 日志

```
GET    /mytheme/v1/seo/404-logs            # 获取 404 日志
POST   /mytheme/v1/seo/404-logs            # 更新 404 日志
```

### 设置

```
GET    /mytheme/v1/seo/settings            # 获取设置
POST   /mytheme/v1/seo/settings            # 更新设置
GET    /mytheme/v1/seo/settings/public     # 获取公开设置
```

### 语言

```
GET    /mytheme/v1/seo/languages           # 获取语言列表
POST   /mytheme/v1/seo/languages           # 更新语言列表
POST   /mytheme/v1/seo/languages/import    # 从 Nuxt 导入语言
```

### Product Schema

```
GET    /mytheme/v1/seo/schema/product/{id}              # 获取商品 Schema
GET    /mytheme/v1/seo/schema/product/by-slug/{slug}    # 通过 slug 获取
GET    /mytheme/v1/seo/schema/product/resolve           # 解析商品 Schema
```

---

## 💻 前端集成

### Nuxt.js 示例

```javascript
// composables/useSEO.js
export const useSEO = () => {
  const { $wpApi } = useNuxtApp()

  const fetchPostSEO = async (postId, locale) => {
    const response = await $wpApi(`/seo/meta/${postId}`, {
      params: { lang: locale }
    })
    return response.data
  }

  const fetchProductSchema = async (productId) => {
    const response = await $wpApi(`/seo/schema/product/${productId}`)
    return response.data.schema
  }

  return {
    fetchPostSEO,
    fetchProductSchema
  }
}
```

### 使用示例

```vue
<script setup>
const route = useRoute()
const { locale } = useI18n()
const { fetchPostSEO, fetchProductSchema } = useSEO()

// 获取 SEO 数据
const seoData = await fetchPostSEO(route.params.id, locale.value)

// 设置 Head
useHead({
  title: seoData.title,
  meta: [
    { name: 'description', content: seoData.description },
    { name: 'keywords', content: seoData.keywords },
    { property: 'og:title', content: seoData.title },
    { property: 'og:description', content: seoData.description }
  ]
})

// 获取 Product Schema
const schema = await fetchProductSchema(route.params.id)
useHead({
  script: [
    {
      type: 'application/ld+json',
      children: JSON.stringify(schema)
    }
  ]
})
</script>
```

---

## 🎯 使用场景

### 1. 多语言电商网站
- 每种语言独立 SEO 优化
- 商品结构化数据
- 多语言站点地图

### 2. 内容营销网站
- 文章 SEO 优化
- 分类页面 SEO
- 404 监控和重定向

### 3. 企业官网
- 首页 SEO
- 组织 Schema
- 面包屑导航

---

## 📝 注意事项

### 1. 性能优化
- 启用对象缓存
- 定期清理 404 日志
- 优化站点地图生成

### 2. 兼容性
- 需要 WooCommerce（商品功能）
- 需要 Nuxt i18n（多语言）
- 需要 URLLink（URL 管理）

### 3. 安全性
- API 权限验证
- 数据清理和验证
- XSS 防护

---

## 🔍 故障排除

### Q: 语言列表不显示？
**A**: 检查 Nuxt i18n 配置文件路径是否正确。

### Q: Schema 不生效？
**A**: 确认在设置中已启用 Product Schema。

### Q: 404 日志过多？
**A**: 定期清理日志，或调整日志记录规则。

### Q: 站点地图不更新？
**A**: 检查站点地图设置，确认已启用自动更新。

---

**最后更新**: 2025-11-11  
**维护者**: Tanzanite Team
