<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import QuickBuyModal from '~/components/QuickBuy.vue'
import WheelsetSelectionAssistantModal from '~/components/WheelsetSelectionAssistantModal.vue'
import WheelsetSelectionAssistantFlow from '~/components/wheelset-selection/WheelsetSelectionAssistantFlow.vue'
import QuickBuyEntryModePanel from '~/components/quick-buy/QuickBuyEntryModePanel.vue'
import { createOverlayInstanceId, useOverlayBackStack } from '~/composables/useOverlayBackStack'
import { useChatWidget } from '~/composables/useChatWidget'
import type { QuickBuyConfig } from '~/utils/quickBuy/types'
import type { WheelsetSelectionRequestDraft } from '~/types/wheelsetSelectionAssistant'

const props = withDefaults(defineProps<{
  config: QuickBuyConfig | null
  anchor?: HTMLElement | null
}>(), {
  anchor: null,
})

const emit = defineEmits<{ close: [] }>()
const overlayBackStack = useOverlayBackStack()
const overlayId = createOverlayInstanceId('quick-buy')
const { openChat } = useChatWidget()

const activeMode = ref<'entry' | 'direct-select' | 'wheelset-selection-assistant'>('entry')
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

  const viewportPadding = 12
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

const openDirectSelect = () => {
  activeMode.value = 'direct-select'
}

const openWheelsetSelectionAssistant = () => {
  activeMode.value = 'wheelset-selection-assistant'
}

const returnToEntryMode = () => {
  activeMode.value = 'entry'
  void updatePopoverPosition()
}

const handleWheelsetSelectionAssistantModelUpdate = (value: boolean) => {
  if (!value) {
    returnToEntryMode()
  }
}

const closeState = () => {
  emit('close')
}

const closeQuickBuyOverlay = async () => {
  const closePromise = overlayBackStack.close(overlayId)
  emit('close')
  await closePromise
}

const openWheelsetSelectionSupportChat = async (draft?: WheelsetSelectionRequestDraft) => {
  // Replace QuickBuy with chat in one overlay transaction so browser-back
  // cannot race the handoff and discard the chat state.
  openChat({
    showAgentList: true,
    source: 'wheelset-selection-assistant',
    pendingSelectionRequest: draft || null,
  })
  await nextTick()
}

const handleClose = () => {
  void closeQuickBuyOverlay()
}

const closeWhenClickingOutside = (event: PointerEvent) => {
  if (activeMode.value !== 'entry') return
  const target = event.target
  if (!(target instanceof Node)) return
  if (popoverRef.value?.contains(target)) return
  if (props.anchor?.contains(target)) return
  handleClose()
}

const closeWhenEscapeIsPressed = (event: KeyboardEvent) => {
  if (activeMode.value !== 'entry') return
  if (event.key === 'Escape') {
    handleClose()
  }
}

onMounted(() => {
  overlayBackStack.open(overlayId, closeState)
  updatePopoverPosition()
  window.addEventListener('resize', updatePopoverPosition)
  window.addEventListener('scroll', updatePopoverPosition, true)
  document.addEventListener('pointerdown', closeWhenClickingOutside, true)
  window.addEventListener('keydown', closeWhenEscapeIsPressed)
})

onBeforeUnmount(() => {
  if (overlayBackStack.isActive(overlayId)) {
    void overlayBackStack.close(overlayId)
  }
  window.removeEventListener('resize', updatePopoverPosition)
  window.removeEventListener('scroll', updatePopoverPosition, true)
  document.removeEventListener('pointerdown', closeWhenClickingOutside, true)
  window.removeEventListener('keydown', closeWhenEscapeIsPressed)
})
</script>

<template>
  <QuickBuyModal
    v-if="activeMode === 'direct-select'"
    :config="config"
    @close="handleClose"
  />

  <WheelsetSelectionAssistantModal
    v-else-if="activeMode === 'wheelset-selection-assistant'"
    :model-value="true"
    source="quick-buy/wheelset-selection-assistant"
    description=""
    :show-steps="false"
    @update:model-value="handleWheelsetSelectionAssistantModelUpdate"
    @close="returnToEntryMode"
  >
    <WheelsetSelectionAssistantFlow
      source="quick-buy/wheelset-selection-assistant"
      @contact-support="openWheelsetSelectionSupportChat"
    />
  </WheelsetSelectionAssistantModal>

  <teleport v-else to="body">
    <Transition name="quickbuy-entry-popover">
      <div
        ref="popoverRef"
        class="quickbuy-entry-router-popover"
        :style="popoverStyle"
        role="dialog"
        aria-modal="false"
      >
        <QuickBuyEntryModePanel
          @direct-select="openDirectSelect"
          @wheelset-selection-assistant="openWheelsetSelectionAssistant"
        />
      </div>
    </Transition>
  </teleport>
</template>

<style scoped>
.quickbuy-entry-popover-enter-active,
.quickbuy-entry-popover-leave-active {
  transition: opacity 160ms ease, transform 160ms ease;
}

.quickbuy-entry-popover-enter-from,
.quickbuy-entry-popover-leave-to {
  opacity: 0;
  transform: translateY(0.25rem) scale(0.98);
}

.quickbuy-entry-router-popover {
  --quickbuy-shell-surface: #050505;
  --quickbuy-panel-surface: var(--tz-card-surface, #111116);
  --quickbuy-panel-surface-soft: #0c0c0e;
  --quickbuy-panel-surface-raised: #17171b;
  --quickbuy-control-surface-raised: #151519;
  --quickbuy-divider: rgba(255, 255, 255, 0.045);
  --quickbuy-entry-accent-edge: color-mix(in srgb, var(--tz-brand-primary, #b5ff6d) 74%, transparent);
  position: fixed;
  z-index: 10002;
  width: min(34rem, calc(100vw - 1.5rem));
  border: 1px solid var(--quickbuy-entry-accent-edge);
  border-radius: 0.875rem;
  background:
    linear-gradient(180deg, #080808 0%, var(--quickbuy-shell-surface) 100%);
  box-shadow:
    0 20px 54px rgba(0, 0, 0, 0.72),
    0 0 0 4px rgba(181, 255, 109, 0.055),
    inset 0 1px 0 rgba(255, 255, 255, 0.025),
    inset 0 0 0 1px rgba(0, 0, 0, 0.72);
}

.quickbuy-entry-router-popover::after {
  position: absolute;
  left: var(--quickbuy-entry-arrow-left, 50%);
  bottom: -0.5rem;
  width: 1rem;
  height: 1rem;
  background: #050505;
  border-right: 1px solid var(--quickbuy-entry-accent-edge);
  border-bottom: 1px solid var(--quickbuy-entry-accent-edge);
  box-shadow:
    8px 8px 18px rgba(0, 0, 0, 0.18);
  content: "";
  transform: translateX(-50%) rotate(45deg);
}

@media (max-width: 767px) {
  .quickbuy-entry-router-popover {
    width: min(22rem, calc(100vw - 1rem));
  }
}
</style>
