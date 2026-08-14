<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import QuickBuyModal from '~/components/QuickBuy.vue'
import QuickBuyEntryModePanel from '~/components/quick-buy/QuickBuyEntryModePanel.vue'
import type { QuickBuyConfig } from '~/utils/quickBuy/types'

const props = withDefaults(defineProps<{
  config: QuickBuyConfig | null
  anchor?: HTMLElement | null
}>(), {
  anchor: null,
})

const emit = defineEmits<{ close: [] }>()

const activeMode = ref<'entry' | 'direct-select'>('entry')
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

const handleClose = () => {
  emit('close')
}

const closeWhenClickingOutside = (event: PointerEvent) => {
  if (activeMode.value !== 'entry') return
  const target = event.target
  if (!(target instanceof Node)) return
  if (popoverRef.value?.contains(target)) return
  if (props.anchor?.contains(target)) return
  emit('close')
}

const closeWhenEscapeIsPressed = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    emit('close')
  }
}

onMounted(() => {
  updatePopoverPosition()
  window.addEventListener('resize', updatePopoverPosition)
  window.addEventListener('scroll', updatePopoverPosition, true)
  document.addEventListener('pointerdown', closeWhenClickingOutside, true)
  window.addEventListener('keydown', closeWhenEscapeIsPressed)
})

onBeforeUnmount(() => {
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

  <teleport v-else to="body">
    <Transition name="quickbuy-entry-popover">
      <div
        ref="popoverRef"
        class="quickbuy-entry-router-popover"
        :style="popoverStyle"
        role="dialog"
        aria-modal="false"
      >
        <QuickBuyEntryModePanel @direct-select="openDirectSelect" />
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
