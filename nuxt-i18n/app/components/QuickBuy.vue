<template>
  <!-- 弹窗模态框 (由 GradientDockMenu 触发) -->
  <teleport to="body">
    <!-- 遮罩层 -->
    <Transition name="fade">
      <div
        class="fixed inset-0 z-[10002] flex items-center justify-center p-0 md:p-4"
        @click.self="handleClose"
      >
        <!-- 半透明背景遮罩 -->
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm"></div>
        <!-- 弹窗内容 -->
        <Transition name="slide-up" appear>
          <div
            class="sidebar-panel quickbuy-modal-shell relative max-w-[1400px] w-full h-[90vh] md:h-[700px] max-h-[85vh] bg-slate-950/80 backdrop-blur-xl border-2 border-[#6b73ff]/40 rounded-2xl shadow-[0_0_30px_rgba(107,115,255,0.6)] box-border flex flex-col overflow-hidden"
            role="dialog"
            aria-modal="true"
          >
        <!-- 背景装饰，与聊天欢迎页一致 -->
        <div class="absolute inset-x-0 top-0 h-[200px] bg-gradient-to-br from-indigo-600/20 to-teal-600/20 blur-3xl pointer-events-none z-0"></div>
        <!-- 头部 -->
        <header class="flex items-center justify-between px-3.5 max-md:px-2 py-2.5 max-md:py-2 border-b border-white/10 rounded-t-2xl overflow-hidden max-md:gap-1.5">
          <nav class="flex-1 min-w-0 overflow-hidden max-md:flex-auto" aria-label="quick-buy-steps">
            <ol class="flex items-center justify-center gap-3 max-md:gap-1.5 list-none m-0 p-0 max-md:flex-nowrap">
              <li
                v-for="n in 5"
                :key="n"
                class="inline-flex items-center gap-3 max-md:gap-1.5"
              >
                <span 
                  class="w-7 h-7 max-md:w-[22px] max-md:h-[22px] rounded-full grid place-items-center font-bold transition-all duration-200"
                  :class="[
                    n === step ? 'bg-[var(--accent-color,#6b73ff)] text-white shadow-[0_0_0_3px_rgba(107,115,255,0.25)]' :
                    n < step ? 'bg-[#3c4454] text-white' :
                    'bg-[#2c2f35] text-white/90'
                  ]"
                >{{ n }}</span>
                <span v-if="n < 5" class="w-8 max-md:w-2.5 h-1 rounded-full bg-white/[0.18]" aria-hidden="true" />
              </li>
            </ol>
          </nav>
          <button 
            class="flex-none ml-1.5 max-md:ml-0 appearance-none border-0 bg-transparent text-white text-[22px] cursor-pointer px-2 py-1 rounded-lg hover:bg-white/10 transition-colors" 
            type="button" 
            aria-label="close" 
            @click="handleClose"
          >×</button>
        </header>

        <!-- 主体内容 -->
        <section class="px-3.5 py-3 flex flex-col gap-3 flex-1 min-h-0 overflow-y-auto overflow-x-hidden">
          <div class="w-full min-w-0 overflow-hidden">
            <div v-if="currentCategoryName" class="flex items-center gap-2 mb-1.5 text-white text-[13px] opacity-90">
              <span class="opacity-70">Category</span>
              <span>{{ currentCategoryName }}</span>
            </div>
            <input
              v-model.trim="query"
              type="text"
              placeholder="Search products... (Enter or pause to search)"
              class="w-full px-3 py-2.5 rounded-lg bg-white/[0.06] text-white border border-white box-border max-w-full focus:outline-none focus:border-[#6b73ff] transition-colors"
              @keydown.enter.prevent="triggerSearch"
              @input="scheduleSearch"
            />
          </div>
          
          <div class="flex-1 min-h-0">
            <div v-if="loading" class="p-2.5 text-white opacity-85">Loading...</div>
            <div v-else-if="error" class="p-2.5 text-white opacity-85">{{ error }}</div>
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
                  <div class="text-[#40ffaa]">${{ (product.prices?.sale || product.prices?.regular || 0).toFixed(2) }}</div>
                </div>
              </li>
            </ul>
            <div v-else class="p-2.5">
              <h2 class="my-2 text-lg text-white">Step {{ step }}</h2>
              <p class="m-0 text-white/90">{{ stepHint }}</p>
            </div>
          </div>
        </section>

        <!-- 底部 -->
        <footer class="relative flex flex-col items-center justify-center gap-1.5 max-md:gap-1 px-3.5 py-2.5 max-md:pt-4 border-t border-white/[0.08] rounded-b-2xl overflow-hidden">
          <div class="text-white/80 text-[13px] text-center max-md:order-1 max-md:-mb-1">{{ footerText }}</div>
          <div class="inline-flex items-center gap-2 text-white font-semibold max-md:order-3 max-md:text-[13px] max-md:-mt-1">
            <span>Items: {{ totalQty }}</span>
            <span class="opacity-50">·</span>
            <span>Weight: {{ totalWeightG }}g</span>
            <span class="opacity-50">·</span>
            <span>Price: ${{ formattedTotalPrice }}</span>
          </div>
          <div class="inline-flex gap-2 justify-center flex-wrap max-md:order-4 max-md:mt-1">
            <button 
              class="appearance-none border border-white bg-white/[0.08] text-white px-3.5 py-2 rounded-full cursor-pointer hover:bg-white/[0.15] disabled:opacity-60 disabled:cursor-not-allowed transition-colors" 
              type="button" 
              :disabled="step <= 1" 
              @click="prev"
            >Prev</button>
            <button 
              v-if="step < 5" 
              class="appearance-none border border-[#6b73ff] bg-[#6b73ff] text-white px-3.5 py-2 rounded-full cursor-pointer hover:brightness-110 transition-all" 
              type="button" 
              @click="next"
            >Next</button>
            <template v-else>
              <button 
                class="appearance-none border border-[#6b73ff] bg-[#6b73ff] text-white px-3.5 py-2 rounded-full cursor-pointer hover:brightness-110 transition-all" 
                type="button" 
                @click="goToCart"
              >To cart</button>
              <button 
                class="appearance-none border border-[#6b73ff] bg-[#6b73ff] text-white px-3.5 py-2 rounded-full cursor-pointer hover:brightness-110 transition-all" 
                type="button" 
                @click="goToCheckout"
              >Payment</button>
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
import { computed, onBeforeUnmount, ref } from 'vue'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import type { CartItem } from '~/types/cart'

type Maybe<T> = T | null | undefined

interface QuickBuyStep {
  id: number
  slug: string
  name: string
}

interface QuickBuyConfig {
  steps?: QuickBuyStep[]
  storeApiBase?: string
  cartUrl?: string
  checkoutUrl?: string
  taxonomy?: string
  buttonText?: string
  enabled?: boolean
}

interface GoProduct {
  id: number
  title: string
  slug: string
  thumbnail: string
  prices: { regular: number; sale: number }
  stock: { quantity: number }
}

interface Selection {
  id: number
  title: string
  slug: string
  thumbnail: string
  qty: number
  weight_g: number
  price: number
}

const props = defineProps<{ config: QuickBuyConfig | null }>()
const emit = defineEmits<{ close: [] }>()

const step = ref(1)
const query = ref('')
const products = ref<GoProduct[]>([])
const loading = ref(false)
const error = ref('')
const selections = ref<Selection[]>([])

let searchTimer: Maybe<number> = null

const auth = useAuth()
const { addToCart } = useCart()

const qbConfig = computed(() => props.config || {})
const steps = computed(() => qbConfig.value.steps || [])
const currentStepConf = computed(() => steps.value[step.value - 1] || { id: 0, slug: '', name: '' })
const currentCategorySlug = computed(() => currentStepConf.value.slug || '')
const currentCategoryName = computed(() => currentStepConf.value.name || '')

const stepHint = computed(() => {
  return (
    {
      1: 'Search or filter products',
      2: 'Select specifications/quantity',
      3: 'Confirm product information',
      4: 'Complete your order',
      5: 'Finish and review'
    } as Record<number, string>
  )[step.value]
})

const footerText = computed(() => {
  return (
    {
      1: 'Enter keywords in the search box to filter products',
      2: 'Select product specifications and quantity to continue',
      3: 'Please confirm the product information is correct',
      4: 'Complete this step to submit/continue your process',
      5: 'All steps complete'
    } as Record<number, string>
  )[step.value]
})

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

const fetchProducts = async () => {
  loading.value = true
  error.value = ''

  try {
    const params = new URLSearchParams()
    if (query.value) params.set('keyword', query.value)
    params.set('per_page', '12')
    params.set('status', 'active')

    const res = await auth.request<any>(`/customer-service/products?${params.toString()}`)
    products.value = res.items || []
  } catch (err) {
    error.value = (err as Error).message || String(err)
    products.value = []
  } finally {
    loading.value = false
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

const next = () => {
  if (step.value < 5) {
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
      title: item.title,
      slug: item.slug,
      thumbnail: item.thumbnail,
      price: item.price,
      weight: item.weight_g
    } as Omit<CartItem, 'quantity'>)
  }
}

const goToCart = () => {
  addSelectionsToCart()
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('open-cart-drawer'))
  }
  emit('close')
}

const goToCheckout = () => {
  addSelectionsToCart()
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('open-checkout-modal'))
  }
  emit('close')
}

const selectProduct = (product: GoProduct) => {
  const price = product.prices?.sale || product.prices?.regular || 0
  selections.value.push({
    id: product.id,
    title: product.title,
    slug: product.slug,
    thumbnail: product.thumbnail,
    qty: 1,
    weight_g: 0,
    price
  })
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
    height: min(95vh, calc(100vh - 16px));
    max-height: min(95vh, calc(100vh - 16px));
  }

  @supports (height: 100svh) {
    .quickbuy-modal-shell {
      height: min(95svh, calc(100svh - 16px));
      max-height: min(95svh, calc(100svh - 16px));
    }
  }

  @supports (height: 100dvh) {
    .quickbuy-modal-shell {
      height: min(95dvh, calc(100dvh - 16px));
      max-height: min(95dvh, calc(100dvh - 16px));
    }
  }
}
</style>
