<template>
  <Teleport to="body">
    <dialog
      v-if="open"
      ref="dialogElement"
      :id="id"
      class="schwalbe-filter-drawer"
      aria-modal="true"
      :aria-labelledby="titleId"
      @cancel.prevent="close"
      @click.self="close"
    >
      <div class="schwalbe-filter-drawer__surface">
        <header class="schwalbe-filter-drawer__header">
          <h2 :id="titleId">{{ title }}</h2>
          <button
            ref="closeButtonElement"
            type="button"
            class="schwalbe-filter-drawer__close"
            :aria-label="closeLabel"
            @click="close"
          >
            <span aria-hidden="true">×</span>
          </button>
        </header>

        <div class="schwalbe-filter-drawer__body">
          <slot />
        </div>

        <footer class="schwalbe-filter-drawer__footer">
          <button type="button" class="schwalbe-filter-drawer__done" @click="apply">
            {{ showResultsLabel }}
          </button>
        </footer>
      </div>
    </dialog>
  </Teleport>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, useId, watch } from 'vue'
import { createDialogStackId, useDialogStack } from '~/composables/useDialogStack'

const props = defineProps<{
  id: string
  open: boolean
  title: string
  closeLabel: string
  showResultsLabel: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  apply: []
  cancel: []
}>()

const instanceId = useId()
const titleId = `schwalbe-filter-title-${instanceId}`
const dialogElement = ref<HTMLDialogElement | null>(null)
const closeButtonElement = ref<HTMLButtonElement | null>(null)
const dialogStack = useDialogStack()
const dialogStackId = createDialogStackId('schwalbe-tire-catalog-filter')
let unregisterDialogStack: (() => void) | null = null
let returnFocusTarget: HTMLElement | null = null
let previousBodyOverflow = ''
let bodyScrollLockActive = false

const close = () => {
  emit('cancel')
  emit('update:open', false)
}

const apply = () => {
  emit('apply')
  emit('update:open', false)
}

const unlockBackgroundScroll = () => {
  if (typeof document === 'undefined' || !bodyScrollLockActive) return
  document.body.style.overflow = previousBodyOverflow
  bodyScrollLockActive = false
}

watch(() => props.open, async (isOpen) => {
  if (typeof document === 'undefined') return

  if (isOpen) {
    returnFocusTarget = document.activeElement instanceof HTMLElement
      ? document.activeElement
      : null
    previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    bodyScrollLockActive = true

    unregisterDialogStack = dialogStack.register(dialogStackId, () => {
      close()
    }, { priority: 12000 })

    await nextTick()
    const dialog = dialogElement.value
    if (!dialog || !props.open) {
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
}, { flush: 'post' })

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
.schwalbe-filter-drawer {
  box-sizing: border-box;
  width: min(58rem, calc(100vw - 2rem));
  height: min(44rem, calc(100dvh - 2rem));
  max-width: none;
  margin: auto;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  padding: 0;
  color: var(--tz-text-primary);
  background: var(--tz-card-surface);
  box-shadow: 0 24px 80px rgb(15 23 42 / 0.24);
}

.schwalbe-filter-drawer::backdrop {
  background: rgb(15 23 42 / 0.42);
  backdrop-filter: blur(3px);
}

.schwalbe-filter-drawer__surface {
  display: flex;
  min-height: 0;
  height: 100%;
  flex-direction: column;
}

.schwalbe-filter-drawer__header {
  display: flex;
  min-height: 3.7rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid var(--tz-border-subtle);
  padding: 0.7rem 1rem 0.7rem 1.2rem;
}

.schwalbe-filter-drawer__header h2 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1.05rem;
  font-weight: 800;
  line-height: 1.3;
}

.schwalbe-filter-drawer__close {
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

.schwalbe-filter-drawer__close:focus-visible,
.schwalbe-filter-drawer__done:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-filter-drawer__body {
  min-height: 0;
  flex: 1 1 auto;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 0.9rem 1rem;
  scrollbar-width: thin;
}

.schwalbe-filter-drawer__footer {
  display: flex;
  flex: 0 0 auto;
  justify-content: flex-end;
  border-top: 1px solid var(--tz-border-subtle);
  background: var(--tz-card-surface);
  padding: 0.75rem 1rem;
}

.schwalbe-filter-drawer__done {
  min-height: 2.65rem;
  border: 1px solid var(--tz-action-primary);
  border-radius: 0.65rem;
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  padding: 0.55rem 1rem;
  font: inherit;
  font-size: 0.85rem;
  font-weight: 750;
  cursor: pointer;
}

@media (max-width: 760px) {
  .schwalbe-filter-drawer {
    position: fixed;
    inset: auto 0 0;
    box-sizing: border-box;
    width: 100%;
    height: min(92dvh, 60rem);
    margin: 0;
    border-right: 0;
    border-bottom: 0;
    border-left: 0;
    border-radius: 1rem 1rem 0 0;
  }

  .schwalbe-filter-drawer__surface {
    height: 100%;
  }

  .schwalbe-filter-drawer__header {
    padding-top: max(0.7rem, env(safe-area-inset-top));
  }

  .schwalbe-filter-drawer__body {
    padding: 0.75rem;
  }

  .schwalbe-filter-drawer__footer {
    padding: 0.7rem 0.85rem calc(0.7rem + env(safe-area-inset-bottom));
  }

  .schwalbe-filter-drawer__done {
    width: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .schwalbe-filter-drawer__close,
  .schwalbe-filter-drawer__done {
    scroll-behavior: auto;
  }
}
</style>
