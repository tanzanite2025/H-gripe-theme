# Maple UI Font Authority

Last updated: 2026-08-18

This project has one font authority for the storefront: Maple UI shards plus a
small set of bundled coverage shards. No OS/system fallback, no CDN fonts, no
`@nuxt/fonts`, and no other secondary font source are allowed to sneak back in.

## Why The Files Are Named This Way

The files carry the `MapleUI-` prefix so they read like owned assets, not random
system leftovers. `MapleUI-Latin` is the first-paint subset, `MapleUI-CJK` is
the full Maple UI core shard, and `MapleUI-Coverage-NotoSans-*` are internal
coverage shards for the enabled language set.

## Approved Stack

```text
MapleUILatin, MapleUICJK
```

That is the storefront default. Coverage shards are only added through
`unicode-range`-scoped `@font-face` blocks when a locale needs characters Maple
UI does not ship.

## Approved Files

```text
nuxt-i18n/public/fonts/MapleUI-Latin.00af3fec5b34.woff2
nuxt-i18n/public/fonts/MapleUI-CJK.f8ce6d72e8cb.woff2
nuxt-i18n/public/fonts/MapleUI-Coverage-NotoSans-Devanagari.3b3cae4d2600.woff2
nuxt-i18n/public/fonts/MapleUI-Coverage-NotoSans-Latin-Accents.e645edc952b6.woff2
nuxt-i18n/public/fonts/MapleUI-Coverage-NotoSans-Arabic.ce85091f0209.woff2
nuxt-i18n/public/fonts/MapleUI-Coverage-NotoSans-Thai.1f5a173641bb.woff2
nuxt-i18n/public/fonts/maple-ui.css
```

## Rules

1. Use `MapleUILatin, MapleUICJK` for default storefront text.
2. Keep every production font filename content-addressed and `.woff2` only.
3. Keep `public/fonts/maple-ui.css` aligned with `app/assets/css/tailwind.css`.
4. Do not reintroduce `StorefrontSystem*`, `storefront-system.css`, `@nuxt/fonts`,
   or any system fallback families.
5. When adding a new locale or character block, add a new `MapleUI-Coverage-*`
   shard with exact `unicode-range` values instead of changing page components.
6. Run `npm run clean` before `build` or `generate` so stale `.nuxt`, `.output`,
   and `dist` artifacts cannot keep serving retired fonts.
7. Run `npm run check:font-policy`, `npm run check:font-coverage`, and
   `npm run check:font-performance` before shipping font changes.
8. Keep the preflight manifest and production artifact checks in the same commit
   as font changes.
9. The admin panel uses the same built-in Maple UI authority. Its
   `font-sans`, `font-mono`, chart canvas text, Tailwind preflight, and
   third-party toast CSS must resolve to `MapleUICJK`; the Admin build must pass
   its source and `dist` font gates.

## Source Of Truth

Storefront CSS:

```text
nuxt-i18n/app/assets/css/tailwind.css
nuxt-i18n/public/fonts/maple-ui.css
nuxt-i18n/tailwind.config.ts
```

Font source archives and licenses:

```text
nuxt-i18n/fonts/source/noto-sans/
nuxt-i18n/fonts/source/noto-sans-arabic/
nuxt-i18n/fonts/source/noto-sans-devanagari/
nuxt-i18n/fonts/source/noto-sans-thai/
```

Do not replace or edit the source OFL files unless the upstream font source
changes. When CMS or catalog content introduces new characters, rerun
`npm run check:font-coverage` and expand a built-in coverage shard before
publishing.

Admin authority:

```text
go-backend/web/admin/src/styles/admin.css
go-backend/web/admin/src/assets/fonts/maple-ui/MapleUI-CJK.f8ce6d72e8cb.woff2
go-backend/web/admin/vite.config.ts
go-backend/web/admin/scripts/check-admin-font-policy.ts
```

The Vite authority plugin rewrites framework and third-party CSS fallback
stacks before development and production output. This is required because a
dependency can carry a system stack even when application source code does not.
