<template>
  <Teleport to="body">
    <Transition name="wa-drawer">
      <div
        v-if="modelValue"
        class="wa-drawer-mask"
        :class="{ '!items-end': props.variant === 'bottom' }"
        @click.self="handleClose"
      >
        <!-- Backdrop -->
        <div 
          class="wa-drawer-backdrop"
          :class="{ 'md:hidden': props.variant === 'bottom' }" 
        ></div>
        <!-- 
           Note: wa-drawer-backdrop is md:hidden by default in CSS. 
           If variant != 'bottom' (e.g. default center modal), we might want backdrop on desktop too?
           TestReportDrawer didn't have desktop backdrop in the code I copied (it was hidden).
           If Wishlist needs desktop backdrop, we can override or leave it as per standard.
           The standard from TestDrawer implies no desktop backdrop or handled by mask?
           TestDrawer: backdrop div has md:hidden.
           I will stick to the standard: md:hidden (no backdrop on desktop, or transparent).
        -->

        <div class="wa-drawer-shell">
          <!-- Background Decoration -->
          <div class="absolute inset-x-0 top-0 h-[200px] bg-gradient-to-br from-indigo-600/20 to-teal-600/20 blur-3xl pointer-events-none z-0"></div>

          <!-- Header -->
          <div class="wa-drawer-header relative z-10">
            <div class="flex flex-col gap-1 min-w-0">
              <div class="wa-drawer-title">
                Wishlist
              </div>
              <div class="text-[11px] text-white/50 truncate">
                Products you add to your wishlist will appear here.
              </div>
            </div>
            <button
              type="button"
              class="wa-drawer-close-btn"
              @click="handleClose"
            >
              <span class="text-lg leading-none">x</span>
            </button>
          </div>

          <!-- Content -->
          <div class="wa-drawer-content relative z-10">
            <div
              v-if="loading"
              class="flex flex-col items-center justify-center h-full text-white/70 text-sm gap-3"
            >
              <svg class="animate-spin h-6 w-6 text-white/60" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
              <span>Loading wishlist...</span>
            </div>

            <div
              v-else-if="error"
              class="flex items-center justify-center h-full text-red-300 text-sm text-center px-4"
            >
              {{ error }}
            </div>

            <div
              v-else-if="!items.length"
              class="flex flex-col items-center justify-center h-full text-white/60 text-sm text-center px-4 gap-2"
            >
              <svg class="w-10 h-10 text-white/30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M12.1 19.3 12 19.4l-.1-.1C7.14 15.24 4 12.39 4 9.2 4 7 5.7 5.3 7.9 5.3c1.4 0 2.8.7 3.6 1.9 0.8-1.2 2.2-1.9 3.6-1.9 2.2 0 3.9 1.7 3.9 3.9 0 3.19-3.14 6.04-7.9 10.1z" />
              </svg>
              <p class="font-medium text-white/80">Your wishlist is empty</p>
              <p class="text-xs text-white/60 max-w-md">
                Save products you like to your wishlist so you can quickly find and share them later.
              </p>
            </div>

            <div
              v-else
              class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 md:gap-4"
            >
              <div
                v-for="item in items"
                :key="item.id"
                class="border border-white/10 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] transition-colors overflow-hidden flex flex-col"
              >
                <img
                  v-if="item.product?.thumbnail"
                  :src="item.product.thumbnail"
                  alt="Product image"
                  class="w-full h-32 object-cover"
                />
                <div class="px-3 pt-2 pb-3 flex-1 flex flex-col">
                  <div class="text-sm font-semibold text-white truncate">
                    {{ item.product?.title || 'Product' }}
                  </div>
                  <div v-if="displayPrice(item)" class="text-xs text-[#40ffaa] mt-1">
                    {{ displayPrice(item) }}
                  </div>
                  <div class="mt-3 flex justify-end gap-2">
                    <button
                      type="button"
                      class="text-xs px-2 py-1 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-white hover:from-[#35e599] hover:to-[#5a62ee] transition-colors shadow-sm"
                      @click="handleShare(item)"
                    >
                      Share to chat
                    </button>
                    <button
                      type="button"
                      class="text-xs px-2 py-1 rounded-full border border-white/30 text-white/80 hover:bg-white/10 transition-colors"
                      @click="handleRemove(item.id)"
                    >
                      Remove
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useWishlist } from '~/composables/useWishlist'

const props = defineProps<{
  modelValue: boolean
  variant?: 'default' | 'bottom'
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'close'): void
  (e: 'share-to-chat', product: any): void
}>()

const { items, loading, error, loadWishlist, removeFromWishlist } = useWishlist()

watch(
  () => props.modelValue,
  (val) => {
    if (val) {
      loadWishlist()
    }
  },
)

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}

const displayPrice = (item: any) => {
  const prices = item?.product?.prices
  if (!prices) return ''
  if (prices.sale && prices.sale > 0) return `$${prices.sale}`
  if (prices.regular && prices.regular > 0) return `$${prices.regular}`
  return ''
}

const handleShare = (item: any) => {
  if (!item || !item.product) return
  const product = item.product
  const price = displayPrice(item)
  const payload = {
    id: product.id ?? item.product_id,
    title: product.title,
    url: product.preview_url || `/shop/${product.slug || product.id}`,
    thumbnail: product.thumbnail,
    price,
  }
  emit('share-to-chat', payload)
}

const handleRemove = async (id: number) => {
  await removeFromWishlist(id)
}
</script>

<style scoped>
/* Scoped styles removed in favor of global .wa-drawer classes */
</style>
