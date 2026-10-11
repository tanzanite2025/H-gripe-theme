<template>
  <div
    class="reusable-page-section-underline-tabs"
    role="tablist"
    aria-orientation="horizontal"
    :aria-label="tabListLabel"
  >
    <button
      v-for="(tab, tabIndex) in tabs"
      :id="tab.tabId"
      :key="tab.id"
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
  </div>
</template>

<script setup lang="ts" generic="TabId extends string">
const props = defineProps<{
  tabs: readonly {
    id: TabId
    label: string
    tabId: string
    panelId: string
  }[]
  activeId: TabId
  tabListLabel: string
}>()

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
  document.getElementById(nextTab.tabId)?.focus()
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
