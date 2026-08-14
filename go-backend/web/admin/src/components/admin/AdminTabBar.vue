<template>
  <nav
    class="admin-tab-bar"
    :aria-label="tabLabel"
    role="tablist"
  >
    <RouterLink
      v-for="tab in tabs"
      :key="itemKey(tab)"
      :to="itemTarget(tab)"
      class="admin-tab-bar__tab"
      :class="{ 'admin-tab-bar__tab--active': isActive(tab) }"
      role="tab"
      :aria-selected="isActive(tab)"
      :aria-current="isActive(tab) ? 'page' : undefined"
    >
      <span class="admin-tab-bar__tab-label">{{ tab.label }}</span>
    </RouterLink>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import type { AdminNavigationItem } from '@/lib/adminNavigation'
import { useAdminI18n } from '@/i18n'

const props = withDefaults(defineProps<{
  tabs: AdminNavigationItem[]
  activePath: string
  label?: string
}>(), {
  label: ''
})

const { t } = useAdminI18n()

const itemKey = (item: AdminNavigationItem): string => item.id || item.path || item.routeName || item.label

const itemTarget = (item: AdminNavigationItem): RouteLocationRaw => {
  if (item.routeName) return { name: item.routeName }
  if (item.path) return item.path
  return { name: 'Dashboard' }
}

const isActive = (item: AdminNavigationItem): boolean => {
  if (!item.path) return false
  if (item.path === '/') return props.activePath === '/'
  return props.activePath === item.path || props.activePath.startsWith(`${item.path}/`)
}

const tabLabel = computed(() => props.label || t('layout.navigation'))
</script>

<style scoped>
.admin-tab-bar {
  display: flex;
  min-width: 0;
  width: 100%;
  min-height: 3.5rem;
  align-items: stretch;
  gap: 0.25rem;
  overflow-x: auto;
  border: 1px solid #e2e8f0;
  border-radius: 1.25rem;
  background: #f8fafc;
  padding: 0.35rem;
  scrollbar-width: thin;
}

.admin-tab-bar::-webkit-scrollbar {
  height: 6px;
}

.admin-tab-bar__tab {
  display: inline-flex;
  min-width: max-content;
  flex: 1 0 auto;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  border: 1px solid transparent;
  border-radius: 0.95rem;
  padding: 0.55rem 1.1rem;
  color: #64748b;
  font-size: 0.75rem;
  font-weight: 850;
  line-height: 1.1;
  text-decoration: none;
  transition:
    background-color 160ms ease,
    border-color 160ms ease,
    box-shadow 160ms ease,
    color 160ms ease,
    transform 160ms ease;
}

.admin-tab-bar__tab:hover,
.admin-tab-bar__tab:focus-visible {
  border-color: #cbd5e1;
  background: #ffffff;
  color: #0f172a;
}

.admin-tab-bar__tab:focus-visible {
  outline: 2px solid rgb(16 185 129 / 0.32);
  outline-offset: 2px;
}

.admin-tab-bar__tab:active {
  transform: scale(0.99);
}

.admin-tab-bar__tab--active {
  border-color: var(--admin-control-selected-border);
  background: var(--admin-control-selected);
  color: var(--admin-control-selected-foreground);
  box-shadow: var(--admin-control-selected-shadow);
}

.admin-tab-bar__tab--active:hover,
.admin-tab-bar__tab--active:focus-visible {
  border-color: var(--admin-control-selected-border);
  background: var(--admin-control-selected-hover);
  color: var(--admin-control-selected-foreground);
}

.admin-tab-bar__tab-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 639px) {
  .admin-tab-bar {
    min-height: 3.25rem;
    border-radius: 1rem;
  }

  .admin-tab-bar__tab {
    min-height: 2.5rem;
    padding-inline: 0.85rem;
  }
}
</style>
