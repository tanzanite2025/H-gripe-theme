# Tanzanite Go Backend

This is the Go API service for the Tanzanite e-commerce project. It serves both the Nuxt storefront and the Vue admin console.

The project is still under active development and has not launched in production. Treat this file as a practical backend entry point, not a production-readiness claim.

## Runtime Shape

- Public/customer API: `/api/v1`
- Admin API: `/api/admin`
- Health checks: `/health`, `/ready`, `/liveness`
- Metrics: `/metrics`

## Tech Stack

- Go `1.25.1`
- Gin HTTP router
- GORM persistence layer
- PostgreSQL primary database
- Redis for cache/session/worker-related infrastructure
- Docker Compose for local infrastructure

## Directory Map

```text
go-backend/
|-- cmd/server/          # API entrypoint
|-- config/              # configuration examples
|-- docs/                # backend module and testing notes
|-- internal/
|   |-- api/             # routes, middleware, HTTP handlers
|   |-- domain/          # domain models
|   |-- repository/      # persistence access
|   |-- service/         # business logic
|   `-- pkg/             # internal shared packages
|-- migrations/          # database migrations
|-- scripts/             # backend utility scripts
|-- web/admin/           # Vue admin console
|-- API.md               # API notes
|-- DEPLOYMENT.md        # deployment notes
`-- QUICK_START.md       # quick start notes
```

## Local Development

From the repository root, start PostgreSQL and Redis:

```powershell
docker compose up -d postgres redis
```

Then run the backend:

```powershell
cd go-backend
Copy-Item .env.example .env -ErrorAction SilentlyContinue
Copy-Item config/config.example.yaml config/config.yaml -ErrorAction SilentlyContinue
go run ./cmd/server
```

Default local address:

- `http://localhost:9000`
- health check: `http://localhost:9000/health`

## Testing

```powershell
cd go-backend
go test ./...
```

## Backend Boundaries

- Handlers should parse requests, call services, and return responses.
- Business rules belong in `internal/service`.
- Persistence details belong in `internal/repository`.
- Database transactions should be owned by a clear service-level use case or a dedicated transaction helper.
- Payment and refund state should only change through verified provider callbacks or controlled admin/service flows.
- New code should not add WordPress compatibility paths unless they are explicit migration tools.

## Related Docs

- API notes: `API.md`
- Quick start: `QUICK_START.md`
- Deployment notes: `DEPLOYMENT.md`
- API testing guide: `docs/API_TESTING_GUIDE.md`
- Blog i18n maintainability notes: `docs/MAINTAINABILITY_GUIDE.md`
- Blog i18n quick reference: `docs/I18N_QUICK_REFERENCE.md`
- Subscription quick reference: `docs/SUBSCRIPTION_QUICK_REFERENCE.md`
- Admin console: `web/admin/README.md`

## Historical Reports

Old completion, security, quality, and refactoring reports live under `../docs/archive/`. They are useful context, but they are not the current source of truth.

Last updated: 2026-07-02.
