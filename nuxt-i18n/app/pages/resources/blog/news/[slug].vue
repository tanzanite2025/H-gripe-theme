<template>
  <BlogPostDetailContent
    v-if="post"
    :post="post"
    category="news"
  />
</template>

<script setup lang="ts">
import { useBlogPostDetail } from '~/composables/blog/useBlogPostDetail'
import { useBlogPostSeo } from '~/composables/seo/useBlogPostSeo'

definePageMeta({
  layout: 'products',
})

const { post } = await useBlogPostDetail({
  category: 'news',
  keyPrefix: 'blog-post-news',
})

useBlogPostSeo({
  post,
  category: 'news',
})
</script>
