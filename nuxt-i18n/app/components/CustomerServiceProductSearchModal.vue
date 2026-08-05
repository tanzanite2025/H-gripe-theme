<template>
  <Teleport to="body">
    <Transition name="customer-service-product-search-modal">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[10030] flex items-end justify-center p-0 md:items-center md:p-6 tz-mobile-safe-modal-mask"
        role="presentation"
        @click.self="handleCloseCustomerServiceProductSearchModal"
      >
        <div
          class="absolute inset-0 bg-black/80 backdrop-blur-sm"
          aria-hidden="true"
          @click="handleCloseCustomerServiceProductSearchModal"
        />

        <section
          ref="modalElement"
          class="customer-service-product-search-modal-shell relative flex tz-mobile-safe-drawer-90-height w-full flex-col overflow-hidden border border-x-0 border-b-0 border-white/20 bg-black text-white shadow-[0_24px_80px_rgba(0,0,0,0.7)] md:h-auto md:max-h-[min(42rem,calc(100dvh-1.5rem))] md:w-[min(56rem,calc(100vw-1.5rem))] md:border"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="modalTitleId"
        >
          <button
            type="button"
            class="flex h-9 w-full items-center justify-center border-b border-white/15 text-white/75 transition-colors hover:text-white md:hidden"
            :aria-label="t('chatModal.productPicker.collapse')"
            :title="t('chatModal.productPicker.collapse')"
            @click="handleCloseCustomerServiceProductSearchModal"
          >
            <Icon name="lucide:chevron-down" class="h-5 w-5" />
          </button>

          <header class="flex items-start justify-between gap-4 border-b border-white/15 px-4 py-4 md:px-6">
            <div class="min-w-0">
              <h2 :id="modalTitleId" class="text-base font-semibold text-white md:text-lg">
                {{ t('chatModal.productPicker.title') }}
              </h2>
              <p class="mt-1 text-sm leading-5 text-white/60">
                {{ t('chatModal.productPicker.description') }}
              </p>
            </div>

            <button
              type="button"
              class="hidden h-9 w-9 shrink-0 items-center justify-center border border-white/20 text-white/70 transition-colors hover:border-white/50 hover:text-white md:flex"
              :aria-label="t('chatModal.actions.close')"
              :title="t('chatModal.actions.close')"
              @click="handleCloseCustomerServiceProductSearchModal"
            >
              <Icon name="lucide:x" class="h-4 w-4" />
            </button>
          </header>

          <form
            class="flex gap-2 border-b border-white/10 px-4 py-3 md:px-6"
            @submit.prevent="handleCustomerServiceProductSearchSubmit"
          >
            <div class="relative min-w-0 flex-1">
              <Icon
                name="lucide:search"
                class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-white/45"
              />
              <input
                ref="searchInputElement"
                v-model="searchQuery"
                type="search"
                :placeholder="t('chatModal.productPicker.placeholder')"
                class="h-11 w-full border border-white/20 bg-white/[0.04] pl-10 pr-10 text-sm text-white outline-none transition-colors placeholder:text-white/35 focus:border-[#B5FF6D]/70"
                autocomplete="off"
              />
              <button
                v-if="searchQuery"
                type="button"
                class="absolute right-2 top-1/2 flex h-7 w-7 -translate-y-1/2 items-center justify-center text-white/45 transition-colors hover:text-white"
                :aria-label="t('chatModal.productPicker.clear')"
                :title="t('chatModal.productPicker.clear')"
                @click="handleClearCustomerServiceProductSearchQuery"
              >
                <Icon name="lucide:x-circle" class="h-4 w-4" />
              </button>
            </div>

            <button
              type="submit"
              class="h-11 shrink-0 bg-white px-4 text-sm font-semibold text-black transition-colors hover:bg-white/90 disabled:cursor-not-allowed disabled:opacity-50 md:px-5"
              :disabled="isSearching"
            >
              {{ isSearching ? t('chatModal.productPicker.searching') : t('chatModal.productPicker.search') }}
            </button>
          </form>

          <div class="min-h-0 flex-1 overflow-y-auto p-4 md:p-6">
            <div
              v-if="isSearching"
              class="flex min-h-48 flex-col items-center justify-center gap-3 text-sm text-white/60"
            >
              <Icon name="lucide:loader-circle" class="h-6 w-6 animate-spin text-[#B5FF6D]" />
              <span>{{ t('chatModal.productPicker.searching') }}</span>
            </div>

            <div
              v-else-if="searchError"
              class="flex min-h-48 items-center justify-center text-center text-sm text-red-200"
            >
              {{ searchError }}
            </div>

            <div
              v-else-if="!hasSearched"
              class="flex min-h-48 items-center justify-center text-center text-sm text-white/50"
            >
              {{ t('chatModal.productPicker.empty') }}
            </div>

            <div
              v-else-if="searchResults.length === 0"
              class="flex min-h-48 items-center justify-center text-center text-sm text-white/50"
            >
              {{ t('chatModal.productPicker.noResults') }}
            </div>

            <div v-else class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              <article
                v-for="product in searchResults"
                :key="product.id"
                class="flex min-h-56 flex-col overflow-hidden border border-white/[0.12] bg-white/[0.035] transition-colors hover:border-[#B5FF6D]/55"
              >
                <div class="aspect-[4/3] w-full bg-white/[0.03]">
                  <img
                    v-if="product.thumbnail"
                    :src="product.thumbnail"
                    :alt="product.title"
                    class="h-full w-full object-cover"
                  />
                </div>

                <div class="flex flex-1 flex-col p-3">
                  <h3 class="line-clamp-2 text-sm font-semibold leading-5 text-white">
                    {{ product.title }}
                  </h3>
                  <p v-if="product.priceLabel" class="mt-1 text-sm font-semibold text-[#B5FF6D]">
                    {{ product.priceLabel }}
                  </p>
                  <p v-if="product.sku" class="mt-1 truncate text-xs text-white/45">
                    {{ product.sku }}
                  </p>

                  <button
                    type="button"
                    class="mt-auto flex h-9 items-center justify-center gap-2 border border-white/30 bg-white/[0.06] px-3 text-xs font-semibold text-white transition-colors hover:border-[#B5FF6D]/70 hover:bg-[#B5FF6D]/10"
                    @click="handleSelectCustomerServiceProductForChat(product)"
                  >
                    <Icon name="lucide:check" class="h-3.5 w-3.5" />
                    {{ t('chatModal.productPicker.select') }}
                  </button>
                </div>
              </article>
            </div>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { useShopProducts, type ShopProduct } from '~/composables/useShopProducts'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
  'select-customer-service-product': [product: ShopProduct]
}>()

const { t } = useI18n()
const { fetchShopProducts } = useShopProducts()
const modalElement = ref<HTMLElement | null>(null)
const searchInputElement = ref<HTMLInputElement | null>(null)
const searchQuery = ref('')
const searchResults = ref<ShopProduct[]>([])
const searchError = ref('')
const isSearching = ref(false)
const hasSearched = ref(false)
const modalTitleId = 'customer-service-product-search-modal-title'

const resetCustomerServiceProductSearchState = () => {
  searchQuery.value = ''
  searchResults.value = []
  searchError.value = ''
  isSearching.value = false
  hasSearched.value = false
}

const handleCloseCustomerServiceProductSearchModal = () => {
  emit('update:modelValue', false)
  emit('close')
}

const handleClearCustomerServiceProductSearchQuery = () => {
  searchQuery.value = ''
  searchResults.value = []
  searchError.value = ''
  hasSearched.value = false
  nextTick(() => searchInputElement.value?.focus())
}

const handleCustomerServiceProductSearchSubmit = async () => {
  const normalizedQuery = searchQuery.value.trim()
  if (!normalizedQuery || isSearching.value) return

  isSearching.value = true
  searchError.value = ''
  hasSearched.value = true

  try {
    const response = await fetchShopProducts({
      keyword: normalizedQuery,
      per_page: 20,
    })
    searchResults.value = response.items
  } catch (error) {
    console.error('搜索客服商品选择器失败:', error)
    searchResults.value = []
    searchError.value = t('chatModal.productPicker.error')
  } finally {
    isSearching.value = false
  }
}

const handleSelectCustomerServiceProductForChat = (product: ShopProduct) => {
  emit('select-customer-service-product', product)
  handleCloseCustomerServiceProductSearchModal()
}

const handleCustomerServiceProductSearchModalKeydown = (event: KeyboardEvent) => {
  if (props.modelValue && event.key === 'Escape') {
    handleCloseCustomerServiceProductSearchModal()
  }
}

watch(
  () => props.modelValue,
  value => {
    if (value) {
      resetCustomerServiceProductSearchState()
      nextTick(() => searchInputElement.value?.focus())
    }
  }
)

onMounted(() => {
  document.addEventListener('keydown', handleCustomerServiceProductSearchModalKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleCustomerServiceProductSearchModalKeydown)
})
</script>

<style scoped>
.customer-service-product-search-modal-enter-active,
.customer-service-product-search-modal-leave-active {
  transition: opacity 0.2s ease;
}

.customer-service-product-search-modal-enter-active section,
.customer-service-product-search-modal-leave-active section {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.customer-service-product-search-modal-enter-from,
.customer-service-product-search-modal-leave-to {
  opacity: 0;
}

.customer-service-product-search-modal-enter-from section,
.customer-service-product-search-modal-leave-to section {
  opacity: 0;
  transform: translateY(1rem);
}

@media (min-width: 768px) {
  .customer-service-product-search-modal-enter-from section,
  .customer-service-product-search-modal-leave-to section {
    transform: translateY(0.5rem) scale(0.98);
  }
}

@media (max-width: 767px) {
  .customer-service-product-search-modal-shell {
    height: min(90dvh, var(--tz-mobile-safe-viewport-height, 90dvh));
    max-height: min(90dvh, var(--tz-mobile-safe-viewport-height, 90dvh));
  }
}
</style>
