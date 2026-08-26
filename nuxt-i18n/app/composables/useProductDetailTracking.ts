import {
  computed,
  onBeforeUnmount,
  onMounted,
  toValue,
  watch,
  type MaybeRefOrGetter,
} from 'vue'
import { useRoute } from '#imports'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import type { GoProduct } from '~/types/productDetail'

export interface ProductDetailTrackingOptions {
  product: MaybeRefOrGetter<GoProduct | null | undefined>
  formattedPrice: MaybeRefOrGetter<string>
  primaryMediaThumbnail: MaybeRefOrGetter<string>
}

type ProductDwellEndReason = 'product_change' | 'visibility_hidden' | 'unmount'

export function useProductDetailTracking(options: ProductDetailTrackingOptions) {
  const route = useRoute()
  const { addToHistory } = useBrowsingHistory()
  const { track: trackBehaviorEvent } = useBehaviorEvents()
  const product = computed(() => toValue(options.product) || null)
  const formattedPrice = computed(() => String(toValue(options.formattedPrice) || ''))
  const primaryMediaThumbnail = computed(() => String(toValue(options.primaryMediaThumbnail) || ''))

  let activeTrackedProductID = 0
  let productVisibleSince = 0

  const trackProductDwell = (reason: ProductDwellEndReason) => {
    if (!import.meta.client || !activeTrackedProductID || !productVisibleSince) return

    const durationSeconds = Math.min(
      1800,
      Math.max(0, Math.round((Date.now() - productVisibleSince) / 1000)),
    )
    if (durationSeconds < 1) {
      productVisibleSince = reason === 'visibility_hidden' ? 0 : productVisibleSince
      return
    }

    trackBehaviorEvent({
      eventType: 'product_dwell',
      productId: activeTrackedProductID,
      metadata: {
        surface: 'product_page',
        duration_seconds: durationSeconds,
        end_reason: reason,
      },
    })
    productVisibleSince = reason === 'visibility_hidden' ? 0 : Date.now()
  }

  const handleProductVisibilityChange = () => {
    if (!import.meta.client) return
    if (document.visibilityState === 'hidden') {
      trackProductDwell('visibility_hidden')
      return
    }
    if (activeTrackedProductID && !productVisibleSince) {
      productVisibleSince = Date.now()
    }
  }

  watch(product, (currentProduct) => {
    if (!import.meta.client || !currentProduct) return

    const productID = Number(currentProduct.id)
    if (!Number.isInteger(productID) || productID <= 0) return

    if (activeTrackedProductID !== productID) {
      trackProductDwell('product_change')
      activeTrackedProductID = productID
      productVisibleSince = Date.now()
      trackBehaviorEvent({
        eventType: 'product_view',
        productId: productID,
        metadata: {
          surface: 'product_page',
          product_specification_template: currentProduct.product_specification_template?.slug || '',
        },
      })
    }

    addToHistory({
      id: productID,
      title: currentProduct.name,
      thumbnail: primaryMediaThumbnail.value,
      price: formattedPrice.value,
      url: route.path,
    })
  }, { immediate: true })

  onMounted(() => {
    document.addEventListener('visibilitychange', handleProductVisibilityChange)
  })

  onBeforeUnmount(() => {
    trackProductDwell('unmount')
    document.removeEventListener('visibilitychange', handleProductVisibilityChange)
  })
}
