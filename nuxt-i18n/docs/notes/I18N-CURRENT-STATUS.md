# Nuxt storefront i18n current status

Last updated: 2026-07-23

This document is the current source of truth for the Nuxt storefront language-file workflow. Older completion reports and archived notes are historical context only.

## Current reality

- The storefront declares 34 locales in `app/i18n/locales.manifest.js`.
- Nuxt reads aggregated locale files from `app/i18n/locales/*.json`.
- The maintainable source files are split by module under `app/i18n/messages/<locale>/*.json`.
- `npm run i18n:build` generates the aggregated `app/i18n/locales/*.json` files from the split `messages` files.
- The i18n scripts now resolve `app/i18n` first, falling back to root `i18n` only if needed.

## What is actually translated today

The project has 34 locale files, but only part of the storefront copy is fully internationalized. Some locale files still contain English fallback values.

Completed or partially structured areas include:

- common storefront chrome and navigation modules already present under `app/i18n/messages/*/`
- product/search/sidebar/member/warranty/support-related modules that existed before this pass
- first new guide module: `guidesTireguides`
  - `InnerTubeGuide.vue` now uses i18n keys instead of hard-coded display text
  - English is the base source
  - Simplified Chinese has reviewed translations
  - the remaining locales currently use English fallback values until each language is reviewed

## Do not blindly internationalize everything

Internationalization should be added by bounded product area, not by bulk replacing every English string.

Good candidates for immediate i18n:

- stable UI labels, buttons, table headings, empty-state text
- static guide content with clear page ownership
- policy/support page shell text
- accessibility labels and image alt text when they describe UI or stable content

Do not rush into i18n yet:

- product names, SKU option names, product dimensions, weights, prices, stock and technical specs
- admin-entered content that should eventually come from the backend as localized content
- FAQ answer bodies if the FAQ source of truth is still changing
- brand names, model names, standards, units, and search keywords used as behavior parameters
- content-heavy pages that should first be split into smaller components

When a file contains too much mixed content, split the component first, then add i18n keys to the stable child component.

## File ownership

Use one module file per bounded area:

```text
app/i18n/messages/<locale>/
├── common.json
├── products.json
├── support.json
├── guidesTireguides.json
└── ...
```

Naming rule:

- page or feature module: `guidesTireguides`, `productInformationTabs`, `shopCategoryMenu`
- broad site areas: `products`, `support`, `company`, `member`
- avoid dumping page-specific copy into `common.json`

After changing any `messages` file, run:

```powershell
cd nuxt-i18n
npm run i18n:build
```

Then verify:

```powershell
npm run build
```

Use `npm run i18n:check` as an audit tool. At the time of this note, it still reports pre-existing missing keys from `ProductInformationTabs.vue`; those should be handled as the next focused i18n block.

## Current known i18n debt

`npm run i18n:check -- --limit=12` currently exposes a focused debt area:

- `app/components/ProductInformationTabs.vue`
  - `sectionLabel`
  - `tabListLabel`
  - `tabs.details`
  - `tabs.afterSales`
  - `tabs.packaging`
  - `tabs.shipping`
  - `empty.details`
  - `empty.afterSales`
  - `empty.packaging`
  - `empty.shipping`

This should be the next block because it is small, user-facing, and already referenced by static `$t(...)` calls.

There are also extra keys in some non-English locales. Treat those as cleanup warnings, not as proof that the language coverage is complete.

## Recommended next steps

1. Add a `productInformationTabs.json` module for the keys listed above.
2. Generate aggregate locale files with `npm run i18n:build`.
3. Run `npm run i18n:check` again and confirm the missing static key count goes down.
4. Continue one bounded component/page at a time:
   - product detail tabs
   - shop category menu
   - stable support/policy shell text
   - individual guide child components only after they are visually and structurally stable
5. Keep a short note in this file after each i18n block:
   - component/page covered
   - source module name
   - languages reviewed vs languages using fallback
   - any known reason a string was intentionally left hard-coded

## Current completed block log

### 2026-07-23 — Tire Guides / Inner Tube

- Component: `app/components/tireguides/InnerTubeGuide.vue`
- Source module: `app/i18n/messages/<locale>/guidesTireguides.json`
- Aggregated into: `app/i18n/locales/*.json`
- Reviewed languages: `en`, `zh_cn`
- Fallback languages: all other configured locales currently use English fallback values
- Intentionally not translated:
  - behavior parameters such as product search keyword presets
  - technical abbreviations such as AV, DV, SV
  - numeric measurements and units such as `40mm`, `32/40mm`, `≤25mm`

