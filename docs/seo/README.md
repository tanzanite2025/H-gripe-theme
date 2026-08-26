# SEO Documentation

This directory is the project-level source of truth for SEO architecture and
ownership boundaries.

The SEO administration domain is a control surface for page metadata. It is
not the owner of product catalog data, media metadata, reviews, shipping
policies, or Google Merchant channel credentials.

## Documents

- [SEO system architecture](./SEO_SYSTEM_ARCHITECTURE.md)
- [E-commerce URL and SEO architecture](./ECOMMERCE_URL_ARCHITECTURE.md)
- [Google Indexing API](./GOOGLE_INDEXING.md)
- [Google Merchant Center implementation plan](../../go-backend/docs/GOOGLE_MERCHANT_CENTER_IMPLEMENTATION_PLAN.md)

## Current Decisions

- Keep SEO as a top-level collapsible administration domain.
- Keep four SEO resource sections: Home, Articles, Products, and Categories.
- Keep the Article, Product, and Category lists linked to their real storefront routes.
- Keep Category SEO metadata attached to the product-category aggregate:
  `Meta Title`, `Meta Description`, and the sanitized rich-text `intro`.
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
- Use hierarchical `/shop/...` routes for category landing pages and the flat
  `/products/:slug` route for the target product detail architecture.
- Keep product permalinks independent from category membership. Express the
  category path through visible navigation and SSR `BreadcrumbList` JSON-LD.
- Generate Product canonical URLs from the localized route by default. Manual
  canonical overrides are exceptional and require explicit validation.
- Use Product's real `localized_routes` projection for hreflang. Never create
  translated product URLs by replacing a locale prefix around the current
  slug.
- Resolve a public Product by the requested locale. Missing translations are
  a real absence, not an instruction to serve another language under the
  wrong URL.

## Maintenance Rule

When implementation and documentation disagree, resolve the discrepancy
against the formal route contract before adding more SEO behavior. Do not add a
second SEO storage model only to make the administration UI easier.

The e-commerce URL document is the active route contract: categories use
`/shop/...`, products use `/products/:slug`, and a non-category `/shop/:slug`
returns 404. URL Management 301 rules are reserved for real URLs created after
the site is operated; pre-launch code must not generate product aliases.
