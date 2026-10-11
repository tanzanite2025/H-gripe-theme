import type { ShopSearchPayload } from '~/composables/useShopSearchSheet'
import { useShopSearchNavigation } from '~/composables/useShopSearchNavigation'

export const TIRE_GUIDE_TIRE_PRODUCT_CATEGORY_SLUG = 'tire'

export const useTireGuideTireProductSearch = () => {
  const { pendingSearch, submit: submitSearch } = useShopSearchNavigation()

  const submit = async (payload: ShopSearchPayload) => {
    await submitSearch(payload, {
      productCategorySlug: TIRE_GUIDE_TIRE_PRODUCT_CATEGORY_SLUG,
    })
  }

  return {
    pendingSearch,
    submit,
  }
}
