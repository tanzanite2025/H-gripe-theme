<template>
  <div
    class="admin-sidebar"
    :class="{ 'admin-sidebar--collapsed': collapsed }"
  >
    <div class="admin-sidebar__main">
      <header class="admin-sidebar__header">
        <RouterLink
          :to="{ name: 'Dashboard' }"
          class="admin-sidebar__brand"
          aria-label="返回控制台"
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
        :title="collapsed ? 'ROLE MATRIX AUTHORIZED' : undefined"
      >
        <span class="admin-sidebar__status-dot" />
        <span v-if="!collapsed" class="admin-sidebar__status-label">ROLE MATRIX AUTHORIZED</span>
      </div>

      <nav class="admin-sidebar__nav" aria-label="后台导航">
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
              v-if="hasChildren(item) && !collapsed"
              type="button"
              class="admin-sidebar__link admin-sidebar__link--button"
              :class="{
                'admin-sidebar__link--active': isGroupActive(item),
                'admin-sidebar__link--secondary': isSecondaryGroupOpen(item)
              }"
              :aria-expanded="isGroupOpen(item)"
              @click="toggleGroup(item)"
            >
              <component :is="item.icon" class="admin-sidebar__icon" aria-hidden="true" />
              <span class="admin-sidebar__label">
                <span class="admin-sidebar__label-code">{{ navCode(item) }}</span>
                <span class="admin-sidebar__label-divider">/</span>
                <span class="admin-sidebar__label-local">{{ item.label }}</span>
              </span>
              <ChevronDown
                class="admin-sidebar__chevron"
                :class="{ 'admin-sidebar__chevron--open': isGroupOpen(item) }"
                aria-hidden="true"
              />
            </button>

            <RouterLink
              v-else
              :to="itemTarget(item)"
              class="admin-sidebar__link"
              :class="{
                'admin-sidebar__link--active': isGroupActive(item),
                'admin-sidebar__link--secondary': isSecondaryGroupOpen(item),
                'admin-sidebar__link--collapsed': collapsed
              }"
              :aria-current="isGroupActive(item) ? 'page' : undefined"
              @blur="hideHoverTip"
              @click="emit('navigate')"
              @focus="showHoverTip(item, $event)"
              @mouseenter="showHoverTip(item, $event)"
              @mouseleave="hideHoverTip"
            >
              <component :is="item.icon" class="admin-sidebar__icon" aria-hidden="true" />
              <span v-if="!collapsed" class="admin-sidebar__label">
                <span class="admin-sidebar__label-code">{{ navCode(item) }}</span>
                <span class="admin-sidebar__label-divider">/</span>
                <span class="admin-sidebar__label-local">{{ item.label }}</span>
              </span>
            </RouterLink>

            <div
              v-if="hasChildren(item) && !collapsed && isGroupOpen(item)"
              class="admin-sidebar__children"
            >
              <RouterLink
                v-for="child in item.children"
                :key="itemKey(child)"
                :to="itemTarget(child, item)"
                class="admin-sidebar__child-link"
                :class="{ 'admin-sidebar__child-link--active': isChildActive(child, item) }"
                :aria-current="isChildActive(child, item) ? 'page' : undefined"
                @click="emit('navigate')"
              >
                <span class="admin-sidebar__child-dot" />
                <span class="admin-sidebar__child-label">{{ child.label }}</span>
              </RouterLink>
            </div>
          </div>
        </div>
      </nav>
    </div>

    <footer class="admin-sidebar__footer">
      <div class="admin-sidebar__user-card">
        <span class="admin-sidebar__avatar">{{ userInitials || 'AD' }}</span>
        <span v-if="!collapsed" class="admin-sidebar__user-copy">
          <strong>{{ displayName }}</strong>
          <small>{{ userEmail || roleLabel || 'backoffice user' }}</small>
        </span>
        <button
          type="button"
          class="admin-sidebar__logout"
          aria-label="退出登录"
          title="退出登录"
          @click="emit('request-logout')"
        >
          <LogOut class="admin-sidebar__logout-icon" />
        </button>
      </div>
    </footer>

    <Teleport to="body">
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

<script setup>
import { computed, ref, watch } from 'vue'
import { ChevronDown, LogOut } from '@lucide/vue'

const props = defineProps({
  collapsed: {
    type: Boolean,
    default: false
  },
  activePath: {
    type: String,
    required: true
  },
  items: {
    type: Array,
    required: true
  },
  brandInitial: {
    type: String,
    default: ''
  },
  brandName: {
    type: String,
    default: ''
  },
  panelLabel: {
    type: String,
    default: ''
  },
  roleLabel: {
    type: String,
    default: ''
  },
  user: {
    type: Object,
    default: null
  },
  userInitials: {
    type: String,
    default: 'AD'
  }
})

const emit = defineEmits(['navigate', 'request-logout'])
const hoveredItem = ref(null)
const hoverTipStyle = ref({})
const secondaryGroupKey = ref(null)

const itemKey = (item) => item.id || item.path || item.routeName || item.label
const hasChildren = (item) => Array.isArray(item.children) && item.children.length > 0

const matchesPath = (path) => {
  if (!path) return false
  if (path === '/') return props.activePath === '/'
  return props.activePath === path || props.activePath.startsWith(`${path}/`)
}

const isChildActive = (child, parent = null) => {
  const path = child.path || parent?.path
  if (!matchesPath(path)) return false
  return path !== parent?.path || props.activePath === path
}

const isGroupActive = (item) => {
  if (matchesPath(item.path)) return true
  return hasChildren(item) && item.children.some((child) => isChildActive(child, item))
}

const activeChild = (item) => (item.children || []).find((child) => isChildActive(child, item))

const activeGroupKey = computed(() => {
  const activeGroup = props.items.find((item) => hasChildren(item) && isGroupActive(item))
  return activeGroup ? itemKey(activeGroup) : null
})

const isGroupOpen = (item) => {
  const key = itemKey(item)
  return key === activeGroupKey.value || key === secondaryGroupKey.value
}

const isSecondaryGroupOpen = (item) => (
  isGroupOpen(item) && !isGroupActive(item)
)

const toggleGroup = (item) => {
  const key = itemKey(item)
  if (!hasChildren(item) || isGroupActive(item)) return
  secondaryGroupKey.value = secondaryGroupKey.value === key ? null : key
}

const itemTarget = (item, parent = null) => {
  const target = hasChildren(item) ? item.children[0] : item
  return {
    name: target.routeName || parent?.routeName || item.routeName,
  }
}

const navCode = (item) => item.code || item.routeName || item.label

const hideHoverTip = () => {
  hoveredItem.value = null
}

const showHoverTip = (item, event) => {
  if (!props.collapsed) return

  const rect = event.currentTarget.getBoundingClientRect()
  hoveredItem.value = item
  hoverTipStyle.value = {
    left: `${rect.right + 8}px`,
    top: `${rect.top + rect.height / 2}px`
  }
}

const hoverTipText = computed(() => {
  if (!hoveredItem.value) return ''
  const child = activeChild(hoveredItem.value)
  const childLabel = child ? ` / ${child.label}` : hasChildren(hoveredItem.value) ? ` / ${hoveredItem.value.children.length} 项` : ''
  return `${navCode(hoveredItem.value)} / ${hoveredItem.value.label}${childLabel}`
})

const brandTitle = computed(() => props.brandName?.trim() || 'SALES CONSOLE')
const brandSubtitle = computed(() => props.panelLabel?.trim() || 'CONTROL PANEL')
const brandMark = computed(() => {
  const configured = props.brandInitial?.trim()
  if (configured) return configured.slice(0, 3).toUpperCase()

  const words = brandTitle.value.split(/[\s_-]+/).filter(Boolean)
  if (words.length > 1) return (words[0][0] + words[words.length - 1][0]).toUpperCase()

  return Array.from(brandTitle.value).slice(0, 2).join('').toUpperCase()
})

const displayName = computed(() => {
  const currentUser = props.user || {}
  return currentUser.username || currentUser.email || 'Admin User'
})

const userEmail = computed(() => props.user?.email || '')

watch(
  activeGroupKey,
  (nextKey, previousKey) => {
    if (previousKey && previousKey !== nextKey) secondaryGroupKey.value = previousKey
    if (nextKey === secondaryGroupKey.value) secondaryGroupKey.value = null

    if (
      secondaryGroupKey.value &&
      !props.items.some((item) => itemKey(item) === secondaryGroupKey.value)
    ) {
      secondaryGroupKey.value = null
    }
  },
  { immediate: true }
)

watch(
  () => props.collapsed,
  () => hideHoverTip()
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
  overflow: hidden;
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
  gap: 0.25rem;
}

.admin-sidebar__group {
  display: grid;
  gap: 0.25rem;
}

.admin-sidebar__link {
  display: flex;
  position: relative;
  min-width: 0;
  width: 100%;
  height: 2.5rem;
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

.admin-sidebar__link--secondary {
  background: #eff6ff;
  color: #1d4ed8;
  box-shadow: inset 3px 0 0 #3b82f6;
}

.admin-sidebar__link--secondary:hover,
.admin-sidebar__link--secondary:focus-visible {
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

.admin-sidebar__link--secondary .admin-sidebar__icon {
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

.admin-sidebar__label-code,
.admin-sidebar__label-local {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.admin-sidebar__label-divider {
  flex-shrink: 0;
  opacity: 0.78;
}

.admin-sidebar__chevron {
  width: 0.9rem;
  height: 0.9rem;
  flex-shrink: 0;
  color: currentColor;
  opacity: 0.58;
  transition: transform 160ms ease, opacity 160ms ease;
}

.admin-sidebar__chevron--open {
  transform: rotate(180deg);
  opacity: 0.9;
}

.admin-sidebar__children {
  display: grid;
  gap: 0.125rem;
  margin: 0 0 0.15rem 1.25rem;
  border-left: 1px dashed #cbd5e1;
  padding-left: 0.55rem;
}

.admin-sidebar__child-link {
  display: flex;
  min-width: 0;
  height: 2rem;
  align-items: center;
  gap: 0.5rem;
  border-radius: 0.65rem;
  padding: 0 0.65rem;
  color: #64748b;
  font-size: 0.72rem;
  font-weight: 780;
  letter-spacing: 0;
  line-height: 1;
  text-decoration: none;
  transition: background-color 160ms ease, color 160ms ease;
}

.admin-sidebar__child-link:hover,
.admin-sidebar__child-link:focus-visible {
  background: #f1f5f9;
  color: #0f172a;
}

.admin-sidebar__child-link:focus-visible {
  outline: 2px solid rgb(16 185 129 / 0.28);
  outline-offset: 2px;
}

.admin-sidebar__child-link--active {
  background: rgb(15 23 42 / 0.06);
  color: #0f172a;
  font-weight: 900;
}

.admin-sidebar__child-dot {
  width: 0.32rem;
  height: 0.32rem;
  flex-shrink: 0;
  border-radius: 999px;
  background: currentColor;
  opacity: 0.38;
}

.admin-sidebar__child-link--active .admin-sidebar__child-dot {
  opacity: 1;
}

.admin-sidebar__child-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  height: 2.75rem;
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
