<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-300 ease-out"
      leave-active-class="transition-opacity duration-200 ease-in"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="modelValue"
        class="tire-rim-search-sheet fixed inset-0 z-[14000] flex items-end justify-center p-0 tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
        @click.self="handleClose"
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
          <section
            v-if="modelValue"
            class="tire-rim-search-sheet__panel tz-mobile-dialog-surface tz-surface-card relative pointer-events-auto w-full max-w-none tz-mobile-safe-full-height md:h-[82vh] md:max-h-[900px] rounded-none flex flex-col overflow-hidden"
            role="dialog"
            aria-modal="true"
            :aria-label="$t('sidebar.searchProducts', 'Search Products')"
          >
            <header class="relative z-10 flex items-center justify-between px-4 py-4 md:px-6">
              <h2 class="text-lg font-semibold text-[#059669] md:text-xl">
                {{ $t('sidebar.searchProducts', 'Search Products') }}
              </h2>
              <button
                type="button"
                class="tz-global-close-btn"
                :aria-label="$t('common.close', 'Close')"
                @click="handleClose"
              >
                <Icon name="lucide:x" class="h-5 w-5 tz-text-primary" />
              </button>
            </header>

            <div class="relative z-10 flex-1 overflow-y-auto px-4 py-4 md:px-6">
              <ShopProductQuickSearchForm
                density="drawer"
                @submit="handleSearch"
              />
            </div>
          </section>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, watch } from 'vue'
import ShopProductQuickSearchForm from '~/components/shop/ShopProductQuickSearchForm.vue'
import type { ShopSearchPayload } from '~/composables/useShopSearchSheet'
import { useTireGuideTireProductSearch } from '~/composables/useTireGuideTireProductSearch'
import { setSidebarHandlesHidden } from '~/utils/sidebarHandles'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
}>()

const { submit } = useTireGuideTireProductSearch()
const SIDEBAR_TOKEN = 'tire-rim-product-search-sheet'

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}

const handleSearch = async (payload: ShopSearchPayload) => {
  handleClose()
  await submit(payload)
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && props.modelValue) {
    handleClose()
  }
}

watch(
  () => props.modelValue,
  (open) => {
    setSidebarHandlesHidden(SIDEBAR_TOKEN, open)
  },
  { immediate: true },
)

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN, false)
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.tire-rim-search-sheet__panel {
  height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
  max-height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
  background: var(--tz-card-surface);
  background-image: none;
  border: 1px solid rgba(5, 150, 105, 0.22);
  box-shadow: 0 24px 80px rgba(20, 32, 43, 0.16);
}

.tire-rim-search-sheet__panel::before {
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

@media (min-width: 768px) {
  .tire-rim-search-sheet__panel {
    height: 82vh;
    max-height: 900px;
  }
}
</style>
