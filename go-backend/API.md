# Go Backend API Documentation

## Base URL
```
http://localhost:9200/api/v1
```

## Version
Current API Version: **1.4.0**

## Authentication

Most endpoints use HttpOnly Cookie authentication. Browser clients must send credentials, and unsafe methods must include the CSRF header:
```
credentials: include
X-CSRF-Token: <csrf_token_cookie_value>
```

---

## Authentication Endpoints

### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "password123"
}
```

**Response:**
```json
{
  "message": "user registered successfully",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "johndoe",
    "role": "user",
    "created_at": "2026-05-25T10:00:00Z"
  }
}
```

### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email_or_username": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "johndoe",
    "role": "user"
  }
}
```

### Get Profile
```http
GET /api/v1/auth/profile
Cookie: auth_token=<http_only_cookie>
```

---

## Content Endpoints

### List Posts
```http
GET /api/v1/content/posts?page=1&page_size=10&status=published
Accept-Language: en
```

**Query Parameters:**
- `page` (default: 1)
- `page_size` (default: 10, max: 100)
- `status` (default: published) - draft, published, archived

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "title": "Sample Post",
      "slug": "sample-post",
      "excerpt": "This is a sample post",
      "locale": "en",
      "featured_image": "https://...",
      "created_at": "2026-05-25T10:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "page_size": 10,
  "total_pages": 5
}
```

### Get Single Post
```http
GET /api/v1/content/posts/:id
Accept-Language: en
```

Can use either ID or slug:
- `/api/v1/content/posts/1`
- `/api/v1/content/posts/sample-post`

### List FAQs
```http
GET /api/v1/content/faqs?category=general&page=1&page_size=20
Accept-Language: en
```

### Get FAQ Categories
```http
GET /api/v1/content/faq-categories
Accept-Language: en
```

---

## Storefront Context Endpoints

### Resolve Market Context
```http
GET /api/v1/storefront/context?country=DE&locale=de&currency=EUR
CF-IPCountry: DE
Accept-Language: fr-FR,fr;q=0.9
Cookie: locale=de; display_currency=EUR
```

**Resolution rules:**
- Country resolves from explicit `country`, then trusted proxy headers such as `CF-IPCountry`, then `ZZ`.
- Locale resolves from request, cookie, `Accept-Language`, market default, then English.
- Display currency resolves from request, cookie, market default, market display currencies, then the backend entry currency.
- Display currency is only for browsing and conversion labels. Product, shipping, checkout, payment, refund, and callbacks use backend product/variant/order pricing data.
- The currency `base` is the backend primary pricing currency used for admin-entered product, SKU, and shipping amounts. The examples below use `CNY`.

**Response:**
```json
{
  "data": {
    "country": {"code": "DE", "source": "cf_ip_country"},
    "market": {
      "code": "EU",
      "default_locale": "en",
      "supported_locales": ["en", "de", "fr", "es", "it", "nl"],
      "default_currency": "EUR",
      "display_currencies": ["EUR", "USD", "GBP"]
    },
    "locale": {"requested": "de", "resolved": "de", "fallback": "en", "source": "request"},
    "currency": {"requested": "EUR", "resolved": "EUR", "base": "CNY", "source": "request"}
  }
}
```

---

## Product Endpoints

### List Products
```http
GET /api/v1/products?page=1&page_size=12&featured=true&currency=EUR&country=DE
Accept-Language: en
```

**Query Parameters:**
- `page` (default: 1)
- `page_size` (default: 12, max: 24)
- `featured` (boolean)
- `country`, `locale`, `currency` resolve the same storefront context as `/api/v1/storefront/context`.

**Response:**
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "sku": "PROD-001",
      "name": "Product Name",
      "slug": "product-name",
      "currency": "CNY",
      "price": 699,
      "sale_price": 599,
      "display_price": {
        "amount": 76.37,
        "currency": "EUR",
        "rate": 0.1275,
        "source": "direct_rate",
        "converted": true
      },
      "availability": "in_stock",
      "media": [{"media_type": "image", "role": "primary", "url": "https://...", "sort_order": 0, "is_primary": true}],
      "variants": [
        {
          "id": 11,
          "sku": "PROD-001-DEFAULT",
          "title": "Default",
          "currency": "CNY",
          "price": 699,
          "sale_price": 599,
          "display_price": {
            "amount": 76.37,
            "currency": "EUR",
            "rate": 0.1275,
            "source": "direct_rate",
            "converted": true
          },
          "is_default": true,
          "availability": "in_stock"
        }
      ]
    }
  ],
  "context": {
    "country": {"code": "DE", "source": "request"},
    "market": {"code": "EU", "default_locale": "en", "supported_locales": ["en", "de"], "default_currency": "EUR", "display_currencies": ["EUR", "USD"]},
    "locale": {"resolved": "en", "fallback": "en", "source": "fallback"},
    "currency": {"requested": "EUR", "resolved": "EUR", "base": "CNY", "source": "request"}
  },
  "page_size": 12,
  "has_more": false
}
```

`display_price` is only a browsing label generated by the backend from cached
exchange rates. Cart, checkout, order, payment, refund, and gateway callbacks
use the product or variant `currency`, `price`, and `sale_price` fields as the
catalog truth. Normal Nuxt product browsing should read display prices from
product responses rather than calling exchange-rate endpoints to price product
cards itself.

### Get Single Product
```http
GET /api/v1/products/:id?currency=EUR&country=DE
Accept-Language: en
```

Can use either ID or slug.

The response shape is `{ "code": 0, "data": <product>, "context": <storefront_context> }`.

---

## Currency Endpoints

### Get Currency Policy
```http
GET /api/v1/settings/currency-policy
```

`primary_currency` is the backend entry currency for admin-entered product,
SKU, shipping, coupon-threshold, and order-threshold amounts. Product and SKU
`currency`, `price`, and `sale_price` stay in this original currency and are not
rewritten by exchange-rate sync.

`display_currencies` remains in this response as a backward-compatible field,
but global currency policy no longer owns storefront display currencies. The
source of truth for storefront display currencies is the admin "市场与本地化语种"
market TAB: each enabled market defines its `default_currency` and
`display_currencies`.

Changing `primary_currency` does not silently convert existing product amounts.
Admins should use the backend currency audit on the currency page to find
products or SKUs whose saved currency no longer matches the backend entry
currency and then correct those records deliberately.

**Response:**
```json
{
  "data": {
    "primary_currency": "CNY",
    "display_currencies": [],
    "available_currencies": [
      {"code": "CNY", "name": "Chinese Yuan", "minor_units": 2}
    ]
  }
}
```

### Admin Backend Entry Currency Audit
```http
GET /api/admin/settings/currency-policy/audit
```

Returns products and SKUs whose saved `currency` differs from the current
backend entry currency. This is the required safety check after changing
`primary_currency`; the endpoint reports mismatches but does not convert or
rewrite amounts.

**Response:**
```json
{
  "audit": {
    "expected_currency": "USD",
    "product_mismatch_count": 12,
    "variant_mismatch_count": 12,
    "total_mismatch_count": 24,
    "samples": [
      {"kind": "product", "id": 101, "sku": "RIM-CNY", "name": "Old RMB Product", "currency": "CNY"}
    ]
  }
}
```

### List Cached Exchange Rates
```http
GET /api/v1/currency/exchange-rates?base=CNY
```

Nuxt and other clients may read cached backend rates for metadata or tooling
screens only. Normal product browsing should consume display prices returned by
product/shipping APIs. Clients must not call the third-party exchange-rate
provider or receive its API key. The exchange-rate API base is the backend entry
currency; quote targets are the unique `default_currency` and
`display_currencies` from enabled storefront markets, excluding the base.

**Response:**
```json
{
  "data": {
    "base_currency": "CNY",
    "provider": "ExchangeRate-API",
    "refresh_minutes": 1440,
    "quote_currencies": ["USD", "EUR", "GBP"],
    "rates": [
      {"base_currency": "CNY", "quote_currency": "EUR", "rate": 0.1275, "source": "ExchangeRate-API"}
    ]
  }
}
```

### Admin Convert Display Prices
```http
POST /api/admin/pricing/exchange-rates/convert
Content-Type: application/json

{
  "amount": 699,
  "base_currency": "CNY",
  "quote_currencies": ["USD", "EUR", "GBP"]
}
```

This admin helper reads backend cached rates and returns secondary display price
values for product, SKU, and shipping pricing forms. It is a pricing helper, not
a payment-method or gateway-currency endpoint. If `quote_currencies` is omitted,
the enabled storefront market display currencies are used through the
ExchangeRate-API quote target resolver.

Product and SKU forms persist converted values as `display_prices` lists.
Shipping templates and rules persist converted values as
`display_price_snapshots` objects keyed by the money field, such as
`default_fee`, `free_threshold`, rule `fee`, and rule `additional`.

**Response:**
```json
{
  "code": 0,
  "data": {
    "amount": 699,
    "base_currency": "CNY",
    "quote_currencies": ["USD", "EUR", "GBP"],
    "prices": [
      {"amount": 96.8, "currency": "USD", "quote_currency": "USD", "rate": 0.1385, "source": "direct_rate", "converted": true}
    ]
  }
}
```

---

## Cart Endpoints

### Get Cart Summary
```http
GET /api/v1/cart/summary
Cookie: auth_token=<http_only_cookie> (optional)
Cookie: session_id=<session_id>
```

**Response:**
```json
{
  "item_count": 3,
  "total": 299.97
}
```

### Add to Cart
```http
POST /api/v1/cart/add
Cookie: auth_token=<http_only_cookie> (optional)
X-CSRF-Token: <csrf_token_cookie_value>
Content-Type: application/json

{
  "product_id": 1,
  "quantity": 2
}
```

### Update Cart Item
```http
PUT /api/v1/cart/items/:id
Content-Type: application/json

{
  "quantity": 3
}
```

### Remove from Cart
```http
DELETE /api/v1/cart/items/:id
```

---

## Shipping and Checkout Quote Endpoints

### Quote Shipping
```http
POST /api/v1/shipping/quote
Content-Type: application/json

{
  "country": "US",
  "currency": "CNY",
  "display_currency": "USD",
  "items": [
    {"product_id": 1, "variant_id": 10, "quantity": 2}
  ]
}
```

`currency` is the order/catalog pricing currency used to calculate the actual
shipping fee. `display_currency` only selects a persisted display snapshot that
admin already filled from cached exchange rates. Nuxt must not call exchange-rate
providers or cached-rate endpoints to price shipping at request time.

**Response excerpt:**
```json
{
  "shipping_fee": 11,
  "currency": "CNY",
  "display_currency": "USD",
  "display_price": {"amount": 1.54, "currency": "USD", "rate": 0.14, "source": "direct_rate", "converted": true},
  "display_prices": [
    {"amount": 1.54, "currency": "USD", "rate": 0.14, "source": "direct_rate", "converted": true}
  ]
}
```

If a shipping quote includes configured surcharges that do not have complete
stored display snapshots, the response omits `display_price` rather than doing a
runtime conversion. The actual `shipping_fee` remains authoritative for cart,
checkout, order, payment, refund, and callbacks.

### Quote Checkout
```http
POST /api/v1/checkout/quote
Content-Type: application/json

{
  "shipping_address": {"country": "US"},
  "display_currency": "USD",
  "coupon_code": "",
  "points_to_use": 0
}
```

The nested `shipping_quote.display_price` follows the same persisted-snapshot
rule as `/shipping/quote`. Payment method buttons are still exposed only from the
configured payment-method list and risk/protection checks, not from
`display_currency`.

---

## Settings Endpoints

### Get Site Settings
```http
GET /api/v1/settings/site
Accept-Language: en
```

**Response:**
```json
{
  "site_name": "Example Site",
  "site_description": "Premium products",
  "site_logo": "https://...",
  "contact_email": "info@example.com",
  "contact_phone": "+1234567890"
}
```

`site_name` is the single storefront site name used by the admin settings panel and public storefront.

### Get Why This Name Page Settings
```http
GET /api/v1/settings/website-name?locale=en
```

Returns the localized content for the Nuxt `Why This Name` page. The public API
falls back to English for a supported locale without saved content.

## i18n Endpoints

### Get Supported Languages
```http
GET /api/v1/i18n/languages
```

**Response:**
```json
{
  "languages": [
    {
      "code": "en",
      "name": "English",
      "native_name": "English",
      "enabled": true
    },
    {
      "code": "zh_cn",
      "name": "Chinese (Simplified)",
      "native_name": "简体中文",
      "enabled": true
    },
    {
      "code": "fr",
      "name": "French",
      "native_name": "Français",
      "enabled": true
    }
  ],
  "total": 20
}
```

**Supported storefront locales (20 fixed codes)**:

`en, zh_cn, fr, de, es, ja, ko, it, pt, ru, ar, nl, tr, id, th, sv, da, fi, hi, ms`

Browser or integration aliases such as `en-US`, `zh-CN`, or `zh` may be accepted as input compatibility and normalized by the API, but persisted storefront content must use the canonical codes above.

### Get Post Translations
```http
GET /api/v1/i18n/translations/:post_id
```

**Example:**
```http
GET /api/v1/i18n/translations/123
```

**Response:**
```json
{
  "post_id": 123,
  "translations": {
    "en": {
      "id": 123,
      "title": "Product Guide",
      "slug": "product-guide",
      "locale": "en",
      "published_at": "2026-05-20T10:00:00Z",
      "url": "/blog/product-guide"
    },
    "zh_cn": {
      "id": 456,
      "title": "产品指南",
      "slug": "product-guide",
      "locale": "zh_cn",
      "published_at": "2026-05-20T10:00:00Z",
      "url": "/zh_cn/blog/product-guide"
    },
    "fr": {
      "id": 789,
      "title": "Guide du produit",
      "slug": "guide-du-produit",
      "locale": "fr",
      "published_at": "2026-05-20T10:00:00Z",
      "url": "/fr/blog/guide-du-produit"
    }
  },
  "count": 3
}
```

### Detect User Language
```http
GET /api/v1/i18n/detect
Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
Cookie: locale=fr
```

**Response:**
```json
{
  "detected_locale": "fr",
  "source": "cookie"
}
```

**Detection Priority**:
1. Cookie (`locale`)
2. Accept-Language header
3. Default (`en`)

### Set User Language
```http
POST /api/v1/i18n/set-language
Content-Type: application/json

{
  "locale": "zh_cn"
}
```

**Response:**
```json
{
  "message": "Language preference saved",
  "locale": "zh_cn"
}
```

**Note**: Sets a cookie with 1-year expiration.

---

## Sitemap Endpoints

### Get Sitemap Index
```http
GET /sitemap.xml
```

**Response**: XML Sitemap Index
```xml
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap>
    <loc>https://example.com/sitemap-hreflang.xml</loc>
    <lastmod>2026-05-25T10:00:00Z</lastmod>
  </sitemap>
  <sitemap>
    <loc>https://example.com/sitemap-en.xml</loc>
    <lastmod>2026-05-25T10:00:00Z</lastmod>
  </sitemap>
  <sitemap>
    <loc>https://example.com/sitemap-zh_cn.xml</loc>
    <lastmod>2026-05-25T10:00:00Z</lastmod>
  </sitemap>
</sitemapindex>
```

### Get Hreflang Sitemap
```http
GET /sitemap-hreflang.xml
```

**Response**: XML Sitemap with Hreflang tags
```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
        xmlns:xhtml="http://www.w3.org/1999/xhtml">
  <url>
    <loc>https://example.com/blog/product-guide</loc>
    <lastmod>2026-05-20T10:00:00Z</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
    <xhtml:link rel="alternate" hreflang="en" 
                href="https://example.com/blog/product-guide"/>
    <xhtml:link rel="alternate" hreflang="zh-CN"
                href="https://example.com/zh_cn/blog/product-guide"/>
    <xhtml:link rel="alternate" hreflang="fr" 
                href="https://example.com/fr/blog/guide-du-produit"/>
    <xhtml:link rel="alternate" hreflang="x-default" 
                href="https://example.com/blog/product-guide"/>
  </url>
</urlset>
```

**Features**:
- Includes all published posts
- Groups translations together
- Adds hreflang tags for each language version
- Includes x-default tag (usually English)

### Get Locale-Specific Sitemap
```http
GET /sitemap-{locale}.xml
```

**Examples**:
- `/sitemap-en.xml` - English posts
- `/sitemap-zh_cn.xml` - Chinese posts
- `/sitemap-fr.xml` - French posts

**Response**: XML Sitemap for specific language
```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/blog/product-guide</loc>
    <lastmod>2026-05-20T10:00:00Z</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>
  <url>
    <loc>https://example.com/blog/another-post</loc>
    <lastmod>2026-05-21T10:00:00Z</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>
</urlset>
```

---

## Warranty Endpoints

Warranty status is based on a user's shipped order. Product registration,
barcode, serial-number, and product-code lookup endpoints do not exist.

### Get Warranty Status for a Shipped Order
```http
GET /api/v1/warranty/orders/:order_number
Cookie: auth_token=<http_only_cookie>
```

Returns the shipped items, ship date, warranty period, expiry date, remaining
time, and service records for the authenticated user's order.

### Request Warranty Claim Verification
```http
POST /api/v1/warranty/verify-order
Content-Type: application/json

{
  "order_number": "TZ202608230001",
  "email": "user@example.com",
  "captcha_token": ""
}
```

The endpoint returns a generic accepted response and sends a one-time email
verification link only when the order and email match.

### Verify Warranty Claim Token
```http
GET /api/v1/warranty/verify/:token
```

### Submit Warranty Claim
```http
POST /api/v1/warranty/claim
Content-Type: multipart/form-data
```

The multipart request contains `order_number`, `email`,
`verification_token`, `issue_description`, and optional image/video fields.

Admin processing is under `/api/admin/warranty/...`. The admin shipped-order
view reads `orders` and only writes optional after-sales evidence to
`shipment_records`.

---

## Internationalization

All endpoints support multiple languages through:

1. **URL Path** (highest priority):
   ```
   /fr/api/v1/products
   ```

2. **Accept-Language Header**:
   ```
   Accept-Language: fr
   ```

3. **Cookie**:
   ```
   Cookie: locale=fr
   ```

Supported locales: en, zh, fr, de, es, ja, ko, it, pt, ru, ar, fi, da, th, sv, id, ms, nl, tr, fil, tl, jv, hi, ur, mr, ta, te, bn, fa, ps, ha, sw, pcm, be

---

## Error Responses

All errors follow this format:
```json
{
  "error": "error message description"
}
```

**HTTP Status Codes:**
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `429` - Too Many Requests
- `500` - Internal Server Error

---

## Rate Limiting

API is rate-limited to **100 requests per minute** per IP address.

When rate limit is exceeded:
```json
{
  "error": "rate limit exceeded"
}
```

---

## Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```
