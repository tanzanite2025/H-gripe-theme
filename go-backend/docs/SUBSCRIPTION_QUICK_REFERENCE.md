# Subscription Quick Reference

This guide covers the newsletter/subscription API surface.

## Setup

Run backend migrations through the normal backend startup/migration flow. The subscription schema is defined in:

- `migrations/008_enhance_subscription_system.up.sql`

For local development:

```powershell
cd go-backend
go run ./cmd/server
```

## Public Endpoints

Subscribe:

```http
POST /api/v1/subscriptions
Content-Type: application/json

{
  "email": "user@example.com",
  "source": "website",
  "locale": "en",
  "tags": ["newsletter", "promotions"]
}
```

Unsubscribe by token:

```http
GET /api/v1/subscriptions/unsubscribe/:token
```

Unsubscribe by email:

```http
POST /api/v1/subscriptions/unsubscribe
Content-Type: application/json

{
  "email": "user@example.com"
}
```

Resubscribe:

```http
POST /api/v1/subscriptions/resubscribe
Content-Type: application/json

{
  "email": "user@example.com"
}
```

Status:

```http
GET /api/v1/subscriptions/status/:email
```

## Admin Endpoints

Admin endpoints require an authenticated admin session under `/api/admin`.

```http
GET /api/admin/subscriptions?page=1&page_size=20&status=active
GET /api/admin/subscriptions/:email
GET /api/admin/subscriptions/stats
GET /api/admin/subscriptions/active-emails
PATCH /api/admin/subscriptions/:email/status
DELETE /api/admin/subscriptions/:email
```

Unsafe admin methods must include:

```text
Cookie: auth_token=<http_only_cookie>
X-CSRF-Token: <csrf_token_cookie_value>
```

## Local Curl Examples

```powershell
$env:API_BASE_URL = "http://localhost:9000"

curl.exe -X POST "$env:API_BASE_URL/api/v1/subscriptions" `
  -H "Content-Type: application/json" `
  -d "{\"email\":\"user@example.com\",\"source\":\"newsletter_form\",\"locale\":\"en\",\"tags\":[\"newsletter\"]}"

curl.exe "$env:API_BASE_URL/api/v1/subscriptions/unsubscribe/abc123def456"

curl.exe "$env:API_BASE_URL/api/admin/subscriptions/stats" -b cookies.txt
```

## Tests

```powershell
cd go-backend
go test ./internal/api/v1/subscription ./internal/service
```
