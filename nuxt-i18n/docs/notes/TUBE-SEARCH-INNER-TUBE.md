# Inner tube search and guide integration

Last updated: 2026-07-25

This note is the current source of truth for the Inner Tube entry on `/guides/tireguides` and its relationship with the storefront shop/search flow.

The old WordPress / site settings / WP REST `tube_specs` plan is no longer the active architecture. our storefront search now follows the Nuxt + Go backend product-type path.

## Current goal

The Inner Tube tab is an education + entry page, not a separate product search system.

- `/guides/tireguides` explains tube size, valve, and model choices.
- The CTA opens the existing shop search sheet.
- Search and product listing remain owned by `/shop`, `ProductSearchPanel`, and the Go product APIs.
- Inner tube products are selected by stable product type slug, currently `inner-tube`.

## Current flow

1. `app/components/tireguides/InnerTubeGuide.vue` renders the Inner Tube content and CTA.
2. The CTA calls `useShopSearchSheet().open({ presetCategorySlug: 'inner-tube', presetKeywords: ['Inner tube'] })`.
3. `app/components/ProductSearchPanel.vue` reads `presetKeywords` and preselects the `Inner tube` popular-search chip.
4. When the selected chip is submitted, `ProductSearchPanel.vue` emits `chipCategorySlug: 'inner-tube'`.
5. `app/pages/shop/index.vue` consumes `presetCategorySlug` / `chipCategorySlug`, matches it against `useShopCategories()` results, and sets the single `selectedCategory`.
6. `/shop` builds the product query with `product_type=<slug>` when the selected category is a real product type.
7. Go backend filters products in `internal/repository/product_query_repository.go` by joining `product_types.slug`.

If `inner-tube` is not present in `GET /api/v1/products/types`, the flow safely degrades to keyword search instead of inventing a fake category.

## Current files

### Nuxt storefront

- `app/components/tireguides/InnerTubeGuide.vue`
  - owns the Inner Tube educational content and CTA only.
- `app/composables/useShopSearchSheet.ts`
  - owns the search-sheet open state, pending search payload, `presetCategorySlug`, and `presetKeywords`.
- `app/components/ProductSearchPanel.vue`
  - owns the search sheet controls and maps the `Inner tube` chip to `chipCategorySlug: 'inner-tube'`.
- `app/pages/shop/index.vue`
  - owns `selectedCategory`, product query building, and the final `product_type=<slug>` request parameter.
- `app/composables/useShopCategories.ts`
  - reads `GET /api/v1/products/types`; DEV fallback is development-only and must not become product-type truth in production.
- `app/utils/popularSearchKeywords.ts`
  - static UI chip labels only; it does not own routing or product facts.

### Go backend

- `internal/api/v1/product/handler.go`
  - public product type and product listing handlers.
- `internal/repository/product_attribute_repository.go`
  - returns enabled product types ordered by `sort_order ASC, id ASC`.
- `internal/repository/product_query_repository.go`
  - filters public products by `product_types.slug` when `product_type` is provided.
- `internal/domain/product/model.go`
  - `Product`, `ProductType`, spec definitions, and variant structures.

## What is closed

- Inner Tube CTA opens the existing shop search sheet instead of creating a second search modal.
- The CTA preselects the `Inner tube` chip and passes the stable `inner-tube` slug through the shared search state.
- `/shop` consumes the slug through the same category mechanism as manual category selection.
- The product request uses the single backend query parameter `product_type=<slug>`.
- No old WordPress `tube_specs` table, WP REST endpoint, or site settings plugin page is part of the current path.

## Still not fully closed

1. Product type data verification
   - DEV / staging must contain an enabled product type with slug `inner-tube` before the CTA can narrow results by category.
   - Verification belongs with the `/shop` category data check documented in `SHOP_CATEGORY_LAYOUT.zh-CN.md`.

2. Precise tube filters
   - Current behavior is category + keyword narrowing.
   - If future tube products need valve family, valve length, execution, ETRTO range, or size filters, define them through the Go product type/spec template model first.
   - Do not add a separate Nuxt-only tube config, old WP `tube_specs`, or per-component special cases.

3. Chip-to-filter mapping
   - The current hardcoded `Inner tube` chip-to-slug mapping is intentionally small.
   - If more chips become structured filters, move the mapping into one shared config object and let `ProductSearchPanel.vue` read it, instead of adding more `if keyword === ...` branches.

## Maintenance rule

- Keep `inner-tube` as a stable slug unless backend product type data is deliberately changed.
- If the slug changes, update product type data, `InnerTubeGuide.vue`, `ProductSearchPanel.vue`, this document, and the `/shop` category verification together.
- Do not reintroduce the archived WordPress tube-spec plan as active architecture.
