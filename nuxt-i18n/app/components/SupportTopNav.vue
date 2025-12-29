<template>
  <nav class="nav-top-bar" aria-label="Support navigation">
    <div class="nav-top-bar__scroll">
      <NuxtLink
        v-for="item in items"
        :key="item.id"
        class="nav-top-bar__link"
        :class="{ 'nav-top-bar__link--active': isActive(item) }"
        :to="localePath(item.to)"
      >
        {{ $t(item.labelKey) }}
      </NuxtLink>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocalePath, useRoute } from '#imports'
import type { SupportNavItem } from '~/utils/supportNav'
import { supportNavItems } from '~/utils/supportNav'

const props = defineProps<{
  /**
   * Optional override for support nav items.
   * If not provided, the default supportNavItems config is used.
   */
  itemsOverride?: SupportNavItem[]
}>()

const route = useRoute()
const localePath = useLocalePath()

const items = computed<SupportNavItem[]>(() => {
  if (props.itemsOverride && props.itemsOverride.length) {
    return props.itemsOverride
  }
  return supportNavItems
})

const isActive = (item: SupportNavItem) => {
  const targetPath = localePath(item.to)
  const currentPath = route.path

  // Exact match or nested path under the target (e.g. /support/payment/paypal)
  return (
    currentPath === targetPath ||
    (currentPath.startsWith(targetPath) && currentPath[targetPath.length] === '/')
  )
}
</script>

<style scoped>
/* Styles moved to global nav.css */
</style>
