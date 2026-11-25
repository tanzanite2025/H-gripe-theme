import type { Component } from 'vue'

type HeaderPrimaryNavigationShopComponents = {
  productCategoryNavigationCards: Component
  homeMainProductCategories: Component
}

let headerPrimaryNavigationShopComponentsRequest: Promise<HeaderPrimaryNavigationShopComponents> | null = null

// Download component definitions without mounting them, requesting their data,
// or prefetching any destination page. Failed downloads remain retryable.
export function preloadHeaderPrimaryNavigationShopComponents(): Promise<HeaderPrimaryNavigationShopComponents> {
  headerPrimaryNavigationShopComponentsRequest ??= Promise.all([
    import('../components/shop/ProductCategoryNavigationCards.vue'),
    import('../components/home/HomeMainProductCategories.vue'),
  ]).then(([productCategoriesModule, mainProductCategoriesModule]) => ({
    productCategoryNavigationCards: productCategoriesModule.default,
    homeMainProductCategories: mainProductCategoriesModule.default,
  })).catch(error => {
    headerPrimaryNavigationShopComponentsRequest = null
    throw error
  })

  return headerPrimaryNavigationShopComponentsRequest
}
