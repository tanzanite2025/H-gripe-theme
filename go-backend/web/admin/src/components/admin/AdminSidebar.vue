<template>
  <div
    ref="sidebarElement"
    class="admin-sidebar"
    :class="{ 'admin-sidebar--collapsed': collapsed }"
  >
    <div class="admin-sidebar__main">
      <header class="admin-sidebar__header">
        <RouterLink
          :to="{ name: 'Dashboard' }"
          class="admin-sidebar__brand"
          :aria-label="t('layout.returnToConsole')"
          @click="emit('navigate')"
        >
          <span class="admin-sidebar__brand-mark">{{ brandMark }}</span>
          <span v-if="!collapsed" class="admin-sidebar__brand-copy">
            <strong>{{ brandTitle }}</strong>
            <small>{{ brandSubtitle }}</small>
          </span>
        </RouterLink>
      </header>

      <div
        class="admin-sidebar__status"
        :title="collapsed ? t('layout.roleMatrixAuthorized') : undefined"
      >
        <span class="admin-sidebar__status-dot" />
        <span v-if="!collapsed" class="admin-sidebar__status-label">{{ t('layout.roleMatrixAuthorized') }}</span>
      </div>

      <nav class="admin-sidebar__nav" :aria-label="t('layout.navigation')">
        <div class="admin-sidebar__menu">
          <div
            v-for="item in items"
            :key="itemKey(item)"
            class="admin-sidebar__group"
            :class="{
              'admin-sidebar__group--active': isGroupActive(item),
              'admin-sidebar__group--open': isGroupOpen(item)
            }"
          >
            <button
              v-if="hasChildren(item)"
              type="button"
              class="admin-sidebar__link admin-sidebar__link--button"
              :class="{
                'admin-sidebar__link--active': isGroupActive(item),
                'admin-sidebar__link--open': isGroupOpen(item)
              }"
              :aria-expanded="isGroupOpen(item)"
              aria-haspopup="dialog"
              :aria-label="collapsed ? item.label : undefined"
              @blur="hideHoverTip"
              @click="toggleGroup(item)"
              @focus="showHoverTip(item, $event)"
              @mouseenter="showHoverTip(item, $event)"
              @mouseleave="hideHoverTip"
            >
              <component :is="item.icon" class="admin-sidebar__icon" aria-hidden="true" />
              <span v-if="!collapsed" class="admin-sidebar__label">
                <span class="admin-sidebar__label-local">{{ item.label }}</span>
              </span>
              <ChevronRight
                class="admin-sidebar__chevron"
                aria-hidden="true"
              />
            </button>

            <RouterLink
              v-else
              :to="itemTarget(item)"
              class="admin-sidebar__link"
              :class="{
                'admin-sidebar__link--active': isGroupActive(item),
                'admin-sidebar__link--collapsed': collapsed
              }"
              :aria-current="isGroupActive(item) ? 'page' : undefined"
              @blur="hideHoverTip"
              @click="handleNavigate"
              @focus="showHoverTip(item, $event)"
              @mouseenter="showHoverTip(item, $event)"
              @mouseleave="hideHoverTip"
            >
              <component :is="item.icon" class="admin-sidebar__icon" aria-hidden="true" />
              <span v-if="!collapsed" class="admin-sidebar__label">
                <span class="admin-sidebar__label-local">{{ item.label }}</span>
              </span>
            </RouterLink>

          </div>
        </div>
      </nav>
    </div>

    <footer class="admin-sidebar__footer">
      <div class="admin-sidebar__user-card">
        <span class="admin-sidebar__avatar">{{ userInitials || 'AD' }}</span>
        <span v-if="!collapsed" class="admin-sidebar__user-copy">
          <strong>{{ displayName }}</strong>
          <small>{{ userEmail || roleLabel || t('roles.backofficeUser') }}</small>
        </span>
        <button
          type="button"
          class="admin-sidebar__logout"
          :aria-label="t('common.logout')"
          :title="t('common.logout')"
          @click="emit('request-logout')"
        >
          <LogOut class="admin-sidebar__logout-icon" />
        </button>
      </div>
    </footer>

    <Teleport to="body">
      <div
        v-if="floatingGroup"
        class="admin-sidebar-floating-overlay"
        aria-hidden="true"
        @click="closeFloatingGroup"
      />

      <section
        v-if="floatingGroup"
        class="admin-sidebar-floating-panel"
        :style="floatingPanelStyle"
        role="dialog"
        :aria-label="floatingGroup.label"
        @click.stop
      >
        <header class="admin-sidebar-floating-panel__header">
          <div class="admin-sidebar-floating-panel__heading">
            <span class="admin-sidebar-floating-panel__eyebrow">{{ t('layout.module') }}</span>
            <strong>{{ floatingGroup.label }}</strong>
          </div>
          <button
            type="button"
            class="admin-sidebar-floating-panel__close"
            :aria-label="t('layout.closeSubmenu')"
            :title="t('layout.closeSubmenu')"
            @click="closeFloatingGroup"
          >
            <X class="admin-sidebar-floating-panel__close-icon" aria-hidden="true" />
          </button>
        </header>

        <nav class="admin-sidebar-floating-panel__children" :aria-label="`${floatingGroup.label} ${t('layout.submenu')}`">
          <RouterLink
            v-for="child in floatingGroup.children"
            :key="itemKey(child)"
            :to="itemTarget(child, floatingGroup)"
            class="admin-sidebar-floating-panel__child-link"
            :class="{ 'admin-sidebar-floating-panel__child-link--active': isChildActive(child, floatingGroup) }"
            :aria-current="isChildActive(child, floatingGroup) ? 'page' : undefined"
            @click="handleNavigate"
          >
            <span class="admin-sidebar-floating-panel__child-mark" />
            <span class="admin-sidebar-floating-panel__child-label">{{ child.label }}</span>
            <ChevronRight class="admin-sidebar-floating-panel__child-arrow" aria-hidden="true" />
          </RouterLink>
        </nav>
      </section>

      <div
        v-if="collapsed && hoveredItem"
        class="admin-sidebar-floating-tip"
        :style="hoverTipStyle"
        aria-hidden="true"
      >
        {{ hoverTipText }}
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { CSSProperties } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { ChevronRight, LogOut, X } from '@lucide/vue'
import type { AdminNavigationItem } from '@/lib/adminNavigation'
import type { AdminUser } from '@/stores/auth'
import { useAdminI18n } from '@/i18n'

const props = withDefaults(defineProps<{
  collapsed?: boolean
  activePath: string
  items: AdminNavigationItem[]
  brandInitial?: string
  brandName?: string
  panelLabel?: string
  roleLabel?: string
  user?: AdminUser | null
  userInitials?: string
}>(), {
  collapsed: false,
  brandInitial: '',
  brandName: '',
  panelLabel: '',
  roleLabel: '',
  user: null,
  userInitials: 'AD'
})

const emit = defineEmits<{
  (event: 'navigate'): void
  (event: 'request-logout'): void
}>()
const { t } = useAdminI18n()
const sidebarElement = ref<HTMLElement | null>(null)
const hoveredItem = ref<AdminNavigationItem | null>(null)
const hoverTipStyle = ref<CSSProperties>({})
const floatingGroupKey = ref<string | null>(null)

const itemKey = (item: AdminNavigationItem): string => item.id || item.path || item.routeName || item.label
const hasChildren = (item: AdminNavigationItem): item is AdminNavigationItem & { children: AdminNavigationItem[] } => (
  Array.isArray(item.children) && item.children.length > 0
)

const matchesPath = (path?: string): boolean => {
  if (!path) return false
  if (path === '/') return props.activePath === '/'
  return props.activePath === path || props.activePath.startsWith(`${path}/`)
}

const isChildActive = (child: AdminNavigationItem, parent: AdminNavigationItem | null = null): boolean => {
  const path = child.path || parent?.path
  if (!matchesPath(path)) return false
  return path !== parent?.path || props.activePath === path
}

const isGroupActive = (item: AdminNavigationItem): boolean => {
  if (matchesPath(item.path)) return true
  return hasChildren(item) && item.children.some((child) => isChildActive(child, item))
}

const activeChild = (item: AdminNavigationItem): AdminNavigationItem | undefined => (
  (item.children || []).find((child) => isChildActive(child, item))
)

const isGroupOpen = (item: AdminNavigationItem): boolean => {
  return itemKey(item) === floatingGroupKey.value
}

const floatingGroup = computed(() => (
  props.items.find((item) => itemKey(item) === floatingGroupKey.value && hasChildren(item)) || null
))

const floatingPanelStyle = computed<CSSProperties>(() => {
  const rect = sidebarElement.value?.getBoundingClientRect()
  return rect ? { left: `${rect.right + 12}px` } : {}
})

const closeFloatingGroup = (): void => {
  floatingGroupKey.value = null
}

const toggleGroup = (item: AdminNavigationItem): void => {
  const key = itemKey(item)
  if (!hasChildren(item)) return

  hideHoverTip()
  floatingGroupKey.value = floatingGroupKey.value === key ? null : key
}

const itemTarget = (item: AdminNavigationItem, parent: AdminNavigationItem | null = null): RouteLocationRaw => {
  const target = hasChildren(item) ? item.children[0] : item
  const routeName = target.routeName || parent?.routeName || item.routeName
  if (routeName) return { name: routeName }

  const path = target.path || parent?.path || item.path
  if (path) return path

  return {
    name: 'Dashboard',
  }
}

const hideHoverTip = (): void => {
  hoveredItem.value = null
}

const handleNavigate = (): void => {
  closeFloatingGroup()
  emit('navigate')
}

const showHoverTip = (item: AdminNavigationItem, event: MouseEvent | FocusEvent): void => {
  if (!props.collapsed) return

  const target = event.currentTarget
  if (!(target instanceof HTMLElement)) return

  const rect = target.getBoundingClientRect()
  hoveredItem.value = item
  hoverTipStyle.value = {
    left: `${rect.right + 8}px`,
    top: `${rect.top + rect.height / 2}px`
  }
}

const hoverTipText = computed(() => {
  const item = hoveredItem.value
  if (!item) return ''

  const child = activeChild(item)
  const childLabel = child ? ` / ${child.label}` : hasChildren(item) ? ` / ${t('common.itemCount', { count: item.children.length })}` : ''
  return `${item.label}${childLabel}`
})

const brandTitle = computed(() => props.brandName?.trim() || t('layout.salesConsole'))
const brandSubtitle = computed(() => props.panelLabel?.trim() || t('layout.panelFallback'))
const brandMark = computed(() => {
  const configured = props.brandInitial?.trim()
  if (configured) return configured.slice(0, 3).toUpperCase()

  const words = brandTitle.value.split(/[\s_-]+/).filter(Boolean)
  if (words.length > 1) return (words[0][0] + words[words.length - 1][0]).toUpperCase()

  return Array.from(brandTitle.value).slice(0, 2).join('').toUpperCase()
})

const displayName = computed(() => {
  return props.user?.username || props.user?.email || t('roles.adminUser')
})

const userEmail = computed(() => props.user?.email || '')

const handleKeydown = (event: KeyboardEvent): void => {
  if (event.key === 'Escape' && floatingGroup.value) closeFloatingGroup()
}

const handleResize = (): void => {
  closeFloatingGroup()
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
  window.removeEventListener('resize', handleResize)
})

watch(
  () => props.collapsed,
  () => {
    hideHoverTip()
    closeFloatingGroup()
  }
)

watch(
  () => props.activePath,
  () => closeFloatingGroup()
)

watch(
  () => props.items,
  () => {
    if (
      floatingGroupKey.value &&
      !props.items.some((item) => itemKey(item) === floatingGroupKey.value && hasChildren(item))
    ) {
      closeFloatingGroup()
    }
  }
)
</script>

<style scoped>
.admin-sidebar {
  position: relative;
  display: flex;
  width: 100%;
  height: 100%;
  min-height: 0;
  flex-direction: column;
  justify-content: space-between;
  overflow: visible;
  background: #ffffff;
  color: #0f172a;
}

.admin-sidebar__main {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
}

.admin-sidebar__header {
  display: flex;
  height: 4rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-bottom: 1px dashed #e2e8f0;
  padding-inline: 1.25rem;
}

.admin-sidebar__brand {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 0.75rem;
  color: inherit;
  text-decoration: none;
}

.admin-sidebar__brand-mark {
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 0.75rem;
  background: #0f172a;
  color: #ffffff;
  box-shadow: 0 12px 24px rgb(15 23 42 / 0.18);
  font-size: 0.875rem;
  font-style: italic;
  font-weight: 950;
  letter-spacing: 0;
}

.admin-sidebar__brand-copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  line-height: 1.05;
  white-space: nowrap;
}

.admin-sidebar__brand-copy strong {
  overflow: hidden;
  color: #0f172a;
  font-size: 0.75rem;
  font-style: italic;
  font-weight: 950;
  letter-spacing: 0;
  text-overflow: ellipsis;
}

.admin-sidebar__brand-copy small {
  margin-top: 0.25rem;
  overflow: hidden;
  color: #94a3b8;
  font-size: 0.5rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-overflow: ellipsis;
  text-transform: uppercase;
}

.admin-sidebar__status {
  display: flex;
  min-height: 3rem;
  flex-shrink: 0;
  align-items: center;
  gap: 0.625rem;
  margin: 1rem 0.75rem 0.5rem;
  border: 1px dashed rgb(16 185 129 / 0.3);
  border-radius: 0.875rem;
  background: rgb(236 253 245 / 0.62);
  padding: 0.625rem 0.75rem;
}

.admin-sidebar__status-dot {
  width: 0.5rem;
  height: 0.5rem;
  flex-shrink: 0;
  border-radius: 999px;
  background: #10b981;
  box-shadow: 0 0 0 4px rgb(16 185 129 / 0.12);
}

.admin-sidebar__status-label {
  min-width: 0;
  overflow: hidden;
  color: #065f46;
  font-size: 0.5625rem;
  font-weight: 850;
  letter-spacing: 0.14em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-sidebar__nav {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem 0.75rem 1rem;
}

.admin-sidebar__menu {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.admin-sidebar__group {
  display: grid;
}

.admin-sidebar__link {
  display: flex;
  position: relative;
  min-width: 0;
  width: 100%;
  height: 2.25rem;
  align-items: center;
  gap: 0.75rem;
  border: 0;
  border-radius: 1rem;
  background: transparent;
  padding-inline: 0.9rem;
  color: #475569;
  font-size: 0.75rem;
  font-weight: 850;
  letter-spacing: 0;
  line-height: 1;
  text-align: left;
  text-decoration: none;
  transition:
    background-color 160ms ease,
    box-shadow 160ms ease,
    color 160ms ease,
    transform 160ms ease;
}

.admin-sidebar__link--button {
  cursor: pointer;
}

.admin-sidebar__link:hover,
.admin-sidebar__link:focus-visible {
  background: #f1f5f9;
  color: #0f172a;
}

.admin-sidebar__link:focus-visible {
  outline: 2px solid rgb(16 185 129 / 0.35);
  outline-offset: 2px;
}

.admin-sidebar__link:active {
  transform: scale(0.985);
}

.admin-sidebar__link--active {
  background: var(--admin-control-selected);
  color: var(--admin-control-selected-foreground);
  box-shadow: var(--admin-control-selected-shadow);
}

.admin-sidebar__link--open:not(.admin-sidebar__link--active) {
  background: #eff6ff;
  color: #1d4ed8;
  box-shadow: inset 3px 0 0 #3b82f6;
}

.admin-sidebar__link--open:not(.admin-sidebar__link--active):hover,
.admin-sidebar__link--open:not(.admin-sidebar__link--active):focus-visible {
  background: #dbeafe;
  color: #1e40af;
}

.admin-sidebar__icon {
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
  color: #64748b;
  stroke-width: 2.2;
  transition: color 160ms ease, transform 160ms ease;
}

.admin-sidebar__link:hover .admin-sidebar__icon,
.admin-sidebar__link:focus-visible .admin-sidebar__icon {
  color: #0f172a;
}

.admin-sidebar__link--active .admin-sidebar__icon {
  color: var(--admin-control-selected-icon);
}

.admin-sidebar__link--open:not(.admin-sidebar__link--active) .admin-sidebar__icon {
  color: #2563eb;
}

.admin-sidebar__label {
  display: inline-flex;
  min-width: 0;
  flex: 1;
  align-items: baseline;
  gap: 0.45rem;
  overflow: hidden;
  white-space: nowrap;
}

.admin-sidebar__label-local {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.admin-sidebar__chevron {
  width: 0.9rem;
  height: 0.9rem;
  flex-shrink: 0;
  color: currentColor;
  opacity: 0.58;
  transition: transform 160ms ease, opacity 160ms ease;
}

.admin-sidebar__footer {
  flex-shrink: 0;
  border-top: 1px dashed #e2e8f0;
  padding: 0.75rem;
}

.admin-sidebar__user-card {
  display: flex;
  position: relative;
  min-width: 0;
  align-items: center;
  gap: 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 1rem;
  background: #f8fafc;
  padding: 0.625rem;
}

.admin-sidebar__avatar {
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: #0f172a;
  color: #ffffff;
  font-size: 0.75rem;
  font-weight: 850;
  letter-spacing: 0;
}

.admin-sidebar__user-copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  line-height: 1.1;
}

.admin-sidebar__user-copy strong,
.admin-sidebar__user-copy small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-sidebar__user-copy strong {
  color: #0f172a;
  font-size: 0.75rem;
  font-weight: 800;
}

.admin-sidebar__user-copy small {
  margin-top: 0.25rem;
  color: #64748b;
  font-size: 0.5625rem;
}

.admin-sidebar__logout {
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  margin-left: auto;
  border: 1px dashed #cbd5e1;
  border-radius: 999px;
  background: #ffffff;
  color: #64748b;
  transition:
    background-color 160ms ease,
    border-color 160ms ease,
    color 160ms ease,
    transform 160ms ease;
}

.admin-sidebar__logout:hover,
.admin-sidebar__logout:focus-visible {
  border-color: rgb(244 63 94 / 0.35);
  background: rgb(255 241 242 / 0.88);
  color: #e11d48;
}

.admin-sidebar__logout:focus-visible {
  outline: 2px solid rgb(244 63 94 / 0.25);
  outline-offset: 2px;
}

.admin-sidebar__logout:active {
  transform: scale(0.96);
}

.admin-sidebar__logout-icon {
  width: 0.95rem;
  height: 0.95rem;
}

.admin-sidebar--collapsed .admin-sidebar__header {
  gap: 0.25rem;
  justify-content: center;
  padding-inline: 0.375rem;
}

.admin-sidebar--collapsed .admin-sidebar__brand {
  gap: 0;
}

.admin-sidebar--collapsed .admin-sidebar__status {
  min-height: 2.5rem;
  justify-content: center;
  margin-inline: 0.75rem;
  padding-inline: 0;
}

.admin-sidebar--collapsed .admin-sidebar__nav {
  padding-inline: 0.625rem;
}

.admin-sidebar--collapsed .admin-sidebar__link {
  width: 2.75rem;
  height: 2.5rem;
  justify-content: center;
  gap: 0;
  padding-inline: 0;
}

.admin-sidebar--collapsed .admin-sidebar__menu {
  align-items: center;
}

.admin-sidebar--collapsed .admin-sidebar__footer {
  padding-inline: 0.625rem;
}

.admin-sidebar--collapsed .admin-sidebar__user-card {
  flex-direction: column;
  gap: 0.5rem;
  justify-content: center;
  padding: 0.5rem 0;
}

.admin-sidebar--collapsed .admin-sidebar__logout {
  margin-left: 0;
}

.admin-sidebar-floating-overlay {
  position: fixed;
  z-index: 89;
  inset: 0;
  background: transparent;
}

.admin-sidebar-floating-panel {
  position: fixed;
  z-index: 90;
  top: 50%;
  display: flex;
  width: clamp(11rem, 12vw, 15rem);
  max-height: min(42rem, calc(100vh - 3rem));
  flex-direction: column;
  overflow: hidden;
  transform: translateY(-50%);
  border: 1px solid var(--border);
  border-radius: 1.25rem;
  background: var(--card);
  color: var(--card-foreground);
  box-shadow:
    0 24px 60px rgb(15 23 42 / 0.18),
    0 8px 24px rgb(15 23 42 / 0.08);
}

.admin-sidebar-floating-panel__header {
  display: flex;
  min-height: 4.25rem;
  flex-shrink: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px dashed var(--border);
  padding: 0.875rem 0.875rem 0.75rem 1rem;
}

.admin-sidebar-floating-panel__heading {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.2rem;
}

.admin-sidebar-floating-panel__heading strong {
  overflow: hidden;
  color: var(--foreground);
  font-size: 0.8125rem;
  font-weight: 950;
  letter-spacing: 0.02em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-sidebar-floating-panel__heading > span:last-child {
  overflow: hidden;
  color: var(--muted-foreground);
  font-size: 0.6875rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-sidebar-floating-panel__eyebrow {
  color: var(--muted-foreground);
  font-size: 0.5rem;
  font-weight: 900;
  letter-spacing: 0.16em;
  line-height: 1;
}

.admin-sidebar-floating-panel__close {
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: var(--background);
  color: var(--muted-foreground);
  cursor: pointer;
  transition: background-color 160ms ease, border-color 160ms ease, color 160ms ease;
}

.admin-sidebar-floating-panel__close:hover,
.admin-sidebar-floating-panel__close:focus-visible {
  border-color: var(--admin-control-selected-border);
  background: var(--admin-control-selected-soft);
  color: var(--foreground);
}

.admin-sidebar-floating-panel__close:focus-visible {
  outline: 2px solid rgb(16 185 129 / 0.3);
  outline-offset: 2px;
}

.admin-sidebar-floating-panel__close-icon {
  width: 0.9rem;
  height: 0.9rem;
}

.admin-sidebar-floating-panel__children {
  display: grid;
  min-height: 0;
  gap: 0.375rem;
  overflow-y: auto;
  padding: 0.75rem;
}

.admin-sidebar-floating-panel__child-link {
  display: flex;
  min-width: 0;
  min-height: 2.25rem;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: var(--background);
  padding: 0.375rem 0.625rem 0.375rem 0.75rem;
  color: var(--muted-foreground);
  font-size: 0.6875rem;
  font-weight: 800;
  line-height: 1.1;
  text-decoration: none;
  transition:
    background-color 160ms ease,
    border-color 160ms ease,
    color 160ms ease,
    transform 160ms ease;
}

.admin-sidebar-floating-panel__child-link:hover,
.admin-sidebar-floating-panel__child-link:focus-visible {
  background: var(--muted);
  border-color: var(--admin-control-selected-border);
  color: var(--foreground);
}

.admin-sidebar-floating-panel__child-link:focus-visible {
  outline: 2px solid rgb(16 185 129 / 0.28);
  outline-offset: 2px;
}

.admin-sidebar-floating-panel__child-link:active {
  transform: scale(0.99);
}

.admin-sidebar-floating-panel__child-link--active {
  border-color: var(--admin-control-selected);
  background: var(--admin-control-selected);
  color: var(--admin-control-selected-foreground);
  font-weight: 950;
  box-shadow: 0 8px 16px rgb(15 23 42 / 0.14);
}

.admin-sidebar-floating-panel__child-mark {
  width: 0.35rem;
  height: 0.35rem;
  flex-shrink: 0;
  border-radius: 999px;
  background: currentColor;
  opacity: 0.34;
}

.admin-sidebar-floating-panel__child-link--active .admin-sidebar-floating-panel__child-mark {
  background: var(--admin-control-selected-icon);
  opacity: 1;
}

.admin-sidebar-floating-panel__child-label {
  min-width: 0;
  flex: 1;
  overflow-wrap: anywhere;
}

.admin-sidebar-floating-panel__child-arrow {
  width: 0.8rem;
  height: 0.8rem;
  flex-shrink: 0;
  opacity: 0.5;
}

@media (max-width: 1023px) {
  .admin-sidebar-floating-overlay {
    background: rgb(15 23 42 / 0.12);
  }

  .admin-sidebar-floating-panel {
    top: auto !important;
    right: 0;
    bottom: 0;
    left: 0 !important;
    width: 100vw;
    max-width: 100vw;
    max-height: min(70vh, 36rem);
    border-radius: 1.25rem 1.25rem 0 0;
    transform: none;
  }
}

.admin-sidebar-floating-panel__child-link,
.admin-sidebar-floating-panel__close {
  touch-action: manipulation;
}

.admin-sidebar-floating-tip {
  position: fixed;
  z-index: 80;
  display: inline-flex;
  align-items: center;
  max-width: min(22rem, calc(100vw - 7rem));
  transform: translateY(-50%);
  border-radius: 999px;
  background: #0f172a;
  color: #ffffff;
  box-shadow: 0 14px 26px rgb(15 23 42 / 0.24);
  font-size: 0.75rem;
  font-weight: 850;
  letter-spacing: 0;
  line-height: 1;
  padding: 0.7rem 0.9rem;
  pointer-events: none;
  white-space: nowrap;
}

.admin-sidebar-floating-tip::before {
  position: absolute;
  top: 50%;
  left: -0.34rem;
  width: 0.7rem;
  height: 0.7rem;
  content: '';
  transform: translateY(-50%) rotate(45deg);
  border-bottom-left-radius: 0.15rem;
  background: #0f172a;
}
</style>
