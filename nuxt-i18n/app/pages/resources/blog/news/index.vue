<template>
  <div>
    <header>
      <h1 class="text-2xl font-semibold tz-text-primary">
        {{ t('blog.pages.news.title') }}
      </h1>
      <p class="mt-2 text-sm tz-text-secondary">
        {{ t('blog.pages.news.intro') }}
      </p>
    </header>

    <section class="mt-6 grid gap-4 md:grid-cols-2 lg:gap-5">
      <NuxtLink
        v-for="post in visiblePosts"
        :key="post.id"
        :to="localePath(`/resources/blog/news/${post.slug}`)"
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

    <p v-else-if="visiblePosts.length === 0" class="mt-6 text-sm tz-text-secondary">
      {{ t('blog.empty') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue'
import { useI18n, useLocalePath, useHead, useState, useAsyncData } from '#imports'
import { useBlogApi } from '~/composables/useBlogApi'
import { useBlogListingSeo } from '~/composables/seo/useBlogListingSeo'
import type { BlogPostSummary } from '~/utils/blogMock'

definePageMeta({
  layout: 'products',
})

useState('alternateLinksOverride').value = null

const { t, locale } = useI18n()
const localePath = useLocalePath()

const blogApi = useBlogApi()
const lang = computed(() => String(locale.value || 'en'))

const PER_PAGE = 5

const page = ref(1)
const total = ref(0)
const posts = ref<BlogPostSummary[]>([])
const loadingMore = ref(false)

const { data: initialResponse } = await useAsyncData(
  `blog-posts-news-${lang.value}`,
  () => blogApi.listPosts({ lang: lang.value, category: 'news', page: 1, perPage: PER_PAGE })
)

watchEffect(() => {
  if (!initialResponse.value?.items) throw new Error("[CRITICAL] items missing")
  posts.value = initialResponse.value.items as BlogPostSummary[]
  page.value = initialResponse.value?.page || 1
  total.value = initialResponse.value?.total || 0
})

const visiblePosts = computed(() => posts.value)
const canLoadMore = computed(() => posts.value.length < total.value)
const pageTitle = computed(() => t('blog.pages.news.title'))
const pageDescription = computed(() => t('blog.pages.news.intro'))

useBlogListingSeo({
  category: 'news',
  title: pageTitle,
  description: pageDescription,
  posts,
})

const loadMore = async () => {
  if (!canLoadMore.value || loadingMore.value) return
  loadingMore.value = true
  try {
    const next = await blogApi.listPosts({
      lang: lang.value,
      category: 'news',
      page: page.value + 1,
      perPage: PER_PAGE,
    })
    if (!next.items) throw new Error("[CRITICAL] next items missing")
    posts.value = [...posts.value, ...(next.items as BlogPostSummary[])]
    page.value = next.page
    total.value = next.total
  } finally {
    loadingMore.value = false
  }
}

const formatDate = (value: string) => {
  try {
    return new Date(value).toLocaleDateString()
  } catch {
    return value
  }
}

useHead(() => ({
  title: t('blog.pages.news.metaTitle'),
}))
</script>

