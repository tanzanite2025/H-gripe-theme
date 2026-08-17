# Self-Hosted UI Fonts

Last updated: 2026-08-17

## Current source

The storefront and admin use self-hosted fonts only. Do not declare system,
generic, CDN, or npm-provided font fallbacks.

```text
StorefrontSystemLatin, StorefrontSystem
```

The current storefront font files are:

```text
StorefrontSystem-Latin.00af3fec5b34.woff2
StorefrontSystem-CJK.f8ce6d72e8cb.woff2
StorefrontSystem-Devanagari.3b3cae4d2600.woff2
StorefrontSystem-Latin-Accents.e645edc952b6.woff2
StorefrontSystem-Arabic.ce85091f0209.woff2
StorefrontSystem-Thai.1f5a173641bb.woff2
```

Because the storefront and admin are built from different Docker contexts, the font file must exist in both source trees:

```text
nuxt-i18n/public/fonts/StorefrontSystem-Latin.00af3fec5b34.woff2
nuxt-i18n/public/fonts/StorefrontSystem-CJK.f8ce6d72e8cb.woff2
nuxt-i18n/public/fonts/StorefrontSystem-Devanagari.3b3cae4d2600.woff2
nuxt-i18n/public/fonts/StorefrontSystem-Latin-Accents.e645edc952b6.woff2
nuxt-i18n/public/fonts/StorefrontSystem-Arabic.ce85091f0209.woff2
nuxt-i18n/public/fonts/StorefrontSystem-Thai.1f5a173641bb.woff2
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

1. Use `StorefrontSystemLatin, StorefrontSystem` for default storefront text. The first file is a small Maple UI Latin subset; the complete Maple UI CJK file is only a fallback for CJK and uncommon glyphs.
2. Use `.woff2` only for production fonts.
3. Do not load Google Fonts, CDN font files, system-font fallbacks, or npm font packages for the system UI.
4. Add future self-hosted language font files as additional `@font-face` blocks with exact `unicode-range` values, instead of changing page components.
5. Run `npm run check:font-coverage` before enabling a locale or replacing a font. The check reads all declared `StorefrontSystem` files and verifies every locale message against their combined glyph coverage.
6. Run `npm run check:font-performance` before publishing. It enforces the default Latin subset budget and content-addressed production font filenames.
7. Use a content hash in every production font filename. `/fonts/**` is immutable; `public/fonts/storefront-system.css` is intentionally revalidated because Stripe loads that stable path in an iframe.
8. When adding or replacing a system font, update this document in the same commit.
9. Keep `public/fonts/storefront-system.css` aligned with the storefront `@font-face` declarations. Stripe Elements uses this public stylesheet because its payment fields render in an iframe.

`StorefrontSystem-CJK.f8ce6d72e8cb.woff2` is the complete Maple UI Regular
file. `StorefrontSystem-Latin.00af3fec5b34.woff2` is a 48KB subset generated
from the same source and covers Latin text and common storefront punctuation,
currency, arrows, and symbols. The CJK fallback has an explicit
`unicode-range`, so unsupported Latin symbols and emoji do not cause its
6.4MB file to download. The full file covers the current Chinese,
Japanese, Korean, and Russian locale messages when those glyphs occur.
Additional self-hosted Noto files cover the remaining current locale scripts:

| File | Purpose |
| --- | --- |
| `StorefrontSystem-Devanagari.3b3cae4d2600.woff2` | Hindi / Devanagari |
| `StorefrontSystem-Arabic.ce85091f0209.woff2` | Arabic |
| `StorefrontSystem-Thai.1f5a173641bb.woff2` | Thai |
| `StorefrontSystem-Latin-Accents.e645edc952b6.woff2` | Missing uppercase Latin accents and Turkish dotted capital I |

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
