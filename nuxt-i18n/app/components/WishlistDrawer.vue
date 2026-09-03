<template>
  <ClientOnly>
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
          <!-- Header -->
          <div class="wa-drawer-header relative z-10">
            <div class="flex flex-col gap-1 min-w-0">
              <div class="wa-drawer-title">
                {{ t('wishlistDrawer.title') }}
              </div>
              <div class="tz-caption tz-text-muted truncate">
                {{ t('wishlistDrawer.subtitle') }}
              </div>
            </div>
            <button
              type="button"
              class="wa-drawer-close-btn"
              :aria-label="t('wishlistDrawer.closeAriaLabel')"
              @click="handleClose"
            >
              <Icon name="lucide:x" class="h-3.5 w-3.5" />
            </button>
          </div>

          <!-- Content -->
          <div class="wa-drawer-content relative z-10">
            <div
              v-if="loading"
              class="flex flex-col items-center justify-center h-full tz-text-secondary text-sm gap-3"
            >
              <svg class="animate-spin h-6 w-6 tz-text-secondary" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
              <span>{{ t('wishlistDrawer.loading') }}</span>
            </div>

            <div
              v-else-if="error"
              class="flex items-center justify-center h-full text-red-300 text-sm text-center px-4"
            >
              {{ error }}
            </div>

            <div
              v-else-if="!items.length"
              class="flex flex-col items-center justify-center h-full tz-text-secondary text-sm text-center px-4 gap-2"
            >
              <svg class="w-10 h-10 tz-text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.7" d="M12.1 19.3 12 19.4l-.1-.1C7.14 15.24 4 12.39 4 9.2 4 7 5.7 5.3 7.9 5.3c1.4 0 2.8.7 3.6 1.9 0.8-1.2 2.2-1.9 3.6-1.9 2.2 0 3.9 1.7 3.9 3.9 0 3.19-3.14 6.04-7.9 10.1z" />
              </svg>
              <p class="font-medium tz-text-primary">{{ t('wishlistDrawer.empty.title') }}</p>
              <p class="text-xs tz-text-secondary max-w-md">
                {{ t('wishlistDrawer.empty.description') }}
              </p>
            </div>

            <div
              v-else
              class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 md:gap-4"
            >
              <div
                v-for="item in items"
                :key="item.id"
                class="border tz-border-subtle rounded-xl tz-surface-subtle hover:tz-surface-subtle transition-colors overflow-hidden flex flex-col"
              >
                <StorefrontImage
                  v-if="item.product?.thumbnail"
                  :src="item.product.thumbnail"
                  :alt="item.product?.name || t('wishlistDrawer.productImageAlt')"
                  class="w-full h-32 object-cover"
                  preset="card"
                />
                <div class="px-3 pt-2 pb-3 flex-1 flex flex-col">
                  <div class="text-sm font-semibold tz-text-primary truncate">
                    {{ item.product?.name || t('wishlistDrawer.productFallback') }}
                  </div>
                  <div v-if="displayPrice(item)" class="text-xs text-[#059669] mt-1">
                    {{ displayPrice(item) }}
                  </div>
                  <div class="mt-3 flex justify-end gap-2">
                    <button
                      type="button"
                      class="text-xs px-2 py-1 rounded-full bg-[var(--tz-action-primary)] text-white hover:bg-[var(--tz-action-primary-hover)] transition-colors shadow-sm"
                      @click="handleShare(item)"
                    >
                      {{ t('wishlistDrawer.actions.shareToChat') }}
                    </button>
                    <button
                      type="button"
                      class="text-xs px-2 py-1 rounded-full border tz-border-strong/30 tz-text-secondary hover:tz-surface-subtle transition-colors"
                      @click="handleRemove(item.id)"
                    >
                      {{ t('wishlistDrawer.actions.remove') }}
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
  </ClientOnly>
</template>

<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { createDialogStackId, useDialogStack } from '~/composables/useDialogStack'
import { useWishlist } from '~/composables/useWishlist'
import { buildProductPath } from '~/utils/seo/urls'

const props = defineProps<{
  variant?: 'default' | 'bottom'
}>()

const modelValue = defineModel<boolean>({ default: false })

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'share-to-chat', product: any): void
}>()

const { items, loading, error, loadWishlist, removeFromWishlist } = useWishlist()
const { t } = useI18n()
const dialogStack = useDialogStack()
const dialogStackId = createDialogStackId('wishlist-drawer')
let unregisterDialogStack: (() => void) | null = null

const handleClose = () => {
  modelValue.value = false
  emit('close')
}

const syncDialogStack = (isOpen: boolean) => {
  if (isOpen && !unregisterDialogStack) {
    unregisterDialogStack = dialogStack.register(dialogStackId, () => {
      handleClose()
    }, {
      priority: 10060,
    })
    return
  }

  if (!isOpen && unregisterDialogStack) {
    unregisterDialogStack()
    unregisterDialogStack = null
  }
}

watch(
  () => modelValue.value,
  (val) => {
    syncDialogStack(val)
    if (val) {
      loadWishlist()
    }
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  syncDialogStack(false)
})

const displayPrice = (item: any) => {
  const product = item?.product
  if (!product) return ''
  if (product.sale_price && product.sale_price > 0) return `$${product.sale_price}`
  if (product.price && product.price > 0) return `$${product.price}`
  return ''
}

const handleShare = (item: any) => {
  if (!item || !item.product) return
  const product = item.product
  const price = displayPrice(item)
  const payload = {
    id: product.id ?? item.product_id,
    title: product.name,
    url: buildProductPath(String(product.slug || product.id || '')),
    thumbnail: product.thumbnail,
    price,
  }
  emit('share-to-chat', payload)
}

const handleRemove = async (id: number) => {
  await removeFromWishlist(id)
}
</script>

<style src="~/assets/css/components/whatsapp-mobile-drawer.css"></style>

<style scoped>
/* Scoped styles removed in favor of global .wa-drawer classes */
</style>

