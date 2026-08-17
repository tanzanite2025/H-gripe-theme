# Storefront Content Security Policy

The storefront owns its CSP in the Nuxt Nitro response pipeline. Nginx forwards
that response header and continues to own transport-oriented headers such as
HSTS, `X-Content-Type-Options`, and `Referrer-Policy`.

## Why hashes instead of nonces

The storefront caches SSR HTML. A per-request nonce would either invalidate
that cache or risk serving a cached nonce with a different CSP header. The
Nitro `render:response` hook instead hashes the final inline script and style
content and emits the matching `Content-Security-Policy` header. The HTML and
its header therefore remain a valid cache pair.

Only inline hashes are derived from final HTML. Resource origins are not
auto-approved from page content, because product/blog/FAQ content can be
admin-authored data. External image, media, script, style, frame, font, form,
and connect origins must come from reviewed runtime configuration or the
small first-party/static SDK allowlists in the CSP module.

## Trusted Types and rich text

The policy requires Trusted Types for script sinks and only permits Vue's
`vue` policy plus the first-party `tanzanite-script-url` policy. Dynamic
third-party SDK scripts must use the helper in
`nuxt-i18n/app/utils/security/trustedScriptUrl.ts`; the helper only accepts
the reviewed Google Identity, Turnstile, Stripe, Google Analytics, and Google
Tag Manager URLs.

Do not add `v-html`. Storefront-authored rich text is rendered through
`SafeRichText.vue`; the sanitizer and VNode renderer live in
`nuxt-i18n/app/utils/security/safeRichText.ts`. It preserves an explicit
formatting subset while removing scriptable elements, inline handlers, inline
styles, and unapproved URLs. Rich-text media may use relative storefront paths,
raster `data:image` URLs, or exact first-party origins derived from
`runtimeConfig.public.siteUrl` and `runtimeConfig.public.apiBase`. External
content media should be brought through the first-party upload/IPX flow or a
dedicated reviewed component, not embedded directly in admin-authored HTML.

## Source changes

Prefer same-origin assets through the existing upload and IPX paths. When a
reviewed deployment needs another exact origin, add it to the matching
comma-separated variable in `deployment/production.env.example`. Values must
be origins such as `https://media.example.com`; paths, wildcard hosts, and
credentials are rejected.

## Verification

Run these after a storefront build:

```powershell
npm run check:security
npm run smoke:html-cache
```

The smoke check starts the production Nitro output, checks the CSP and Trusted
Types directives, validates every final HTML hash against the header, and
verifies every absolute resource origin in the final HTML against its CSP
directive. It then requests the cached page twice to confirm the header
remains paired with the cached HTML.
