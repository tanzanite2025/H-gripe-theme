import {
  computed,
  ref,
  toValue,
  watch,
  type MaybeRefOrGetter,
} from 'vue'
import type {
  GoProduct,
  ProductGalleryItem,
  ProductPreviewMedia,
} from '~/types/productDetail'
import { parseProductVariantOptions } from '~/utils/productDetail'

export interface ProductDetailMediaOptions {
  product: MaybeRefOrGetter<GoProduct | null | undefined>
  selectedVariantId: MaybeRefOrGetter<number | null>
  metaTitle?: MaybeRefOrGetter<string>
}

const PRODUCT_DETAIL_MEDIA_SLOT_COUNT = 5

export function useProductDetailMedia(options: ProductDetailMediaOptions) {
  const product = computed(() => toValue(options.product) || null)
  const selectedVariantId = computed(() => {
    const value = Number(toValue(options.selectedVariantId) || 0)
    return Number.isFinite(value) && value > 0 ? value : null
  })
  const metaTitle = computed(() => String(toValue(options.metaTitle) || ''))
  const selectedMediaId = ref<string | null>(null)

  const productGalleryItems = computed<ProductGalleryItem[]>(() => {
    const currentProduct = product.value
    if (!currentProduct) return []

    const productThumbnail = String(currentProduct.thumbnail || '').trim()
    const currentVariant = (currentProduct.variants || []).find((variant) => (
      variant.id === selectedVariantId.value
    )) || null
    const currentOptions = parseProductVariantOptions(currentVariant)
    const optionValueIds = new Set(
      (currentProduct.variant_option_values || [])
        .filter((option) => (
          option.is_enabled !== false
          && currentOptions[option.spec_slug] === option.value_key
        ))
        .map((option) => Number(option.id))
        .filter((id) => Number.isFinite(id) && id > 0),
    )
    const visibleMedia = (currentProduct.media || []).filter((item) => (
      item.url && item.is_visible !== false
    ))
    const exactVariantMedia = selectedVariantId.value
      ? visibleMedia.filter((item) => Number(item.variant_id || 0) === Number(selectedVariantId.value))
      : []
    const optionMedia = visibleMedia.filter((item) => (
      optionValueIds.has(Number(item.variant_option_value_id || 0))
    ))
    const productMedia = visibleMedia.filter((item) => !item.variant_id && !item.variant_option_value_id)
    const scopedMedia = exactVariantMedia.length
      ? exactVariantMedia
      : optionMedia.length
        ? optionMedia
        : productMedia.length
          ? productMedia
          : visibleMedia
    const shouldIncludeProductThumbnail = exactVariantMedia.length === 0 && optionMedia.length === 0
    const mediaItems = scopedMedia
      .map((item, index): ProductGalleryItem | null => {
        const kind = String(item.media_type || '').toLowerCase() === 'video' ? 'video' : 'image'
        const sourceUrl = String(item.url || '').trim()
        if (!sourceUrl) return null

        const poster = String(item.poster_url || item.thumbnail_url || '').trim()
        const largeVariantUrl = kind === 'image'
          ? String(item.image_variants?.large?.url || '').trim()
          : ''
        const url = largeVariantUrl || sourceUrl
        const thumbnailUrl = kind === 'video'
          ? poster
          : String(item.thumbnail_url || item.image_variants?.thumbnail?.url || sourceUrl).trim()
        const matchesProductThumbnail = sourceUrl === productThumbnail
          || Object.values(item.image_variants || {}).some((variant) => (
            String(variant?.url || '').trim() === productThumbnail
          ))

        return {
          id: String(item.id ?? `product-media-${kind}-${index}-${sourceUrl}`),
          kind,
          url,
          thumbnailUrl,
          poster: kind === 'video' ? poster || undefined : undefined,
          alt: item.alt || item.title || currentProduct.name,
          isPrimary: Boolean(item.is_primary || item.role === 'primary' || matchesProductThumbnail),
          sourceIndex: index,
        }
      })
      .filter((item): item is ProductGalleryItem => Boolean(item))
      .sort((left, right) => {
        if (left.kind !== right.kind) return left.kind === 'image' ? -1 : 1
        if (left.isPrimary !== right.isPrimary) return left.isPrimary ? -1 : 1
        return left.sourceIndex - right.sourceIndex
      })

    if (
      !shouldIncludeProductThumbnail
      || !productThumbnail
      || mediaItems.some((item) => (
        item.url === productThumbnail
        || item.thumbnailUrl === productThumbnail
      ))
    ) {
      return mediaItems
    }

    return [
      {
        id: `product-thumbnail-${currentProduct.id}`,
        kind: 'image',
        url: productThumbnail,
        thumbnailUrl: productThumbnail,
        alt: currentProduct.name,
        isPrimary: true,
        sourceIndex: -1,
      },
      ...mediaItems,
    ]
  })

  const productMediaSlots = computed<Array<ProductGalleryItem | null>>(() => {
    const slotCount = Math.max(
      PRODUCT_DETAIL_MEDIA_SLOT_COUNT,
      productGalleryItems.value.length,
    )
    return Array.from(
      { length: slotCount },
      (_, index) => productGalleryItems.value[index] || null,
    )
  })

  const productMediaSlotsOverflowing = computed(() => (
    productGalleryItems.value.length > PRODUCT_DETAIL_MEDIA_SLOT_COUNT
  ))

  const primaryGalleryItem = computed(() => (
    productGalleryItems.value.find((item) => item.isPrimary)
    || productGalleryItems.value[0]
    || null
  ))

  const selectedMediaIndex = computed(() => {
    if (!productGalleryItems.value.length) return -1
    const index = productGalleryItems.value.findIndex((item) => item.id === selectedMediaId.value)
    return index >= 0 ? index : 0
  })

  const selectedMedia = computed(() => (
    productGalleryItems.value[selectedMediaIndex.value] || null
  ))

  const selectMedia = (mediaId: string) => {
    if (productGalleryItems.value.some((item) => item.id === mediaId)) {
      selectedMediaId.value = mediaId
    }
  }

  const selectPreviousMedia = () => {
    const items = productGalleryItems.value
    if (items.length < 2) return
    const currentIndex = selectedMediaIndex.value >= 0 ? selectedMediaIndex.value : 0
    const item = items[(currentIndex - 1 + items.length) % items.length]
    if (item) selectMedia(item.id)
  }

  const selectNextMedia = () => {
    const items = productGalleryItems.value
    if (items.length < 2) return
    const currentIndex = selectedMediaIndex.value >= 0 ? selectedMediaIndex.value : 0
    const item = items[(currentIndex + 1) % items.length]
    if (item) selectMedia(item.id)
  }

  watch(productGalleryItems, (items) => {
    if (!items.length) {
      selectedMediaId.value = null
      return
    }

    if (selectedMediaId.value && items.some((item) => item.id === selectedMediaId.value)) {
      return
    }

    const nextItem = items.find((item) => item.isPrimary) || items[0]
    if (nextItem) selectedMediaId.value = nextItem.id
  }, { immediate: true })

  const previewMedia = computed<ProductPreviewMedia | null>(() => {
    const media = selectedMedia.value || primaryGalleryItem.value
    if (media?.kind === 'image') {
      return {
        kind: 'image',
        url: media.url,
        alt: media.alt || product.value?.name || metaTitle.value,
      }
    }

    if (media?.kind === 'video') {
      return {
        kind: 'video',
        url: media.url,
        poster: media.poster,
      }
    }

    return null
  })

  const primaryMediaThumbnail = computed(() => {
    const media = primaryGalleryItem.value
    if (!media) return ''
    if (media.kind === 'video') return media.poster || media.thumbnailUrl || ''
    return media.thumbnailUrl || media.url
  })

  return {
    selectedMediaId,
    productGalleryItems,
    productMediaSlots,
    productMediaSlotsOverflowing,
    previewMedia,
    primaryMediaThumbnail,
    selectMedia,
    selectPreviousMedia,
    selectNextMedia,
  }
}
