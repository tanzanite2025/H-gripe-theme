# System Fonts

Last updated: 2026-07-25

## Current source

The storefront and admin share the same system font family name:

```text
CommercePlatformSystem
```

The current CN + Latin font file is:

```text
CommercePlatformSystem-CN-Latin.woff2
```

Because the storefront and admin are built from different Docker contexts, the font file must exist in both source trees:

```text
nuxt-i18n/app/assets/fonts/system/CommercePlatformSystem-CN-Latin.woff2
go-backend/web/admin/src/assets/fonts/system/CommercePlatformSystem-CN-Latin.woff2
```

Do not put source fonts in generated folders:

```text
dist/
.nuxt/
.output/
node_modules/
go-backend/web/admin/dist/
```

## CSS entry points

Storefront:

```text
nuxt-i18n/app/assets/css/tailwind.css
nuxt-i18n/tailwind.config.js
```

Admin:

```text
go-backend/web/admin/src/styles/admin.css
```

## Rules

1. Keep one font-family token: `CommercePlatformSystem`.
2. Use `.woff2` only for production fonts.
3. Do not load Google Fonts or npm font packages for the system UI.
4. Add future language fonts as additional `@font-face` blocks with `unicode-range`, instead of changing page components.
5. When adding or replacing a system font, update this document in the same commit.

## Next phase

The current font covers Chinese and English. The remaining 32 languages should be grouped by script before adding font files, for example Latin Extended, Cyrillic, Greek, Arabic, Thai, Japanese, Korean, and other required scripts. Avoid loading all language fonts globally without `unicode-range`.
