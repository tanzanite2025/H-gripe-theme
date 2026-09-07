<template>
  <div>
    <header class="flex flex-col gap-2">
      <h1 class="text-2xl font-semibold tz-text-primary">
        {{ post.title }}
      </h1>
      <p class="text-sm tz-text-muted">
        {{ formatDate(post.date) }}
      </p>
      <div v-if="post.categories.length" class="flex flex-wrap gap-2">
        <span
          v-for="category in post.categories"
          :key="category.slug"
          class="rounded-full tz-surface-subtle px-3 py-1 text-xs tz-text-secondary"
        >
          {{ category.name }}
        </span>
      </div>
    </header>

    <article
      class="mt-6 rounded-2xl border tz-border-strong tz-surface-card p-4 tz-text-secondary"
    >
      <SafeRichText class="space-y-4" :html="post.contentHtml" />
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
import { useI18n } from '#imports'
import type { BlogPostDetail } from '~/utils/blog/types'

interface Props {
  post: BlogPostDetail
}

defineProps<Props>()
const { t } = useI18n()

const formatDate = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString()
}
</script>
