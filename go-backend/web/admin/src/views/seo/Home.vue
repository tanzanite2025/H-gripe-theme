<template>
  <SEOSettingsEditor
    title="SEO / 首页"
    description="首页默认搜索元数据"
    section-title="首页页面信息"
    page-route="/"
    :api="seoHomeApi"
  />
</template>

<script setup lang="ts">
import SEOSettingsEditor from '@/components/admin/seo/SEOSettingsEditor.vue'
import { seoHomeApi } from '@/modules/seo/home'
</script>
