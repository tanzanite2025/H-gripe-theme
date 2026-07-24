# Site search popular keywords current source

Last updated: 2026-07-24

This note describes the current “Popular searches” chips used by the Nuxt storefront search flow.

## Current goal

Popular chips are only UI state helpers. They must not own routing or product fetching.

The one rule is:

> chips update selected keywords; existing search submit logic builds the query and performs the search.

## Current files

- `nuxt-i18n/app/utils/popularSearchKeywords.ts`
  - current keyword source:
    - `Carbon rim`
    - `Sapim`
    - `Spoke`
    - `Carbon Wheels`
    - `Inner tube`
- `nuxt-i18n/app/components/PopularSearchChips.vue`
  - renders chips
  - supports `v-model` via `modelValue` / `update:modelValue`
  - does not call router or fetch APIs
- `nuxt-i18n/app/components/ProductSearchPanel.vue`
  - uses popular chips in the header/search sheet flow
- `nuxt-i18n/app/pages/shop/index.vue`
  - uses the same keyword source and chip component for `/shop`

## Current behavior

- Clicking a chip toggles it in the selected keyword list.
- Selected chips are merged with free-text input into one final search string.
- `/shop` and the search sheet keep their own local selected keyword state, but share the keyword list and chip component.
- `Inner tube` also participates in the broader preset-category flow through `presetCategorySlug` where applicable.

## Still not fully closed

1. i18n title
   - `PopularSearchChips.vue` still hardcodes `aria-label="Popular searches"` and the visible title `Popular searches`.
   - This should move to i18n keys such as `search.popularTitle` / `search.popularAriaLabel`.

2. Keyword ownership
   - Current keywords are static frontend config.
   - If marketing wants to change them from admin later, define a backend endpoint and keep this utility as fallback only.

3. Chip-to-filter mapping
   - Keywords currently behave like search text.
   - If chips become structured filters, keep the mapping in a single config object and do not add per-component special cases.

## Maintenance rule

Do not let `PopularSearchChips.vue` directly navigate, submit forms, or fetch products. It remains a presentational toggle component.
