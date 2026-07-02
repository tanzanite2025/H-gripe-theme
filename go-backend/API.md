# Tanzanite Go Backend API Documentation

## Base URL
```
http://localhost:9000/api/v1
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

## Product Endpoints

### List Products
```http
GET /api/v1/products?page=1&page_size=12&featured=true&status=active
Accept-Language: en
```

**Query Parameters:**
- `page` (default: 1)
- `page_size` (default: 12, max: 100)
- `status` (default: active)
- `featured` (boolean)

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "sku": "PROD-001",
      "name": "Product Name",
      "slug": "product-name",
      "price": 99.99,
      "sale_price": 79.99,
      "stock": 50,
      "featured_image": "https://...",
      "locale": "en"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 12,
  "total_pages": 9
}
```

### Get Single Product
```http
GET /api/v1/products/:id
Accept-Language: en
```

Can use either ID or slug.

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

## Settings Endpoints

### Get Site Settings
```http
GET /api/v1/settings/site
Accept-Language: en
```

**Response:**
```json
{
  "site_name": "Tanzanite",
  "site_description": "Premium products",
  "site_logo": "https://...",
  "contact_email": "info@tanzanite.site",
  "contact_phone": "+1234567890"
}
```

### Get Quick Buy Settings
```http
GET /api/v1/settings/quick-buy
Accept-Language: en
```

**Response:**
```json
{
  "enabled": true,
  "button_text": "Quick Buy",
  "success_message": "Added to cart!",
  "require_login": false
}
```

---

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
      "code": "zh",
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
  "total": 34
}
```

**Supported Languages (34 total)**:
- European: en, fr, de, es, it, pt, ru, nl, pl, tr, sv, no, da, fi, cs, hu, ro
- Asian: zh, zh-TW, ja, ko, vi, th, id, ms, hi, bn, ta, te, mr, ur
- Middle Eastern: ar, fa, he

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
    "zh": {
      "id": 456,
      "title": "产品指南",
      "slug": "product-guide",
      "locale": "zh",
      "published_at": "2026-05-20T10:00:00Z",
      "url": "/zh/blog/product-guide"
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
  "locale": "zh"
}
```

**Response:**
```json
{
  "message": "Language preference saved",
  "locale": "zh"
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
    <loc>https://tanzanite.site/sitemap-hreflang.xml</loc>
    <lastmod>2026-05-25T10:00:00Z</lastmod>
  </sitemap>
  <sitemap>
    <loc>https://tanzanite.site/sitemap-en.xml</loc>
    <lastmod>2026-05-25T10:00:00Z</lastmod>
  </sitemap>
  <sitemap>
    <loc>https://tanzanite.site/sitemap-zh.xml</loc>
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
    <loc>https://tanzanite.site/blog/product-guide</loc>
    <lastmod>2026-05-20T10:00:00Z</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
    <xhtml:link rel="alternate" hreflang="en" 
                href="https://tanzanite.site/blog/product-guide"/>
    <xhtml:link rel="alternate" hreflang="zh" 
                href="https://tanzanite.site/zh/blog/product-guide"/>
    <xhtml:link rel="alternate" hreflang="fr" 
                href="https://tanzanite.site/fr/blog/guide-du-produit"/>
    <xhtml:link rel="alternate" hreflang="x-default" 
                href="https://tanzanite.site/blog/product-guide"/>
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
- `/sitemap-zh.xml` - Chinese posts
- `/sitemap-fr.xml` - French posts

**Response**: XML Sitemap for specific language
```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://tanzanite.site/blog/product-guide</loc>
    <lastmod>2026-05-20T10:00:00Z</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>
  <url>
    <loc>https://tanzanite.site/blog/another-post</loc>
    <lastmod>2026-05-21T10:00:00Z</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>
</urlset>
```

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
