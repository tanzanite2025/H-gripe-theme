# API Testing Guide

Use this guide for local smoke tests and focused backend API checks.

## Local Base URL

```powershell
$env:API_BASE_URL = "http://localhost:9000"
```

## Health Checks

```powershell
Invoke-RestMethod "$env:API_BASE_URL/health"
Invoke-RestMethod "$env:API_BASE_URL/ready"
Invoke-RestMethod "$env:API_BASE_URL/liveness"
```

## Cookie Auth Flow

Register:

```powershell
curl.exe -X POST "$env:API_BASE_URL/api/v1/auth/register" `
  -H "Content-Type: application/json" `
  -d "{\"email\":\"test@example.com\",\"username\":\"testuser\",\"password\":\"Password123!\"}"
```

Login and store cookies:

```powershell
curl.exe -X POST "$env:API_BASE_URL/api/v1/auth/login" `
  -c cookies.txt `
  -H "Content-Type: application/json" `
  -d "{\"email_or_username\":\"test@example.com\",\"password\":\"Password123!\"}"
```

Read profile:

```powershell
curl.exe "$env:API_BASE_URL/api/v1/auth/profile" -b cookies.txt
```

Unsafe methods that rely on cookie auth must include the CSRF token from the `csrf_token` cookie:

```text
X-CSRF-Token: <csrf_token_cookie_value>
```

## Common Public Checks

```powershell
curl.exe "$env:API_BASE_URL/api/v1/products?page=1&page_size=10"
curl.exe "$env:API_BASE_URL/api/v1/settings/site"
curl.exe "$env:API_BASE_URL/api/v1/subscriptions/status/test@example.com"
```

## Admin Checks

Admin endpoints live under `/api/admin` and require an authenticated admin session.

```powershell
curl.exe "$env:API_BASE_URL/api/admin/subscriptions/stats" -b cookies.txt
```

For state-changing admin requests, include both cookies and `X-CSRF-Token`.

## WebSocket Checks

Use the currently registered chat route from the backend router before testing. If the route requires authentication, connect with valid cookies or headers from the same origin expected by the backend CORS/WebSocket policy.

Do not test WebSockets by disabling origin or auth checks.

## Automated Tests

```powershell
cd go-backend
go test ./...
```
