<template>
  <nav class="schwalbe-wheel-size-tabs" :aria-label="label">
    <div class="schwalbe-wheel-size-tabs__desktop">
      <button
        type="button"
        class="schwalbe-wheel-size-tabs__tab"
        :class="{ 'schwalbe-wheel-size-tabs__tab--active': selectedWheelSizeKeys.length === 0 }"
        :aria-pressed="selectedWheelSizeKeys.length === 0"
        @click="emit('select', null)"
      >
        {{ allLabel }}
      </button>
      <button
        v-for="option in options"
        :key="option.value"
        type="button"
        class="schwalbe-wheel-size-tabs__tab"
        :class="{ 'schwalbe-wheel-size-tabs__tab--active': selectedWheelSizeKeys.includes(option.value) }"
        :aria-pressed="selectedWheelSizeKeys.includes(option.value)"
        @click="emit('select', option.value)"
      >
        {{ optionLabel(option) }}
      </button>
    </div>

    <div class="schwalbe-wheel-size-tabs__mobile">
      <button
        ref="mobileTriggerElement"
        type="button"
        class="schwalbe-wheel-size-tabs__mobile-trigger"
        :aria-label="mobileTriggerLabel"
        aria-haspopup="dialog"
        :aria-expanded="mobileDialogOpen"
        :aria-controls="mobileDialogId"
        @click="openMobileDialog"
      >
        <span class="schwalbe-wheel-size-tabs__mobile-value">{{ mobileSelectedLabel }}</span>
        <Icon name="lucide:chevron-down" aria-hidden="true" />
      </button>
    </div>
  </nav>

  <Teleport to="body">
    <dialog
      v-if="mobileDialogOpen"
      ref="dialogElement"
      :id="mobileDialogId"
      class="schwalbe-wheel-size-dialog"
      aria-modal="true"
      :aria-labelledby="mobileDialogTitleId"
      @cancel.prevent="closeMobileDialog"
      @click.self="closeMobileDialog"
    >
      <div class="schwalbe-wheel-size-dialog__surface">
        <header class="schwalbe-wheel-size-dialog__header">
          <h2 :id="mobileDialogTitleId">{{ dialogTitle }}</h2>
          <button
            ref="closeButtonElement"
            type="button"
            class="schwalbe-wheel-size-dialog__close"
            :aria-label="closeLabel"
            :title="closeLabel"
            @click="closeMobileDialog"
          >
            <span aria-hidden="true">×</span>
          </button>
        </header>

        <div class="schwalbe-wheel-size-dialog__body">
          <p v-if="selectedWheelSizeKeys.length > 1" class="schwalbe-wheel-size-dialog__status">
            {{ multipleLabel }}
          </p>
          <div class="schwalbe-wheel-size-dialog__options" role="listbox" :aria-label="dialogTitle">
            <button
              type="button"
              role="option"
              class="schwalbe-wheel-size-dialog__option"
              :class="{ 'schwalbe-wheel-size-dialog__option--active': selectedWheelSizeKeys.length === 0 }"
              :aria-selected="selectedWheelSizeKeys.length === 0"
              @click="selectMobileWheelSize(null)"
            >
              <span>{{ allLabel }}</span>
              <Icon v-if="selectedWheelSizeKeys.length === 0" name="lucide:check" aria-hidden="true" />
            </button>
            <button
              v-for="option in options"
              :key="option.value"
              type="button"
              role="option"
              class="schwalbe-wheel-size-dialog__option"
              :class="{ 'schwalbe-wheel-size-dialog__option--active': selectedWheelSizeKeys.includes(option.value) }"
              :aria-selected="selectedWheelSizeKeys.includes(option.value)"
              @click="selectMobileWheelSize(option.value)"
            >
              <span>{{ optionLabel(option) }}</span>
              <Icon v-if="selectedWheelSizeKeys.includes(option.value)" name="lucide:check" aria-hidden="true" />
            </button>
          </div>
        </div>
      </div>
    </dialog>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, useId, watch } from 'vue'
import type { SchwalbeTireCatalogWheelSizeOption } from '~/data/tireguides/schwalbeTireCatalogFilterModel'
import { createDialogStackId, useDialogStack } from '~/composables/useDialogStack'

const props = defineProps<{
  label: string
  allLabel: string
  multipleLabel: string
  dialogTitle: string
  closeLabel: string
  optionTemplate: string
  options: readonly SchwalbeTireCatalogWheelSizeOption[]
  selectedWheelSizeKeys: readonly string[]
}>()

const emit = defineEmits<{
  select: [wheelSizeKey: string | null]
}>()

const instanceId = useId()
const mobileDialogId = `schwalbe-wheel-size-dialog-${instanceId}`
const mobileDialogTitleId = `${mobileDialogId}-title`
const mobileDialogOpen = ref(false)
const dialogElement = ref<HTMLDialogElement | null>(null)
const mobileTriggerElement = ref<HTMLButtonElement | null>(null)
const closeButtonElement = ref<HTMLButtonElement | null>(null)
const dialogStack = useDialogStack()
const dialogStackId = createDialogStackId('schwalbe-wheel-size-picker')
let unregisterDialogStack: (() => void) | null = null
let returnFocusTarget: HTMLElement | null = null
let previousBodyOverflow = ''
let bodyScrollLockActive = false

const optionLabel = (option: SchwalbeTireCatalogWheelSizeOption) => (
  // The template is localized by the page. The value itself remains the
  // stable wheel-diameter + BSD key used by the URL and API.
  props.optionTemplate
    .replace('{diameter}', option.wheelDiameterIn)
    .replace('{bsd}', String(option.beadSeatDiameterMm))
)

const mobileSelectedLabel = computed(() => {
  if (props.selectedWheelSizeKeys.length > 1) return props.multipleLabel
  const selectedValue = props.selectedWheelSizeKeys[0]
  if (!selectedValue) return props.allLabel
  const selectedOption = props.options.find(option => option.value === selectedValue)
  if (selectedOption) return optionLabel(selectedOption)
  return selectedValue
})

const mobileTriggerLabel = computed(() => `${props.label}: ${mobileSelectedLabel.value}`)

const openMobileDialog = () => {
  mobileDialogOpen.value = true
}

const closeMobileDialog = () => {
  mobileDialogOpen.value = false
}

const selectMobileWheelSize = (wheelSizeKey: string | null) => {
  emit('select', wheelSizeKey)
  closeMobileDialog()
}

const unlockBackgroundScroll = () => {
  if (typeof document === 'undefined' || !bodyScrollLockActive) return
  document.body.style.overflow = previousBodyOverflow
  bodyScrollLockActive = false
}

watch(mobileDialogOpen, async (isOpen) => {
  if (typeof document === 'undefined') return

  if (isOpen) {
    returnFocusTarget = document.activeElement instanceof HTMLElement
      ? document.activeElement
      : mobileTriggerElement.value
    previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    unregisterDialogStack = dialogStack.register(dialogStackId, () => {
      closeMobileDialog()
    }, { priority: 12000 })

    await nextTick()
    const dialog = dialogElement.value
    if (!dialog || !mobileDialogOpen.value) {
      unregisterDialogStack?.()
      unregisterDialogStack = null
      unlockBackgroundScroll()
      return
    }
    if (!dialog.open) dialog.showModal()
    closeButtonElement.value?.focus()
    return
  }

  unregisterDialogStack?.()
  unregisterDialogStack = null
  unlockBackgroundScroll()

  const target = returnFocusTarget
  returnFocusTarget = null
  await nextTick()
  target?.focus({ preventScroll: true })
})

onBeforeUnmount(() => {
  unregisterDialogStack?.()
  unregisterDialogStack = null
  const dialog = dialogElement.value
  if (dialog?.open) dialog.close()
  unlockBackgroundScroll()
  const target = returnFocusTarget
  returnFocusTarget = null
  target?.focus({ preventScroll: true })
})
</script>

<style scoped>
.schwalbe-wheel-size-tabs {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.schwalbe-wheel-size-tabs__desktop {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  gap: 0.35rem;
  padding: 0 0.1rem 0.45rem;
}

.schwalbe-wheel-size-tabs__mobile {
  display: none;
}

.schwalbe-wheel-size-tabs__tab {
  min-height: 2.35rem;
  flex: 0 0 auto;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 999px;
  background: var(--tz-card-surface);
  color: var(--tz-text-secondary);
  padding: 0.45rem 0.8rem;
  font: inherit;
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
  cursor: pointer;
}

.schwalbe-wheel-size-tabs__tab:hover {
  border-color: var(--tz-border-strong);
  color: var(--tz-text-primary);
}

.schwalbe-wheel-size-tabs__tab:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-wheel-size-tabs__tab--active {
  border-color: var(--tz-action-primary);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
}

@media (max-width: 760.5px) {
  .schwalbe-wheel-size-tabs__desktop {
    display: none;
  }

  .schwalbe-wheel-size-tabs__mobile {
    display: block;
    padding: 0.1rem 0 0.45rem;
  }

  .schwalbe-wheel-size-tabs__mobile-trigger {
    display: flex;
    box-sizing: border-box;
    width: 100%;
    max-width: 100%;
    min-width: 0;
    min-height: 2.5rem;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    border: 1px solid var(--tz-border-strong);
    border-radius: 0.7rem;
    background: var(--tz-card-surface);
    color: var(--tz-text-primary);
    padding: 0.55rem 0.75rem;
    font: inherit;
    font-size: 0.84rem;
    font-weight: 700;
    text-align: left;
    cursor: pointer;
  }

  .schwalbe-wheel-size-tabs__mobile-value {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .schwalbe-wheel-size-tabs__mobile-trigger :deep(svg) {
    width: 1rem;
    height: 1rem;
    flex: 0 0 auto;
  }

  .schwalbe-wheel-size-tabs__mobile-trigger:focus-visible {
    outline: 2px solid var(--tz-action-primary);
    outline-offset: 2px;
  }
}

.schwalbe-wheel-size-dialog {
  box-sizing: border-box;
  width: min(32rem, calc(100vw - 1rem));
  height: min(80vh, 38rem);
  height: min(80dvh, 38rem);
  max-width: none;
  max-height: calc(100vh - 1rem);
  max-height: calc(100dvh - 1rem);
  margin: auto;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  padding: 0;
  color: var(--tz-text-primary);
  background: var(--tz-card-surface);
  box-shadow: 0 24px 80px rgb(15 23 42 / 0.24);
}

.schwalbe-wheel-size-dialog::backdrop {
  background: rgb(15 23 42 / 0.42);
  backdrop-filter: blur(3px);
}

.schwalbe-wheel-size-dialog__surface {
  display: flex;
  min-height: 0;
  height: 100%;
  flex-direction: column;
}

.schwalbe-wheel-size-dialog__header {
  display: flex;
  min-height: 3.7rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid var(--tz-border-subtle);
  padding: 0.7rem 1rem 0.7rem 1.2rem;
}

.schwalbe-wheel-size-dialog__header h2 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1.05rem;
  font-weight: 800;
  line-height: 1.3;
}

.schwalbe-wheel-size-dialog__close {
  display: inline-grid;
  width: 2.35rem;
  height: 2.35rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 999px;
  background: var(--tz-surface-subtle);
  color: var(--tz-text-primary);
  font: inherit;
  font-size: 1.55rem;
  line-height: 1;
  cursor: pointer;
}

.schwalbe-wheel-size-dialog__close:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-wheel-size-dialog__body {
  min-height: 0;
  flex: 1 1 auto;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 0.85rem;
  scrollbar-width: thin;
}

.schwalbe-wheel-size-dialog__status {
  margin: 0 0 0.65rem;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
}

.schwalbe-wheel-size-dialog__options {
  display: grid;
  gap: 0.4rem;
}

.schwalbe-wheel-size-dialog__option {
  display: flex;
  width: 100%;
  min-height: 2.8rem;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.7rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-primary);
  padding: 0.55rem 0.75rem;
  font: inherit;
  font-size: 0.84rem;
  font-weight: 700;
  text-align: left;
  cursor: pointer;
}

.schwalbe-wheel-size-dialog__option:hover,
.schwalbe-wheel-size-dialog__option--active {
  border-color: var(--tz-action-primary);
  background: var(--tz-site-accent-selected-surface, #d1fae5);
}

.schwalbe-wheel-size-dialog__option:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-wheel-size-dialog__option > span {
  min-width: 0;
  overflow-wrap: anywhere;
}

.schwalbe-wheel-size-dialog__option :deep(svg) {
  width: 1rem;
  height: 1rem;
  flex: 0 0 auto;
  color: var(--tz-action-primary);
}

@media (max-width: 760.5px) {
  .schwalbe-wheel-size-dialog {
    width: min(32rem, calc(100vw - 1rem));
    height: min(80vh, 38rem);
    height: min(80dvh, 38rem);
    max-height: calc(100vh - 1rem);
    max-height: calc(100dvh - 1rem);
    border-radius: 1rem;
  }
}

</style>
