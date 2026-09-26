import type { ShopSearchPayload } from '~/composables/useShopSearchSheet'
import { useShopSearchNavigation } from '~/composables/useShopSearchNavigation'
import { useState } from '#imports'

export const TIRE_GUIDE_TIRE_PRODUCT_CATEGORY_SLUG = 'tire'
export const TIRE_GUIDE_TIRE_SEARCH_PENDING_KEY = 'tireGuideTireSearchPending'

export const useTireGuideTireProductSearch = () => {
  const { submit: submitSearch } = useShopSearchNavigation()
  const pendingSearch = useState<ShopSearchPayload | null>(
    TIRE_GUIDE_TIRE_SEARCH_PENDING_KEY,
    () => null,
  )

  const submit = async (payload: ShopSearchPayload) => {
    await submitSearch(payload, {
      productCategorySlug: TIRE_GUIDE_TIRE_PRODUCT_CATEGORY_SLUG,
      pendingState: pendingSearch,
    })
  }

  return {
    pendingSearch,
    submit,
  }
}
