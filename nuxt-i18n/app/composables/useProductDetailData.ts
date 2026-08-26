import { computed } from 'vue'
import {
  createError,
  useAsyncData,
  useI18n,
  useRoute,
  useRuntimeConfig,
} from '#imports'
import {
  normalizeShopProduct,
  useShopProducts,
} from '~/composables/useShopProducts'
import { useStorefrontContext } from '~/composables/useStorefrontContext'
import {
  createStorefrontMediaContext,
  normalizeStorefrontProductMedia,
} from '~/utils/storefrontMedia'
import {
  resolveProductMetaDescription,
  resolveProductMetaTitle,
} from '~/utils/seo/product'
import {
  extractProductDetailPayload,
  stripProductHtml,
} from '~/utils/productDetail'
import type {
  GoProduct,
  ProductMedia,
  ProductMediaImage,
} from '~/types/productDetail'

export async function useProductDetailData() {
  const route = useRoute()
  const config = useRuntimeConfig()
  const { locale } = useI18n()
  const { displayCurrency, countryCode, baseCurrency } = useStorefrontContext()
  const mediaContext = createStorefrontMediaContext(config)
  const { baseURL } = useShopProducts()

  const slug = computed(() => String(route.params.slug || ''))
  // Keep the SSR identity tied to the route. Market context resolves asynchronously
  // and belongs in the refresh watchers, not in a key that can change mid-render.
  const { data: product, pending } = await useAsyncData<GoProduct>(
    () => [
      'go-product',
      locale.value || 'default',
      slug.value,
    ].map((part) => encodeURIComponent(String(part))).join(':'),
    async () => {
      if (!slug.value) {
        throw createError({
          statusCode: 404,
          statusMessage: 'Product not found',
        })
      }

      const response = await $fetch<any>(
        `${baseURL}/products/${encodeURIComponent(slug.value)}`,
        {
          headers: {
            accept: 'application/json',
            ...(locale.value ? { 'Accept-Language': String(locale.value) } : {}),
            ...(displayCurrency.value ? { 'X-Display-Currency': displayCurrency.value } : {}),
            ...(countryCode.value && countryCode.value !== 'ZZ'
              ? { 'X-Market-Country': countryCode.value }
              : {}),
          },
          params: {
            locale: locale.value || undefined,
            currency: displayCurrency.value || undefined,
            country: countryCode.value !== 'ZZ' ? countryCode.value : undefined,
          },
        },
      )
      const data = extractProductDetailPayload(response, slug.value)
      if (!data) {
        throw createError({
          statusCode: 404,
          statusMessage: 'Product not found',
        })
      }
      return normalizeStorefrontProductMedia(data as GoProduct, mediaContext)
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

  const shopProduct = computed(() => {
    return product.value
      ? normalizeShopProduct(product.value, baseCurrency.value, mediaContext)
      : null
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
