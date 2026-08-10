<template>
  <div>
    <header class="flex flex-col gap-2">
      <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
        {{ categoryLabel }}
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

    <PostTranslations
      v-if="post.id"
      :post-id="post.id"
      :localized-routes="post.localizedRoutes"
      :title="t('blog.availableInOtherLanguages')"
      class="mt-6"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import type { BlogCategory, BlogPostDetail } from '~/utils/blog/types'

interface Props {
  post: BlogPostDetail
  category: BlogCategory | null
}

const props = defineProps<Props>()
const { t } = useI18n()

const categoryLabel = computed(() => {
  if (props.category === 'news') return t('blog.nav.news')
  if (props.category === 'wheelsbuild') return t('blog.nav.wheelsbuild')
  return t('blog.nav.all')
})

const formatDate = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString()
}
</script>
