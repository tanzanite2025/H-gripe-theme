<template>
  <nav class="support-top-nav" aria-label="Support navigation">
    <div class="support-top-nav__scroll">
      <NuxtLink
        v-for="item in items"
        :key="item.id"
        class="support-top-nav__link"
        :class="{ 'support-top-nav__link--active': isActive(item) }"
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
.support-top-nav {
  width: 100%;
  border-bottom: 1px solid rgba(148, 163, 184, 0.3);
  background: rgba(15, 23, 42, 0.92);
  -webkit-backdrop-filter: blur(12px);
  backdrop-filter: blur(12px);
}

.support-top-nav__scroll {
  max-width: 960px;
  margin: 0 auto;
  padding: 0.75rem 1.25rem;
  display: flex;
  align-items: center;
  gap: 1.5rem;
  overflow-x: auto;
  scrollbar-width: thin;
}

.support-top-nav__link {
  flex-shrink: 0;
  font-size: 1rem;
  font-weight: 500;
  color: #ffffff !important;
  text-decoration: none;
  padding-bottom: 0.3rem;
  border-bottom: 3px solid transparent;
  transition: color 0.15s ease, border-color 0.15s ease;
}

.support-top-nav__link:hover,
.support-top-nav__link:focus-visible {
  color: #e5f2ff;
}

.support-top-nav__link--active {
  color: #ffffff;
  font-weight: 600;
  border-color: #38bdf8;
}

@media (max-width: 768px) {
  .support-top-nav__scroll {
    padding-inline: 0.75rem;
  }

  .support-top-nav__link {
    font-size: 0.875rem;
  }
}
</style>
