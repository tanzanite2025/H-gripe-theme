<template>
  <BlogPostDetailContent
    v-if="post"
    :post="post"
    category="wheelsbuild"
  />
</template>

<script setup lang="ts">
import { useBlogPostDetail } from '~/composables/blog/useBlogPostDetail'
import { useBlogPostSeo } from '~/composables/seo/useBlogPostSeo'

definePageMeta({
  layout: 'products',
})

const { post } = await useBlogPostDetail({
  category: 'wheelsbuild',
  keyPrefix: 'blog-post-wheelsbuild',
})

useBlogPostSeo({
  post,
  category: 'wheelsbuild',
})
</script>
