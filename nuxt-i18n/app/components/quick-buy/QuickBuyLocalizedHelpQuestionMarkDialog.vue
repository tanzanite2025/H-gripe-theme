<template>
  <span
    v-if="hasLocalizedHelpContent"
    ref="quickBuyLocalizedHelpDialogRoot"
    class="quickbuy-localized-help-dialog-root"
  >
    <button
      class="quickbuy-localized-help-question-trigger"
      type="button"
      :aria-label="triggerAriaLabel || dialogTitle"
      :title="triggerAriaLabel || dialogTitle"
      :aria-expanded="isQuickBuyLocalizedHelpDialogOpen"
      @click.stop="toggleQuickBuyLocalizedHelpDialog"
    >
      <span aria-hidden="true">?</span>
    </button>

    <Teleport to="body">
      <Transition name="quickbuy-localized-help-dialog-fade">
        <section
          v-if="isQuickBuyLocalizedHelpDialogOpen"
          ref="quickBuyLocalizedHelpDialogRef"
          class="quickbuy-localized-help-dialog"
          :style="quickBuyLocalizedHelpDialogStyle"
          role="dialog"
          :aria-modal="false"
          :aria-label="dialogTitle"
          @click.stop
        >
          <header class="quickbuy-localized-help-dialog__header">
            <h3 class="quickbuy-localized-help-dialog__title">
              {{ dialogTitle }}
            </h3>
            <button
              class="quickbuy-localized-help-dialog__close tz-global-close-btn"
              type="button"
              :aria-label="closeLabel"
              :title="closeLabel"
              @click="closeQuickBuyLocalizedHelpDialog"
            >
              <Icon name="lucide:x" class="h-4 w-4" aria-hidden="true" />
            </button>
          </header>

          <div class="quickbuy-localized-help-dialog__content">
            {{ normalizedLocalizedHelpContent }}
          </div>
        </section>
      </Transition>
    </Teleport>
  </span>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  title: string
  content: string
  triggerAriaLabel?: string
  closeLabel?: string
}>(), {
  triggerAriaLabel: '',
  closeLabel: 'Close',
})

const isQuickBuyLocalizedHelpDialogOpen = ref(false)
const quickBuyLocalizedHelpDialogRoot = ref<HTMLElement | null>(null)
const quickBuyLocalizedHelpDialogRef = ref<HTMLElement | null>(null)
const quickBuyLocalizedHelpDialogStyle = ref<Record<string, string>>({})

const normalizedLocalizedHelpContent = computed(() => String(props.content || '').trim())
const hasLocalizedHelpContent = computed(() => normalizedLocalizedHelpContent.value.length > 0)
const dialogTitle = computed(() => props.title || props.triggerAriaLabel || 'Help')

const toggleQuickBuyLocalizedHelpDialog = () => {
  if (!hasLocalizedHelpContent.value) return
  isQuickBuyLocalizedHelpDialogOpen.value = !isQuickBuyLocalizedHelpDialogOpen.value
}

const closeQuickBuyLocalizedHelpDialog = () => {
  isQuickBuyLocalizedHelpDialogOpen.value = false
}

const updateQuickBuyLocalizedHelpDialogPosition = async () => {
  await nextTick()
  if (typeof window === 'undefined') return

  const dialog = quickBuyLocalizedHelpDialogRef.value
  const root = quickBuyLocalizedHelpDialogRoot.value
  if (!dialog || !root) return

  const rootRect = root.getBoundingClientRect()
  const boundary = root.closest<HTMLElement>('[role="dialog"]')
  const boundaryRect = boundary?.getBoundingClientRect()
  const viewportInset = window.matchMedia('(max-width: 767px)').matches ? 8 : 2
  const boundaryInset = 8
  const gap = 8
  const isMobile = window.matchMedia('(max-width: 767px)').matches
  const dialogRect = dialog.getBoundingClientRect()
  const preferredTop = rootRect.bottom + gap
  const boundaryTop = Math.max(viewportInset, boundaryRect?.top ?? viewportInset)
  const boundaryRight = Math.min(window.innerWidth - viewportInset, boundaryRect?.right ?? window.innerWidth - viewportInset)
  const boundaryBottom = Math.min(window.innerHeight - viewportInset, boundaryRect?.bottom ?? window.innerHeight - viewportInset)
  const boundaryLeft = Math.max(viewportInset, boundaryRect?.left ?? viewportInset)
  const availableWidth = Math.max(0, boundaryRight - boundaryLeft)
  const width = Math.min(dialogRect.width, availableWidth - boundaryInset * 2)
  const leftBoundary = boundaryLeft + boundaryInset
  const rightBoundary = boundaryRight - boundaryInset
  const maxTop = boundaryBottom - dialogRect.height - boundaryInset
  const top = Math.max(
    boundaryTop + boundaryInset,
    Math.min(preferredTop, maxTop),
  )

  if (isMobile) {
    quickBuyLocalizedHelpDialogStyle.value = {
      top: `${top}px`,
    }
    return
  }

  const preferredLeft = rootRect.left
  const maxLeft = rightBoundary - width
  const left = Math.max(
    leftBoundary,
    Math.min(preferredLeft, maxLeft),
  )

  quickBuyLocalizedHelpDialogStyle.value = {
    top: `${top}px`,
    right: 'auto',
    left: `${left}px`,
    width: `${width}px`,
  }
}

const closeQuickBuyLocalizedHelpDialogWhenEscapeIsPressed = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    closeQuickBuyLocalizedHelpDialog()
  }
}

const closeQuickBuyLocalizedHelpDialogWhenClickingOutside = (event: PointerEvent) => {
  const target = event.target
  if (
    target instanceof Node
    && !quickBuyLocalizedHelpDialogRoot.value?.contains(target)
    && !quickBuyLocalizedHelpDialogRef.value?.contains(target)
  ) {
    closeQuickBuyLocalizedHelpDialog()
  }
}

watch(isQuickBuyLocalizedHelpDialogOpen, (isOpen) => {
  if (typeof window === 'undefined') return
  if (isOpen) {
    window.addEventListener('keydown', closeQuickBuyLocalizedHelpDialogWhenEscapeIsPressed)
    document.addEventListener('pointerdown', closeQuickBuyLocalizedHelpDialogWhenClickingOutside, true)
    window.addEventListener('resize', updateQuickBuyLocalizedHelpDialogPosition)
    window.addEventListener('scroll', updateQuickBuyLocalizedHelpDialogPosition, true)
    void updateQuickBuyLocalizedHelpDialogPosition()
  } else {
    window.removeEventListener('keydown', closeQuickBuyLocalizedHelpDialogWhenEscapeIsPressed)
    document.removeEventListener('pointerdown', closeQuickBuyLocalizedHelpDialogWhenClickingOutside, true)
    window.removeEventListener('resize', updateQuickBuyLocalizedHelpDialogPosition)
    window.removeEventListener('scroll', updateQuickBuyLocalizedHelpDialogPosition, true)
    quickBuyLocalizedHelpDialogStyle.value = {}
  }
})

onBeforeUnmount(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('keydown', closeQuickBuyLocalizedHelpDialogWhenEscapeIsPressed)
    document.removeEventListener('pointerdown', closeQuickBuyLocalizedHelpDialogWhenClickingOutside, true)
    window.removeEventListener('resize', updateQuickBuyLocalizedHelpDialogPosition)
    window.removeEventListener('scroll', updateQuickBuyLocalizedHelpDialogPosition, true)
  }
})
</script>

<style scoped>
.quickbuy-localized-help-dialog-root {
  position: relative;
  display: inline-flex;
  flex: 0 0 auto;
}

.quickbuy-localized-help-question-trigger {
  display: inline-grid;
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  place-items: center;
  border: 0;
  border-radius: 999px;
  color: rgba(5, 150, 105, 0.96);
  background:
    linear-gradient(180deg, rgba(5, 150, 105, 0.13), rgba(5, 150, 105, 0.06)),
    var(--quickbuy-control-surface-raised, var(--tz-surface-subtle));
  box-shadow:
    0 0 0 0 rgba(5, 150, 105, 0.18),
    inset 0 0 0 1px var(--tz-border-subtle);
  font-size: 0.9375rem;
  font-weight: 800;
  line-height: 1;
  animation: quickbuy-localized-help-question-nudge 4.8s ease-in-out infinite;
  transition: background-color 160ms ease, color 160ms ease;
}

.quickbuy-localized-help-question-trigger:hover {
  color: white;
  background:
    linear-gradient(180deg, rgba(5, 150, 105, 0.2), rgba(5, 150, 105, 0.09)),
    var(--quickbuy-control-surface-raised, var(--tz-surface-subtle));
}

.quickbuy-localized-help-dialog {
  position: fixed;
  top: calc(100% + 0.5rem);
  right: 0;
  z-index: 10060;
  display: flex;
  width: min(28rem, calc(100vw - 2rem));
  max-height: min(72vh, 34rem);
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
  min-width: 0;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.875rem;
  color: var(--tz-text-primary);
  background: var(--tz-card-surface);
  box-shadow:
    0 22px 60px rgba(20, 32, 43, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.quickbuy-localized-help-dialog__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  box-shadow: inset 0 -1px 0 var(--quickbuy-divider, rgba(255, 255, 255, 0.045));
}

.quickbuy-localized-help-dialog__title {
  min-width: 0;
  margin: 0;
  overflow: hidden;
  font-size: 0.9375rem;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-localized-help-dialog__content {
  overflow-y: auto;
  padding: 1rem;
  color: var(--tz-text-secondary);
  font-size: 0.875rem;
  line-height: 1.65;
  white-space: pre-line;
}

.quickbuy-localized-help-dialog-fade-enter-active,
.quickbuy-localized-help-dialog-fade-leave-active {
  transition: opacity 180ms ease;
}

.quickbuy-localized-help-dialog-fade-enter-from,
.quickbuy-localized-help-dialog-fade-leave-to {
  opacity: 0;
}

@keyframes quickbuy-localized-help-question-nudge {
  0%,
  70%,
  100% {
    box-shadow: 0 0 0 0 rgba(5, 150, 105, 0);
    transform: translateY(0) scale(1);
  }

  76% {
    box-shadow: 0 0 0 0.35rem rgba(5, 150, 105, 0.08);
    transform: translateY(-0.125rem) scale(1.06);
  }

  82% {
    box-shadow: 0 0 0 0 rgba(5, 150, 105, 0);
    transform: translateY(0) scale(1);
  }
}

@media (prefers-reduced-motion: reduce) {
  .quickbuy-localized-help-question-trigger {
    animation: none;
  }
}

@media (max-width: 767px) {
  .quickbuy-localized-help-question-trigger {
    width: 1.875rem;
    height: 1.875rem;
  }

  .quickbuy-localized-help-dialog {
    position: fixed;
    top: 0;
    right: var(--tz-mobile-dialog-inset, 2px);
    left: var(--tz-mobile-dialog-inset, 2px);
    width: auto;
    max-width: none;
    max-height: min(60vh, 30rem);
    border-radius: 0.75rem;
  }
}
</style>
