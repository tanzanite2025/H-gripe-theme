# Google Merchant Center Integration Plan

The cross-domain SEO and storefront rendering boundary is documented in
[`../../docs/seo/SEO_SYSTEM_ARCHITECTURE.md`](../../docs/seo/SEO_SYSTEM_ARCHITECTURE.md).
This document owns the Google Merchant channel implementation only.

## Purpose

Integrate Commerce Platform products with Google Merchant Center (GMC) Free Listings in
controlled stages. Product data will be published from the Go backend to
Google, while product clicks continue to land on the Commerce Platform storefront.

This document is the implementation checklist and architectural decision record
for the integration. Do not start a later phase until the previous phase meets
its exit criteria.

## Current Decision

Use the Google Merchant API as the production synchronization channel.

- Do not start new work on Content API for Shopping v2.1. Google has deprecated
  it and states that it will shut down in August 2026.
- Use the Merchant API Products sub-API for product, price, stock, and
  availability changes.
- Treat a scheduled file feed as an optional bulk-import and recovery channel,
  not the long-term source of truth for frequent updates.
- Treat Google's Merchant Data MCP service as an Alpha diagnostic and
  low-risk data-source tool. It must not be the production product-write path.
- Do not build a custom write-capable MCP server in the first implementation.

## Current Implementation Boundary

The channel workspace is now separated from the core Commerce Platform catalog:

- Google-specific fields are stored in `google_merchant_offers`, not in the
  product or variant editor.
- The administration panel has a dedicated Google Merchant page. Catalog
  operators enter it from a per-product "Sync to Google" action or directly
  from the catalog domain.
- A single store-level `google_merchant_connections` record holds the selected
  Merchant Account ID, Data Source ID, Storefront Base URL, connection health,
  and encrypted OAuth refresh token.
- The browser only receives a masked connection status, selected Google
  account email, and non-secret account identifiers. OAuth client secrets and
  refresh tokens remain server-only.
- The backend uses one-time, time-bounded OAuth state values and stores only
  their SHA-256 hashes. A callback cannot reuse a consumed state value.
- OAuth is configured through `GOOGLE_MERCHANT_*` server environment
  variables. The redirect URI must exactly match the URI registered in Google
  Cloud.
- After connection, the page can read the Google Merchant `products` resource
  and display remote offer IDs, titles, prices, availability, destination
  status, and item-level issues without importing those records into the
  storefront catalog.
- A validated, ready offer can be manually submitted with Merchant API
  `productInputs:insert` through the backend. The API call is scoped to one
  selected SKU, uses the configured Merchant Account ID and Data Source ID,
  builds landing/media URLs from the configured Storefront Base URL, and
  records `sync_status`, `last_sync_at`, and `last_error` on the channel offer
  only.
- Product create/update/delete/status and variant/media source changes publish
  typed Merchant outbox events after the catalog mutation. The outbox worker
  dispatches upsert, withdrawal, and offer revalidation handlers with retry and
  dead-letter behavior.
- The dedicated Merchant page exposes a permission-protected full
  reconciliation command at `POST /api/admin/google-merchant/reconcile`.

This boundary does not claim external Google approval. The local queue and
withdrawal behavior are implemented, while real Merchant Center processing,
policy approval, and production monitoring remain external acceptance work.

## Non-Goals for the First Release

- Google Ads campaign management.
- Automatic bid or budget changes.
- AI agents that can directly publish, modify, or remove products.
- Multi-channel product syndication beyond Google Merchant Center.
- Automatic remediation of policy violations without administrator review.

## Existing Project Foundation

The catalog already provides much of the basic data needed for a product offer:

- Product SKU, slug, title, short description, full description, price, sale
  price, stock, status, locale, and media.
- Variant SKU, title, option values, price, sale price, stock, active status,
  and weight.
- Product or variant shipping-template assignment.
- Storefront product pages with canonical URLs, Product JSON-LD containing an
  offer, price currency, and availability. The target URL architecture uses
  the flat `/products/:slug` permalink; variant query support is UI state unless
  a separately approved variant landing-page strategy makes it canonical.
- Currency settings, shipping templates, shipping rules, shipping zones, and
  public refund/return, terms, privacy, and contact pages.

The existing Product JSON-LD is helpful for SEO. It is not a Merchant Center
product source and it does not replace Merchant API synchronization.

## Gaps Before Synchronization

The current product model does not expose compliance-critical GMC fields as
first-class validated data:

- Brand.
- GTIN and MPN, with an explicit identifier-exists decision for genuinely
  exempt products.
- Product condition.
- Google product category.
- Target country, content language, and feed label.
- A publication switch separate from general storefront product status.
- Broader per-variant landing-page UX beyond the current `?variant=` selection
  support.

The system still lacks a full external synchronization history:

- Google product resource name beyond the stable local offer ID.
- Product content hash or source version.
- Last API response, rejection reason, and issue severity history.
- Retry count and next retry time.

## Proposed Data Ownership

Commerce Platform remains the source of truth for all catalog, price, inventory,
shipping, and storefront URL data.

Google Merchant Center is the distribution and diagnostics system. Google must
never become the primary source for Commerce Platform product descriptions, pricing, or
stock.

One Commerce Platform product or variant must map deterministically to one external
offer ID. The offer ID must be stable and non-sequential from a public business
perspective; changing a title or price must not create a new Google offer.

## Recommended Architecture

### Catalog Mapping Layer

Create a dedicated Merchant catalog mapper in the Go backend. It receives a
published product or variant and produces a validated Merchant API
`ProductInput`.

The mapper must:

- Select the correct SKU and stable offer ID.
- Select a canonical public HTTPS landing URL.
- Select visible product images only.
- Emit a fixed price and ISO currency for the target country.
- Map stock state to Google's availability values.
- Map sale price only when it is active and consistent with the storefront.
- Map shipping and return behavior from the actual configured policies.
- Reject incomplete or contradictory data before any Google API call.

### Synchronization Queue

Use an asynchronous, idempotent job rather than calling Google inside the
administrator's save request.

Each product mutation that affects Google data should enqueue a sync request
after the database transaction succeeds. A job must be safe to run multiple
times and must record its outcome in the synchronization record.

Required job types:

- Upsert offer.
- Delete or withdraw offer.
- Refresh offer status and issues.
- Retry transient failure.
- Reconcile Commerce Platform offers against Google on a scheduled basis.

### Administrator Surface

The administration panel needs a dedicated Google Merchant Center domain,
rather than scattering fields across general product settings.

Minimum views:

- Merchant connection and account health.
- Product readiness checklist before publication.
- Per-product and per-variant publication toggle.
- Sync state, last sync time, external offer ID, and response summary.
- Google-reported errors and warnings with direct field attribution.
- Retry and revalidate commands, protected by administrator permissions.

## Implementation Phases

### Phase 0: Decoupled Channel Workspace

Owner: backend and administration panel.

Status: complete in code.

Tasks:

- Do not add Google-specific fields to the product editor, product table, or
  product-variant table.
- Add a dedicated Google Merchant administration page.
- Let an administrator select individual existing products and purchasable
  variants from that page.
- Store channel fields in a separate Google offer record.
- Validate the channel offer before any external synchronization.

Exit criteria:

- The normal product editor remains unchanged by Merchant-specific fields.
- An administrator can select one SKU and save its independent Google offer.
- An incomplete channel offer cannot be marked ready.
- Removing a channel offer does not modify the storefront product.

### Phase 1: Business and Google Account Readiness

Owner: business administrator.

Status: pending external business and Google environment work.

Tasks:

- Create or confirm the Merchant Center account.
- Complete website verification and claim the production domain.
- Create a Google Cloud project.
- Complete Merchant Center developer registration for the Cloud project.
- Decide the first target country, content language, and checkout currency.
- Confirm the public contact, shipping, return/refund, terms, privacy, and
  checkout pages reflect real operating policies.
- Decide whether every initial product has a GTIN/MPN or qualifies for a
  documented identifier exemption.
- Confirm that Google can access storefront product pages without login,
  geofence, bot challenge, or accidental commercial-crawler blocking.

Exit criteria:

- A real Merchant Center account is connected to the production domain.
- One target market and one authoritative currency are documented.
- The owner signs off on the legal, delivery, and return information.
- A test user or service account can authenticate without credentials stored in
  Git.

### Phase 2: Channel Connection and Offer Configuration

Owner: backend and administration panel.

Status: complete in code; production account selection remains environment
dependent.

Tasks:

- Add a Google OAuth connection area on the dedicated page.
- Bind the Google account and Merchant account/data source without exposing
  credentials to the browser.
- Display selected local offers separately from remote Merchant Center
  products.
- Define a stable offer-ID format and keep one record per purchasable variant.
- Add target country, content language, currency, market price, shipping,
  returns, and publication data to the channel workflow.
- Add connection, synchronization, and issue-history tables.

Exit criteria:

- One selected SKU can be configured without changing the global product
  record.
- The Google connection page can show connection health and selected offers.
- No incomplete channel offer can be sent to Google.

### Phase 3: Merchant API Connection and Dry Run

Owner: backend.

Status: code-complete; live dry-run exit criteria remain pending a real Google
Merchant account.

Tasks:

- Add Merchant API authentication configuration through environment variables
  or a secret manager.
- Add the official Go Merchant API client dependency.
- Implement account connectivity checks.
- Implement mapper preview mode without writing to Google.
- Submit three approved test offers through the Merchant API.
- Retrieve product status and issues after processing.
- Verify delete/withdraw behavior in a test environment.

Exit criteria:

- The test offers appear in Merchant Center with correct title, image, price,
  currency, availability, URL, and shipping behavior.
- The recorded Google offer IDs and issue reports are visible in the
  administration panel.
- A failed call produces an actionable error without exposing secrets.

### Phase 4: Production Incremental Synchronization

Owner: backend and operations.

Status: local automation code-complete; production monitoring and live
acceptance remain pending.

Tasks:

- Add idempotent queued upsert, withdraw, retry, and reconciliation jobs.
- Enqueue sync after approved product, variant, price, stock, image, shipping,
  and publication-status changes.
- Add exponential backoff for transient Google API failures.
- Add rate limiting and bounded retry policies.
- Add audit logging for manual retries and publication changes.
- Add alerts for sustained synchronization failures and high issue counts.

Exit criteria:

- Repeating a sync job does not create duplicate offers.
- Stock and price changes are reflected in Google without manual intervention.
- Failed products remain visible to administrators with a clear remediation
  path.
- Production monitoring shows the health of the integration.

Implemented local pieces:

- `internal/domain/outbox/event.go` defines typed Merchant event payloads.
- `internal/service/merchant_outbox_publisher.go` publishes source-change
  events.
- `internal/service/product_merchant_events.go` and
  `internal/service/product_admin_service.go` bridge catalog mutations.
- `internal/service/google_merchant_outbox.go` handles upsert, withdrawal,
  revalidation, and full reconciliation.
- `internal/api/admin/google_merchant_handler.go` exposes the protected
  reconciliation command.

The remaining exit criteria require a real account, approved market, live API
responses, and operational monitoring. They cannot be honestly completed by
local code alone.

### Phase 5: Optional File Data Source

Owner: backend and operations.

Tasks:

- Generate a validated file feed only if a bulk-import or recovery workflow is
  needed.
- Register it as a Merchant Center data source.
- Give Google a stable, HTTPS-accessible fetch location.
- Restrict the file to the minimum necessary merchant attributes.
- Verify that CDN, WAF, robots configuration, and backend middleware allow
  Google to fetch the file.
- Define precedence rules so the file source and API source cannot overwrite
  the same fields unpredictably.

Exit criteria:

- Google can fetch and process the source repeatedly.
- The file is not treated as an unauthenticated public catalog API.
- Data-source reports match the Go mapper output.

### Phase 6: MCP for Diagnostics Only

Owner: operations.

Tasks:

- Evaluate the official Merchant Data MCP service in a non-production or
  operator-only environment.
- Permit read-only catalog status, issue inspection, and reporting use cases.
- Do not grant AI agents direct authority to change product data.
- Require administrator confirmation for any supported data-source fetch
  action.

Exit criteria:

- MCP access cannot disclose credentials or change Commerce Platform catalog data.
- Operators can use it to explain why a product is not approved or visible.

## Critical Consistency Rules

- Google price, sale price, currency, availability, image, and landing-page
  variant must match the customer-visible storefront at crawl time.
- Do not use a converted display currency in Google unless the landing page and
  checkout use that same authoritative currency for the target market.
- Do not expose internal stock counts, sales metrics, order IDs, or private
  product administration endpoints through a feed.
- Do not assign a product to Google solely because it is storefront-active.
  Merchant publication requires separate approval.
- Do not silently replace a rejected product. Preserve its issue history and
  fix the mapped field or policy cause.
- Do not place service-account JSON, OAuth refresh tokens, or Merchant account
  secrets in source control, browser code, or frontend configuration.

## Known Risks

| Risk | Why it matters | Required control |
| --- | --- | --- |
| Price or currency mismatch | Google can disapprove an offer or show an incorrect price. | Use one authoritative target-market price and verify against the landing page. |
| Variant mismatch | A shopper may land on a different price or configuration. | Give each purchasable variant a deterministic offer ID and landing state. |
| Missing identifiers | Products may be rejected or have reduced data quality. | Store GTIN/MPN/brand explicitly and document true exemptions. |
| Inaccurate delivery or return details | This is a merchant policy and customer-trust risk. | Publish only policies confirmed by the business owner and map actual shipping rules. |
| Google processing delay | API success is not immediate listing eligibility. | Track processing status and issues asynchronously. |
| Duplicate or conflicting sources | Feed and API updates can overwrite each other. | Define one primary source and explicit precedence before enabling a second source. |
| Credential leakage | Merchant access can modify catalog distribution. | Use server-side secrets, least privilege, rotation, and audit logs. |
| Anti-bot interference | Google may be unable to load a landing page or feed. | Test production access and explicitly exclude approved Google workflows from commercial-crawler rules. |

## Required Test Matrix

Before production enablement, test at least:

- A simple in-stock product.
- An out-of-stock product.
- A discounted product with sale price.
- A product with multiple purchasable variants.
- A product without a GTIN but with a documented exemption.
- A product with region-specific shipping.
- An image replacement.
- A price increase and a stock transition.
- Product withdrawal and re-publication.
- Authentication failure, API timeout, API validation failure, and retry.

## Official References

- Merchant API overview:
  https://developers.google.com/merchant/api/overview
- Merchant API product inputs:
  https://developers.google.com/merchant/api/reference/rest/products_v1beta/accounts.productInputs
- Merchant API data sources:
  https://developers.google.com/merchant/api/guides/data-sources/overview
- Merchant Data MCP access service:
  https://developers.google.com/merchant/api/guides/agentic-tools/merchant-data-mcp
- Content API release notes:
  https://developers.google.com/shopping-content/guides/rel-notes
- Google Merchant Center free listing policies:
  https://support.google.com/merchants/answer/13889434
