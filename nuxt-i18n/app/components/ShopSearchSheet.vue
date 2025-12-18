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
        class="fixed inset-0 z-[14000] flex items-end justify-center p-0 md:p-4"
        @click.self="close"
      >
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm" @click="close"></div>

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
            v-if="isOpen"
            class="sidebar-panel relative pointer-events-auto w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] bg-slate-950/80 backdrop-blur-xl border-2 border-[#6b73ff]/40 rounded-2xl shadow-[0_0_30px_rgba(107,115,255,0.6)] flex flex-col overflow-hidden"
            aria-modal="true"
            role="dialog"
            :aria-label="$t('sidebar.searchProducts', 'Search Products')"
          >
            <div class="absolute inset-x-0 top-0 h-[200px] bg-gradient-to-br from-indigo-600/20 to-teal-600/20 blur-3xl pointer-events-none z-0"></div>

            <header class="relative z-10 flex items-center justify-between px-4 md:px-6 py-4 border-b border-white/10">
              <h2 class="text-lg md:text-xl font-semibold text-white">
                {{ $t('sidebar.searchProducts', 'Search Products') }}
              </h2>
              <button
                type="button"
                class="w-9 h-9 flex items-center justify-center rounded-full hover:bg-white/10 transition-colors"
                :aria-label="$t('common.close', 'Close')"
                @click="close"
              >
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </header>

            <div class="relative z-10 flex-1 overflow-y-auto px-4 md:px-6 py-4">
              <ProductSearchPanel @search="handleSearch" />
            </div>
          </section>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch, onMounted, onBeforeUnmount } from 'vue'
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

const onKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    close()
  }
}

const onGlobalPopup = (event: Event) => {
  try {
    const custom = event as CustomEvent<{ id?: string }>
    const id = custom?.detail?.id
    if (id && id !== 'shop-search') {
      close()
    }
  } catch {}
}

onMounted(() => {
  if (typeof window === 'undefined') return
  window.addEventListener('keydown', onKeydown)
  window.addEventListener('ui:popup-open', onGlobalPopup as EventListener)
})

onBeforeUnmount(() => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_SHOP_SEARCH, false)
  if (typeof window === 'undefined') return
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('ui:popup-open', onGlobalPopup as EventListener)
})
</script>
