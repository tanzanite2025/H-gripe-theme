# Storefront locale registry

Last audited: 2026-08-09

This document defines the language boundary for storefront-facing localized content.

## Source of truth

The cross-stack registry is:

- `shared/storefront-locales.json`

It currently contains exactly 20 fixed storefront locales. The registry owns:

- `code`: canonical application locale, such as `en` or `zh_cn`
- `iso`: Nuxt/i18n ISO locale, such as `en-US`
- `name`: English backend/Admin name
- `native_name`: storefront/Admin display name
- `file`: Nuxt aggregate locale file
- `dir`: optional text direction
- `enabled`: whether the locale is usable by storefront-facing content

Do not add page-level, feature-level, or Admin-only language arrays. A new storefront language must be added to the registry first.

## Required alignment

These consumers must stay aligned with the registry:

- `nuxt-i18n/app/i18n/locales.manifest.ts`
- `go-backend/internal/pkg/locales/locales.go`
- `go-backend/config/config.example.yaml`
- `go-backend/config/config.production.yaml`
- `go-backend/web/admin/src/lib/languages.ts`
- `GET /api/v1/i18n/languages`

Run this check after changing locale definitions:

```powershell
cd nuxt-i18n
npm run check-locales
```

The check compares the shared registry, Nuxt manifest, Go backend locales, Go config locale lists, and Admin fallback locales.
It also scans active Nuxt `.ts` and `.vue` source for legacy unsupported storefront locale literals and for page/component-level hardcoded locale lists.

Admin also has a UI guard for storefront-facing locale fields:

```powershell
cd go-backend/web/admin
npm run check:storefront-locale-ui
```

This guard is part of `npm run typecheck` and blocks free-text or native `<select v-model="...locale">` controls from coming back in editor surfaces.

## Concept boundaries

Storefront locales are not the same as every language-like value in the system.

- **Nuxt static copy** lives in Nuxt i18n JSON files and is built with the storefront.
- **Nuxt locale utilities** must derive labels, flags, aliases, and canonical codes from `nuxt-i18n/app/i18n/locales.manifest.ts` via `nuxt-i18n/app/utils/storefrontLocales.ts`.
- **Admin editable localized content** lives in the backend database and must use one of the 20 storefront locale codes.
- **Runtime visitor language signals** such as `Accept-Language` may be raw request metadata until resolved.
- **Channel-specific fields** such as Google Merchant `content_language` follow the channel contract and should not be confused with storefront content locale.
- **Market configuration** can reference supported storefront locales, but market/currency/country rules are separate facts.

## Admin editable content

These content types must use controlled storefront locale selection in Admin and backend validation through `requireSupportedLocale` or an equivalent guard:

- FAQ pages, categories, and FAQ items
- Customer service auto replies and FAQ references
- Content/posts shown on the storefront
- Product type translations
- Product information templates, including After-sales and Packaging
- Any future storefront CMS/editor content

In Admin forms, prefer `go-backend/web/admin/src/components/admin/StorefrontLocaleSelect.vue` for single-locale storefront content fields. Reuse `useSupportedLanguages()` for labels, filters, and read-only displays instead of creating page-local arrays.

For records whose identity includes locale, editing an existing record must not silently change locale. Create a new localized record instead.

## Nuxt frontend rules

Nuxt has two different localization surfaces:

- Static UI copy is compiled into `nuxt-i18n/app/i18n/locales/*.json` from `nuxt-i18n/app/i18n/messages/<locale>/*.json`.
- Backend-owned content such as FAQ data, product type labels, post translations, and product information templates must request and render backend content with canonical storefront locale codes.

Do not create component-local language maps such as flag maps, language-name maps, or compatibility alias arrays. Use `~/utils/storefrontLocales` for:

- `normalizeStorefrontLocaleCode()` when accepting route, backend, or i18n runtime locale values
- `getStorefrontLocaleName()` and `getStorefrontLocaleFlag()` for display
- `isSimplifiedChineseStorefrontLocale()` for legacy backend fields that only split English/Chinese content

Aliases such as `en-US` or `zh` are input compatibility only. They must normalize to canonical codes such as `en` and `zh_cn`; they are not selectable storefront locales.

## Anti-patterns

Avoid these patterns:

- Free-text locale inputs for storefront-facing content
- `*` meaning "all languages" at runtime
- Copying one language's content into all 20 locales and treating it as translated
- Letting Nuxt, Admin, and Go maintain separate language lists without a registry check
- Falling back from a missing localized FAQ or auto reply to another language
- Hardcoded Nuxt component maps containing old locales such as `zh-TW`, `pl`, `vi`, `bn`, or `ro`

## Current fixed locale codes

```text
en, zh_cn, fr, de, es, ja, ko, it, pt, ru,
ar, nl, tr, id, th, sv, da, fi, hi, ms
```
