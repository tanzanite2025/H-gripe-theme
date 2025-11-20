<template>
  <Teleport to="body">
    <Transition name="whatsapp-product-drawer">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[10001] flex items-end justify-center p-0 md:p-4 pointer-events-none"
      >
        <div
          class="pointer-events-auto w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh]
                 rounded-2xl border-2 border-[#6b73ff] bg-black shadow-[0_0_30px_rgba(107,115,255,0.6)]
                 flex flex-col overflow-hidden"
        >
          <!-- Header -->
          <div class="flex items-center justify-between px-4 py-3 border-b border-white/10">
            <div class="flex flex-col gap-1 min-w-0">
              <div class="text-sm font-semibold text-white/90 truncate">
                Wishlist
              </div>
              <div class="text-[11px] text-white/50 truncate">
                Products you add to your wishlist will appear here.
              </div>
            </div>
            <button
              type="button"
              class="w-8 h-8 rounded-full border border-white/40 text-white flex items-center justify-center hover:bg-white/10 transition-colors"
              @click="handleClose"
            >
              <span class="text-lg leading-none">x</span>
            </button>
          </div>

          <!-- Content (placeholder) -->
          <div class="flex-1 min-h-0 overflow-y-auto p-4 md:p-6">
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
    url: product.preview_url || `/product/${product.slug || product.id}`,
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
.whatsapp-product-drawer-enter-active,
.whatsapp-product-drawer-leave-active {
  transition: transform 0.3s ease-out, opacity 0.3s ease-out;
}

.whatsapp-product-drawer-enter-from,
.whatsapp-product-drawer-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.whatsapp-product-drawer-enter-to,
.whatsapp-product-drawer-leave-from {
  transform: translateY(0%);
  opacity: 1;
}
</style>
