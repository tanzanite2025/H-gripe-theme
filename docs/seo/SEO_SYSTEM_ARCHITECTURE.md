# SEO System Architecture

## 1. Purpose

This document defines the long-term boundary between:

- storefront page SEO;
- product and article source data;
- media and image metadata;
- Product structured data;
- Google Merchant Center distribution data;
- reviews, shipping, returns, and policy data;
- SEO administration workflows.

The goal is a stable data chain with one source of truth for each fact. SEO
must remain easy to extend without turning the Settings domain, Product domain,
Content domain, or Google Merchant channel into a shared miscellaneous bucket.

This document covers the website and administration architecture. Google
Merchant synchronization details remain in
[GOOGLE_MERCHANT_CENTER_IMPLEMENTATION_PLAN.md](../../go-backend/docs/GOOGLE_MERCHANT_CENTER_IMPLEMENTATION_PLAN.md).

## 2. Google Model

The project must distinguish the following concepts:

| Concept | Purpose | Owner in this project |
| --- | --- | --- |
| Visible page content | What users and crawlers can read on the page | Product or Content domain |
| HTML page metadata | `title`, description, canonical, hreflang, robots | Storefront SEO renderer |
| Product structured data | JSON-LD describing a product and its offers | Storefront SEO renderer from catalog data |
| Merchant channel data | Google-specific offers, identifiers, target market, sync state | Google Merchant domain |
| Reviews and ratings | Genuine customer feedback and aggregate values | Review domain |
| Shipping and returns | Actual delivery and policy information | Shipping and policy domains |

Product JSON-LD is not a Merchant Center feed. Merchant Center is not the
source of truth for catalog title, price, stock, or storefront content.

Google's current documentation should be checked before changing the renderer:

- [Product structured data](https://developers.google.com/search/docs/appearance/structured-data/product)
- [Merchant listings structured data](https://developers.google.com/search/docs/appearance/structured-data/merchant-listing)
- [Product variants structured data](https://developers.google.com/search/docs/appearance/structured-data/product-variants)
- [Title links](https://developers.google.com/search/docs/appearance/title-link)
- [Search snippets](https://developers.google.com/search/docs/appearance/snippet)
- [Canonicalization and duplicate URLs](https://developers.google.com/search/docs/crawling-indexing/consolidate-duplicate-urls)

Fixed character counts such as "50-60 characters" or "150 characters" are
project guidance only. They are not Google data-model requirements. The
administrator UI may provide soft warnings and previews, but should not
pretend that a single character count guarantees a search result appearance.

## 3. Target Architecture

```text
Product catalog --------------------+
Product variants -------------------+--> Storefront SEO renderer
Content / blog ---------------------+       |
Media / images ---------------------+       +--> HTML head
Reviews ----------------------------+       +--> Visible page content
Shipping / returns / policies ------+       +--> Product JSON-LD
                                                 |
Google Merchant channel fields -----------------+--> Merchant mapper
                                                         |
                                                         +--> Merchant API
                                                         +--> Sync status
                                                         +--> Google diagnostics
```

The storefront renderer is the only layer that decides what is emitted in a
public page. The administration domains edit source data and configuration;
they do not write arbitrary HTML head fragments.

## 4. Domain Ownership

### 4.1 Product Catalog Domain

Owns:

- product name;
- product slug and locale;
- product description and short description;
- SKU;
- base price, sale price, currency;
- stock and publication status;
- variants, variant SKU, option values, variant price, and variant stock;
- product specification template and technical specifications;
- product-to-translation relationships.

Does not own:

- Google Merchant OAuth credentials;
- Merchant sync status;
- Google target country or feed label;
- arbitrary canonical URL overrides;
- search title editing workflow.

Current implementation references:

- `go-backend/internal/domain/product/product.go`
- `go-backend/internal/api/v1/product/public_response.go`
- `nuxt-i18n/app/pages/shop/[slug].vue`

### 4.2 Content and Blog Domain

Owns:

- article title;
- article slug, locale, status, publication time;
- article body and visible headings;
- article tags and category routing;
- article media references.

Article SEO is edited from the SEO administration domain, but the article
record remains the source of truth.

The Content create/edit contract must not accept or write article Meta Title,
Meta Description, Meta Keywords, or Canonical URL. Those fields are edited
only through the SEO / 文章 workflow. Image URL, alt text, title/caption, and
media role remain content/media responsibilities.

The public Blog route contract is:

- `/blog/:slug` for uncategorized articles;
- `/blog/news/:slug` for News articles;
- `/blog/wheelsbuild/:slug` for Wheelbuild articles;
- the same route family with the locale prefix for non-default locales.

`wheelsbuild` is a content category, not the name of the overall content
center. The storefront and administration domain use `Blog` as the top-level
name so future article categories do not inherit a product-specific label.

### 4.3 Media Domain

Owns:

- media URL;
- media type;
- visibility;
- primary/media role;
- alt text;
- media title or caption;
- variant or option association.

The SEO administration domain must not duplicate image records or become an
image metadata editor. Product JSON-LD may consume the public primary image and
visible product images from this domain.

### 4.4 SEO Control Plane

Owns:

- Home page SEO defaults;
- SEO administration APIs and permissions;
- resource discovery, filtering, pagination, and route display;
- Meta Title and Meta Description editing workflows;
- exceptional Article canonical override;
- validation and diagnostics for public SEO output;
- SEO-specific audit events in the future.

Does not own:

- product name, H1, price, stock, SKU, or images;
- article body or article title;
- Google Merchant brand, GTIN, MPN, category, or target market;
- review data;
- shipping or return policy content;
- arbitrary custom JSON-LD entered by administrators.

The current `SEOResourceService` is correctly shaped as an administrative
facade over Post and Product services. It should not become a second product or
post repository.

### 4.5 Storefront SEO Renderer

Owns the public rendering contract:

- page `<title>`;
- meta description;
- canonical URL;
- hreflang and `x-default`;
- Open Graph and Twitter metadata;
- Product JSON-LD;
- omission of invalid or incomplete structured data.

This code must use server-renderable data so the initial HTML contains the
important metadata and structured data. Client-only updates after hydration
must not be the only way for a crawler to discover price, stock, title, or
canonical information.

### 4.6 Google Merchant Domain

Owns channel-specific data:

- Merchant Account and Data Source configuration;
- brand;
- GTIN and MPN;
- identifier-exists decision;
- product condition;
- Google product category;
- target country;
- content language;
- feed label;
- Merchant publication switch;
- stable external offer ID;
- validation status;
- sync status;
- last sync time;
- remote resource ID;
- issue and retry history.

The Google Merchant domain must stay separate from the Product editor. A
Google-specific field must not be added to the catalog merely because Google
requires it.

Current implementation references:

- `go-backend/internal/domain/merchant/`
- `go-backend/internal/service/google_merchant_service.go`
- `go-backend/internal/api/admin/google_merchant_handler.go`
- `go-backend/web/admin/src/views/GoogleMerchant.vue`
- `go-backend/migrations/066_google_merchant_offers.up.sql`

## 5. Administration Structure

The administration navigation should remain:

```text
SEO
├── 首页
├── 文章
└── 产品

Google Merchant
```

### 5.1 SEO / 首页

Purpose:

- edit the localized homepage title and description;
- show the locked homepage route;
- show the active locale;
- provide a public-page link.

The homepage editor should not expose product fields, schema fields, or
Merchant fields.

### 5.2 SEO / 文章

Purpose:

- list real article resources;
- show article title, locale, status, and actual storefront route;
- open the article in a new tab;
- edit Meta Title and Meta Description;
- support a carefully validated canonical override for migration or exceptional
  duplicate-content cases.

The article editor must not edit the article H1 or body. Those belong to the
Content domain.

### 5.3 SEO / 产品

Purpose:

- list real localized product resources;
- show product name, locale, status, and actual storefront route;
- open the product page;
- edit Meta Title and Meta Description;
- show whether the public product page has enough data for Product JSON-LD.

Important: the Product JSON-LD fields listed in this document describe the
public page output contract. They are not a list of editable fields for the
Product SEO dialog. The dialog edits only the SEO-owned Meta Title and Meta
Description. Product facts must continue to be projected from their source
domains and may be shown in the dialog as read-only values or readiness
diagnostics.

The Product SEO editor must not contain editable copies of:

- SKU;
- price;
- inventory;
- variant values;
- brand;
- GTIN or MPN;
- image metadata;
- shipping or return rules;
- review or rating values;
- arbitrary JSON-LD.

Those fields belong to their own domains. The Product SEO editor may show a
read-only structured-data projection and readiness diagnostics, but it must
not create editable copies of those values. This keeps the SEO editor small
and prevents operators from creating a second, conflicting product record.

The general Product create/edit dialog must also omit Meta Title and Meta
Description. Those values are edited only through the SEO / 产品 workflow;
the Product API remains responsible for catalog data and must not become a
second SEO editing entry point.

## 6. Page and Route Model

### 6.1 Homepage

| Resource | Default route | Localized route | SEO source |
| --- | --- | --- | --- |
| Home | `/` | locale-aware storefront route | SEO home settings |

### 6.2 Article

| Resource | Route family | SEO source |
| --- | --- | --- |
| Article | `/blog/:slug` and category routes | Article record through SEO facade |

The route shown in the admin table must come from the read-only `route_path`
projection returned by the SEO resource API. The projection is generated from
the same storefront route rules used by the public application. Do not make
the admin UI infer paths from tags or hand-type paths. Public article
responses also expose `localized_routes`, which are the only valid translated
article URLs for the current published translation group.

The current Nuxt locale strategy is `prefix_except_default`: `en` uses no
prefix and `zh_cn` uses `/zh_cn`. The admin SEO route builder must preserve
these exact locale codes when constructing deep links.

The admin API returns the read-only `route_path` projection for Article and
Product resources. The admin UI consumes that projection and must not rebuild
resource routes from tags, slugs, or locale prefixes.

### 6.3 Product

| Resource | Route family | SEO source |
| --- | --- | --- |
| Product | `/shop/:slug` with locale prefix when applicable | Product record through SEO facade |

The product canonical URL must be generated from the localized public route.
The normal product workflow should not require a manually entered canonical
URL.

Product translations use the existing Product-domain relationship with one
root row and direct translated children:

```text
root product:        parent_id = NULL
translated product:  parent_id = root product id
```

The Product service rejects self-links, nested translation children, repeated
locales in one group, and a translation with the same locale as its root.
Public product detail responses expose only the route projection:

```json
{
  "localized_routes": [
    { "locale": "en", "slug": "example-product" },
    { "locale": "zh_cn", "slug": "example-product-zh" }
  ]
}
```

Only active localized products are included. The storefront must not infer
missing translated slugs from the current slug or emit hreflang links to
languages that have no public product row.

A product page with `/zh-cn/shop/example` should normally canonicalize to the
same localized URL, not silently canonicalize to `/shop/example`, unless the
business has explicitly chosen a single-language canonical strategy and has
documented the consequences for hreflang.

## 7. Field Ownership and Data Flow

| Field or signal | Source of truth | SEO admin edits it? | Storefront uses it? | Merchant uses it? |
| --- | --- | ---: | ---: | ---: |
| Product name | Product catalog | No | Yes, H1 and Schema name | Yes, through mapper |
| Article title | Blog/content | No | Yes, H1 and title fallback | No, unless separately syndicated |
| Meta Title | Product/article/home SEO source | Yes | Yes | Usually no |
| Meta Description | Product/article/home SEO source | Yes | Yes | Usually no |
| H1 | Product/article source content | No | Yes | No |
| Canonical URL | Route renderer | Article exceptions only | Yes | Landing URL input |
| Hreflang | Locale and route registry | No | Yes | Target language input |
| Primary image | Media domain | No | Yes | Yes |
| Image alt text | Media domain | No | Yes | Optional channel input |
| SKU | Product/variant catalog | No | Schema and commerce UI | Yes |
| Brand | Approved site brand configuration, with Merchant validation where applicable | No | Optional Schema input | Yes |
| GTIN / MPN | Merchant channel | No | Optional Schema input | Yes |
| Price and currency | Product/market pricing | No | Visible price and Offer | Yes |
| Availability | Inventory/catalog | No | Visible stock and Offer | Yes |
| Reviews and rating | Review domain | No | Visible review block and Schema | Only if supported |
| Shipping and returns | Shipping/policy domain | No | Policy links/content | Yes |
| Sync status | Merchant domain | No | No | Yes |

The same fact may be projected into several outputs, but it must have one
source of truth. For example, the product SKU may appear in the catalog UI,
Product JSON-LD, and Merchant API payload, but it must not be independently
edited in all three places.

## 8. Storefront Rendering Contract

### 8.1 HTML Metadata

For every indexable page:

- emit one effective page title;
- emit a description when meaningful content exists;
- emit one canonical URL;
- emit locale alternates and `x-default` when translations exist;
- keep page title and visible primary heading semantically aligned;
- do not use `meta keywords` as a ranking feature;
- do not emit a canonical URL that points to a different language by accident.

The renderer may use fallbacks, but fallbacks must be explicit:

```text
Product Meta Title
-> Product name
-> Site-level product fallback

Product Meta Description
-> Product Meta Description
-> Product short description
-> Product description
-> Omit description when no meaningful text exists
```

### 8.2 Product JSON-LD

For a purchasable product page, the renderer should build structured data from
the public product view model:

```text
Product
├── name                 <- visible product name, not Meta Title
├── description          <- meaningful product description
├── image                <- visible primary/product images
├── sku                  <- catalog SKU when available
├── brand                <- approved brand source when available
├── gtin / mpn           <- verified identifiers only
├── offers
│   ├── price
│   ├── priceCurrency
│   ├── availability
│   └── url
└── aggregateRating/review <- genuine visible review data only
```

Required implementation rules:

1. `Product.name` must use the visible product name.
2. The main image must be an absolute public URL.
3. Price, currency, availability, and visible default variant must agree.
4. Do not output SKU, GTIN, MPN, brand, or ratings when the value is unknown or
   unverified.
5. Do not fabricate reviews or ratings to obtain rich-result decoration.
6. Do not emit an incomplete `Offer` when price or currency is unavailable.
7. Do not claim that JSON-LD alone enables Merchant Center distribution.

### 8.3 Variant Products

The project supports variants, so the renderer must choose one explicit
strategy:

- for a product with one purchasable configuration, emit a normal Product and
  Offer;
- for a true variant family, emit ProductGroup/hasVariant data and stable
  variant identifiers;
- if the current page cannot expose a deterministic variant landing state, do
  not claim a variant-specific price or stock that the URL cannot reproduce.

The current single dynamic Offer is acceptable as a transitional baseline, but
it is not the final variant model.

### 8.4 Server Rendering

The product data request and head generation should be SSR-compatible. Tests
must inspect the initial HTML response rather than only a hydrated browser DOM.

At minimum, a rendered product page must be testable for:

- title;
- description;
- H1;
- canonical;
- hreflang;
- JSON-LD presence and JSON validity;
- product name;
- price and currency;
- availability;
- absolute image URLs.

The route-link implementation is centralized in
`nuxt-i18n/app/composables/seo/useStorefrontSeoLinks.ts`. Both the default and
products layouts use it. A resource page may provide a route override through
the shared `alternateLinksOverride` state when its backend response contains
real translated routes. This keeps canonical and hreflang output in the
layout boundary while allowing Product and Article pages to provide
resource-specific route data.

Blog detail pages use:

- `nuxt-i18n/app/composables/blog/useBlogPostDetail.ts` for locale-aware
  loading, category validation, and 404 behavior;
- `nuxt-i18n/app/composables/seo/useBlogPostSeo.ts` for title, description,
  canonical, and alternate route data;
- `nuxt-i18n/app/components/blog/BlogPostDetailContent.vue` for the shared
  article presentation;
- `nuxt-i18n/app/components/PostTranslations.vue` for translation links,
  preferring server-provided `localized_routes`.

## 9. Canonical and Locale Rules

Canonical and hreflang are routing signals, not editable marketing copy.

Rules:

1. Every locale page has a stable public route.
2. The default canonical is the current locale's clean route.
3. Query parameters used only for UI state must not become canonical.
4. A variant query parameter may be used only when the landing page reliably
   restores that variant.
5. Hreflang links must point to equivalent translated resources.
6. `x-default` must point to the documented default-market route.
7. A manual canonical override must be absolute, HTTPS, same-site unless an
   approved migration requires otherwise, and recorded with an audit reason.
8. A product canonical override is not part of the normal Product SEO editor.

For Product pages specifically:

- the API lookup uses the requested locale when one is present;
- a slug from another locale is not silently returned as a fallback;
- a missing localized product page returns HTTP 404 and must not render a
  successful empty product shell;
- the current locale's product row supplies the canonical slug;
- `localized_routes` supplies the only valid translated slugs;
- the default storefront locale is used for `x-default` only when that route
  actually exists; otherwise the first real localized route is used.

The current storefront layout already has a general alternate-link mechanism.
Product pages must use the same locale-aware canonical strategy rather than
replacing it with a hard-coded `/shop/:slug` path.

## 10. Change Propagation

The following table defines what should happen after a change:

| Change | Storefront head | Product JSON-LD | Merchant channel |
| --- | ---: | ---: | ---: |
| Meta Title | Update | Usually no | No |
| Meta Description | Update | Update only if used as Schema description | Usually no |
| Product name | Update H1/title/schema | Update | Queue sync if published |
| Article title/body | Update visible page/title fallback | Article schema if implemented | No |
| Primary image | Update social image/schema | Update | Queue sync if published |
| Price or sale price | Visible price update | Update Offer | Queue sync if published |
| Stock or active status | Visible availability | Update Offer | Queue sync or withdraw |
| Variant SKU/price/stock | Variant UI and schema | Update variant data | Queue affected offer |
| Slug or locale | Route, canonical, hreflang | Update URL | Reconcile landing URL |
| Brand/GTIN/MPN | Optional Schema projection | Update only verified values | Update Merchant offer |
| Shipping/return policy | Policy links/content | Optional policy projection | Update Merchant data |
| Review aggregate | Visible review block | Update only genuine values | Usually no |

SEO Meta updates must not trigger Google Merchant synchronization unless a
channel-specific title or description is explicitly derived from that field.
Conversely, a Merchant-specific brand or identifier update must not rewrite the
catalog product name.

### 10.1 Implemented Data Chains

SEO-only product update:

```text
SEO / 产品
  -> SEOResourceService.UpdateProduct
  -> ProductService.UpdateProductSEO
  -> products.meta_title / products.meta_description
  -> storefront cache invalidation
  -> no Merchant outbox event
```

SEO-only article update:

```text
SEO / 文章
  -> SEOResourceService.UpdateArticle
  -> PostService.UpdatePostSEO
  -> posts.meta_title / posts.meta_description / posts.canonical_url
  -> storefront cache invalidation
```

Content article create/update:

```text
内容管理
  -> ContentHandler
  -> rejects article SEO request fields
  -> PostService.CreateAdminPost / PostService.UpdateAdminPost
  -> article title/body/status/media/tags only
```

Catalog-to-Merchant update:

```text
Product or Variant/Media mutation
  -> ProductAdminService
  -> MerchantCatalogEventPublisher
  -> outbox_events
  -> OutboxService retry/dead-letter worker
  -> GoogleMerchantOutboxHandler
  -> GoogleMerchantService
  -> Merchant API
  -> google_merchant_offers sync state
```

Merchant-only update:

```text
Google Merchant page
  -> GoogleMerchantHandler
  -> GoogleMerchantService
  -> google_merchant_offers
  -> optional manual sync/reconcile
```

The SEO control plane never writes Merchant channel fields, and the Merchant
channel never writes product Meta fields. Product create/update contracts omit
Meta fields entirely; the dedicated `UpdateProductSEO` boundary is the only
product Meta write path. Content create/update contracts omit and reject
article SEO fields; the dedicated `UpdatePostSEO` boundary is the only article
SEO Meta write path.

## 11. Storage and API Rules

### 11.1 Do Not Duplicate Resource SEO Data

The current model is intentionally mixed by resource:

- Home SEO values are stored as SEO settings.
- Article SEO values remain on the article record.
- Product SEO values remain on the product record.
- The SEO domain exposes resource-oriented administrative APIs.
- Merchant values remain in Merchant channel tables.

Keep this model until the system needs one of the following:

- SEO revision history;
- scheduled SEO publication;
- reusable SEO templates with inheritance;
- approval workflow separate from content publication;
- page types that do not have a natural resource aggregate.

Only then consider a dedicated `seo_resource_overrides` or revision model. A
new table must identify resource type, resource ID, locale, version, status,
and effective time. It must not copy product price, stock, image, or body data.

### 11.2 Typed Contracts

New SEO implementation code should use TypeScript types on the admin and
storefront sides and explicit Go request/response types on the backend.

Recommended contracts:

- `SEOHomeSettings`
- `SEOArticleResource`
- `SEOProductResource`
- `SEOPageRoute`
- `ProductSEOViewModel`
- `ProductStructuredData`
- `MerchantOfferView`

Do not pass loosely shaped `any` objects between the SEO API, renderer, and
Merchant mapper when a typed contract can express the boundary.

### 11.3 Public Output Must Be Safe

The public SEO renderer must never expose:

- OAuth tokens;
- Merchant account secrets;
- internal stock counts;
- private customer or order data;
- administrator-only URLs;
- unapproved review values;
- raw internal error details.

## 12. Validation Rules

### 12.1 SEO Editor

- Meta Title and Meta Description use soft length warnings, not Google
  guarantees.
- Empty values clearly show when a fallback is active.
- Localized values are edited in the selected locale.
- The route is read-only and links to the real storefront URL.
- Product SEO editing does not expose Merchant fields.

### 12.2 Product Structured Data

Block or omit structured data when:

- the product is not public;
- the product name is empty;
- no public image exists when an image is required for the selected result type;
- price or currency is missing for an Offer;
- availability cannot be reconciled with the visible page;
- an identifier is malformed or unverified.

### 12.3 Merchant Offer

Merchant validation belongs to the Merchant domain and follows the separate
Merchant implementation plan. It must validate:

- target market;
- content language;
- currency;
- stable offer ID;
- public HTTPS landing URL;
- price and availability;
- brand and identifier status;
- shipping and return policy;
- publication approval.

## 13. Implementation Roadmap

### Phase A: Architecture Baseline

Status: complete for the current pre-launch architecture.

- Keep SEO as a top-level administration domain.
- Keep Home, Articles, and Products as child routes.
- Keep Google Merchant as a separate top-level channel.
- Keep Product and Article resource fields in their source aggregates.
- Keep media metadata in the media/content domain.
- Keep the SEO API as a facade, not a duplicate database.

### Phase B: Storefront SEO Contract

Status: complete for the current storefront contract.

Initial implementation started in the dedicated storefront SEO utility
boundary:

- `nuxt-i18n/app/utils/seo/types.ts`
- `nuxt-i18n/app/utils/seo/urls.ts`
- `nuxt-i18n/app/utils/seo/product.ts`

The first slice now:

- generates Product canonical URLs through Nuxt's locale-aware `localePath`;
- keeps Product Meta fallback and JSON-LD construction outside the page file;
- uses the visible product name for Schema `Product.name`;
- includes the selected variant SKU or product SKU when available;
- uses the page's display price and currency for the Offer;
- keeps Product JSON-LD free of unverified brand, GTIN, MPN, and review claims.
- uses the approved public site `brandTitle` as the Product JSON-LD brand when
  configured;
- resolves Product detail data by requested locale;
- exposes active Product translation routes through a typed public projection;
- validates Product translation parent relationships at the Product write
  boundary;
- centralizes canonical and hreflang generation for the default and products
  layouts;
- emits JSON-LD with an SSR-safe Unhead `textContent` script contract.

Verification is provided by:

- `nuxt-i18n/scripts/seo/test-product-output.ts`;
- `nuxt-i18n/scripts/seo/check-product-ssr.ts`;
- the product API and SSR-safe JSON-LD tests.

The SSR check requires a running public product URL through `SEO_PRODUCT_URL`;
that is an environment verification step, not a second renderer. Set
`SEO_EXPECT_JSON_LD=false` when verifying the intentional no-image case; the
check then requires the initial HTML to omit Product JSON-LD.

### Phase C: Variant Structured Data

Status: complete for the current variant landing model.

- Define stable variant landing URLs.
- Define ProductGroup identity.
- Map variant SKU, price, availability, and media.
- Add tests for default and non-default variants.
- Do not emit variant claims that cannot be reproduced from a public URL.

The current implementation emits `Product` for one purchasable configuration
and `ProductGroup` with `hasVariant` for two or more active variants. Variant
landing URLs use the stable `/shop/:slug?variant=:id` contract.

### Phase D: SEO Quality and Diagnostics

Status: complete for the current resource model.

- Show fallback state in the SEO editor.
- Add route and locale validation.
- Add structured-data preview or validation output.
- Add duplicate canonical detection.
- Add missing-image, missing-SKU, and missing-price diagnostics.
- Add audit events for SEO changes.

The current route model derives canonical paths from database-backed resource
routes. A separate duplicate-canonical scanner is intentionally not introduced:
Home, Blog, and Shop occupy distinct route namespaces, while resource route
projections and locale/slug constraints prevent the current classes of
duplicates. If arbitrary canonical overrides or additional page types are
introduced, a dedicated scanner becomes a required follow-up.

### Phase E: Merchant Automation

Status: code-complete for local asynchronous automation; real Google account
and production-market verification remain external acceptance steps.

- Keep manual validation and synchronization until production data is proven.
- Product source changes publish typed outbox events.
- The existing outbox worker provides retry, dead-letter, and bounded backoff.
- Product upsert, product withdrawal, offer revalidation, and full reconciliation
  are implemented in the separate Google Merchant domain.
- Do not make product-editor saves call Google synchronously.
- Keep Merchant issue history and publication decisions separate from SEO Meta
  history.

The local implementation does not claim that a product is approved by Google.
OAuth, website claim, target-market policy approval, live API processing, and
Google issue resolution still require a real Merchant Center environment.

## 14. Required Test Matrix

The SEO system should be tested with:

- English product;
- non-English product;
- product with no custom Meta Title;
- product with no custom Meta Description;
- product with no image;
- product with SKU;
- product without SKU;
- product with one variant;
- product with multiple variants;
- product with sale price;
- out-of-stock product;
- inactive product;
- query parameter used for UI state;
- localized route with hreflang;
- slug change;
- image replacement;
- Meta-only update;
- Merchant-only update;
- Merchant offer withdrawal.

For every case, verify both:

1. the visible storefront page;
2. the initial HTML and JSON-LD emitted to crawlers.

## 15. Current Code Map

| Responsibility | Current code |
| --- | --- |
| SEO navigation | `go-backend/web/admin/src/lib/adminNavigation.ts` |
| SEO routes | `go-backend/web/admin/src/router/index.ts` |
| SEO route projection | `go-backend/internal/domain/seo/route.go` |
| Public article route projection | `go-backend/internal/domain/post/translation.go`, `go-backend/internal/repository/post_translation_repository.go` |
| Public article response | `go-backend/internal/api/v1/content/public_response.go` |
| Blog detail data contract | `nuxt-i18n/app/composables/blog/useBlogPostDetail.ts` |
| Blog SEO renderer | `nuxt-i18n/app/composables/seo/useBlogPostSeo.ts` |
| Blog detail presentation | `nuxt-i18n/app/components/blog/BlogPostDetailContent.vue` |
| SEO route display helpers | `go-backend/web/admin/src/modules/seo/routes.ts` |
| Home SEO view | `go-backend/web/admin/src/views/seo/Home.vue` |
| Article SEO view | `go-backend/web/admin/src/views/seo/Articles.vue` |
| Product SEO view | `go-backend/web/admin/src/views/seo/Products.vue` |
| SEO editor dialog | `go-backend/web/admin/src/components/admin/seo/SEOResourceEditorDialog.vue` |
| SEO backend contracts | `go-backend/internal/domain/seo/` |
| SEO resource facade | `go-backend/internal/service/seo_resource_service.go` |
| Product public response | `go-backend/internal/api/v1/product/public_response.go` |
| Product translation route contract | `go-backend/internal/domain/product/translation.go` |
| Product translation query | `go-backend/internal/repository/product_translation_repository.go` |
| Product translation validation | `go-backend/internal/service/product_translation_validation.go` |
| Product storefront renderer | `nuxt-i18n/app/pages/shop/[slug].vue` |
| Product SEO output types | `nuxt-i18n/app/utils/seo/types.ts` |
| Product SEO output builder | `nuxt-i18n/app/utils/seo/product.ts` |
| Storefront SEO URL helpers | `nuxt-i18n/app/utils/seo/urls.ts` |
| Storefront canonical/hreflang composable | `nuxt-i18n/app/composables/seo/useStorefrontSeoLinks.ts` |
| Product SEO-only write boundary | `go-backend/internal/service/product_seo_update.go` |
| Merchant outbox event contract | `go-backend/internal/domain/outbox/event.go` |
| Merchant outbox publisher | `go-backend/internal/service/merchant_outbox_publisher.go` |
| Product-to-Merchant event bridge | `go-backend/internal/service/product_merchant_events.go`, `go-backend/internal/service/product_admin_service.go` |
| Merchant outbox handler and reconciliation | `go-backend/internal/service/google_merchant_outbox.go` |
| Merchant reconciliation endpoint | `go-backend/internal/api/admin/google_merchant_handler.go`, `go-backend/internal/api/admin/router.go` |
| Global locale links | `nuxt-i18n/app/layouts/default.vue` |
| Product layout locale links | `nuxt-i18n/app/layouts/products.vue` |
| Merchant channel UI | `go-backend/web/admin/src/views/GoogleMerchant.vue` |
| Merchant channel backend | `go-backend/internal/service/google_merchant_service.go` |
| Merchant channel plan | `go-backend/docs/GOOGLE_MERCHANT_CENTER_IMPLEMENTATION_PLAN.md` |

## 16. Non-Goals

This architecture does not require:

- a generic SEO editor for every arbitrary HTML element;
- manual JSON-LD editing;
- product image metadata duplication in SEO;
- `meta keywords` as a new ranking feature;
- putting Google Merchant fields into the Product editor;
- using Google Merchant as a catalog database;
- synchronous Google API calls inside normal catalog or SEO saves.
