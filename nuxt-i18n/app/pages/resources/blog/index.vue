<template>
  <div>
    <header>
      <h1 class="text-2xl font-semibold tz-text-primary">
        {{ t('blog.pages.all.title') }}
      </h1>
      <p class="mt-2 text-sm tz-text-secondary">
        {{ t('blog.pages.all.intro') }}
      </p>
    </header>

    <nav v-if="categories.length" class="mt-6 flex flex-wrap gap-2" aria-label="Blog categories">
      <button
        type="button"
        class="rounded-full border px-3 py-1.5 text-sm transition"
        :class="selectedCategory ? 'tz-border-strong tz-text-secondary' : 'border-emerald-600 bg-emerald-600 text-white'"
        :aria-pressed="!selectedCategory"
        @click="selectedCategory = ''"
      >
        {{ t('blog.nav.all') }}
      </button>
      <button
        v-for="category in categories"
        :key="category.slug"
        type="button"
        class="rounded-full border px-3 py-1.5 text-sm transition"
        :class="selectedCategory === category.slug
          ? 'border-emerald-600 bg-emerald-600 text-white'
          : 'tz-border-strong tz-text-secondary'"
        :aria-pressed="selectedCategory === category.slug"
        @click="selectedCategory = category.slug"
      >
        {{ category.name }}
      </button>
    </nav>

    <section class="mt-6 grid gap-4 md:grid-cols-2 lg:gap-5">
      <NuxtLink
        v-for="post in posts"
        :key="post.id"
        :to="localePath(buildBlogPath(post.slug))"
        class="group rounded-2xl border tz-border-strong tz-surface-card p-4 shadow-[0_10px_26px_-14px_rgb(15_23_42_/_0.16)] transition-all duration-200 hover:-translate-y-[1px] hover:tz-surface-subtle hover:shadow-[0_14px_32px_-16px_rgb(15_23_42_/_0.2)]"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0 flex-1">
            <h2 class="text-sm font-semibold tz-text-primary sm:text-base">
              {{ post.title }}
            </h2>
            <p class="mt-2 text-xs tz-text-secondary sm:text-sm">
              {{ post.excerpt }}
            </p>
            <div v-if="post.categories.length" class="mt-3 flex flex-wrap gap-1.5">
              <span
                v-for="category in post.categories"
                :key="category.slug"
                class="rounded-full tz-surface-subtle px-2 py-0.5 text-[11px] tz-text-secondary"
              >
                {{ category.name }}
              </span>
            </div>
          </div>
          <span class="shrink-0 rounded-full tz-surface-subtle px-3 py-1 tz-caption font-medium tz-text-secondary">
            {{ formatDate(post.date) }}
          </span>
        </div>

        <div class="mt-4 inline-flex items-center tz-caption font-medium text-emerald-600 group-hover:text-emerald-700">
          {{ t('blog.actions.openArticle') }}
        </div>
      </NuxtLink>
    </section>

    <div v-if="canLoadMore" class="mt-6 flex justify-center">
      <button
        type="button"
        class="rounded-full bg-[var(--tz-action-primary)] px-5 py-2 text-sm font-semibold text-white shadow-[0_10px_22px_-14px_rgb(15_23_42_/_0.28)] transition hover:bg-[var(--tz-action-primary-hover)]"
        @click="loadMore"
      >
        {{ t('blog.actions.readMore') }}
      </button>
    </div>

    <p v-else-if="posts.length === 0" class="mt-6 text-sm tz-text-secondary">
      {{ t('blog.empty') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, watchEffect } from 'vue'
import { useAsyncData, useHead, useI18n, useLocalePath, useState } from '#imports'
import { useBlogApi } from '~/composables/useBlogApi'
import { useBlogListingSeo } from '~/composables/seo/useBlogListingSeo'
import { buildBlogPath } from '~/utils/seo/blog'
import type { BlogCategory, BlogPostSummary } from '~/utils/blog/types'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'footer.links.blog',
  footerLabelFallback: 'Blog',
})

useState('alternateLinksOverride').value = null

const { t, locale } = useI18n()
const localePath = useLocalePath()
const blogApi = useBlogApi()
const lang = computed(() => String(locale.value || 'en'))

const PER_PAGE = 5
const selectedCategory = ref('')
const page = ref(1)
const total = ref(0)
const posts = ref<BlogPostSummary[]>([])
const categories = ref<BlogCategory[]>([])
const loadingMore = ref(false)

const { data: categoryResponse } = await useAsyncData(
  `blog-categories-${lang.value}`,
  () => blogApi.listCategories({ lang: lang.value }),
  { watch: [lang] },
)

const { data: initialResponse, refresh: refreshPosts } = await useAsyncData(
  `blog-posts-all-${lang.value}`,
  () => blogApi.listPosts({
    lang: lang.value,
    category: selectedCategory.value || undefined,
    page: 1,
    perPage: PER_PAGE,
  }),
  { watch: [lang] },
)

watchEffect(() => {
  categories.value = categoryResponse.value || []
  if (!initialResponse.value) return
  posts.value = initialResponse.value.items || []
  page.value = initialResponse.value.page || 1
  total.value = initialResponse.value.total || 0
})

watch(lang, () => {
  selectedCategory.value = ''
})

watch(selectedCategory, async () => {
  page.value = 1
  await refreshPosts()
})

const canLoadMore = computed(() => posts.value.length < total.value)

const loadMore = async () => {
  if (!canLoadMore.value || loadingMore.value) return
  loadingMore.value = true
  try {
    const next = await blogApi.listPosts({
      lang: lang.value,
      category: selectedCategory.value || undefined,
      page: page.value + 1,
      perPage: PER_PAGE,
    })
    posts.value = [...posts.value, ...next.items]
    page.value = next.page
    total.value = next.total
  } finally {
    loadingMore.value = false
  }
}

const formatDate = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString()
}

useBlogListingSeo({
  title: computed(() => t('blog.pages.all.metaTitle')),
  description: computed(() => t('blog.pages.all.intro')),
  posts,
})

useHead(() => ({
  title: t('blog.pages.all.metaTitle'),
}))
</script>
