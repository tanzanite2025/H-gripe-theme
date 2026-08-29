import { computed } from 'vue'
import {
  createError,
  useAsyncData,
  useI18n,
  useRoute,
} from '#imports'
import { useProductDetailLookup, type ProductDetailSnapshot } from '~/composables/useProductDetailLookup'
import { useStorefrontContext } from '~/composables/useStorefrontContext'
import {
  resolveProductMetaDescription,
  resolveProductMetaTitle,
} from '~/utils/seo/product'
import {
  stripProductHtml,
} from '~/utils/productDetail'
import type {
  ProductMedia,
  ProductMediaImage,
} from '~/types/productDetail'

export async function useProductDetailData() {
  const route = useRoute()
  const { locale } = useI18n()
  const { displayCurrency, countryCode } = useStorefrontContext()
  const { fetchProductDetailSnapshot } = useProductDetailLookup()

  const slug = computed(() => String(route.params.slug || ''))
  // Keep the SSR identity tied to the route. Market context resolves asynchronously
  // and belongs in the refresh watchers, not in a key that can change mid-render.
  const { data: detailSnapshot, pending } = await useAsyncData<ProductDetailSnapshot>(
    () => [
      'go-product',
      locale.value || 'default',
      slug.value,
    ].map((part) => encodeURIComponent(String(part))).join(':'),
    async () => {
      const snapshot = await fetchProductDetailSnapshot(slug.value)
      if (!snapshot?.rawProduct) {
        throw createError({
          statusCode: 404,
          statusMessage: 'Product not found',
        })
      }
      return snapshot
    },
    {
      server: true,
      watch: [
        () => slug.value,
        () => locale.value,
        () => displayCurrency.value,
        () => countryCode.value,
      ],
    },
  )

  const product = computed(() => detailSnapshot.value?.rawProduct || null)
  const shopProduct = computed(() => detailSnapshot.value?.shopProduct || null)

  if (!pending.value && !product.value) {
    throw createError({
      statusCode: 404,
      statusMessage: 'Product not found',
    })
  }

  const metaTitle = computed(() => resolveProductMetaTitle(
    product.value?.meta_title,
    product.value?.name,
  ))

  const metaDescription = computed(() => resolveProductMetaDescription({
    metaDescription: product.value?.meta_description,
    shortDescription: product.value?.short_description,
    description: product.value?.description,
  }))

  const productSummaryDescription = computed(() => {
    const text = stripProductHtml(product.value?.description || '')
    if (text.length <= 220) return text
    return `${text.slice(0, 217)}...`
  })

  const productMediaImages = computed<ProductMedia[]>(() => {
    return (product.value?.media || []).filter((item) => (
      item.media_type === 'image'
      && item.url
      && item.is_visible !== false
    ))
  })

  const productImages = computed<ProductMediaImage[]>(() => {
    return productMediaImages.value.map((item) => ({
      id: item.id,
      url: item.url,
      alt: item.alt || item.title,
    }))
  })

  return {
    slug,
    product,
    pending,
    shopProduct,
    metaTitle,
    metaDescription,
    productSummaryDescription,
    productMediaImages,
    productImages,
    displayCurrency,
    countryCode,
    locale,
  }
}
