# Nuxt Frontend Example

This directory contains example Nuxt integration files for backend blog i18n features.

The main storefront now lives in `../../../nuxt-i18n/`. Treat these files as examples only; do not use them as the source of truth for the current storefront implementation.

## Files

- `useI18n.ts` - example composable for backend i18n calls
- `LanguageSwitcher.vue` - example language switcher
- `PostTranslations.vue` - example post translation links
- `blog-post-page.vue` - example blog page
- `nuxt.config.example.ts` - example Nuxt config

## Backend API

Set the backend URL in the consuming Nuxt app:

```env
NUXT_PUBLIC_API_BASE=http://localhost:9000
```

Relevant backend docs:

- `../../docs/I18N_QUICK_REFERENCE.md`
- `../../API.md`

## Maintenance Rules

- Keep examples small and copyable.
- Do not duplicate current storefront business logic here.
- If the real storefront changes, update `../../../nuxt-i18n/` first and only refresh these examples when still useful.
