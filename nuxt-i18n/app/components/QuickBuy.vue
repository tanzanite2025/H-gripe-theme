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
        <div
          class="absolute inset-0 bg-[#030406]/76 backdrop-blur-sm"
          aria-hidden="true"
          @click="handleClose"
        ></div>
        <!-- 弹窗内容 -->
        <Transition name="slide-up" appear>
          <div
            class="sidebar-panel quickbuy-modal-shell tz-mobile-dialog-surface relative w-[90vw] max-w-none h-[80vh] max-h-[80vh] bg-[#101116] backdrop-blur-xl rounded-2xl shadow-[0_18px_44px_rgba(0,0,0,0.72)] box-border flex flex-col overflow-hidden"
            role="dialog"
            aria-modal="true"
            @click.stop
          >
        <!-- 头部 -->
        <header class="quickbuy-modal-header grid items-center px-3.5 max-md:px-2 py-2.5 max-md:py-2 rounded-t-2xl max-md:gap-1.5">
          <span class="quickbuy-mobile-header-title">QUICKBUY</span>
          <div class="quickbuy-step-header-rail">
            <div class="quickbuy-step-control">
              <nav class="quickbuy-step-nav" :aria-label="t('quickBuy.stepsAriaLabel')">
                <ol class="flex items-center justify-center gap-5 max-md:gap-3 list-none m-0 p-0 max-md:flex-nowrap">
                  <li
                    v-for="n in totalSteps"
                    :key="n"
                    class="inline-flex items-center gap-5 max-md:gap-3"
                  >
                    <button
                      class="quickbuy-step-button w-8 h-8 max-md:w-7 max-md:h-7 rounded-full grid place-items-center font-bold transition-all duration-200"
                      type="button"
                      :aria-current="n === currentStepIndex ? 'step' : undefined"
                      :aria-label="steps[n - 1]?.name || t('quickBuy.placeholder.stepTitle', { step: n })"
                      :title="steps[n - 1]?.name || t('quickBuy.placeholder.stepTitle', { step: n })"
                      :class="[
                        n === currentStepIndex ? 'bg-white text-black shadow-[0_0_0_3px_rgba(255,255,255,0.16)]' :
                        n < currentStepIndex ? 'bg-[#3c4454] text-white' :
                        'bg-[#2c2f35] text-white/90'
                      ]"
                      @click="goToStep(n)"
                    >{{ n }}</button>
                    <span v-if="n < totalSteps" class="w-12 max-md:w-5 h-1 rounded-full bg-white/[0.18]" aria-hidden="true" />
                  </li>
                </ol>
              </nav>
              <span
                v-if="showStepHint"
                class="quickbuy-step-hint"
                aria-hidden="true"
              >
                <Icon name="lucide:pointer" />
              </span>
            </div>
          </div>
          <div class="quickbuy-modal-header-actions">
            <QuickBuyLocalizedHelpQuestionMarkDialog
              class="quickbuy-mobile-help"
              :title="t('quickBuy.help.title', 'QUICK instructions')"
              :content="quickBuyFlowHelpContent"
              :trigger-aria-label="t('quickBuy.selection.label', 'Selected products')"
              :close-label="t('common.close', 'Close')"
            />
            <button
              class="quickbuy-modal-close tz-global-close-btn relative z-10"
              type="button"
              :aria-label="t('common.close', 'Close')"
              :title="t('common.close', 'Close')"
              @click="handleClose"
            >
              <Icon name="lucide:x" aria-hidden="true" />
            </button>
          </div>
        </header>

        <!-- 主体内容 -->
        <section class="quickbuy-modal-body px-3.5 py-3 flex-1 min-h-0 overflow-y-auto overflow-x-hidden">
          <div class="quickbuy-workspace quickbuy-workspace--desktop">
            <QuickBuyCandidateProductSelectionPanel
              v-model:query="query"
              :title="currentStep.name"
              :fallback-title="t('quickBuy.placeholder.stepTitle', { step: currentStepIndex })"
              :help-title="t('quickBuy.help.title', 'QUICK instructions')"
              :help-content="quickBuyFlowHelpContent"
              :search-placeholder="t('quickBuy.search.placeholder')"
              :products="candidateProducts"
              :filters="candidateFilters"
              :selected-filters="currentSpecFilters"
              :filters-label="t('quickBuy.filters.label', 'Filters')"
              :clear-filters-label="t('quickBuy.filters.clear', 'Clear')"
              :no-filter-values-label="t('quickBuy.filters.empty', 'No values available')"
              :error-message="quickBuyError"
              :loading="loadingCandidates"
              :can-go-to-previous-product-page="canGoToPreviousQuickBuyCandidateProductPage"
              :can-go-to-next-product-page="canGoToNextQuickBuyCandidateProductPage"
              :show-product-pagination="showCandidatePagination"
              :product-page="candidatePage"
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
              @toggle-filter="toggleSpecFilter"
              @clear-filters="clearSpecFilters"
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
              :view-details-label="t('products.detail.viewDetails', 'View details')"
              :close-label="t('common.close', 'Close')"
              :done-label="t('common.done', 'Done')"
              :previous-product-label="t('common.previous', 'Previous')"
              :next-product-label="t('common.next', 'Next')"
              @remove-product="removeSelectedProduct"
              @change-quantity="changeSelectedProductQuantity"
              @update-quantity="updateSelectedProductQuantity"
              @add-to-cart="addSelectedProductsAndOpenCart"
              @direct-payment="addSelectedProductsAndOpenCheckout"
            />
          </div>

          <!-- Mobile uses step rows in the main dialog; product browsing opens in the bottom picker below. -->
          <div class="quickbuy-mobile-workspace">
            <QuickBuySelectedProductsSummaryPanel
              :title="t('quickBuy.selection.label', 'Selected products')"
              :slots="mobileQuickBuySelectedProductSlots"
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
              :view-details-label="t('products.detail.viewDetails', 'View details')"
              :close-label="t('common.close', 'Close')"
              :done-label="t('common.done', 'Done')"
              :previous-product-label="t('common.previous', 'Previous')"
              :next-product-label="t('common.next', 'Next')"
              :show-header="false"
              mobile-step-list
              @select-step="openMobileProductPicker"
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

    <!-- Mobile-only product picker: a bottom sheet sized to roughly 95% of the viewport. -->
    <Transition name="quickbuy-mobile-picker">
      <div
        v-if="mobileProductPickerOpen"
        class="quickbuy-mobile-picker-mask"
        tabindex="-1"
        @click.self="closeMobileProductPicker"
        @keydown.esc="closeMobileProductPicker"
      >
        <div
          class="quickbuy-mobile-picker-backdrop"
          aria-hidden="true"
          @click="closeMobileProductPicker"
        ></div>
        <section
          class="quickbuy-mobile-picker-shell"
          role="dialog"
          aria-modal="true"
          :aria-label="mobilePickerStep.name || t('quickBuy.placeholder.stepTitle', { step: mobilePickerStepIndex })"
          @click.stop
        >
          <div class="quickbuy-mobile-picker-handle" aria-hidden="true"></div>
          <header class="quickbuy-mobile-picker-header">
            <div class="quickbuy-mobile-picker-heading">
              <h2>{{ mobilePickerStep.name || t('quickBuy.placeholder.stepTitle', { step: mobilePickerStepIndex }) }}</h2>
            </div>
            <button
              type="button"
              class="quickbuy-mobile-picker-close tz-global-close-btn relative z-10"
              :aria-label="t('common.close', 'Close')"
              :title="t('common.close', 'Close')"
              @click="closeMobileProductPicker"
            >
              <Icon name="lucide:x" aria-hidden="true" />
            </button>
          </header>
          <div class="quickbuy-mobile-picker-body">
            <QuickBuyCandidateProductSelectionPanel
              v-model:query="query"
              :title="currentStep.name"
              :fallback-title="t('quickBuy.placeholder.stepTitle', { step: currentStepIndex })"
              :help-title="t('quickBuy.help.title', 'QUICK instructions')"
              :help-content="quickBuyFlowHelpContent"
              :search-placeholder="t('quickBuy.search.placeholder')"
              :products="candidateProducts"
              :filters="candidateFilters"
              :selected-filters="currentSpecFilters"
              :filters-label="t('quickBuy.filters.label', 'Filters')"
              :clear-filters-label="t('quickBuy.filters.clear', 'Clear')"
              :no-filter-values-label="t('quickBuy.filters.empty', 'No values available')"
              :error-message="quickBuyError"
              :loading="loadingCandidates"
              :can-go-to-previous-product-page="canGoToPreviousQuickBuyCandidateProductPage"
              :can-go-to-next-product-page="canGoToNextQuickBuyCandidateProductPage"
              :show-product-pagination="showCandidatePagination"
              :product-page="candidatePage"
              :previous-label="t('common.previous', 'Previous')"
              :next-label="t('common.next', 'Next')"
              :product-rail-label="t('quickBuy.productRail.label', 'Available products')"
              :empty-label="t('quickBuy.productRail.empty', 'No products are available for this step.')"
              :loading-label="t('common.loading', 'Loading...')"
              :current-page-label="t('common.currentPage', 'Current page')"
              :help-trigger-aria-label="t('quickBuy.selection.label', 'Selected products')"
              :close-label="t('common.close', 'Close')"
              :is-product-selected="isCurrentStepProductSelected"
              hide-title
              :show-help="false"
              @query-input="scheduleSearch"
              @submit-search="triggerSearch"
              @previous-product-page="goToPreviousQuickBuyCandidateProductPage"
              @next-product-page="goToNextQuickBuyCandidateProductPage"
              @select-product="toggleCandidateSelection"
              @open-product-details="openQuickBuyCandidateProductDetails"
              @toggle-filter="toggleSpecFilter"
              @clear-filters="clearSpecFilters"
            />
          </div>
        </section>
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
import QuickBuyLocalizedHelpQuestionMarkDialog from '~/components/quick-buy/QuickBuyLocalizedHelpQuestionMarkDialog.vue'
import QuickBuySelectedProductsSummaryPanel from '~/components/quick-buy/QuickBuySelectedProductsSummaryPanel.vue'
import { useShopProducts } from '~/composables/useShopProducts'
import type { ShopProduct } from '~/composables/useShopProducts'
import type {
  QuickBuyConfig,
  QuickBuySpecFilter,
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
const showStepHint = ref(true)
const mobileProductPickerOpen = ref(false)
const mobilePickerStepIndex = ref(1)
const query = ref('')
const candidateProducts = ref<ShopProduct[]>([])
const candidateFilters = ref<QuickBuySpecFilter[]>([])
const stepFilterSelections = ref<Record<string, Record<string, string[]>>>({})
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
const mobilePickerStep = computed<QuickBuyStep>(() =>
  steps.value[mobilePickerStepIndex.value - 1]
  || currentStep.value
  || { id: 0, slug: '', name: '' },
)
const currentStepKey = computed(() => currentStep.value.stepKey || currentStep.value.slug || `step-${currentStepIndex.value}`)
const currentSpecFilters = computed(() => stepFilterSelections.value[currentStepKey.value] || {})
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

const mobileQuickBuySelectedProductSlots = computed(() =>
  quickBuySelectedProductSlots.value.slice(0, totalSteps.value),
)

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
  candidateFilters.value = []

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
        specFilters: currentSpecFilters.value,
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
    candidateFilters.value = result.filters
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
  quickBuyFilters?: QuickBuySpecFilter[]
}) => ({
  items: result.items,
  page: result.page || 1,
  hasMore: Boolean(result.hasMore),
  filters: result.quickBuyFilters || [],
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
    return
  }
  if (mobileProductPickerOpen.value) {
    closeMobileProductPicker()
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

const toggleSpecFilter = (slug: string, value: string) => {
  const stepKey = currentStepKey.value
  const current = { ...(stepFilterSelections.value[stepKey] || {}) }
  const nextValues = new Set(current[slug] || [])
  if (nextValues.has(value)) {
    nextValues.delete(value)
  } else {
    nextValues.add(value)
  }
  if (nextValues.size > 0) {
    current[slug] = Array.from(nextValues)
  } else {
    delete current[slug]
  }
  stepFilterSelections.value = {
    ...stepFilterSelections.value,
    [stepKey]: current,
  }
  loadQuickBuyCandidateProductsByPage(1)
}

const clearSpecFilters = () => {
  const stepKey = currentStepKey.value
  const next = { ...stepFilterSelections.value }
  delete next[stepKey]
  stepFilterSelections.value = next
  loadQuickBuyCandidateProductsByPage(1)
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

const dismissStepHint = () => {
  showStepHint.value = false
}

const goToStep = (stepIndex: number) => {
  dismissStepHint()
  if (stepIndex < 1 || stepIndex > totalSteps.value) return
  currentStepIndex.value = stepIndex
}

const openMobileProductPicker = (stepIndex: number) => {
  dismissStepHint()
  if (stepIndex < 1 || stepIndex > totalSteps.value) return
  mobilePickerStepIndex.value = stepIndex
  currentStepIndex.value = stepIndex
  mobileProductPickerOpen.value = true
}

const closeMobileProductPicker = () => {
  mobileProductPickerOpen.value = false
}

const handleClose = () => {
  closeMobileProductPicker()
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
  background:
    linear-gradient(180deg, #1d1f26 0%, var(--quickbuy-shell-surface) 54%, var(--quickbuy-shell-surface-soft) 100%);
}

.quickbuy-mobile-header-title {
  display: none;
}

.quickbuy-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(18rem, 0.65fr);
  gap: 0.75rem;
  width: 100%;
  min-height: 100%;
}

.quickbuy-mobile-workspace {
  display: none;
}

.quickbuy-mobile-picker-mask {
  display: none;
}

.quickbuy-mobile-picker-enter-active,
.quickbuy-mobile-picker-leave-active {
  transition: opacity 180ms ease;
}

.quickbuy-mobile-picker-enter-active .quickbuy-mobile-picker-shell,
.quickbuy-mobile-picker-leave-active .quickbuy-mobile-picker-shell {
  transition: transform 220ms ease, opacity 180ms ease;
}

.quickbuy-mobile-picker-enter-from,
.quickbuy-mobile-picker-leave-to {
  opacity: 0;
}

.quickbuy-mobile-picker-enter-from .quickbuy-mobile-picker-shell,
.quickbuy-mobile-picker-leave-to .quickbuy-mobile-picker-shell {
  transform: translateY(100%);
  opacity: 0;
}

.quickbuy-modal-shell {
  --quickbuy-shell-surface: #17181f;
  --quickbuy-shell-surface-soft: #111218;
  --quickbuy-panel-surface: #1b1c23;
  --quickbuy-panel-surface-soft: #16171d;
  --quickbuy-panel-surface-raised: #22242c;
  --quickbuy-control-surface: #111219;
  --quickbuy-control-surface-raised: #1a1c24;
  --quickbuy-divider: rgba(255, 255, 255, 0.085);
  --quickbuy-divider-strong: rgba(255, 255, 255, 0.12);
  --quickbuy-dark-edge: rgba(0, 0, 0, 0.42);
  --quickbuy-focus-ring: rgba(181, 255, 109, 0.12);
  --quickbuy-shadow: 0 18px 54px rgba(0, 0, 0, 0.5);
  border: 0 !important;
  background:
    linear-gradient(180deg, #24262e 0%, #1d1f26 34%, var(--quickbuy-shell-surface) 68%, var(--quickbuy-shell-surface-soft) 100%) !important;
  box-shadow:
    0 30px 90px rgba(0, 0, 0, 0.66),
    inset 0 1px 0 rgba(255, 255, 255, 0.075),
    inset 0 0 0 1px rgba(255, 255, 255, 0.045) !important;
}

.quickbuy-modal-shell > header {
  background:
    linear-gradient(180deg, #272a32 0%, #1f2128 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    inset 0 -1px 0 var(--quickbuy-divider);
}

.quickbuy-modal-header {
  position: relative;
  grid-template-columns: minmax(0, 1fr) auto;
  overflow: visible;
}

.quickbuy-modal-header-actions {
  grid-column: 2;
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
}

.quickbuy-mobile-help {
  display: none;
}

.quickbuy-step-header-rail {
  grid-column: 1;
  z-index: 1;
  display: flex;
  width: 100%;
  min-width: 0;
  align-items: center;
  justify-content: center;
  justify-self: stretch;
}

.quickbuy-step-control {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
}

.quickbuy-step-hint {
  display: grid;
  width: 1.9rem;
  height: 1.9rem;
  flex: 0 0 auto;
  place-items: center;
  color: #b5ff6d;
  pointer-events: none;
  filter: drop-shadow(0 0 0.55rem rgba(181, 255, 109, 0.44));
  animation: quickbuy-step-hint-tap 1.8s ease-in-out infinite;
}

.quickbuy-step-hint :deep(svg) {
  width: 1.45rem;
  height: 1.45rem;
  transform: rotate(-45deg);
}

@keyframes quickbuy-step-hint-tap {
  0%,
  100% {
    opacity: 0.72;
    transform: translate(0.12rem, 0.08rem) scale(0.9);
  }

  42% {
    opacity: 1;
    transform: translate(-0.24rem, -0.14rem) scale(1.08);
  }

  58% {
    opacity: 0.94;
    transform: translate(-0.08rem, -0.04rem) scale(0.98);
  }
}

.quickbuy-modal-close {
  justify-self: end;
}

.quickbuy-step-nav {
  position: static;
  min-width: 0;
  overflow: visible;
}

.quickbuy-step-nav ol {
  width: max-content;
}

.quickbuy-step-button {
  padding: 0;
  border: 0;
  cursor: pointer;
  outline: none;
}

.quickbuy-step-button:hover {
  filter: brightness(1.12);
  transform: translateY(-1px);
}

.quickbuy-step-button:focus-visible {
  box-shadow:
    0 0 0 3px rgba(181, 255, 109, 0.22),
    0 0 0 5px rgba(181, 255, 109, 0.08);
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

  .quickbuy-modal-header {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .quickbuy-modal-header-actions {
    gap: 0.35rem;
  }

  .quickbuy-mobile-header-title {
    display: block;
    grid-column: 1;
    color: rgba(181, 255, 109, 0.88);
    font-size: 0.7rem;
    font-weight: 850;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .quickbuy-step-header-rail {
    grid-column: 1 / -1;
    margin-top: 0.15rem;
    padding-bottom: 0.15rem;
    justify-content: center;
  }

  .quickbuy-step-hint {
    display: none;
  }

  .quickbuy-mobile-help {
    display: inline-flex;
  }

  .quickbuy-step-nav {
    min-width: 0;
    max-width: none;
    overflow: visible;
  }

  .quickbuy-step-nav ol {
    max-width: none;
  }

  .quickbuy-modal-body {
    padding: 0.5rem;
    overflow: hidden;
  }

  .quickbuy-workspace--desktop {
    display: none;
  }

  .quickbuy-mobile-workspace {
    display: flex;
    width: 100%;
    min-height: 0;
  }

  .quickbuy-mobile-workspace :deep(.quickbuy-summary-panel) {
    width: 100%;
    min-height: 0;
  }

  .quickbuy-mobile-picker-mask {
    position: fixed;
    inset: 0;
    z-index: 10010;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    box-sizing: border-box;
    padding-top: 5dvh;
  }

  .quickbuy-mobile-picker-backdrop {
    position: absolute;
    inset: 0;
    background: rgba(3, 4, 6, 0.78);
    backdrop-filter: blur(5px);
  }

  .quickbuy-mobile-picker-shell {
    position: relative;
    z-index: 1;
    display: flex;
    width: 100%;
    height: 95dvh;
    max-height: 95dvh;
    min-height: 0;
    flex-direction: column;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-bottom: 0;
    border-radius: 1rem 1rem 0 0;
    background:
      linear-gradient(180deg, #24262e 0%, #1d1f26 34%, #17181f 68%, #111218 100%);
    box-shadow:
      0 -18px 54px rgba(0, 0, 0, 0.58),
      inset 0 1px 0 rgba(255, 255, 255, 0.075);
  }

  .quickbuy-mobile-picker-handle {
    width: 2.75rem;
    height: 0.25rem;
    flex: 0 0 auto;
    margin: 0.45rem auto 0;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.28);
  }

  .quickbuy-mobile-picker-header {
    display: flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.55rem 0.75rem 0.65rem;
  }

  .quickbuy-mobile-picker-heading {
    min-width: 0;
  }

  .quickbuy-mobile-picker-heading h2 {
    margin: 0;
    overflow: hidden;
    color: #fff;
    font-size: 1rem;
    font-weight: 850;
    line-height: 1.2;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .quickbuy-mobile-picker-close {
    margin-left: auto;
  }

  .quickbuy-mobile-picker-body {
    display: flex;
    flex: 1;
    min-height: 0;
    overflow: hidden;
    padding: 0 0.5rem max(0.5rem, var(--tz-safe-area-bottom, 0px));
  }

  .quickbuy-mobile-picker-body :deep(.quickbuy-selection-panel) {
    width: 100%;
    min-height: 0;
    padding: 0.5rem;
    border-radius: 0.75rem;
  }

  .quickbuy-mobile-picker-body :deep(.quickbuy-panel-heading-row) {
    justify-content: flex-end;
  }

  .quickbuy-mobile-picker-body :deep(.quickbuy-candidate-area) {
    flex: 1;
    min-height: 0;
  }

  .quickbuy-mobile-picker-body :deep(.quickbuy-product-grid-shell) {
    min-height: 0;
    flex: 1;
  }
}
</style>
