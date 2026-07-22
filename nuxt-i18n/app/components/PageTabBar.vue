<template>
  <div class="page-tab-bar" :class="{ 'page-tab-bar--open': mobileOpen }">
    <button
      type="button"
      class="page-tab-bar__mobile-trigger"
      :aria-expanded="mobileOpen"
      :aria-label="mobileOpen ? 'Collapse page sections' : 'Expand page sections'"
      @click="mobileOpen = !mobileOpen"
    >
      <span class="page-tab-bar__hamburger" aria-hidden="true">
        <span />
        <span />
        <span />
      </span>
      <span class="page-tab-bar__mobile-text">
        {{ activeLabel }}
      </span>
    </button>

    <div
      class="page-tab-bar__list"
      role="tablist"
      :aria-label="ariaLabel"
    >
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="page-tab-bar__item"
        :class="{ 'page-tab-bar__item--active': activeId === tab.id }"
        role="tab"
        :aria-selected="activeId === tab.id"
        :tabindex="activeId === tab.id ? 0 : -1"
        @click="selectTab(tab.id)"
      >
        {{ getLabel(tab) }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

type PageTabItem = {
  id: string
  label?: string
  labelKey?: string
  fallback?: string
}

const props = withDefaults(defineProps<{
  tabs: PageTabItem[]
  activeId: string
  ariaLabel?: string
}>(), {
  ariaLabel: 'Page sections',
})

const emit = defineEmits<{
  select: [id: string]
}>()

const { t } = useI18n()
const mobileOpen = ref(false)

const getLabel = (tab: PageTabItem) => {
  if (tab.labelKey) return t(tab.labelKey, tab.fallback || tab.label || tab.id)
  return tab.label || tab.fallback || tab.id
}

const activeLabel = computed(() => {
  const active = props.tabs.find((tab) => tab.id === props.activeId)
  return active ? getLabel(active) : props.ariaLabel
})

const selectTab = (id: string) => {
  emit('select', id)
  mobileOpen.value = false
}

watch(
  () => props.activeId,
  () => {
    mobileOpen.value = false
  }
)
</script>

<style scoped>
.page-tab-bar {
  width: 100%;
  margin: 0 0 2rem;
  padding: 0.45rem;
  border: 1px solid rgba(45, 212, 191, 0.14);
  border-radius: 1.65rem;
  background:
    linear-gradient(135deg, rgba(15, 23, 42, 0.92), rgba(2, 6, 23, 0.78)),
    radial-gradient(circle at 50% 0%, rgba(45, 212, 191, 0.12), transparent 46%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.05),
    0 16px 34px rgba(0, 0, 0, 0.34);
}

.page-tab-bar__mobile-trigger {
  display: none;
}

.page-tab-bar__list {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: center;
  gap: clamp(0.35rem, 0.55vw, 0.7rem);
}

.page-tab-bar__item {
  min-width: 0;
  flex: 0 1 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
  border: 0;
  border-radius: 9999px;
  padding: 0.72rem clamp(0.72rem, 1.08vw, 1.25rem);
  background: rgba(15, 23, 42, 0.82);
  color: rgba(226, 232, 240, 0.88);
  font-size: clamp(0.72rem, 0.73vw, 0.88rem);
  font-weight: 700;
  line-height: 1;
  cursor: pointer;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.04),
    0 10px 22px rgba(0, 0, 0, 0.24);
  transition:
    background-color 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease,
    box-shadow 0.18s ease;
}

.page-tab-bar__item:hover {
  color: #ffffff;
  background: rgba(30, 41, 59, 0.94);
  transform: translateY(-1px);
}

.page-tab-bar__item--active {
  color: #0f172a;
  background: #ffffff;
  box-shadow:
    0 14px 28px rgba(255, 255, 255, 0.12),
    0 10px 24px rgba(0, 0, 0, 0.28);
}

@media (min-width: 901px) {
  .page-tab-bar__list {
    overflow: visible;
  }
}

@media (max-width: 900px) {
  .page-tab-bar {
    padding: 0.45rem;
    border-radius: 1.15rem;
  }

  .page-tab-bar__mobile-trigger {
    display: flex;
    width: 100%;
    align-items: center;
    gap: 0.75rem;
    border: 0;
    border-radius: 0.9rem;
    padding: 0.82rem 0.95rem;
    background: rgba(15, 23, 42, 0.88);
    color: #f8fafc;
    font-size: 0.92rem;
    font-weight: 800;
    cursor: pointer;
    text-align: left;
  }

  .page-tab-bar__hamburger {
    display: inline-flex;
    width: 1.05rem;
    flex: 0 0 auto;
    flex-direction: column;
    gap: 0.22rem;
  }

  .page-tab-bar__hamburger span {
    display: block;
    height: 2px;
    border-radius: 9999px;
    background: currentColor;
    transition:
      opacity 0.18s ease,
      transform 0.18s ease;
  }

  .page-tab-bar--open .page-tab-bar__hamburger span:nth-child(1) {
    transform: translateY(0.34rem) rotate(45deg);
  }

  .page-tab-bar--open .page-tab-bar__hamburger span:nth-child(2) {
    opacity: 0;
  }

  .page-tab-bar--open .page-tab-bar__hamburger span:nth-child(3) {
    transform: translateY(-0.34rem) rotate(-45deg);
  }

  .page-tab-bar__mobile-text {
    min-width: 0;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .page-tab-bar__list {
    display: none;
    flex-direction: column;
    align-items: stretch;
    justify-content: flex-start;
    gap: 0.42rem;
    padding-top: 0.48rem;
  }

  .page-tab-bar--open .page-tab-bar__list {
    display: flex;
  }

  .page-tab-bar__item {
    width: 100%;
    justify-content: flex-start;
    padding: 0.82rem 0.95rem;
    font-size: 0.88rem;
  }
}
</style>
