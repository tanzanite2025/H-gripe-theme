import { useI18n, useRuntimeConfig } from '#imports'
import { computed } from 'vue'
import { normalizeShopProduct, useShopProducts, type ShopProduct } from '~/composables/useShopProducts'
import { useStorefrontContext } from '~/composables/useStorefrontContext'
import {
  createStorefrontMediaContext,
  normalizeStorefrontProductMedia,
} from '~/utils/storefrontMedia'
import { extractProductDetailPayload } from '~/utils/productDetail'
import type { GoProduct } from '~/types/productDetail'

export interface ProductDetailSnapshot {
  rawProduct: GoProduct
  shopProduct: ShopProduct
}

const createProductDetailRequestHeaders = (
  locale: string,
  displayCurrency: string,
  countryCode: string,
) => {
  const headers: Record<string, string> = {}
  const normalizedLocale = String(locale || '').trim()
  if (normalizedLocale) headers['Accept-Language'] = normalizedLocale
  if (displayCurrency) headers['X-Display-Currency'] = displayCurrency
  if (countryCode && countryCode !== 'ZZ') headers['X-Market-Country'] = countryCode
  return Object.keys(headers).length ? headers : undefined
}

const createProductDetailRequestParams = (
  locale: string,
  displayCurrency: string,
  countryCode: string,
) => ({
  locale: locale || undefined,
  currency: displayCurrency || undefined,
  country: countryCode !== 'ZZ' ? countryCode : undefined,
})

export function useProductDetailLookup() {
  const config = useRuntimeConfig()
  const { locale } = useI18n()
  const { displayCurrency, countryCode, baseCurrency } = useStorefrontContext()
  const { baseURL } = useShopProducts()
  const mediaContext = createStorefrontMediaContext(config)
  const currentLocale = computed(() => String(locale.value || '').trim())

  const fetchProductDetailSnapshot = async (slug: string): Promise<ProductDetailSnapshot | null> => {
    const normalizedSlug = String(slug || '').trim()
    if (!normalizedSlug) return null

    const requestedLocale = currentLocale.value
    const requestedCurrency = String(displayCurrency.value || '').trim()
    const requestedCountry = String(countryCode.value || '').trim()

    const response = await $fetch<any>(`${baseURL}/products/${encodeURIComponent(normalizedSlug)}`, {
      headers: createProductDetailRequestHeaders(
        requestedLocale,
        requestedCurrency,
        requestedCountry,
      ),
      params: createProductDetailRequestParams(
        requestedLocale,
        requestedCurrency,
        requestedCountry,
      ),
    })

    const data = extractProductDetailPayload(response, normalizedSlug)
    if (!data) return null

    const rawProduct = normalizeStorefrontProductMedia(data as GoProduct, mediaContext)
    return {
      rawProduct,
      shopProduct: normalizeShopProduct(rawProduct, baseCurrency.value, mediaContext),
    }
  }

  return {
    fetchProductDetailSnapshot,
  }
}
