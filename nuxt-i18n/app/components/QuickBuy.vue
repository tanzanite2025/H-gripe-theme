<template>
  <!-- 弹窗模态框 (由 GradientDockMenu 触发) -->
  <teleport to="body">
    <!-- 遮罩层 -->
    <Transition name="fade">
      <div
        class="quickbuy-modal-mask fixed inset-0 z-[10002] flex items-center justify-center p-0 md:p-4 tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
        @click.self="handleClose"
      >
        <!-- 半透明背景遮罩 -->
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm"></div>
        <!-- 弹窗内容 -->
        <Transition name="slide-up" appear>
          <div
            class="sidebar-panel quickbuy-modal-shell tz-mobile-dialog-surface relative w-[90vw] max-w-none h-[80vh] max-h-[80vh] bg-black backdrop-blur-xl rounded-2xl shadow-[0_18px_44px_rgba(0,0,0,0.72)] box-border flex flex-col overflow-hidden"
            role="dialog"
            aria-modal="true"
          >
        <!-- 头部 -->
        <header class="flex items-center justify-between px-3.5 max-md:px-2 py-2.5 max-md:py-2 rounded-t-2xl overflow-hidden max-md:gap-1.5">
          <nav class="flex-1 min-w-0 overflow-hidden max-md:flex-auto" :aria-label="t('quickBuy.stepsAriaLabel')">
            <ol class="flex items-center justify-center gap-3 max-md:gap-1.5 list-none m-0 p-0 max-md:flex-nowrap">
              <li
                v-for="n in totalSteps"
                :key="n"
                class="inline-flex items-center gap-3 max-md:gap-1.5"
              >
                <span 
                  class="w-7 h-7 max-md:w-[22px] max-md:h-[22px] rounded-full grid place-items-center font-bold transition-all duration-200"
                  :class="[
                    n === currentStepIndex ? 'bg-white text-black shadow-[0_0_0_3px_rgba(255,255,255,0.16)]' :
                    n < currentStepIndex ? 'bg-[#3c4454] text-white' :
                    'bg-[#2c2f35] text-white/90'
                  ]"
                >{{ n }}</span>
                <span v-if="n < totalSteps" class="w-8 max-md:w-2.5 h-1 rounded-full bg-white/[0.18]" aria-hidden="true" />
              </li>
            </ol>
          </nav>
          <button 
            class="tz-global-close-btn flex-none ml-1.5 max-md:ml-0"
            type="button" 
            :aria-label="t('common.close', 'Close')"
            @click="handleClose"
          >×</button>
        </header>

        <!-- 主体内容 -->
        <section class="quickbuy-modal-body px-3.5 py-3 flex-1 min-h-0 overflow-y-auto overflow-x-hidden">
          <div class="quickbuy-workspace">
            <QuickBuyCandidateProductSelectionPanel
              v-model:query="query"
              :title="currentStep.name"
              :fallback-title="t('quickBuy.placeholder.stepTitle', { step: currentStepIndex })"
              :help-title="t('quickBuy.help.title', 'QUICK instructions')"
              :help-content="quickBuyFlowHelpContent"
              :search-placeholder="t('quickBuy.search.placeholder')"
              :products="candidateProducts"
              :error-message="quickBuyError"
              :loading="loadingCandidates"
              :can-go-to-previous-product-page="canGoToPreviousQuickBuyCandidateProductPage"
              :can-go-to-next-product-page="canGoToNextQuickBuyCandidateProductPage"
              :show-product-pagination="showCandidatePagination"
              :product-page="candidatePage"
              :current-step-index="currentStepIndex"
              :total-steps="totalSteps"
              :previous-label="t('common.previous', 'Previous')"
              :next-label="t('common.next', 'Next')"
              :product-rail-label="t('quickBuy.productRail.label', 'Available products')"
              :empty-label="t('quickBuy.productRail.empty', 'No products are available for this step.')"
              :loading-label="t('common.loading', 'Loading...')"
              :current-page-label="t('common.currentPage', 'Current page')"
              :help-trigger-aria-label="t('quickBuy.selection.label', 'Selected products')"
              :close-label="t('common.close', 'Close')"
              :is-product-selected="isCurrentStepProductSelected"
              @query-input="scheduleSearch"
              @submit-search="triggerSearch"
              @previous-product-page="goToPreviousQuickBuyCandidateProductPage"
              @next-product-page="goToNextQuickBuyCandidateProductPage"
              @select-product="toggleCandidateSelection"
              @open-product-details="openQuickBuyCandidateProductDetails"
              @previous-step="goToPreviousStep"
              @next-step="goToNextStep"
            />

            <QuickBuySelectedProductsSummaryPanel
              :title="t('quickBuy.selection.label', 'Selected products')"
              :slots="quickBuySelectedProductSlots"
              :total-qty="totalQty"
              :total-weight-g="totalWeightG"
              :formatted-total-price="formattedTotalPrice"
              :has-selected-products="selectedProducts.length > 0"
              :items-label="t('quickBuy.summary.items', 'Items')"
              :weight-label="t('quickBuy.summary.weight', 'Weight')"
              :price-label="t('quickBuy.summary.price', 'Price')"
              :add-to-cart-label="t('quickBuy.actions.addToCart', 'Add to cart')"
              :direct-payment-label="t('quickBuy.actions.directPayment', 'Pay now')"
              :decrease-label="t('common.decrease', 'Decrease')"
              :increase-label="t('common.increase', 'Increase')"
              :remove-label="t('common.remove', 'Remove')"
              @remove-product="removeSelectedProduct"
              @change-quantity="changeSelectedProductQuantity"
              @update-quantity="updateSelectedProductQuantity"
              @add-to-cart="addSelectedProductsAndOpenCart"
              @direct-payment="addSelectedProductsAndOpenCheckout"
            />
          </div>
        </section>
          </div>
        </Transition>
      </div>
    </Transition>
  </teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useCart } from '~/composables/useCart'
import { useGlobalProductDetailBottomSheet } from '~/composables/useGlobalProductDetailBottomSheet'
import { useQuickBuySelectionState } from '~/composables/useQuickBuySelectionState'
import { useQuickBuySession } from '~/composables/useQuickBuySession'
import QuickBuyCandidateProductSelectionPanel from '~/components/quick-buy/QuickBuyCandidateProductSelectionPanel.vue'
import QuickBuySelectedProductsSummaryPanel from '~/components/quick-buy/QuickBuySelectedProductsSummaryPanel.vue'
import { useShopProducts } from '~/composables/useShopProducts'
import type { ShopProduct } from '~/composables/useShopProducts'
import type {
  QuickBuyConfig,
  QuickBuyStep,
} from '~/utils/quickBuy/types'

type Maybe<T> = T | null | undefined

const props = defineProps<{ config: QuickBuyConfig | null }>()
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()

const defaultQuickBuySteps = computed<QuickBuyStep[]>(() => [
  {
    id: 1,
    slug: 'product-search',
    name: t('quickBuy.defaultSteps.step1', 'Step 1'),
  },
  {
    id: 2,
    slug: 'specifications',
    name: t('quickBuy.defaultSteps.step2', 'Step 2'),
  },
  {
    id: 3,
    slug: 'quantity',
    name: t('quickBuy.defaultSteps.step3', 'Step 3'),
  },
])

const currentStepIndex = ref(1)
const query = ref('')
const candidateProducts = ref<ShopProduct[]>([])
const loadingCandidates = ref(false)
const quickBuyError = ref('')
const candidatePage = ref(1)
const hasMoreCandidates = ref(false)

const QUICK_BUY_CANDIDATE_PAGE_SIZE = 6

let searchTimer: Maybe<number> = null

const { addToCart, openCheckout } = useCart()
const { openGlobalProductDetailBottomSheet } = useGlobalProductDetailBottomSheet()
const quickBuySession = useQuickBuySession('dock')
const { session, fetchStepCandidates, createSession, error: quickBuySessionError } = quickBuySession
const { fetchPublicShopProducts } = useShopProducts()

const propConfiguredSteps = computed<QuickBuyStep[]>(() => props.config?.steps ?? [])
const sessionConfiguredSteps = computed<QuickBuyStep[]>(() => session.value?.flow?.steps ?? [])
const configuredSteps = computed<QuickBuyStep[]>(() =>
  propConfiguredSteps.value.length ? propConfiguredSteps.value : sessionConfiguredSteps.value,
)
const hasConfiguredFlow = computed(() => configuredSteps.value.length > 0)
const steps = computed(() => hasConfiguredFlow.value ? configuredSteps.value : defaultQuickBuySteps.value)
const totalSteps = computed(() => steps.value.length)
const currentStep = computed<QuickBuyStep>(() => steps.value[currentStepIndex.value - 1] || { id: 0, slug: '', name: '' })
const currentStepKey = computed(() => currentStep.value.stepKey || currentStep.value.slug || `step-${currentStepIndex.value}`)
const quickBuyFlowHelpContent = computed(() =>
  props.config?.flowHelpText?.trim()
  || session.value?.flow?.helpText?.trim()
  || t('quickBuy.selection.editHint', 'Review and adjust your selections'),
)

const {
  selectedProducts,
  selectedProductSlots: quickBuySelectedProductSlots,
  totalQty,
  totalWeightG,
  totalPrice,
  hydrateSelectedProductsFromSession,
  isProductSelectedForStep,
  toggleProductSelection,
  removeSelectedProduct,
  changeSelectedProductQuantity,
  updateSelectedProductQuantity,
} = useQuickBuySelectionState(steps, hasConfiguredFlow, quickBuySession)

const formattedTotalPrice = computed(() => {
  return formatAmount(totalPrice.value)
})

const canGoToPreviousQuickBuyCandidateProductPage = computed(() => candidatePage.value > 1)
const canGoToNextQuickBuyCandidateProductPage = computed(() => hasMoreCandidates.value)
const showCandidatePagination = computed(() =>
  canGoToPreviousQuickBuyCandidateProductPage.value || canGoToNextQuickBuyCandidateProductPage.value,
)

let fetchSequence = 0

const formatAmount = (value: number) => {
  try {
    return new Intl.NumberFormat(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(Number(value || 0))
  } catch {
    return String(value)
  }
}

const loadQuickBuyCandidateProductsByPage = async (page = 1) => {
  const sequence = fetchSequence + 1
  fetchSequence = sequence
  loadingCandidates.value = true
  quickBuyError.value = ''
  candidatePage.value = page
  hasMoreCandidates.value = false
  candidateProducts.value = []

  try {
    const activeSession = await createSession()
    if (sequence !== fetchSequence) return
    hydrateSelectedProductsFromSession(activeSession)

    let result: ReturnType<typeof normalizeCandidatePage>
    const shouldUseConfiguredCandidates = Boolean(
      propConfiguredSteps.value.length || activeSession?.flow?.steps?.length,
    )
    if (shouldUseConfiguredCandidates) {
      const res = await fetchStepCandidates(currentStepKey.value, {
        keyword: query.value.trim(),
        page,
        pageSize: QUICK_BUY_CANDIDATE_PAGE_SIZE,
      })
      if (sequence !== fetchSequence) return
      if (!res) {
        throw new Error(quickBuySessionError.value || 'Unable to load QUICK candidates')
      }
      result = normalizeCandidatePage(res)
    } else {
      const res = await fetchPublicShopProducts({
        page,
        page_size: QUICK_BUY_CANDIDATE_PAGE_SIZE,
        status: 'active',
        ...(query.value.trim() ? { keyword: query.value.trim() } : {}),
      })
      if (sequence !== fetchSequence) return
      result = normalizeCandidatePage(res)
    }
    candidateProducts.value = result.items
    candidatePage.value = result.page
    hasMoreCandidates.value = result.hasMore
  } catch (err) {
    if (sequence !== fetchSequence) return
    quickBuyError.value = (err as Error).message || String(err)
    candidateProducts.value = []
  } finally {
    if (sequence === fetchSequence) {
      loadingCandidates.value = false
    }
  }
}

const normalizeCandidatePage = (result: {
  items: ShopProduct[]
  page?: number
  hasMore?: boolean
}) => ({
  items: result.items,
  page: result.page || 1,
  hasMore: Boolean(result.hasMore),
})

const scheduleSearch = () => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }
  searchTimer = window.setTimeout(() => {
    loadQuickBuyCandidateProductsByPage()
    searchTimer = null
  }, 300)
}

const triggerSearch = () => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
    searchTimer = null
  }
  loadQuickBuyCandidateProductsByPage()
}

const goToPreviousQuickBuyCandidateProductPage = () => {
  if (!canGoToPreviousQuickBuyCandidateProductPage.value || loadingCandidates.value) return
  loadQuickBuyCandidateProductsByPage(candidatePage.value - 1)
}

const goToNextQuickBuyCandidateProductPage = () => {
  if (!canGoToNextQuickBuyCandidateProductPage.value || loadingCandidates.value) return
  loadQuickBuyCandidateProductsByPage(candidatePage.value + 1)
}

const isCurrentStepProductSelected = (product: ShopProduct) =>
  isProductSelectedForStep(product, currentStepKey.value)

const toggleCandidateSelection = async (product: ShopProduct) => {
  const saved = await toggleProductSelection(product, currentStepKey.value)
  if (saved === false) {
    quickBuyError.value = quickBuySessionError.value || 'Unable to save QUICK selection'
  }
}

const openQuickBuyCandidateProductDetails = (product: ShopProduct) => {
  openGlobalProductDetailBottomSheet({
    id: product.id,
    slug: product.slug,
    title: product.title,
    thumbnail: product.thumbnail,
  })
}

watch([currentStepKey, hasConfiguredFlow], () => {
  query.value = ''
  quickBuyError.value = ''
  loadQuickBuyCandidateProductsByPage()
}, { immediate: true })

watch(totalSteps, (nextTotalSteps) => {
  if (currentStepIndex.value > nextTotalSteps) {
    currentStepIndex.value = nextTotalSteps
  }
})

const goToNextStep = () => {
  if (currentStepIndex.value < totalSteps.value) {
    currentStepIndex.value += 1
  }
}

const goToPreviousStep = () => {
  if (currentStepIndex.value > 1) {
    currentStepIndex.value -= 1
  }
}

const handleClose = () => {
  emit('close')
}

const addSelectedProductsToCart = () => {
  for (const item of selectedProducts.value) {
    addToCart({
      id: item.productId,
      product_id: item.productId,
      variant_id: item.variantId,
      title: item.title,
      name: item.title,
      slug: item.slug,
      sku: item.sku,
      thumbnail: item.thumbnail,
      image: item.thumbnail,
      price: item.unitPrice,
      currency: item.currency,
      weight_grams: item.weightGrams,
    }, item.quantity)
  }
}

const addSelectedProductsAndOpenCart = () => {
  addSelectedProductsToCart()
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('open-cart-drawer'))
  }
  emit('close')
}

const addSelectedProductsAndOpenCheckout = () => {
  addSelectedProductsToCart()
  emit('close')
  openCheckout()
}

onBeforeUnmount(() => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }
})
</script>

<style scoped>
/* 遮罩层淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 弹窗滑入滑出动画：与 WishlistDrawer 一致，从底部整块滑入 */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.3s ease-out, opacity 0.3s ease-out;
}

.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.slide-up-enter-to,
.slide-up-leave-from {
  transform: translateY(0%);
  opacity: 1;
}

.quickbuy-modal-body {
  display: flex;
  min-height: 0;
  background: var(--quickbuy-shell-surface);
}

.quickbuy-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(18rem, 0.65fr);
  gap: 0.75rem;
  width: 100%;
  min-height: 100%;
}

.quickbuy-modal-shell {
  --quickbuy-shell-surface: #050505;
  --quickbuy-panel-surface: var(--tz-card-surface, #111116);
  --quickbuy-panel-surface-soft: #0c0c0e;
  --quickbuy-panel-surface-raised: #17171b;
  --quickbuy-control-surface: #0a0a0c;
  --quickbuy-control-surface-raised: #151519;
  --quickbuy-divider: rgba(255, 255, 255, 0.045);
  --quickbuy-divider-strong: rgba(255, 255, 255, 0.065);
  --quickbuy-dark-edge: rgba(0, 0, 0, 0.68);
  --quickbuy-focus-ring: rgba(181, 255, 109, 0.12);
  --quickbuy-shadow: 0 18px 54px rgba(0, 0, 0, 0.64);
  border: 0 !important;
  background:
    linear-gradient(180deg, #080808 0%, var(--quickbuy-shell-surface) 100%) !important;
  box-shadow:
    0 30px 90px rgba(0, 0, 0, 0.82),
    inset 0 1px 0 rgba(255, 255, 255, 0.025),
    inset 0 0 0 1px rgba(0, 0, 0, 0.72) !important;
}

.quickbuy-modal-shell > header {
  background:
    linear-gradient(180deg, #0b0b0c, #070707);
  box-shadow: inset 0 -1px 0 var(--quickbuy-divider);
}

@media (max-width: 767px) {
  .quickbuy-modal-mask {
    align-items: stretch;
    justify-content: stretch;
    width: 100vw;
    min-width: 100vw;
    height: 100vh;
    min-height: 100vh;
    box-sizing: border-box;
  }

  @supports (width: 100dvw) {
    .quickbuy-modal-mask {
      width: 100dvw;
      min-width: 100dvw;
    }
  }

  @supports (height: 100svh) {
    .quickbuy-modal-mask {
      height: 100svh;
      min-height: 100svh;
    }
  }

  @supports (height: 100dvh) {
    .quickbuy-modal-mask {
      height: 100dvh;
      min-height: 100dvh;
    }
  }

  .quickbuy-modal-shell {
    width: 100%;
    max-width: 100%;
    height: 100%;
    max-height: 100%;
    min-height: 0;
  }

  .quickbuy-modal-shell > header {
    border-radius: 0.75rem 0.75rem 0 0;
  }

  .quickbuy-modal-body {
    overflow-y: auto;
  }

  .quickbuy-workspace {
    grid-template-columns: minmax(0, 1fr);
    min-height: auto;
  }
}
</style>
