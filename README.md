# Tanzanite Theme

Tanzanite Theme is a monorepo for an e-commerce site with three main parts:

- `nuxt-i18n/` - customer-facing storefront built with Nuxt 4 and Vue 3.
- `go-backend/` - Go API service used by both storefront and admin.
- `go-backend/web/admin/` - Vue 3 + Vite admin console.

This project is still under active development and has not been launched in production. Treat the README as a practical developer entry point, not as a production readiness claim.

## Current Architecture

```text
tanzanite-theme/
|-- nuxt-i18n/              # Storefront app
|-- go-backend/
|   |-- cmd/server/          # Go API entrypoint
|   |-- internal/            # API, service, repository, domain packages
|   |-- migrations/          # Database migrations
|   |-- config/              # Backend config examples
|   `-- web/admin/           # Admin console
|-- docs/                   # Project notes; some files may be older
`-- docker-compose.yml      # Local Postgres, Redis, backend, storefront
```

The backend exposes:

- Public/customer API under `/api/v1`.
- Admin API under `/api/admin`.
- Health checks under `/health`, `/ready`, and `/liveness`.
- Metrics under `/metrics`.

## Tech Stack

| Area | Stack |
| --- | --- |
| Backend | Go `1.25.1`, Gin, GORM, PostgreSQL, Redis |
| Storefront | Nuxt 4, Vue 3, Pinia, Tailwind CSS |
| Admin | Vue 3, Vite, Pinia, Element Plus, Axios |
| Infra | Docker Compose for local services; Kubernetes manifests exist under `go-backend/k8s/` |

## Local Development

### Prerequisites

- Go matching `go-backend/go.mod` (`1.25.1` at the time of writing).
- Node.js 24 recommended for current workflows.
- Docker Desktop, or local PostgreSQL + Redis.

### Start infrastructure

From the repository root:

```powershell
docker compose up -d postgres redis
```

This starts:

- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`

### Start backend

```powershell
cd go-backend
Copy-Item .env.example .env -ErrorAction SilentlyContinue
Copy-Item config/config.example.yaml config/config.yaml -ErrorAction SilentlyContinue
go run ./cmd/server
```

Default backend address:

- `http://localhost:9000`
- health check: `http://localhost:9000/health`

### Start storefront

Use a different port from the admin app if running both at the same time:

```powershell
cd nuxt-i18n
npm install
$env:NUXT_PUBLIC_API_BASE = "http://localhost:9000"
npm run dev -- --port 3001
```

Storefront address:

- `http://localhost:3001`

### Start admin console

```powershell
cd go-backend/web/admin
npm install
npm run dev
```

Admin address:

- `http://localhost:3000`
- backend API base: `/api/admin`

## Docker Compose

The root `docker-compose.yml` can start PostgreSQL, Redis, backend, and the Nuxt storefront:

```powershell
docker compose up -d
```

Notes:

- The compose storefront maps to `http://localhost:3000`.
- The admin console is not a service in the root compose file; run it manually from `go-backend/web/admin/`.
- Optional database/Redis tools are behind the `tools` profile:

```powershell
docker compose --profile tools up -d adminer redis-commander
```

## Testing

Backend:

```powershell
cd go-backend
go test ./...
```

Storefront:

```powershell
cd nuxt-i18n
npm run build
```

Admin:

```powershell
cd go-backend/web/admin
npm run build
```

If a frontend build fails immediately after checkout, run `npm install` in that app directory first.

## Important Backend Boundaries

- Payment and refund state should be changed through verified payment provider callbacks or controlled service methods, not by direct handler/repository writes.
- Admin order status must not manually write payment-owned states such as `paid` or `refunded`.
- Handlers should stay thin: parse requests, call services, and return responses.
- Business logic belongs in `internal/service`; persistence details belong in `internal/repository`.
- The current project is being simplified toward one source of truth. Avoid adding legacy WordPress compatibility paths unless they are explicit migration tools.

## Documentation Map

- Backend guide: `go-backend/README.md`
- Backend API notes: `go-backend/API.md`
- Backend maintainability notes: `go-backend/MAINTAINABILITY_GUIDE.md`
- Admin app guide: `go-backend/web/admin/README.md`
- Kubernetes manifests and notes: `go-backend/k8s/`

Some older documentation files still contain historical or aspirational wording. Prefer the current code and this README when there is a conflict.

## Project Status

What is real today:

- Go backend with customer and admin APIs.
- Nuxt storefront app.
- Vue admin console.
- PostgreSQL and Redis local infrastructure.
- Tests for key backend packages.

What should not be assumed without verification:

- Production readiness.
- Claimed benchmark numbers.
- Complete CI/CD coverage.
- Full Kubernetes deployment readiness.
- Any "microservice", "CQRS", "edge", or "AI search" claim unless proven in the current code path.

Last updated: 2026-07-02.
