import { useLocalePath, useRoute, useRouter, useState } from '#imports'
import type { Ref } from 'vue'
import type { ShopSearchPayload } from '~/composables/useShopSearchSheet'

export interface ShopSearchNavigationOptions {
  productCategorySlug?: string | null
  pendingState?: Ref<ShopSearchPayload | null>
}

export const useShopSearchNavigation = () => {
  const pendingSearch = useState<ShopSearchPayload | null>('shopSearchSheetPending', () => null)
  const localePath = useLocalePath()
  const router = useRouter()
  const route = useRoute()

  const submit = async (
    payload: ShopSearchPayload,
    options: ShopSearchNavigationOptions = {},
  ) => {
    if (options.pendingState) {
      options.pendingState.value = payload
    } else {
      pendingSearch.value = payload
    }

    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('ui:shop-search-submit', { detail: payload }))
    }

    const shopPath = localePath('/shop')
    const chipCategorySlug = String(payload?.chipCategorySlug || '').trim()
    const productCategorySlug = String(options.productCategorySlug || '').trim()
    const query: Record<string, string> = {}

    if (chipCategorySlug) {
      query.product_specification_template = chipCategorySlug
    }
    if (productCategorySlug) {
      query.product_category = productCategorySlug
    }

    const currentProductCategory = String(route.query.product_category || '').trim()
    const currentTemplateCategory = String(route.query.product_specification_template || '').trim()
    const hasSameQuery = (
      currentProductCategory === productCategorySlug
      && currentTemplateCategory === chipCategorySlug
    )

    if (route.path !== shopPath || !hasSameQuery) {
      await router.push(Object.keys(query).length ? { path: shopPath, query } : shopPath)
    }
  }

  return {
    pendingSearch,
    submit,
  }
}
