<template>
  <BlogPostDetailContent
    v-if="post"
    :post="post"
  />
</template>

<script setup lang="ts">
import { useBlogPostDetail } from '~/composables/blog/useBlogPostDetail'
import { useBlogPostSeo } from '~/composables/seo/useBlogPostSeo'

definePageMeta({
  layout: 'products',
})

const { post } = await useBlogPostDetail({
  keyPrefix: 'blog-post',
})

useBlogPostSeo({
  post,
})
</script>
