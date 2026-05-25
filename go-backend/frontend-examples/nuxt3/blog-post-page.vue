<template>
  <div class="blog-post-page">
    <!-- 文章内容 -->
    <article class="post-article">
      <header class="post-header">
        <h1 class="post-title">{{ post?.title }}</h1>
        <div class="post-meta">
          <time :datetime="post?.published_at">
            {{ formatDate(post?.published_at) }}
          </time>
          <span class="post-locale">{{ currentLanguageName }}</span>
        </div>
      </header>

      <div class="post-content" v-html="post?.content"></div>
    </article>

    <!-- 翻译链接 -->
    <PostTranslations 
      v-if="post?.id" 
      :post-id="post.id" 
      title="Read this article in other languages"
    />
  </div>
</template>

<script setup lang="ts">
/**
 * 博客文章页面示例
 * 
 * 路由: /blog/[slug].vue 或 /[locale]/blog/[slug].vue
 * 
 * 功能:
 * - 显示文章内容
 * - 显示翻译链接
 * - 自动设置 SEO 元标签（包括 Hreflang）
 */

const route = useRoute()
const config = useRuntimeConfig()
const apiBase = config.public.apiBase || 'http://localhost:9000'

const { locale, getPostTranslations, getLanguageName } = useI18n()

// 获取文章数据
const { data: post } = await useFetch(`${apiBase}/api/v1/content/posts/${route.params.slug}`, {
  headers: {
    'Accept-Language': locale.value
  }
})

// 获取翻译版本
const translations = ref<Record<string, any>>({})
if (post.value?.id) {
  translations.value = await getPostTranslations(post.value.id)
}

// 获取当前语言名称
const currentLanguageName = ref('')
onMounted(async () => {
  currentLanguageName.value = await getLanguageName(locale.value)
})

// 格式化日期
const formatDate = (dateString?: string) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleDateString(locale.value, {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

// 设置 SEO 元标签
useHead({
  title: post.value?.meta_title || post.value?.title,
  meta: [
    {
      name: 'description',
      content: post.value?.meta_description || post.value?.excerpt
    },
    {
      name: 'keywords',
      content: post.value?.meta_keywords
    },
    // Open Graph
    {
      property: 'og:title',
      content: post.value?.title
    },
    {
      property: 'og:description',
      content: post.value?.excerpt
    },
    {
      property: 'og:image',
      content: post.value?.featured_image
    },
    {
      property: 'og:type',
      content: 'article'
    },
    {
      property: 'og:locale',
      content: locale.value
    },
    // Twitter Card
    {
      name: 'twitter:card',
      content: 'summary_large_image'
    },
    {
      name: 'twitter:title',
      content: post.value?.title
    },
    {
      name: 'twitter:description',
      content: post.value?.excerpt
    },
    {
      name: 'twitter:image',
      content: post.value?.featured_image
    }
  ],
  link: [
    // Canonical URL
    {
      rel: 'canonical',
      href: post.value?.canonical_url || `${config.public.siteUrl}${route.path}`
    },
    // Hreflang 标签
    ...Object.entries(translations.value).map(([locale, trans]: [string, any]) => ({
      rel: 'alternate',
      hreflang: locale,
      href: `${config.public.siteUrl}${trans.url}`
    })),
    // x-default (通常指向英文版本)
    {
      rel: 'alternate',
      hreflang: 'x-default',
      href: translations.value.en?.url 
        ? `${config.public.siteUrl}${translations.value.en.url}`
        : post.value?.canonical_url || `${config.public.siteUrl}${route.path}`
    }
  ],
  script: [
    // JSON-LD 结构化数据
    {
      type: 'application/ld+json',
      children: JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'BlogPosting',
        headline: post.value?.title,
        description: post.value?.excerpt,
        image: post.value?.featured_image,
        datePublished: post.value?.published_at,
        dateModified: post.value?.updated_at,
        author: {
          '@type': 'Person',
          name: 'Tanzanite'
        },
        publisher: {
          '@type': 'Organization',
          name: 'Tanzanite',
          logo: {
            '@type': 'ImageObject',
            url: `${config.public.siteUrl}/logo.png`
          }
        },
        inLanguage: locale.value,
        // 添加翻译版本
        ...(Object.keys(translations.value).length > 1 && {
          translationOfWork: Object.entries(translations.value)
            .filter(([loc]) => loc !== locale.value)
            .map(([loc, trans]: [string, any]) => ({
              '@type': 'BlogPosting',
              url: `${config.public.siteUrl}${trans.url}`,
              inLanguage: loc
            }))
        })
      })
    }
  ]
})
</script>

<style scoped>
.blog-post-page {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.post-article {
  margin-bottom: 3rem;
}

.post-header {
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid #e5e7eb;
}

.post-title {
  font-size: 2.5rem;
  font-weight: 700;
  line-height: 1.2;
  margin: 0 0 1rem 0;
  color: #111827;
}

.post-meta {
  display: flex;
  gap: 1rem;
  font-size: 0.875rem;
  color: #6b7280;
}

.post-locale {
  padding: 0.25rem 0.75rem;
  background-color: #eff6ff;
  color: #3b82f6;
  border-radius: 0.25rem;
  font-weight: 500;
}

.post-content {
  font-size: 1.125rem;
  line-height: 1.75;
  color: #374151;
}

.post-content :deep(h2) {
  font-size: 1.875rem;
  font-weight: 600;
  margin: 2rem 0 1rem 0;
  color: #111827;
}

.post-content :deep(h3) {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 1.5rem 0 0.75rem 0;
  color: #111827;
}

.post-content :deep(p) {
  margin: 1rem 0;
}

.post-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
  margin: 1.5rem 0;
}

.post-content :deep(a) {
  color: #3b82f6;
  text-decoration: underline;
}

.post-content :deep(a:hover) {
  color: #2563eb;
}

/* 响应式 */
@media (max-width: 640px) {
  .post-title {
    font-size: 1.875rem;
  }
  
  .post-content {
    font-size: 1rem;
  }
}
</style>
