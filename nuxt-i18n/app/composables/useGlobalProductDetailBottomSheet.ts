import { computed, ref } from 'vue'
import { useOverlayBackStack } from '~/composables/useOverlayBackStack'

export interface GlobalProductDetailBottomSheetProductReference {
  id?: number | null
  slug: string
  title?: string
  thumbnail?: string
}

const isGlobalProductDetailBottomSheetOpen = ref(false)
const globalProductDetailBottomSheetProductReference = ref<GlobalProductDetailBottomSheetProductReference | null>(null)
const overlayBackStack = useOverlayBackStack()

export const useGlobalProductDetailBottomSheet = () => {
  const globalProductDetailBottomSheetProductSlug = computed(() =>
    globalProductDetailBottomSheetProductReference.value?.slug || '',
  )

  const closeGlobalProductDetailBottomSheetState = () => {
    isGlobalProductDetailBottomSheetOpen.value = false
    globalProductDetailBottomSheetProductReference.value = null
  }

  const openGlobalProductDetailBottomSheet = (
    product: GlobalProductDetailBottomSheetProductReference | string,
  ) => {
    const reference = typeof product === 'string'
      ? { slug: product }
      : product
    const slug = String(reference.slug || '').trim()
    if (!slug) return

    globalProductDetailBottomSheetProductReference.value = {
      id: reference.id ?? null,
      slug,
      title: reference.title || '',
      thumbnail: reference.thumbnail || '',
    }
    isGlobalProductDetailBottomSheetOpen.value = true
    overlayBackStack.open('product-detail-sheet', closeGlobalProductDetailBottomSheetState)
  }

  const closeGlobalProductDetailBottomSheet = () => {
    void overlayBackStack.close('product-detail-sheet')
    closeGlobalProductDetailBottomSheetState()
  }

  return {
    isGlobalProductDetailBottomSheetOpen,
    globalProductDetailBottomSheetProductReference,
    globalProductDetailBottomSheetProductSlug,
    openGlobalProductDetailBottomSheet,
    closeGlobalProductDetailBottomSheet,
  }
}
