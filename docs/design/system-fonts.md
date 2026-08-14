# Self-Hosted UI Fonts

Last updated: 2026-08-13

## Current source

The storefront and admin use self-hosted fonts only. Do not declare system,
generic, CDN, or npm-provided font fallbacks.

```text
StorefrontSystem
```

The current storefront font file is:

```text
StorefrontSystem-CN-Latin.woff2
StorefrontSystem-Devanagari.woff2
StorefrontSystem-Latin-Accents.woff2
StorefrontSystem-Arabic.woff2
StorefrontSystem-Thai.woff2
```

Because the storefront and admin are built from different Docker contexts, the font file must exist in both source trees:

```text
nuxt-i18n/public/fonts/StorefrontSystem-CN-Latin.woff2
nuxt-i18n/public/fonts/StorefrontSystem-Devanagari.woff2
nuxt-i18n/public/fonts/StorefrontSystem-Latin-Accents.woff2
nuxt-i18n/public/fonts/StorefrontSystem-Arabic.woff2
nuxt-i18n/public/fonts/StorefrontSystem-Thai.woff2
go-backend/web/admin/src/assets/fonts/system/AdminSystem-CN-Latin.woff2
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
nuxt-i18n/tailwind.config.ts
```

Admin:

```text
go-backend/web/admin/src/styles/admin.css
```

## Rules

1. Keep one storefront font-family token: `StorefrontSystem`.
2. Use `.woff2` only for production fonts.
3. Do not load Google Fonts, CDN font files, system-font fallbacks, or npm font packages for the system UI.
4. Add future self-hosted language font files as additional `@font-face` blocks for `StorefrontSystem` with exact `unicode-range` values, instead of changing page components.
5. Run `npm run check:font-coverage` before enabling a locale or replacing a font. The check reads all declared `StorefrontSystem` files and verifies every locale message against their combined glyph coverage.
6. When adding or replacing a system font, update this document in the same commit.
7. Keep `public/fonts/storefront-system.css` aligned with the storefront `@font-face` declarations. Stripe Elements uses this public stylesheet because its payment fields render in an iframe.

The current `StorefrontSystem-CN-Latin.woff2` is Maple UI Regular. It covers the
current English, Chinese, Japanese, Korean, and Russian locale messages.
Additional self-hosted Noto files cover the remaining current locale scripts:

| File | Purpose |
| --- | --- |
| `StorefrontSystem-Devanagari.woff2` | Hindi / Devanagari |
| `StorefrontSystem-Arabic.woff2` | Arabic |
| `StorefrontSystem-Thai.woff2` | Thai |
| `StorefrontSystem-Latin-Accents.woff2` | Missing uppercase Latin accents and Turkish dotted capital I |

Each file uses exact `unicode-range` declarations. The Latin files are limited
to glyphs missing from Maple UI so the existing English and Chinese visual
identity remains unchanged.

The original Noto source fonts, OFL licenses, and upstream metadata live under:

```text
nuxt-i18n/fonts/source/noto-sans/
nuxt-i18n/fonts/source/noto-sans-arabic/
nuxt-i18n/fonts/source/noto-sans-devanagari/
nuxt-i18n/fonts/source/noto-sans-thai/
```

Do not replace or modify the source license files when updating these fonts.
When CMS or catalog content introduces new characters, rerun
`npm run check:font-coverage` and expand a script's self-hosted subset before
publishing.
