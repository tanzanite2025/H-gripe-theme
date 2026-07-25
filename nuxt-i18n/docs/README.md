# Nuxt storefront documentation hub

Last updated: 2026-07-25

This folder is the live documentation hub for the Nuxt storefront. Active notes stay in `notes/`. Superseded implementation plans and old checklists stay in `archive/notes/` and should not be used as the current source of truth.

## Current active sources

- i18n status and workflow: `notes/I18N-CURRENT-STATUS.md`
- FAQ ownership and current backend/frontend boundary: `../../go-backend/docs/FAQ_MANAGEMENT_SOURCE.md`
- Chat and customer-service flow: `notes/CHAT-SYSTEM-ANALYSIS.md`
- Chat product configuration confirmation flow: `notes/CHAT_CONFIG_CONFIRM_DESIGN.md`
- Shop category layout: `notes/SHOP_CATEGORY_LAYOUT.zh-CN.md`
- Search chips and product-search behavior: `notes/SEARCH-POPULAR-KEYWORDS.md`
- Inner-tube search and guide integration: `notes/TUBE-SEARCH-INNER-TUBE.md`
- Breadcrumb and page hash-tab behavior: `notes/BREADCRUMB-PAGE-SUBNAV.md`
- Product warranty system notes: `notes/PRODUCT-WARRANTY-SYSTEM.md`
- Spoke calculator data notes: `notes/SPOKE-CALCULATOR-SYSTEM.md`

## Older active notes that need re-audit before implementation

- `notes/CHAT_CONFIG_CONFIRM_DESIGN.md` — Phase 1 shell is complete, but Phase 2/3 must be revalidated after the product/SKU data structure is final.
- `notes/SPOKE-CALCULATOR-SYSTEM.md` — system manual for the static spoke database + Go export sync. It is not an open backlog by itself, but should be refreshed after the next spoke-data workflow change.

## Still not fully closed

- FAQ: keep `FAQ_MANAGEMENT_SOURCE.md` synchronized whenever FAQ admin, route-based insertion, `PageFaqSlot`, static fallback, rich answer rendering, or FAQ loading styling changes. Route-based insertion is now single-source: storefront pages no longer retain duplicated page-level `<PageFaq />`; current products/support layout route coverage is seeded by migration `031`, product detail pages use stable `/shop/:slug` and legacy `/products/:slug` FAQ lookups, FAQ aggregation pages skip auto-inserting a second FAQ block, and `PageFaqSlot` now uses a unified FAQ skeleton loading state instead of bare `LOAD` text.
- Chat: `config_confirm` and `order` messages now persist through the customer-service ticket source and render as cards in both Nuxt customer chat and Admin staff chat. Remaining work is real non-SKU configurable fields from the final product/SKU contract and optional history-order-to-configuration refill.
- Chat realtime: current transport is confirmed in `notes/CHAT-SYSTEM-ANALYSIS.md`. HTTP remains the durable source of truth; SSE is wired for admin inbox and Nuxt customer chat refresh/invalidation. Scoped typing indicators now use HTTP publish endpoints plus SSE delivery. The Go WebSocket route remains only a guarded keepalive foundation and must not be treated as the active chat protocol.
- i18n: static key coverage is clean, but long-tail locale wording still needs native-language review before marketing-sensitive production use.
- Warranty: current Nuxt + Go backend boundary is documented in `notes/PRODUCT-WARRANTY-SYSTEM.md`. Admin backend ownership, public upload validation, admin UI consumption, and nullable order-based claim linkage are closed; open items are claim-to-registration linkage for order-based claims, real service records, richer claim processing UX, and e2e tests.
- Shop categories: current implementation is documented in `notes/SHOP_CATEGORY_LAYOUT.zh-CN.md`. DEV fallback policy and empty-state copy are closed; remaining work is verification against real product-type data and future backend-owned i18n handling for dynamic category names.
- Search/tube notes: popular-search chips and inner-tube preset slug mapping are documented. Fixed `Popular searches` UI text now uses i18n keys; remaining work is optional admin-owned keyword config and optional precision tube filtering (`tube_execution` / `tube_valve_*`) after backend data is ready.
- Breadcrumb/hash tabs: current behavior is documented in `notes/BREADCRUMB-PAGE-SUBNAV.md`. Tire Guides file responsibility now matches `/guides/tireguides`.

## Archived notes

The following old implementation plans were moved out of `notes/` on 2026-07-24 because they were either completed, superseded by current docs, or too stale to be used as active backlog:

- `archive/notes/task.md`
- `archive/notes/implementation_plan.md`
- `archive/notes/FAQ-REFACTOR-PLAN.md`
- `archive/notes/HOME-FAQ-INTEGRATION.md`
- `archive/notes/WHATSAPP-PRODUCT-SEARCH-DRAWER.md`

When a note is completed or replaced by a more current source, move it to `archive/notes/` and update this file.
