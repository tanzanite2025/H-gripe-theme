<template>
  <component
    :is="usesRouteLinks ? 'nav' : 'div'"
    class="reusable-page-section-underline-tabs"
    :role="usesRouteLinks ? undefined : 'tablist'"
    :aria-orientation="usesRouteLinks ? undefined : 'horizontal'"
    :aria-label="tabListLabel"
  >
    <template v-for="(tab, tabIndex) in tabs" :key="tab.id">
      <NuxtLink
        v-if="tab.to"
        :id="tab.tabId"
        :to="tab.to"
        class="reusable-page-section-underline-tabs__tab"
        :class="{ 'reusable-page-section-underline-tabs__tab--active': activeId === tab.id }"
        :aria-current="activeId === tab.id ? 'page' : undefined"
      >
        {{ tab.label }}
      </NuxtLink>
      <button
        v-else
        :id="tab.tabId"
        type="button"
        role="tab"
        class="reusable-page-section-underline-tabs__tab"
        :class="{ 'reusable-page-section-underline-tabs__tab--active': activeId === tab.id }"
        :aria-selected="activeId === tab.id"
        :aria-controls="tab.panelId"
        :tabindex="activeId === tab.id ? 0 : -1"
        @click="selectPageSectionTab(tab.id)"
        @keydown="handlePageSectionTabKeyboardNavigation($event, tabIndex)"
      >
        {{ tab.label }}
      </button>
    </template>
  </component>
</template>

<script setup lang="ts" generic="TabId extends string">
import { computed } from 'vue'

const props = defineProps<{
  tabs: readonly {
    id: TabId
    label: string
    tabId?: string
    panelId?: string
    to?: string
  }[]
  activeId: TabId
  tabListLabel: string
}>()

const usesRouteLinks = computed(() => props.tabs.some(tab => Boolean(tab.to)))

const emit = defineEmits<{
  select: [id: TabId]
}>()

const selectPageSectionTab = (tabId: TabId) => {
  emit('select', tabId)
}

const handlePageSectionTabKeyboardNavigation = (event: KeyboardEvent, currentTabIndex: number) => {
  let nextTabIndex: number | null = null
  const tabDirection = document.documentElement.dir === 'rtl' ? -1 : 1

  if (event.key === 'ArrowRight') {
    nextTabIndex = (currentTabIndex + tabDirection + props.tabs.length) % props.tabs.length
  } else if (event.key === 'ArrowLeft') {
    nextTabIndex = (currentTabIndex - tabDirection + props.tabs.length) % props.tabs.length
  } else if (event.key === 'Home') {
    nextTabIndex = 0
  } else if (event.key === 'End') {
    nextTabIndex = props.tabs.length - 1
  }

  if (nextTabIndex === null) return
  event.preventDefault()
  const nextTab = props.tabs[nextTabIndex]
  if (!nextTab) return
  selectPageSectionTab(nextTab.id)
  if (nextTab.tabId) document.getElementById(nextTab.tabId)?.focus()
}
</script>

<style scoped>
.reusable-page-section-underline-tabs {
  display: flex;
  gap: 0.25rem;
  margin-bottom: 0.8rem;
  overflow-x: auto;
  border-bottom: 1px solid var(--tz-border-subtle);
  scrollbar-width: none;
}

.reusable-page-section-underline-tabs::-webkit-scrollbar {
  display: none;
}

.reusable-page-section-underline-tabs__tab {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  border: 0;
  border-radius: 0.5rem 0.5rem 0 0;
  background: transparent;
  color: var(--tz-text-secondary);
  cursor: pointer;
  padding: 0.55rem 0.85rem 0.65rem;
  font-size: 0.875rem;
  font-weight: 800;
  line-height: 1.2;
  text-decoration: none;
  white-space: nowrap;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.reusable-page-section-underline-tabs__tab::after {
  position: absolute;
  right: 0.55rem;
  bottom: 0;
  left: 0.55rem;
  height: 2px;
  border-radius: 999px;
  background: transparent;
  content: '';
}

.reusable-page-section-underline-tabs__tab:hover,
.reusable-page-section-underline-tabs__tab:focus-visible {
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
  outline: none;
}

.reusable-page-section-underline-tabs__tab--active {
  color: var(--tz-text-accent);
}

.reusable-page-section-underline-tabs__tab--active::after {
  background: var(--tz-site-accent);
}
</style>
