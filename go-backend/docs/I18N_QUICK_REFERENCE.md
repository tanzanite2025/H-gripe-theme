# Blog I18n Quick Reference

This note covers the backend blog i18n endpoints and sitemap routes.

## Public API

```http
GET /api/v1/i18n/languages
GET /api/v1/i18n/translations/:post_id
GET /api/v1/i18n/detect
POST /api/v1/i18n/set-language
```

`POST /api/v1/i18n/set-language` accepts:

```json
{
  "locale": "en"
}
```

## Sitemap Routes

```http
GET /sitemap.xml
GET /sitemap-hreflang.xml
GET /sitemap-:locale.xml
```

## Main Code Paths

- Handler: `internal/api/v1/i18n/handler.go`
- Posts service: `internal/service/post_service.go`
- Sitemap service: `internal/service/sitemap_service.go`
- Post repository: `internal/repository/post_repository.go`
- Post model: `internal/domain/post/model.go`

## Maintenance Notes

- Keep locale validation in one place.
- Avoid duplicating URL-building logic between handlers and services.
- Prefer service methods over handler-level database queries.
- If frontend routing changes, update sitemap behavior and storefront links together.

## Testing

```powershell
cd go-backend
go test ./internal/api/v1/i18n ./internal/service
```
