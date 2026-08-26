# Google URL Notification

This project provides one narrow admin action: select one published product
and send one `URL_UPDATED` notification for that product's public URL.

The action is a notification only. A successful response means the upstream
request was accepted; it does not report Google's later crawl or indexing
result.

## Scope contract

The following boundaries are intentional and must not be expanded as part of
this feature:

- The admin selects one product from the SEO product list.
- The backend builds that product's public localized URL.
- The backend sends one `URL_UPDATED` request containing that URL.
- Sitemap generation and updates remain a separate existing workflow.
- The button does not redirect to or submit a sitemap.
- Product save and publish transactions do not call Google automatically.
- There is no batch push, delete notification, crawl-status query, or
  indexing-status guarantee in this feature.

The request is allowed only for products with `status=active`. The API
credentials and HTTP request are implementation details of this one action,
not a second product publishing pipeline.

## Protection rules

This is a synchronous single-request action, so it does not use the existing
Outbox worker or a queue.

- The admin API requires an `Idempotency-Key` header. The admin UI generates
  one key per notification operation; a transport retry must reuse that key,
  and the same key replays the original response without calling Google again.
- Redis applies a shared per-admin limit of one notification request per
  minute. A blocked request returns HTTP `429` and a `Retry-After` header.
- Redis also keeps a shared cooldown for the exact public URL for 15 minutes.
  This prevents a second admin, a second browser tab, or a new idempotency key
  from submitting the same URL again during that period.
- After product validation, the URL cooldown is reserved before any Google
  upstream request. A normal token-acquisition failure releases the reservation
  so the operation can be retried; once the publish request has started, the
  cooldown is retained even if the upstream response is an error. This avoids
  an uncertain network retry sending a duplicate notification.
- Redis protection fails closed. If Redis is unavailable, the backend does not
  call Google and returns HTTP `503`.

These controls protect duplicate submissions only. They do not query Google's
later crawl or indexing state.

## Google Cloud setup

1. Create or select the Google Cloud project used for this site.
2. Enable the **Web Search Indexing API**.
3. Create a service account and download its JSON key, or mount the JSON key as
   a secret file on the backend host.
4. Add the service account email as an owner or delegated owner of the
   corresponding property in Google Search Console.
5. Request the required Indexing API quota if the project needs more than the
   default onboarding/test allowance.
6. Make sure the configured storefront base URL is the verified Search Console
   property and is publicly reachable.

Do not commit the service-account JSON or private key to the repository.

## Backend configuration

The feature is disabled by default:

```env
GOOGLE_INDEXING_ENABLED=false
GOOGLE_INDEXING_SERVICE_ACCOUNT_FILE=/run/secrets/google-indexing-service-account.json
GOOGLE_INDEXING_SERVICE_ACCOUNT_JSON=
GOOGLE_INDEXING_REQUEST_TIMEOUT_SECONDS=15
```

Use exactly one of `GOOGLE_INDEXING_SERVICE_ACCOUNT_FILE` and
`GOOGLE_INDEXING_SERVICE_ACCOUNT_JSON`. A mounted file is preferred for
production deployments. Set `GOOGLE_INDEXING_ENABLED=true` only after the
service account and Search Console access have been configured.

The same values can be supplied through the `google_indexing` section in the
backend YAML configuration. Environment variables are bound by the backend
configuration loader.

## Admin workflow

1. Open **SEO / 产品** in the admin panel.
2. Confirm the Google notification status is **已就绪**.
3. Click **通知 Google** for one active product.
4. The backend builds the localized public URL and sends one `URL_UPDATED`
   request using the service account.

Only products with `status=active` are accepted. Draft and inactive products
are rejected by the backend before the external request is made.

Each attempt handled by the backend is written to the SEO audit log with
action `indexing_notify`, product ID, public URL when available,
`URL_UPDATED`, HTTP status when available, success/failure status, and the
error message when applicable. Service-account private keys and access tokens
are never written to the audit record.

## API endpoints

The admin API exposes:

- `GET /api/admin/seo/indexing/status`
- `POST /api/admin/seo/products/:id/indexing`

The status endpoint requires `seo:view`. The push endpoint requires
`seo:edit` and an `Idempotency-Key` header. The push endpoint returns HTTP
`409` with `Retry-After` when the same public URL is still in its cooldown.

## Official references

- [Indexing API quickstart](https://developers.google.com/search/apis/indexing-api/v3/quickstart)
- [Indexing API reference](https://developers.google.com/search/apis/indexing-api/v3/reference)
