<template>
  <!-- 弹窗模态框 (由 GradientDockMenu 触发) -->
  <teleport to="body">
    <!-- 遮罩层 -->
    <Transition name="fade">
      <div
        class="fixed inset-0 z-[10002] flex items-center justify-center p-0 md:p-4 tz-mobile-safe-modal-mask"
        @click.self="handleClose"
      >
        <!-- 半透明背景遮罩 -->
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm"></div>
        <!-- 弹窗内容 -->
        <Transition name="slide-up" appear>
          <div
            class="sidebar-panel quickbuy-modal-shell relative w-[90vw] max-w-none h-[80vh] max-h-[80vh] bg-black backdrop-blur-xl border border-white/15 rounded-2xl shadow-[0_18px_44px_rgba(0,0,0,0.72)] box-border flex flex-col overflow-hidden"
            role="dialog"
            aria-modal="true"
          >
        <!-- 头部 -->
        <header class="flex items-center justify-between px-3.5 max-md:px-2 py-2.5 max-md:py-2 border-b border-white/10 rounded-t-2xl overflow-hidden max-md:gap-1.5">
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
                    n === step ? 'bg-white text-black shadow-[0_0_0_3px_rgba(255,255,255,0.16)]' :
                    n < step ? 'bg-[#3c4454] text-white' :
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
        <section class="px-3.5 py-3 flex flex-col gap-3 flex-1 min-h-0 overflow-y-auto overflow-x-hidden">
          <div class="w-full min-w-0 overflow-hidden">
            <div v-if="currentCategoryName" class="flex items-center gap-2 mb-1.5 tz-text-secondary text-[13px]">
              <span class="tz-text-muted">{{ t('quickBuy.search.categoryLabel') }}</span>
              <span>{{ currentCategoryName }}</span>
            </div>
            <input
              v-model.trim="query"
              type="text"
              :placeholder="t('quickBuy.search.placeholder')"
              class="quickbuy-search-input w-full px-3 py-2.5 rounded-lg bg-white/[0.06] text-white border border-white box-border max-w-full focus:outline-none transition-colors"
              @keydown.enter.prevent="triggerSearch"
              @input="scheduleSearch"
            />
          </div>
          
          <div class="flex-1 min-h-0">
            <div v-if="loading" class="p-2.5 tz-text-secondary">{{ t('common.loading', 'Loading...') }}</div>
            <div v-else-if="error" class="p-2.5 tz-text-secondary">{{ error }}</div>
            <ul v-else-if="products.length" class="list-none grid grid-cols-[repeat(auto-fill,minmax(180px,1fr))] gap-2.5 m-0 p-0">
              <li 
                v-for="product in products" 
                :key="product.id" 
                class="flex gap-2.5 p-2 border border-white rounded-[10px] bg-white/[0.06] cursor-pointer hover:bg-white/[0.12] transition-colors"
                @click="selectProduct(product)"
              >
                <img
                  v-if="product.thumbnail"
                  :src="product.thumbnail"
                  :alt="product.title"
                  class="w-14 h-14 object-cover rounded-lg"
                />
                <div class="flex flex-col gap-1">
                  <div class="text-sm text-white">{{ product.title }}</div>
                  <div class="text-white/90">{{ product.priceLabel || '$0' }}</div>
                </div>
              </li>
            </ul>
            <div v-else class="p-2.5">
              <h2 class="my-2 text-lg text-white">{{ t('quickBuy.placeholder.stepTitle', { step }) }}</h2>
              <p class="m-0 tz-text-secondary">
                {{ currentStepConf.helpText || currentStepConf.description || t('quickBuy.emptyStep', 'No products are available for this step.') }}
              </p>
            </div>
          </div>
        </section>

        <!-- 底部 -->
        <footer class="quickbuy-modal-footer relative flex flex-col items-center justify-center gap-1.5 max-md:gap-1 px-3.5 py-2.5 max-md:pt-4 border-t border-white/[0.08] rounded-b-2xl overflow-hidden">
          <div class="tz-text-secondary text-[13px] text-center max-md:order-1 max-md:-mb-1">{{ footerText }}</div>
          <div class="inline-flex items-center gap-2 tz-text-primary font-semibold max-md:order-3 max-md:text-[13px] max-md:-mt-1">
            <span>{{ t('quickBuy.summary.items') }}: {{ totalQty }}</span>
            <span class="opacity-50">·</span>
            <span>{{ t('quickBuy.summary.weight') }}: {{ totalWeightG }}g</span>
            <span class="opacity-50">·</span>
            <span>{{ t('quickBuy.summary.price') }}: ${{ formattedTotalPrice }}</span>
          </div>
          <div class="inline-flex gap-2 justify-center flex-wrap max-md:order-4 max-md:mt-1">
            <button 
              class="appearance-none border border-white bg-white/[0.08] text-white px-3.5 py-2 rounded-full cursor-pointer hover:bg-white/[0.15] disabled:opacity-60 disabled:cursor-not-allowed transition-colors" 
              type="button" 
              :disabled="step <= 1" 
              @click="prev"
            >{{ t('common.previous', 'Previous') }}</button>
            <button
              v-if="step < totalSteps"
              class="appearance-none border border-white bg-white text-black px-3.5 py-2 rounded-full cursor-pointer hover:bg-white/90 transition-colors"
              type="button" 
              @click="next"
            >{{ t('common.next', 'Next') }}</button>
            <template v-else>
              <button 
                class="appearance-none border border-white bg-white text-black px-3.5 py-2 rounded-full cursor-pointer hover:bg-white/90 transition-colors"
                type="button" 
                @click="goToCart"
              >{{ t('quickBuy.actions.toCart') }}</button>
            </template>
          </div>
        </footer>
          </div>
        </Transition>
      </div>
    </Transition>
  </teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useCart } from '~/composables/useCart'
import { useShopProducts } from '~/composables/useShopProducts'
import type { ShopProduct } from '~/composables/useShopProducts'
import type {
  QuickBuyConfig,
  QuickBuyStep,
} from '~/utils/quickBuy/types'
import type { CartItem } from '~~/types/cart'

type Maybe<T> = T | null | undefined

interface Selection {
  id: number
  stepKey: string
  variant_id?: number | null
  title: string
  slug: string
  thumbnail: string
  qty: number
  weight_g: number
  price: number
}

const props = defineProps<{ config: QuickBuyConfig | null }>()
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()

const defaultQuickBuySteps = computed<QuickBuyStep[]>(() => [
  {
    id: 1,
    slug: 'product-search',
    name: t('quickBuy.defaultSteps.productSearch', 'Search products'),
    description: t('quickBuy.hints.step1', 'Search or filter products'),
  },
  {
    id: 2,
    slug: 'specifications',
    name: t('quickBuy.defaultSteps.specifications', 'Choose specifications'),
    description: t('quickBuy.hints.step2', 'Select product specifications and quantity to continue'),
  },
  {
    id: 3,
    slug: 'quantity',
    name: t('quickBuy.defaultSteps.quantity', 'Confirm quantity'),
    description: t('quickBuy.hints.step3', 'Confirm product information'),
  },
  {
    id: 4,
    slug: 'cart-review',
    name: t('quickBuy.defaultSteps.cartReview', 'Review cart'),
    description: t('quickBuy.hints.step4', 'Complete this step to submit or continue your process'),
  },
  {
    id: 5,
    slug: 'checkout',
    name: t('quickBuy.defaultSteps.checkout', 'Checkout'),
    description: t('quickBuy.hints.step5', 'Finish and review'),
  },
])

const step = ref(1)
const query = ref('')
const products = ref<ShopProduct[]>([])
const loading = ref(false)
const error = ref('')
const selections = ref<Selection[]>([])

let searchTimer: Maybe<number> = null

const { addToCart } = useCart()
const {
  fetchStepCandidates,
  updateSelections: updateQuickBuySelections,
  error: quickBuySessionError,
} = useQuickBuySession('dock')
const { fetchShopProducts } = useShopProducts()

const configuredSteps = computed<QuickBuyStep[]>(() => props.config?.steps ?? [])
const hasConfiguredFlow = computed(() => configuredSteps.value.length > 0)
const steps = computed(() => hasConfiguredFlow.value ? configuredSteps.value : defaultQuickBuySteps.value)
const totalSteps = computed(() => steps.value.length)
const currentStepConf = computed<QuickBuyStep>(() => steps.value[step.value - 1] || { id: 0, slug: '', name: '' })
const currentCategoryName = computed(() => currentStepConf.value.name || '')
const currentStepKey = computed(() => currentStepConf.value.stepKey || currentStepConf.value.slug || `step-${step.value}`)
const footerText = computed(() =>
  currentStepConf.value.helpText
  || currentStepConf.value.description
  || t('quickBuy.footer.default', 'Complete this step to continue')
)

const totalQty = computed(() => selections.value.reduce((sum, item) => sum + (Number(item.qty) || 0), 0))

const totalWeightG = computed(() =>
  selections.value.reduce((sum, item) => sum + (Number(item.weight_g) || 0) * (Number(item.qty) || 0), 0)
)

const totalPrice = computed(() =>
  selections.value.reduce((sum, item) => sum + (Number(item.price) || 0) * (Number(item.qty) || 0), 0)
)

const formattedTotalPrice = computed(() => {
  try {
    const n = Number(totalPrice.value || 0)
    return new Intl.NumberFormat(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n)
  } catch (err) {
    return String(totalPrice.value)
  }
})

let fetchSequence = 0

const fetchProducts = async () => {
  const sequence = fetchSequence + 1
  fetchSequence = sequence
  loading.value = true
  error.value = ''

  try {
    if (hasConfiguredFlow.value) {
      const res = await fetchStepCandidates(currentStepKey.value, {
        keyword: query.value,
        page: 1,
        pageSize: 12,
      })
      if (sequence !== fetchSequence) return
      if (!res) {
        throw new Error(quickBuySessionError.value || 'Unable to load QUICK candidates')
      }
      products.value = res.items
    } else {
      const res = await fetchShopProducts({
        per_page: 12,
        status: 'active',
        ...(query.value ? { keyword: query.value } : {}),
      })
      if (sequence !== fetchSequence) return
      products.value = res.items
    }
  } catch (err) {
    if (sequence !== fetchSequence) return
    error.value = (err as Error).message || String(err)
    products.value = []
  } finally {
    if (sequence === fetchSequence) {
      loading.value = false
    }
  }
}

const scheduleSearch = () => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }
  searchTimer = window.setTimeout(() => {
    fetchProducts()
    searchTimer = null
  }, 300)
}

const triggerSearch = () => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
    searchTimer = null
  }
  fetchProducts()
}

watch([currentStepKey, hasConfiguredFlow], () => {
  query.value = ''
  products.value = []
  error.value = ''
  fetchProducts()
}, { immediate: true })

watch(totalSteps, (nextTotalSteps) => {
  if (step.value > nextTotalSteps) {
    step.value = nextTotalSteps
  }
})

const next = () => {
  if (step.value < totalSteps.value) {
    step.value += 1
  }
}

const prev = () => {
  if (step.value > 1) {
    step.value -= 1
  }
}

const handleClose = () => {
  emit('close')
}

const addSelectionsToCart = () => {
  for (const item of selections.value) {
    addToCart({
      id: item.id,
      product_id: item.id,
      variant_id: item.variant_id || null,
      title: item.title,
      slug: item.slug,
      thumbnail: item.thumbnail,
      price: item.price,
      weight: item.weight_g
    } as Omit<CartItem, 'quantity'>, item.qty)
  }
}

const goToCart = () => {
  addSelectionsToCart()
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('open-cart-drawer'))
  }
  emit('close')
}

const selectProduct = async (product: ShopProduct) => {
  const stepKeyForSelection = currentStepKey.value
  const selectedVariant = product.defaultVariantId
    ? product.variants.find(variant => Number(variant.id) === Number(product.defaultVariantId)) || null
    : product.variants.find(variant => variant.isDefault) || product.variants[0] || null
  const selectedVariantId = selectedVariant?.id ? Number(selectedVariant.id) : Number(product.defaultVariantId) || null
  const selection = {
    id: product.id,
    stepKey: stepKeyForSelection,
    variant_id: selectedVariantId,
    title: product.title,
    slug: product.slug,
    thumbnail: product.thumbnail || '',
    qty: currentStepConf.value.defaultQuantity || 1,
    weight_g: selectedVariant?.weightGrams || 0,
    price: product.priceNumber
  }

  const selectionMode = currentStepConf.value.selectionMode || 'single'
  const currentStepSelections = selections.value.filter(item => item.stepKey === stepKeyForSelection)
  const nextStepSelections = selectionMode === 'multiple'
    ? [...currentStepSelections.filter(item => item.id !== selection.id || item.variant_id !== selection.variant_id), selection]
    : [selection]
  if (hasConfiguredFlow.value) {
    const session = await updateQuickBuySelections(nextStepSelections.map(item => ({
      stepKey: item.stepKey,
      productId: item.id,
      variantId: item.variant_id || null,
      quantity: item.qty,
    })))
    if (!session) {
      error.value = quickBuySessionError.value || 'Unable to save QUICK selection'
      return
    }
  }

  if (selectionMode === 'multiple') {
    selections.value = [
      ...selections.value.filter(item => item.stepKey !== stepKeyForSelection),
      ...nextStepSelections,
    ]
  } else {
    const existingIndex = selections.value.findIndex(item => item.stepKey === stepKeyForSelection)
    if (existingIndex >= 0) {
      selections.value.splice(existingIndex, 1, selection)
    } else {
      selections.value.push(selection)
    }
  }
  next()
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

@media (max-width: 767px) {
  .quickbuy-modal-shell {
    height: min(80vh, var(--tz-mobile-safe-viewport-height, 80vh));
    max-height: min(80vh, var(--tz-mobile-safe-viewport-height, 80vh));
  }

  @supports (height: 100svh) {
    .quickbuy-modal-shell {
      height: min(80svh, var(--tz-mobile-safe-viewport-height, 80svh));
      max-height: min(80svh, var(--tz-mobile-safe-viewport-height, 80svh));
    }
  }

  @supports (height: 100dvh) {
    .quickbuy-modal-shell {
      height: min(80dvh, var(--tz-mobile-safe-viewport-height, 80dvh));
      max-height: min(80dvh, var(--tz-mobile-safe-viewport-height, 80dvh));
    }
  }

  .quickbuy-modal-footer {
    padding-bottom: var(--tz-mobile-modal-safe-padding-bottom, 0.75rem);
  }
}
</style>
