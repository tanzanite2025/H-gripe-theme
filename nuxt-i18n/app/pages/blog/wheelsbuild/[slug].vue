<template>
  <div v-if="post">
    <header class="flex flex-col gap-2">
      <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
        {{ t('blog.nav.wheelsbuild') }}
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
  </div>

  <div v-else class="text-sm tz-text-secondary">
    {{ t('blog.notFound') }}
  </div>
</template>

<script setup lang="ts">
import { computed, watchEffect } from 'vue'
import { useRoute, useI18n, useHead, useLocalePath, useState, useAsyncData } from '#imports'
import { useBlogApi } from '~/composables/useBlogApi'
import type { BlogPostDetail } from '~/utils/blogMock'

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
  `blog-post-wheelsbuild-${lang.value}-${slug.value}`,
  async () => {
    try {
      return await blogApi.getPost({ lang: lang.value, slug: slug.value })
    } catch {
      return null
    }
  }
)

const post = computed(() => {
  const resolved = (postData.value || null) as BlogPostDetail | null
  const categories = resolved && Array.isArray(resolved.categories) ? resolved.categories : []
  return categories.includes('wheelsbuild') ? resolved : null
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

  alternateLinksOverride.value = Object.entries(translations).map(([code, entry]) => {
    return {
      code,
      path: localePath(`/blog/wheelsbuild/${entry.slug}`, code as any),
    }
  })
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
