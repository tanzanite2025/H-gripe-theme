<template>
  <div v-if="post">
    <header class="flex flex-col gap-2">
      <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
        {{
          t(
            postCategory === 'news'
              ? 'blog.nav.news'
              : postCategory === 'wheelsbuild'
                ? 'blog.nav.wheelsbuild'
                : 'blog.nav.all'
          )
        }}
      </p>
      <h1 class="text-2xl font-semibold tz-text-primary">
        {{ post.title }}
      </h1>
      <p class="text-sm tz-text-muted">
        {{ formatDate(post.date) }}
      </p>
    </header>

    <article
      class="mt-6 rounded-2xl border border-slate-700/60 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] p-4 tz-text-secondary"
    >
      <div class="space-y-4" v-html="post.contentHtml"></div>
    </article>

    <!-- 文章翻译链接 -->
    <PostTranslations 
      v-if="post.id" 
      :post-id="post.id" 
      :title="t('blog.availableInOtherLanguages')"
      class="mt-6"
    />
  </div>

  <div v-else class="text-sm tz-text-secondary">
    {{ t('blog.notFound') }}
  </div>
</template>

<script setup lang="ts">
import { computed, watchEffect } from 'vue'
import { useRoute, useI18n, useHead, useLocalePath, useState, useAsyncData } from '#imports'
import { useBlogApi } from '~/composables/useBlogApi'
import type { BlogCategory, BlogPostDetail } from '~/utils/blogMock'

definePageMeta({
  layout: 'products',
})

const route = useRoute()
const { t, locale } = useI18n()
const localePath = useLocalePath()

const blogApi = useBlogApi()

const slug = computed(() => String(route.params.slug || ''))

const lang = computed(() => String(locale.value || 'en'))

const { data: postData } = await useAsyncData(
  `blog-post-${lang.value}-${slug.value}`,
  async () => {
    try {
      return await blogApi.getPost({ lang: lang.value, slug: slug.value })
    } catch {
      return null
    }
  }
)

const post = computed(() => (postData.value || null) as BlogPostDetail | null)

const postCategory = computed<BlogCategory | null>(() => {
  if (!post.value) return null

  const categories = Array.isArray(post.value.categories) ? post.value.categories : []
  if (categories.includes('news')) return 'news'
  if (categories.includes('wheelsbuild')) return 'wheelsbuild'
  return null
})

const alternateLinksOverride = useState<{ code: string; path: string }[] | null>(
  'alternateLinksOverride'
)

watchEffect(() => {
  if (!post.value) {
    alternateLinksOverride.value = null
    return
  }

  const translations = post.value.translations as Record<string, { id: number; slug: string }>

  const category = postCategory.value
  const prefix = category ? `/blog/${category}` : '/blog'

  const entries = Object.entries(translations).map(([code, entry]) => {
    return {
      code,
      path: localePath(`${prefix}/${entry.slug}`, code as any),
    }
  })

  alternateLinksOverride.value = entries
})

useHead(() => {
  const title = post.value?.title || t('blog.pages.detail.metaTitle')
  const description = post.value?.excerpt || t('blog.pages.detail.metaDescription')

  return {
    title,
    meta: [
      { name: 'description', content: description, key: 'description' },
      { property: 'og:type', content: 'article', key: 'og:type' },
      { property: 'og:title', content: title, key: 'og:title' },
      { property: 'og:description', content: description, key: 'og:description' },
    ],
  }
})

const formatDate = (value: string) => {
  try {
    return new Date(value).toLocaleDateString()
  } catch {
    return value
  }
}
</script>
