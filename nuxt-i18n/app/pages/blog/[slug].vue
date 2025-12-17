<template>
  <div v-if="post">
    <header class="flex flex-col gap-2">
      <p class="text-xs font-medium uppercase tracking-wide text-slate-400">
        {{ t(postCategory === 'news' ? 'blog.nav.news' : 'blog.nav.wheelsbuild') }}
      </p>
      <h1 class="text-2xl font-semibold text-slate-50">
        {{ post.title }}
      </h1>
      <p class="text-sm text-slate-300">
        {{ formatDate(post.date) }}
      </p>
    </header>

    <article
      class="mt-6 rounded-2xl border border-slate-700/60 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] p-4 text-slate-200"
    >
      <div class="space-y-4" v-html="post.contentHtml"></div>
    </article>
  </div>

  <div v-else class="text-sm text-slate-300">
    {{ t('blog.notFound') }}
  </div>
</template>

<script setup lang="ts">
import { computed, watchEffect } from 'vue'
import { useRoute, useI18n, useHead, useLocalePath, useState } from '#imports'
import {
  getBlogPostBySlug,
  getBlogCategoryFromPost,
  type BlogCategory,
} from '~/utils/blogMock'

definePageMeta({
  layout: 'products',
})

const route = useRoute()
const { t, locale } = useI18n()
const localePath = useLocalePath()

const slug = computed(() => String(route.params.slug || ''))

const post = computed(() =>
  getBlogPostBySlug({ lang: String(locale.value || 'en'), slug: slug.value })
)

const postCategory = computed<BlogCategory | null>(() => {
  if (!post.value) return null
  return getBlogCategoryFromPost(post.value)
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

  const entries = Object.entries(translations).map(([code, entry]) => {
    return {
      code,
      path: localePath(`/blog/${entry.slug}`, code as any),
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
