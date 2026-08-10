import { useLocalePath, useRoute, useRouter, useState } from '#imports'

export type ShopSearchFiltersPayload = Record<string, any> & {
  priceRange?: [number, number]
  currency?: string
  attributes?: Record<string, string[]>
}

export interface ShopSearchPayload {
  query: string
  filters: ShopSearchFiltersPayload
  chipCategorySlug?: string
}

export interface ShopSearchOpenOptions {
  presetCategorySlug?: string | null
  presetKeywords?: string[]
}

export const useShopSearchSheet = () => {
  const isOpen = useState<boolean>('shopSearchSheetOpen', () => false)
  const pendingSearch = useState<ShopSearchPayload | null>('shopSearchSheetPending', () => null)
  const presetCategorySlug = useState<string | null>('shopSearchSheetPresetCategory', () => null)
  const presetKeywords = useState<string[]>('shopSearchSheetPresetKeywords', () => [])

  const localePath = useLocalePath()
  const router = useRouter()
  const route = useRoute()

  const close = () => {
    isOpen.value = false
  }

  const open = (options?: ShopSearchOpenOptions) => {
    if (options && typeof options.presetCategorySlug !== 'undefined') {
      presetCategorySlug.value = options.presetCategorySlug || null
    } else {
      presetCategorySlug.value = null
    }

    if (options && Array.isArray(options.presetKeywords)) {
      presetKeywords.value = [...options.presetKeywords]
    } else {
      presetKeywords.value = []
    }

    isOpen.value = true
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'shop-search' } }))
    }
  }

  const submit = async (payload: ShopSearchPayload) => {
    pendingSearch.value = payload
    close()

    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('ui:shop-search-submit', { detail: payload }))
    }
    const shopPath = localePath('/shop')
    const chipCategorySlug = String(payload?.chipCategorySlug || '').trim()
    const query = chipCategorySlug ? { product_type: chipCategorySlug } : undefined

    if (route.path !== shopPath || String(route.query.product_type || '') !== chipCategorySlug) {
      await router.push(query ? { path: shopPath, query } : shopPath)
    }
  }

  const consumePending = () => {
    const payload = pendingSearch.value
    pendingSearch.value = null
    return payload
  }

  return {
    isOpen,
    pendingSearch,
    presetCategorySlug,
    presetKeywords,
    open,
    close,
    submit,
    consumePending,
  }
}
