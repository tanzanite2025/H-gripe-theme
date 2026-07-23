<template>
  <div>
    <header>
      <h1 class="text-2xl font-semibold text-slate-50">
        {{ t('blog.pages.wheelsbuild.title') }}
      </h1>
      <p class="mt-2 text-sm tz-text-secondary">
        {{ t('blog.pages.wheelsbuild.intro') }}
      </p>
    </header>

    <section class="mt-6 grid gap-4 md:grid-cols-2 lg:gap-5">
      <NuxtLink
        v-for="post in visiblePosts"
        :key="post.id"
        :to="localePath(`/blog/wheelsbuild/${post.slug}`)"
        class="group rounded-2xl border border-slate-700/60 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] p-4 shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)] transition-all duration-200 hover:-translate-y-[1px] hover:border-slate-600/70 hover:bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.98),rgba(15,23,42,0.99))] hover:shadow-[0_14px_32px_-16px_rgba(0,0,0,1)]"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0 flex-1">
            <h2 class="text-sm font-semibold text-white sm:text-base">
              {{ post.title }}
            </h2>
            <p class="mt-2 text-xs tz-text-secondary sm:text-sm">
              {{ post.excerpt }}
            </p>
          </div>
          <span class="shrink-0 rounded-full bg-white/10 px-3 py-1 text-[11px] font-medium tz-text-secondary">
            {{ formatDate(post.date) }}
          </span>
        </div>

        <div class="mt-4 inline-flex items-center text-[11px] font-medium text-sky-300 group-hover:text-sky-200">
          {{ t('blog.actions.openArticle') }}
        </div>
      </NuxtLink>
    </section>

    <div v-if="canLoadMore" class="mt-6 flex justify-center">
      <button
        type="button"
        class="rounded-full bg-sky-500/90 px-5 py-2 text-sm font-semibold text-slate-950 shadow-[0_10px_22px_-14px_rgba(56,189,248,0.9)] transition hover:bg-sky-400"
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
  `blog-posts-wheelsbuild-${lang.value}`,
  () => blogApi.listPosts({ lang: lang.value, category: 'wheelsbuild', page: 1, perPage: PER_PAGE })
)

watchEffect(() => {
  if (!initialResponse.value?.items) throw new Error("[CRITICAL] items missing")
  posts.value = initialResponse.value.items as BlogPostSummary[]
  page.value = initialResponse.value?.page || 1
  total.value = initialResponse.value?.total || 0
})

const visiblePosts = computed(() => posts.value)
const canLoadMore = computed(() => posts.value.length < total.value)

const loadMore = async () => {
  if (!canLoadMore.value || loadingMore.value) return
  loadingMore.value = true
  try {
    const next = await blogApi.listPosts({
      lang: lang.value,
      category: 'wheelsbuild',
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

useHead({
  title: t('blog.pages.wheelsbuild.metaTitle'),
})
</script>
