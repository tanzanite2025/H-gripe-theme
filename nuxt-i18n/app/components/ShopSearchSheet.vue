<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-300 ease-out"
      leave-active-class="transition-opacity duration-200 ease-in"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isOpen"
        class="fixed inset-0 z-[14000] flex items-end justify-center p-0 tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
        @click.self="close"
      >
        <Transition
          enter-active-class="transition-all duration-300 ease-out"
          leave-active-class="transition-all duration-200 ease-in"
          enter-from-class="translate-y-full opacity-0"
          enter-to-class="translate-y-0 opacity-100"
          leave-from-class="translate-y-0 opacity-100"
          leave-to-class="translate-y-full opacity-0"
          appear
        >
          <div
            v-if="isOpen"
            class="shop-search-sheet-shell w-full"
          >
            <section
            class="shop-search-sheet-panel tz-mobile-dialog-surface tz-surface-card relative pointer-events-auto w-full max-w-none tz-mobile-safe-full-height md:h-[82vh] md:max-h-[900px] rounded-none flex flex-col overflow-hidden"
              :class="{ 'shop-search-sheet-panel--dragging': isDraggingPanel }"
              ref="panelRef"
              aria-modal="true"
              role="dialog"
              :aria-label="$t('sidebar.searchProducts', 'Search Products')"
            >
              <div
                class="shop-search-sheet-drag-edge absolute inset-x-0 top-0 z-30 h-4"
                aria-hidden="true"
                @pointerdown="startPanelDrag"
              ></div>

              <header class="relative z-10 flex items-center justify-between px-4 md:px-6 py-4">
                <h2 class="text-lg md:text-xl font-semibold text-[#059669]">
                  {{ $t('sidebar.searchProducts', 'Search Products') }}
                </h2>
                <button
                  type="button"
                  class="tz-global-close-btn"
                  :aria-label="$t('common.close', 'Close')"
                  @click="close"
                >
                  <svg class="w-5 h-5 tz-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </header>

              <div class="relative z-10 flex-1 overflow-y-auto px-4 md:px-6 py-4">
                <ProductSearchPanel @search="handleSearch" />
              </div>
            </section>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import ProductSearchPanel from '~/components/ProductSearchPanel.vue'
import { setSidebarHandlesHidden } from '~/utils/sidebarHandles'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'
import type { ShopSearchPayload } from '~/composables/useShopSearchSheet'

const { isOpen, close, submit } = useShopSearchSheet()

const SIDEBAR_TOKEN_SHOP_SEARCH = 'shop-search-sheet'

watch(
  isOpen,
  (open) => {
    setSidebarHandlesHidden(SIDEBAR_TOKEN_SHOP_SEARCH, open)
  },
  { immediate: true }
)

const handleSearch = async (payload: ShopSearchPayload) => {
  await submit(payload)
}

const dragOffsetY = ref(0)
const isDraggingPanel = ref(false)
const panelRef = ref<HTMLElement | null>(null)
let activeDragPointerId: number | null = null
let dragStartY = 0

const setPanelDragOffset = (offset: number) => {
  dragOffsetY.value = offset
  panelRef.value?.style.setProperty('--shop-search-sheet-drag-y', `${offset}px`)
}

const stopPanelDragListeners = () => {
  if (typeof window === 'undefined') return
  window.removeEventListener('pointermove', onPanelDragMove)
  window.removeEventListener('pointerup', finishPanelDrag)
  window.removeEventListener('pointercancel', cancelPanelDrag)
}

const resetPanelDrag = () => {
  stopPanelDragListeners()
  activeDragPointerId = null
  dragStartY = 0
  setPanelDragOffset(0)
  panelRef.value?.style.removeProperty('transition')
  isDraggingPanel.value = false
}

const onPanelDragMove = (event: PointerEvent) => {
  if (activeDragPointerId !== null && event.pointerId !== activeDragPointerId) return

  event.preventDefault()
  const maxOffset = typeof window === 'undefined' ? 360 : window.innerHeight * 0.65
  setPanelDragOffset(Math.min(Math.max(event.clientY - dragStartY, 0), maxOffset))
}

const finishPanelDrag = (event: PointerEvent) => {
  if (activeDragPointerId !== null && event.pointerId !== activeDragPointerId) return

  const shouldClose = dragOffsetY.value > 96
  resetPanelDrag()

  if (shouldClose) {
    close()
  }
}

const cancelPanelDrag = () => {
  resetPanelDrag()
}

const startPanelDrag = (event: PointerEvent) => {
  if (event.pointerType === 'mouse' && event.button !== 0) return
  if (typeof window === 'undefined') return

  event.preventDefault()
  const target = event.currentTarget as HTMLElement | null
  target?.setPointerCapture?.(event.pointerId)
  activeDragPointerId = event.pointerId
  dragStartY = event.clientY
  panelRef.value?.style.setProperty('transition', 'none')
  setPanelDragOffset(0)
  isDraggingPanel.value = true
  window.addEventListener('pointermove', onPanelDragMove)
  window.addEventListener('pointerup', finishPanelDrag)
  window.addEventListener('pointercancel', cancelPanelDrag)
}

watch(isOpen, (open) => {
  if (!open) {
    resetPanelDrag()
  }
})

const onKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    close()
  }
}

onMounted(() => {
  if (typeof window === 'undefined') return
  window.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_SHOP_SEARCH, false)
  resetPanelDrag()
  if (typeof window === 'undefined') return
  window.removeEventListener('keydown', onKeydown)
})
</script>

<style scoped>
.shop-search-sheet-panel {
  height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
  max-height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
  background: var(--tz-card-surface);
  background-image: none;
  border: 1px solid rgba(5, 150, 105, 0.22);
  transform: translateY(var(--shop-search-sheet-drag-y, 0));
  transition: transform 0.2s ease;
  will-change: transform;
  box-shadow:
    0 24px 80px rgba(20, 32, 43, 0.16);
}

.shop-search-sheet-panel::before {
  position: absolute;
  top: 0;
  right: 0;
  left: 0;
  z-index: 20;
  height: 1px;
  content: '';
  background: #059669;
  opacity: 0.62;
}

.shop-search-sheet-panel--dragging {
  transition: none !important;
}

.shop-search-sheet-drag-edge {
  cursor: grab;
  touch-action: none;
  user-select: none;
}

.shop-search-sheet-panel--dragging .shop-search-sheet-drag-edge {
  cursor: grabbing;
}

@media (min-width: 768px) {
  .shop-search-sheet-panel {
    height: 82vh;
    max-height: 900px;
  }
}
</style>
