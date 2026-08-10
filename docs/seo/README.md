# SEO Documentation

This directory is the project-level source of truth for SEO architecture and
ownership boundaries.

The SEO administration domain is a control surface for page metadata. It is
not the owner of product catalog data, media metadata, reviews, shipping
policies, or Google Merchant channel credentials.

## Documents

- [SEO system architecture](./SEO_SYSTEM_ARCHITECTURE.md)
- [Google Merchant Center implementation plan](../../go-backend/docs/GOOGLE_MERCHANT_CENTER_IMPLEMENTATION_PLAN.md)

## Current Decisions

- Keep SEO as a top-level collapsible administration domain.
- Keep three SEO resource sections: Home, Articles, and Products.
- Keep the Article and Product lists linked to their real storefront routes.
- Return those routes as a read-only `route_path` projection from the SEO API;
  the admin UI must not infer article categories or locale prefixes.
- Use `Blog` as the content-center name. Keep `news` and `wheelsbuild` as
  article categories only, so future content categories do not inherit a
  product-specific top-level route name.
- Make public Blog detail responses carry real `localized_routes`; the
  storefront must use those routes for hreflang and translation links.
- Treat a missing localized article or a category/path mismatch as HTTP 404.
- Keep Product and Article SEO values attached to their resource aggregates.
  The SEO API is an administrative use-case boundary, not a duplicate content
  database.
- Keep Google Merchant as a separate channel domain and administration page.
- Keep image, alt text, caption, and media role ownership in the content/media
  domain.
- Generate Product JSON-LD from the same public product data shown on the
  storefront. Do not make JSON-LD fields manually editable in the SEO dialog.
- Generate Product canonical URLs from the localized route by default. Manual
  canonical overrides are exceptional and require explicit validation.
- Use Product's real `localized_routes` projection for hreflang. Never create
  translated product URLs by replacing a locale prefix around the current
  slug.
- Resolve a public Product by the requested locale. Missing translations are
  a real absence, not an instruction to serve another language under the
  wrong URL.

## Maintenance Rule

When implementation and documentation disagree, the current code is the
runtime source of truth until the architecture decision is updated and the
code is migrated. Do not add a second SEO storage model only to make the
administration UI easier.
