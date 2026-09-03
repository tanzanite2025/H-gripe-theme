<template>
  <Teleport to="body">
    <Transition name="fitment-catalog-lookup">
      <div
        v-if="modelValue"
        class="fitment-catalog-lookup__mask tz-mobile-safe-modal-mask"
        @click.self="close"
      >
        <div class="fitment-catalog-lookup__backdrop" aria-hidden="true" @click="close" />

        <section
          :id="dialogId"
          class="fitment-catalog-lookup__dialog"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="titleId"
          :aria-describedby="descriptionId"
        >
          <header class="fitment-catalog-lookup__header">
            <div class="fitment-catalog-lookup__heading">
              <p class="fitment-catalog-lookup__eyebrow">
                {{ t('fitmentCatalog.dialog.eyebrow') }}
              </p>
              <h2 :id="titleId">{{ t('fitmentCatalog.dialog.title') }}</h2>
              <p :id="descriptionId">
                {{ t('fitmentCatalog.dialog.description') }}
              </p>
            </div>

            <button
              type="button"
              class="fitment-catalog-lookup__close"
              :aria-label="t('fitmentCatalog.dialog.close')"
              :title="t('fitmentCatalog.dialog.close')"
              @click="close"
            >
              <Icon name="lucide:x" aria-hidden="true" />
            </button>
          </header>

          <nav
            class="fitment-catalog-lookup__tabs"
            role="tablist"
            :aria-label="t('fitmentCatalog.dialog.tabListLabel')"
          >
            <button
              v-for="(tab, index) in tabs"
              :id="tabId(tab.id)"
              :key="tab.id"
              type="button"
              role="tab"
              :aria-selected="activeTab === tab.id"
              :aria-controls="panelId(tab.id)"
              :tabindex="activeTab === tab.id ? 0 : -1"
              :class="{ 'fitment-catalog-lookup__tab--active': activeTab === tab.id }"
              @click="activeTab = tab.id"
              @keydown="handleTabKeydown($event, index)"
            >
              {{ tab.label }}
            </button>
          </nav>

          <div class="fitment-catalog-lookup__body">
            <div
              v-show="activeTab === 'frame'"
              :id="panelId('frame')"
              role="tabpanel"
              :aria-labelledby="tabId('frame')"
            >
              <FitmentFrameLookupPanel />
            </div>
            <div
              v-show="activeTab === 'fork'"
              :id="panelId('fork')"
              role="tabpanel"
              :aria-labelledby="tabId('fork')"
            >
              <FitmentForkLookupPanel />
            </div>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, useId, watch } from 'vue'
import { useI18n } from '#imports'
import FitmentForkLookupPanel from '~/components/fitment-catalog/FitmentForkLookupPanel.vue'
import FitmentFrameLookupPanel from '~/components/fitment-catalog/FitmentFrameLookupPanel.vue'
import { createDialogStackId, useDialogStack } from '~/composables/useDialogStack'
import type { FitmentCatalogResourceType } from '~/types/fitmentCatalog'

const props = withDefaults(defineProps<{
  modelValue: boolean
  defaultTab?: FitmentCatalogResourceType
}>(), {
  defaultTab: 'frame',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
}>()

const { t } = useI18n()
const instanceId = useId()
const dialogStack = useDialogStack()
const dialogStackId = createDialogStackId('fitment-catalog-lookup')
const activeTab = ref<FitmentCatalogResourceType>(props.defaultTab)
const dialogId = `fitment-catalog-dialog-${instanceId}`
const titleId = `fitment-catalog-title-${instanceId}`
const descriptionId = `fitment-catalog-description-${instanceId}`
let unregisterDialogStack: (() => void) | null = null

const tabs = computed(() => [
  {
    id: 'frame' as const,
    label: t('fitmentCatalog.tabs.frame'),
  },
  {
    id: 'fork' as const,
    label: t('fitmentCatalog.tabs.fork'),
  },
])

const tabId = (tab: FitmentCatalogResourceType) => `${dialogId}-tab-${tab}`
const panelId = (tab: FitmentCatalogResourceType) => `${dialogId}-panel-${tab}`

const close = () => {
  emit('update:modelValue', false)
  emit('close')
}

const handleTabKeydown = (event: KeyboardEvent, currentIndex: number) => {
  let nextIndex: number | null = null
  const direction = document.documentElement.dir === 'rtl' ? -1 : 1

  if (event.key === 'ArrowRight') {
    nextIndex = (currentIndex + direction + tabs.value.length) % tabs.value.length
  } else if (event.key === 'ArrowLeft') {
    nextIndex = (currentIndex - direction + tabs.value.length) % tabs.value.length
  } else if (event.key === 'Home') {
    nextIndex = 0
  } else if (event.key === 'End') {
    nextIndex = tabs.value.length - 1
  }

  if (nextIndex === null) return
  event.preventDefault()
  const tab = tabs.value[nextIndex]
  if (!tab) return
  activeTab.value = tab.id
  document.getElementById(tabId(tab.id))?.focus()
}

const syncDialogStack = (isOpen: boolean) => {
  if (isOpen && !unregisterDialogStack) {
    unregisterDialogStack = dialogStack.register(dialogStackId, () => {
      close()
    }, {
      priority: 12000,
    })
    return
  }

  if (!isOpen && unregisterDialogStack) {
    unregisterDialogStack()
    unregisterDialogStack = null
  }
}

watch(
  () => [props.modelValue, props.defaultTab] as const,
  ([isOpen, defaultTab]) => {
    syncDialogStack(isOpen)
    if (defaultTab) activeTab.value = defaultTab
    if (!isOpen) return
    activeTab.value = defaultTab || 'frame'
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  syncDialogStack(false)
})
</script>

<style scoped>
.fitment-catalog-lookup__mask {
  position: fixed;
  inset: 0;
  z-index: 12000;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  padding: 1rem;
}

.fitment-catalog-lookup__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.2);
  -webkit-backdrop-filter: blur(4px);
  backdrop-filter: blur(4px);
}

.fitment-catalog-lookup__dialog {
  position: relative;
  z-index: 1;
  display: flex;
  width: min(58rem, 100%);
  max-height: min(88vh, 52rem);
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  color: var(--tz-text-primary);
  background: var(--tz-card-surface);
  box-shadow: 0 24px 80px rgba(15, 23, 42, 0.18);
}

.fitment-catalog-lookup__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.1rem 1.25rem;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.fitment-catalog-lookup__heading {
  min-width: 0;
}

.fitment-catalog-lookup__eyebrow {
  margin: 0 0 0.25rem;
  color: var(--tz-action-primary);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.fitment-catalog-lookup__heading h2 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1.2rem;
  line-height: 1.3;
}

.fitment-catalog-lookup__heading > p:last-child {
  max-width: 42rem;
  margin: 0.45rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.82rem;
  line-height: 1.55;
}

.fitment-catalog-lookup__close {
  display: inline-grid;
  width: 2.25rem;
  height: 2.25rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.45rem;
  color: var(--tz-text-secondary);
  background: var(--tz-surface-subtle);
  cursor: pointer;
}

.fitment-catalog-lookup__close:hover {
  color: var(--tz-text-primary);
  background: var(--tz-surface-muted);
}

.fitment-catalog-lookup__tabs {
  display: flex;
  overflow-x: auto;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.fitment-catalog-lookup__tabs button {
  position: relative;
  min-height: 3rem;
  flex: 0 0 auto;
  border: 0;
  border-bottom: 2px solid transparent;
  padding: 0.75rem 1rem;
  color: var(--tz-text-secondary);
  background: transparent;
  cursor: pointer;
  font: inherit;
  font-size: 0.85rem;
  font-weight: 800;
  white-space: nowrap;
}

.fitment-catalog-lookup__tabs button:hover {
  color: var(--tz-text-primary);
}

.fitment-catalog-lookup__tab--active {
  border-bottom-color: var(--tz-action-primary) !important;
  color: var(--tz-action-primary) !important;
}

.fitment-catalog-lookup__tabs button:focus-visible,
.fitment-catalog-lookup__close:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: -2px;
}

.fitment-catalog-lookup__body {
  min-height: 0;
  overflow-y: auto;
  padding: 1.1rem 1.25rem 1.25rem;
}

.fitment-catalog-lookup-enter-active,
.fitment-catalog-lookup-leave-active {
  transition: opacity 180ms ease;
}

.fitment-catalog-lookup-enter-active .fitment-catalog-lookup__dialog,
.fitment-catalog-lookup-leave-active .fitment-catalog-lookup__dialog {
  transition: transform 180ms ease, opacity 180ms ease;
}

.fitment-catalog-lookup-enter-from,
.fitment-catalog-lookup-leave-to {
  opacity: 0;
}

.fitment-catalog-lookup-enter-from .fitment-catalog-lookup__dialog,
.fitment-catalog-lookup-leave-to .fitment-catalog-lookup__dialog {
  opacity: 0;
  transform: translateY(0.75rem) scale(0.985);
}

@media (max-width: 767px) {
  .fitment-catalog-lookup__mask {
    align-items: flex-end;
    padding: 0;
  }

  .fitment-catalog-lookup__dialog {
    width: 100%;
    max-height: min(92dvh, 52rem);
    border-right: 0;
    border-bottom: 0;
    border-left: 0;
    border-radius: 0.75rem 0.75rem 0 0;
  }

  .fitment-catalog-lookup__header {
    padding: 1rem;
  }

  .fitment-catalog-lookup__body {
    padding: 1rem;
  }

  .fitment-catalog-lookup-enter-from .fitment-catalog-lookup__dialog,
  .fitment-catalog-lookup-leave-to .fitment-catalog-lookup__dialog {
    transform: translateY(1rem);
  }
}
</style>
