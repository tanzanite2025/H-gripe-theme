<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import QuickBuyEntryModePanel from '~/components/quick-buy/QuickBuyEntryModePanel.vue'
import { createOverlayInstanceId, useOverlayBackStack } from '~/composables/useOverlayBackStack'
import type { QuickBuyConfig } from '~/utils/quickBuy/types'

const props = withDefaults(defineProps<{
  config: QuickBuyConfig | null
  anchor?: HTMLElement | null
}>(), {
  anchor: null,
})

const emit = defineEmits<{
  close: []
  'direct-select': []
  'contact-service': []
  'wheelset-selection-assistant': []
}>()
const overlayBackStack = useOverlayBackStack()
const overlayId = createOverlayInstanceId('quick-buy')

const popoverRef = ref<HTMLElement | null>(null)
const popoverStyle = ref<Record<string, string>>({
  left: '50%',
  top: 'auto',
  bottom: '5rem',
  transform: 'translateX(-50%)',
})

const updatePopoverPosition = async () => {
  await nextTick()
  const anchorRect = props.anchor?.getBoundingClientRect()
  const popoverRect = popoverRef.value?.getBoundingClientRect()
  if (!anchorRect || !popoverRect) {
    popoverStyle.value = {
      left: '50%',
      top: 'auto',
      bottom: '5rem',
      transform: 'translateX(-50%)',
    }
    return
  }

  const viewportPadding = window.innerWidth <= 767 ? 2 : 12
  const gap = 12
  const anchorCenter = anchorRect.left + anchorRect.width / 2
  const preferredLeft = anchorCenter - popoverRect.width / 2
  const maxLeft = window.innerWidth - popoverRect.width - viewportPadding
  const left = Math.min(Math.max(preferredLeft, viewportPadding), Math.max(viewportPadding, maxLeft))
  const top = Math.max(viewportPadding, anchorRect.top - popoverRect.height - gap)
  const arrowLeft = anchorCenter - left

  popoverStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
    '--quickbuy-entry-arrow-left': `${Math.min(Math.max(arrowLeft, 24), popoverRect.width - 24)}px`,
    transform: 'none',
  }
}

let isClosing = false

const closeState = () => {
  emit('close')
}

const closeQuickBuyOverlay = async () => {
  if (isClosing) return
  isClosing = true

  if (overlayBackStack.isActive(overlayId)) {
    await overlayBackStack.close(overlayId)
    return
  }

  emit('close')
}

const handleClose = () => {
  void closeQuickBuyOverlay()
}

const handleDirectSelect = () => {
  emit('direct-select')
}

const handleContactService = () => {
  emit('contact-service')
}

const handleWheelsetSelectionAssistant = () => {
  emit('wheelset-selection-assistant')
}

const closeWhenClickingOutside = (event: MouseEvent) => {
  const target = event.target
  if (!(target instanceof Node)) return
  if (popoverRef.value?.contains(target)) return
  if (props.anchor?.contains(target)) return
  handleClose()
}

const closeWhenEscapeIsPressed = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    handleClose()
  }
}

onMounted(() => {
  overlayBackStack.open(overlayId, closeState)
  updatePopoverPosition()
  window.addEventListener('resize', updatePopoverPosition)
  window.addEventListener('scroll', updatePopoverPosition, true)
  document.addEventListener('click', closeWhenClickingOutside, true)
  window.addEventListener('keydown', closeWhenEscapeIsPressed)
})

onBeforeUnmount(() => {
  if (overlayBackStack.isActive(overlayId)) {
    void overlayBackStack.close(overlayId)
  }
  window.removeEventListener('resize', updatePopoverPosition)
  window.removeEventListener('scroll', updatePopoverPosition, true)
  document.removeEventListener('click', closeWhenClickingOutside, true)
  window.removeEventListener('keydown', closeWhenEscapeIsPressed)
})
</script>

<template>
  <teleport to="body">
      <div
        class="quickbuy-entry-router-layer"
        role="presentation"
        @click.self="handleClose"
      >
        <div
          ref="popoverRef"
          class="quickbuy-entry-router-popover"
          :style="popoverStyle"
          role="dialog"
          aria-modal="false"
          @click.stop
        >
          <QuickBuyEntryModePanel
            @direct-select="handleDirectSelect"
            @contact-service="handleContactService"
            @wheelset-selection-assistant="handleWheelsetSelectionAssistant"
          />
        </div>
      </div>
  </teleport>
</template>

<style scoped>
.quickbuy-entry-router-layer {
  position: fixed;
  inset: 0;
  z-index: 10002;
  background: transparent;
  pointer-events: auto;
}

.quickbuy-entry-router-popover {
  --quickbuy-shell-surface: var(--tz-card-surface);
  --quickbuy-panel-surface: var(--tz-card-surface);
  --quickbuy-panel-surface-soft: var(--tz-surface-subtle);
  --quickbuy-panel-surface-raised: var(--tz-surface-muted);
  --quickbuy-control-surface-raised: var(--tz-surface-subtle);
  --quickbuy-divider: var(--tz-border-subtle);
  --quickbuy-entry-accent-edge: color-mix(in srgb, var(--tz-site-accent, #059669) 74%, transparent);
  position: fixed;
  z-index: 1;
  width: min(34rem, calc(100vw - 1.5rem));
  box-sizing: border-box;
  border: 1px solid var(--quickbuy-entry-accent-edge);
  border-radius: 0.875rem;
  background: var(--quickbuy-shell-surface);
  box-shadow:
    0 20px 54px rgba(0, 0, 0, 0.72),
    0 0 0 4px rgba(5, 150, 105, 0.055),
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 0 0 1px var(--tz-border-subtle);
}

.quickbuy-entry-router-popover::after {
  position: absolute;
  left: var(--quickbuy-entry-arrow-left, 50%);
  bottom: -0.5rem;
  width: 1rem;
  height: 1rem;
  background: var(--quickbuy-shell-surface);
  border-right: 1px solid var(--quickbuy-entry-accent-edge);
  border-bottom: 1px solid var(--quickbuy-entry-accent-edge);
  box-shadow:
    8px 8px 18px rgba(0, 0, 0, 0.18);
  content: "";
  transform: translateX(-50%) rotate(45deg);
}

@media (max-width: 767px) {
  .quickbuy-entry-router-popover {
    width: calc(100vw - 4px);
  }
}
</style>
