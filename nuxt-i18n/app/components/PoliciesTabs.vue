<template>
  <div class="nav-pill-tabs" role="tablist">
    <NuxtLink
      v-for="tab in tabs"
      :key="tab.slug"
      :to="localePath(`/policies/${tab.slug}`)"
      class="nav-pill-item"
      :class="{ 'nav-pill-item--active': resolvedCurrentSlug === tab.slug }"
    >
      {{ $t(tab.labelKey, tab.fallback) }}
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocalePath, useRoute } from '#imports'

const props = defineProps<{
  currentSlug?: string
}>()

const localePath = useLocalePath()
const route = useRoute()

const tabs = [
  { slug: 'cookie', labelKey: 'policyTabs.cookie', fallback: 'Cookie' },
  { slug: 'privacy', labelKey: 'policyTabs.privacy', fallback: 'Privacy' },
  { slug: 'refund-return', labelKey: 'policyTabs.refundReturn', fallback: 'Refund & Return' },
  { slug: 'terms', labelKey: 'policyTabs.terms', fallback: 'Terms' },
] as const

const resolvedCurrentSlug = computed(() => {
  if (typeof props.currentSlug === 'string' && props.currentSlug.length) {
    return props.currentSlug
  }

  const parts = route.path.split('/').filter(Boolean)
  return parts[parts.length - 1] || 'privacy'
})
</script>

<style scoped>
/* Styles moved to global nav.css */
</style>
