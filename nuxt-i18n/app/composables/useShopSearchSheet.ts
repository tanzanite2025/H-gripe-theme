import { useState } from '#imports'
import { useOverlayBackStack } from '~/composables/useOverlayBackStack'
import { useShopSearchNavigation } from '~/composables/useShopSearchNavigation'
import { activateStorefrontClientOverlays } from '~/utils/clientOverlays'

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
  const presetCategorySlug = useState<string | null>('shopSearchSheetPresetCategory', () => null)
  const presetKeywords = useState<string[]>('shopSearchSheetPresetKeywords', () => [])

  const overlayBackStack = useOverlayBackStack()
  const { pendingSearch, submit: navigateToSearch } = useShopSearchNavigation()

  const closeState = () => {
    isOpen.value = false
  }

  const close = async () => {
    await overlayBackStack.close('shop-search')
    closeState()
  }

  const open = (options?: ShopSearchOpenOptions) => {
    activateStorefrontClientOverlays()
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
    overlayBackStack.open('shop-search', closeState)
  }

  const submit = async (payload: ShopSearchPayload) => {
    await close()
    await navigateToSearch(payload)
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
