import { computed, ref } from 'vue'

export interface GlobalProductDetailBottomSheetProductReference {
  id?: number | null
  slug: string
  title?: string
  thumbnail?: string
}

const isGlobalProductDetailBottomSheetOpen = ref(false)
const globalProductDetailBottomSheetProductReference = ref<GlobalProductDetailBottomSheetProductReference | null>(null)

export const useGlobalProductDetailBottomSheet = () => {
  const globalProductDetailBottomSheetProductSlug = computed(() =>
    globalProductDetailBottomSheetProductReference.value?.slug || '',
  )

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
  }

  const closeGlobalProductDetailBottomSheet = () => {
    isGlobalProductDetailBottomSheetOpen.value = false
  }

  return {
    isGlobalProductDetailBottomSheetOpen,
    globalProductDetailBottomSheetProductReference,
    globalProductDetailBottomSheetProductSlug,
    openGlobalProductDetailBottomSheet,
    closeGlobalProductDetailBottomSheet,
  }
}
