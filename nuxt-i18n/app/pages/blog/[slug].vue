<template>
  <BlogPostDetailContent
    v-if="post"
    :post="post"
    :category="null"
  />
</template>

<script setup lang="ts">
import { useBlogPostDetail } from '~/composables/blog/useBlogPostDetail'
import { useBlogPostSeo } from '~/composables/seo/useBlogPostSeo'

definePageMeta({
  layout: 'products',
})

const { post } = await useBlogPostDetail({
  category: null,
  keyPrefix: 'blog-post',
})

useBlogPostSeo({
  post,
  category: null,
})
</script>
