<template>
  <div>
    <header>
      <h1 class="text-2xl font-semibold text-slate-50">
        {{ t('blog.pages.news.title') }}
      </h1>
      <p class="mt-2 text-sm text-slate-300">
        {{ t('blog.pages.news.intro') }}
      </p>
    </header>

    <section class="mt-6 grid gap-4 md:grid-cols-2 lg:gap-5">
      <NuxtLink
        v-for="post in visiblePosts"
        :key="post.id"
        :to="localePath(`/blog/news/${post.slug}`)"
        class="group rounded-2xl border border-slate-700/60 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] p-4 shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)] transition-all duration-200 hover:-translate-y-[1px] hover:border-slate-600/70 hover:bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.98),rgba(15,23,42,0.99))] hover:shadow-[0_14px_32px_-16px_rgba(0,0,0,1)]"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0 flex-1">
            <h2 class="text-sm font-semibold text-white sm:text-base">
              {{ post.title }}
            </h2>
            <p class="mt-2 text-xs text-slate-300 sm:text-sm">
              {{ post.excerpt }}
            </p>
          </div>
          <span class="shrink-0 rounded-full bg-white/10 px-3 py-1 text-[11px] font-medium text-slate-200">
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

    <p v-else-if="allPosts.length === 0" class="mt-6 text-sm text-slate-300">
      {{ t('blog.empty') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n, useLocalePath, useHead, useState } from '#imports'
import { listBlogPosts } from '~/utils/blogMock'

definePageMeta({
  layout: 'products',
})

useState('alternateLinksOverride').value = null

const { t, locale } = useI18n()
const localePath = useLocalePath()

const allPosts = computed(() =>
  listBlogPosts({ lang: String(locale.value || 'en'), category: 'news' })
)

const visibleCount = ref(5)
const visiblePosts = computed(() => allPosts.value.slice(0, visibleCount.value))
const canLoadMore = computed(() => visibleCount.value < allPosts.value.length)

const loadMore = () => {
  visibleCount.value += 5
}

const formatDate = (value: string) => {
  try {
    return new Date(value).toLocaleDateString()
  } catch {
    return value
  }
}

useHead({
  title: t('blog.pages.news.metaTitle'),
})
</script>
