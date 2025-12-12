<template>
  <div class="company-tabs" role="tablist">
    <NuxtLink
      v-for="tab in tabs"
      :key="tab.slug"
      :to="localePath(`/policies/${tab.slug}`)"
      class="company-tabs__item"
      :class="{ 'company-tabs__item--active': resolvedCurrentSlug === tab.slug }"
    >
      {{ tab.label }}
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
  { slug: 'cookie', label: 'cookie' },
  { slug: 'privacy', label: 'privacy' },
  { slug: 'refund-return', label: 'refund-return' },
  { slug: 'terms', label: 'terms' },
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
.company-tabs {
  display: flex;
  overflow-x: auto;
  gap: 12px;
  padding: 4px 16px;
  margin: 0 -16px 1rem;
  max-width: calc(100% + 32px);
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  touch-action: pan-x;
}

.company-tabs::-webkit-scrollbar {
  display: none;
}

.company-tabs__item {
  flex-shrink: 0;
  border: none;
  border-radius: 9999px;
  padding: 8px 18px;
  font-size: 0.85rem;
  font-weight: 500;
  color: #ffffff;
  background: rgba(31, 41, 55, 0.9);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  backdrop-filter: blur(4px);
  box-shadow: 6px 8px 18px -12px rgba(0, 0, 0, 0.85);
  text-decoration: none;
}

.company-tabs__item:active {
  transform: scale(0.96);
}

.company-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.company-tabs__item--active {
  background: #ffffff;
  color: #0f172a;
  border: none;
  font-weight: 600;
  box-shadow: 8px 10px 22px -10px rgba(0, 0, 0, 0.9);
}

@media (min-width: 768px) {
  .company-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
}

@media (max-width: 768px) {
  .company-tabs {
    justify-content: flex-start;
  }
}
</style>
