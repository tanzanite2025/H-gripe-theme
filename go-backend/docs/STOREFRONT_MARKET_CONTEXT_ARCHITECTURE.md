# Storefront Market Context Architecture

This document defines the long-term boundary for product pricing, exchange-rate
assisted display prices, country, locale, checkout, and payment.

## Principles

- Product, SKU, shipping, tax threshold, coupon threshold, and other commercial
  amounts are configured in the backend primary pricing currency. If the store's
  primary pricing currency is CNY, admin-entered product and shipping prices are
  CNY.
- ExchangeRate-API is a backend pricing/display helper. It keeps secondary
  display currency values in sync from cached rates. It is not a payment method
  selector, and Nuxt must not call the third-party exchange-rate provider
  directly.
- Display currency is not an order or provider-settlement currency. Display
  prices are labels for browsing; checkout, payment, refund, and provider
  callbacks use the product or variant order currency and amount.
- Payment buttons come from configured and enabled payment methods. If PayPal is
  enabled in admin, Nuxt can show PayPal. If WeChat Pay is enabled in admin,
  Nuxt can show WeChat Pay. IP, language, display currency, and exchange-rate
  conversion must not decide which buttons are exposed.
- Country is not language. Country can seed the locale resolver, but the user can
  explicitly choose a language. If no localized content exists, fall back to
  English.
- Product master data is not duplicated by country. Translations, optional
  market-specific overrides, and storefront visibility live outside the core
  product record.

## Pricing Currency Model

The backend should expose a dedicated currency/pricing management area with:

- Primary pricing currency: the currency used when admin users enter product,
  variant, shipping, tax, coupon, and order-threshold amounts.
- Secondary display currencies: currencies shown to shoppers as converted
  labels.
- Exchange-rate provider settings: API key, provider endpoint, sync cadence, and
  last successful sync metadata.
- Cached exchange rates: backend-owned rates refreshed by manual sync or a daily
  scheduler.

The currency/FX center owns the primary pricing currency, provider settings, and
cached-rate status. The storefront market tab owns the secondary display
currencies through each market's `default_currency` and `display_currencies`.
`ExchangeRateService.GetConfig()` combines those two domains into one API base
and one set of quote targets. The API management card stores provider enable/API
key settings and can trigger cache sync; it does not define another currency list.

Example:

- Primary pricing currency: `CNY`
- Secondary display currencies: `USD`, `EUR`, `GBP`
- Product admin enters: `699 CNY`
- Backend cached rates produce display labels such as `USD 96.80`, `EUR 89.10`,
  and `GBP 76.20`

These secondary values are for display. They do not mean PayPal, Stripe, or
WeChat Pay is selected, and they do not describe the customer's card/account
funding currency.

## Admin Pricing UX

When editing a product, SKU, or shipping rule, the admin UI should make the
pricing boundary obvious:

- The main editable amount is the primary pricing currency amount. Admin users
  type the primary amount directly, for example `699 CNY`.
- Secondary display currency fields are filled or refreshed by a one-click
  action that reads the backend ExchangeRate-API cache.
- This should be an inline fill/refresh action beside the secondary display
  price fields, not a help-icon popup workflow.
- The same one-click fill rule applies to product, SKU, and shipping pricing
  surfaces.
- Product and SKU secondary display prices are materialized backend data for
  storefront display. Nuxt reads them from product API responses. Nuxt should
  not separately read ExchangeRate-API cache just to price a product card.
  If a requested display currency has no stored snapshot, the public product API
  should leave that display price absent instead of converting it at request
  time.
- The daily/manual exchange-rate sync rebuilds product and SKU snapshots from
  current source amounts. It only refreshes rows whose source currency matches
  the current backend entry currency.
- If the backend entry currency changes, existing product and SKU source amounts
  are preserved. Rows in the old currency are reported by the backend currency
  audit and are not silently reinterpreted as the new currency.
- Shipping uses the same one-click conversion helper in admin and persists
  secondary display prices by exact money field. Template snapshots are keyed by
  `default_fee` and `free_threshold`; rule snapshots are keyed by `fee`,
  `additional`, and, for price-based templates only, `min_value` and
  `max_value`.

Target contract: primary pricing amounts are the commercial truth; secondary
display prices are backend-filled snapshots derived from cached rates.

## Locale Boundary

Locale resolution is separate from pricing and payment:

1. Explicit request locale, such as `?locale=de` or `X-Locale`.
2. Existing locale cookie or user profile preference.
3. Browser `Accept-Language`.
4. Market default locale inferred from country/IP.
5. English fallback.

Locale does not change the product's primary order price, shipping price, or
payment button list.

## Display Currency Resolution

Display currency is resolved for browsing labels only:

1. Explicit request currency, such as `?currency=EUR` or `X-Display-Currency`.
2. Existing currency cookie or user profile preference.
3. Market default display currency.
4. First configured secondary display currency.
5. Primary pricing currency fallback.

Display currency affects product cards, product detail labels, cart labels, and
other browsing UI. It does not select payment methods and it does not change the
final order amount unless an explicit backend pricing override says so.

## API Contract

The stable public context entry point is:

```http
GET /api/v1/storefront/context
```

Response shape:

```json
{
  "data": {
    "country": {"code": "DE", "source": "cf_ip_country"},
    "market": {
      "code": "EU",
      "default_locale": "de",
      "supported_locales": ["en", "de", "fr", "es", "it", "nl"],
      "default_currency": "EUR",
      "display_currencies": ["EUR", "USD", "GBP"]
    },
    "locale": {"requested": "de", "resolved": "de", "fallback": "en"},
    "currency": {"requested": "EUR", "resolved": "EUR", "base": "CNY"}
  }
}
```

Product browsing APIs also resolve the same context. List responses include a
top-level `context` sibling, and single-product responses keep `data` as the
product while adding a top-level `context` sibling.

Current product and variant responses include one resolved display price:

```json
"display_price": {
  "amount": 89.10,
  "currency": "EUR",
  "rate": 0.1275,
  "source": "direct_rate",
  "converted": true
}
```

`display_price` is a storefront label only. The original `currency`, `price`,
and `sale_price` fields remain the catalog truth used by cart, checkout, order,
payment, and refund. If a cached snapshot is missing, `display_price` is absent;
the source currency and source amount remain available in the catalog fields.

Product and variant responses also expose a `display_prices` list for all
filled secondary currencies. That lets Nuxt render a currency dropdown from the
product API response without separately fetching exchange-rate cache data.

Shipping template and rule responses expose `display_price_snapshots` objects
keyed by the money field. Weight and quantity rule thresholds are not currency
amounts, so their `min_value` and `max_value` snapshots are not persisted unless
the template type is `price`.

Shipping quote responses may expose `display_price` and `display_prices` derived
only from those stored shipping snapshots. The actual `shipping_fee` and
`currency` remain the checkout/order truth. If a selected shipping option uses a
configured fee component without a complete stored display snapshot, the quote
omits the display price instead of calling exchange rates at request time.

Display currency never controls payment-method exposure. PayPal, Stripe, Alipay,
WeChat, bank transfer, and similar buttons are shown from the configured
payment-method list plus risk/protection checks. Provider order currency
validation belongs only to provider order creation/runtime protection.

The public cached-rate endpoint is:

```http
GET /api/v1/currency/exchange-rates?base=CNY
```

Nuxt reads this backend endpoint only for metadata or tooling screens. Normal
product browsing should read display prices from product/shipping APIs. Nuxt
must not call the third-party exchange-rate provider directly.

## Admin Market Module

The admin menu label is "市场与本地化语种". Market management is its own admin
module, not a pile of generic settings. The database owns market code, enabled
countries, default locale, supported locales, default display currency, display
currencies, logistics policy, tax policy, enabled state, and priority.

Admin endpoints:

```http
GET    /api/admin/storefront/markets
GET    /api/admin/storefront/markets/options
GET    /api/admin/storefront/markets/:id
POST   /api/admin/storefront/markets
PUT    /api/admin/storefront/markets/:id
DELETE /api/admin/storefront/markets/:id
```

The public context endpoint keeps the same response contract whether market data
comes from the database or the built-in bootstrap fallback.

## Exchange-Rate Module

Exchange-rate settings live under the admin API settings module because the
provider key and endpoint are operational credentials, not market definitions.
The implementation supports manual sync through the admin panel/admin API and a
daily backend scheduler:

```http
GET  /api/admin/settings/exchange-rates
POST /api/admin/settings/exchange-rates/sync
```

The sync endpoint stores rates in `currency_exchange_rates`. Admin pricing
helpers and storefront product/shipping responses read this cached table through
Go services. The scheduler calls the same sync service at startup and then at
`worker.exchange_rate_sync_interval_seconds`, so manual and scheduled refreshes
share one code path and one audit trail.

Admin pricing forms use `POST /api/admin/pricing/exchange-rates/convert` to
fill inline secondary display price previews from the cached rates. This helper
does not select payment methods and does not call the third-party provider.

The currency policy is stored as:

- `currency_primary_currency`: the backend entry currency for product/SKU source
  amounts entered by admins.
- `currency_display_currencies`: legacy compatibility field only. Storefront
  display currencies no longer come from this global setting.

Storefront display currencies belong to the market/localization TAB. Each
enabled market contributes its `default_currency` and `display_currencies` to
the exchange-rate target list. Exchange-rate sync uses
`currency_primary_currency` as the API base currency and those enabled market
currencies as quote targets.

Changing the backend entry currency does not silently rewrite historical product
or SKU amounts. The admin currency policy page exposes
`GET /api/admin/settings/currency-policy/audit` to detect products/SKUs whose
saved source currency differs from the current backend entry currency. See
`docs/CURRENCY_EXCHANGE_OPERATING_MODEL.md` for the maintenance rules and
function boundaries.

## Future Product Modules

Product administration should keep master product fields separate from:

- `product_translations` for locale-specific content and SEO.
- `market_price_overrides` for optional localized pricing.
- `market_product_visibility` for regional availability.

Exchange-rate configuration should remain a backend API settings module. Go
syncs provider rates into storage/cache and storefront APIs read the cached
values only.

## Payment Boundary

Payment button exposure comes from configured and enabled payment methods. Market
context, display currency, and ExchangeRate-API data must not hide or show
PayPal, Stripe, Alipay, or WeChat Pay buttons.

Binding a provider is the button policy. If PayPal, Stripe, Alipay, or WeChat
Pay is configured and enabled, Nuxt can expose that button. The provider then
handles whether the customer's chosen card/account/funding currency can complete
the payment under that provider account's own rules.

Backend order/payment creation still uses the local order amount and order
currency. Provider callbacks/webhooks must verify signature, idempotency, amount,
and currency against that local order before marking it paid. This validation is
part of the real payment flow, not product pricing, display currency, IP market,
or ExchangeRate-API logic.

The current implementation intentionally does not auto-refund, auto-switch
providers, or hard-close checkout from market context. Those actions belong to
payment risk controls and must remain explicit, auditable, and reversible.
