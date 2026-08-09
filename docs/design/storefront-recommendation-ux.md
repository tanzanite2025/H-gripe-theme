# Storefront Recommendation UX

## Source Of Truth

- Recommendation decisions come from the backend API: `POST /api/v1/recommendations`.
- Storefront product recommendation UI must use `nuxt-i18n/app/components/shop/ProductRecommendations.vue`.
- The shared loader is `nuxt-i18n/app/composables/useSmartRecommendations.ts`.
- Page files should only pass placement context such as `surface`, `productId`, `categoryId`, `query`, `limit`, and `excludeProductIds`.

## Current Placements

- Product detail pages use `surface: product_detail_bottom` after product specifications and information tabs. They pass the current product ID and exclude that same product from the result set.
- The main shop page uses `surface: shop_index_bottom` after the catalog section and before the feedback thread. It passes the active category, active query context, and currently visible product IDs as exclusions.
- `SmartRecommendationPanel.vue` is currently category navigation for the search drawer. It should not be treated as the reusable product recommendation component.

## Rules

- Do not hardcode recommendation products directly inside page templates.
- Do not create another recommendation card layout unless the shared component cannot support the placement.
- Every new placement needs a stable `surface` value so impressions and clicks can be attributed.
- Recommendation impressions and clicks should continue to use the existing behavior event pipeline.
- Empty production responses should not show placeholder marketing content. The component may use catalog fallback data only through the shared loader.
