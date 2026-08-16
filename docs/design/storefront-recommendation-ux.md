# Storefront Recommendation UX And Algorithm Contract

## Source Of Truth

- Recommendation decisions come from the backend API: `POST /api/v1/recommendations`.
- Storefront product recommendation UI must use `nuxt-i18n/app/components/shop/ProductRecommendations.vue`.
- The shared loader is `nuxt-i18n/app/composables/useSmartRecommendations.ts`.
- Page files should only pass placement context such as `surface`, `productId`, `categoryId`, `query`, `limit`, and `excludeProductIds`.
- The backend algorithm version is declared in `go-backend/internal/service/recommendation_service.go`.

## Current Placements

- Product detail pages use `surface: product_detail_bottom` after product specifications and information tabs. They pass the current product ID and exclude that same product from the result set.
- The main shop page uses `surface: shop_index_bottom` after the catalog section and before the feedback thread. It passes the active category, active query context, and currently visible product IDs as exclusions.
- `SmartRecommendationPanel.vue` is currently category navigation for the search drawer. It should not be treated as the reusable product recommendation component.

## Algorithm Contract

- Current backend version: `contextual-v1`.
- Candidate products must be active, must have at least one active variant, and must have at least one active variant with stock above zero.
- `product_detail_bottom` prioritizes products from the same product specification template and then boosts matching filterable or variant-option specifications.
- `shop_index_bottom` prioritizes the active category and active search query, then fills with available trending products.
- The backend returns only real, active, purchasable products. If fewer than five recommendations are returned but the public catalog has more products, the shared loader may fill the remaining slots with other real catalog products. An empty catalog remains empty.
- Recommendation responses expose `slot` and `reason` for analytics only. Do not show those fields as storefront copy unless a dedicated UX decision is made.

## Behavior Signals

- Recommendation cards report `recommendation_impression` and `recommendation_click` through the shared behavior event pipeline.
- Ranking uses recent product behavior from `recommendation_events`, including product views, dwell, recommendation clicks, add to cart, wishlist add, checkout start, and trusted purchase events when present.
- Anonymous and session IDs may personalize ranking for the same visitor. This personalization must stay first-party and must not depend on customer-service transcripts, IP addresses, user agents, or payment data.
- Client-side storefront code must not emit `purchase`; purchase is a trusted server-side commerce fact.

## Rules

- Do not hardcode recommendation products directly inside page templates.
- Do not create another recommendation card layout unless the shared component cannot support the placement.
- Every new placement needs a stable `surface` value so impressions and clicks can be attributed.
- Recommendation impressions and clicks should continue to use the existing behavior event pipeline.
- Recommendation sections should remain visible even when no items are returned. The component should show its empty state instead of disappearing.
- Empty responses with an empty catalog should show the component's empty state. Catalog-fill cards must be real product records only; category navigation entries and search keywords must never become product cards.
- New recommendation placements must reuse the backend API and the shared component unless there is a written reason to add a new surface-specific component.
