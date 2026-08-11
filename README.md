# Commerce Operations Platform

This repository is a monorepo for a configurable commerce software platform. It is not tied to a single storefront brand.

The platform currently contains:

- `nuxt-i18n/` - customer-facing storefront built with Nuxt 4 and Vue 3.
- `go-backend/` - Go API service for public storefront APIs and admin APIs.
- `go-backend/web/admin/` - Vue 3 + Vite admin console.
- `shared/` - cross-stack registry data shared by frontend, backend, and admin checks.

The project is still under active development. Treat this README as the current developer entry point, not as a production readiness claim.

## Repository Map

```text
repo-root/
|-- nuxt-i18n/              # Storefront app
|-- go-backend/
|   |-- cmd/server/          # Go API entrypoint
|   |-- internal/            # API, service, repository, domain packages
|   |-- migrations/          # Database migrations
|   |-- config/              # Backend config examples
|   `-- web/admin/           # Admin console
|-- shared/                  # Shared fixed registries
|-- docs/                    # Current docs plus archive index
|-- docker-compose.yml       # Local Docker services
`-- start-dev.ps1            # Windows local development launcher
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
| Admin | Vue 3, Vite, Pinia, Tailwind CSS, shadcn-vue, Reka UI, Axios |
| Local infra | Docker Compose for PostgreSQL, Redis, API, and storefront |

## Local Development

### Prerequisites

- Go matching `go-backend/go.mod` (`1.25.1` at the time of writing).
- Node.js 24 recommended for current frontend workflows.
- Docker Desktop, or local PostgreSQL + Redis.

### Start the full local stack

From the repository root:

```powershell
npm run dev
```

The root development command starts local infrastructure and the three app servers:

- Storefront Nuxt: `http://localhost:9199`
- Go API: `http://localhost:9200`
- Admin console: `http://localhost:9300`
- PostgreSQL host port: `localhost:9400`
- Redis host port: `localhost:9500`

It also clears the app ports before starting and writes logs under `output/dev/`.

Useful root commands:

```powershell
npm run dev
npm run dev:stop
npm run dev:ports
```

The local dev launcher ensures a first backoffice account exists after the API health check passes:

- Email: `admin@example.com`
- Password: `Admin123456!`
- Role: `admin`

Override these with `DEV_ADMIN_EMAIL`, `DEV_ADMIN_USERNAME`, `DEV_ADMIN_PASSWORD`, and `DEV_ADMIN_ROLE` before running `npm run dev`. If a backoffice user already exists, the bootstrap is skipped. Set `DEV_ADMIN_RESET=true` to reset or create the configured dev admin account.

### Start only infrastructure

```powershell
docker compose up -d postgres redis
```

### Start backend manually

```powershell
cd go-backend
Copy-Item .env.example .env -ErrorAction SilentlyContinue
Copy-Item config/config.example.yaml config/config.yaml -ErrorAction SilentlyContinue
go run ./cmd/server
```

Default backend address:

- `http://localhost:9200`
- Health check: `http://localhost:9200/health`

### Start storefront manually

```powershell
cd nuxt-i18n
npm install
npm run dev
```

Default storefront address:

- `http://localhost:9199`

### Start admin manually

```powershell
cd go-backend/web/admin
npm install
npm run dev
```

Default admin address:

- `http://localhost:9300`
- Backend API base: `/api/admin`

## Docker Compose

The root `docker-compose.yml` can start PostgreSQL, Redis, backend, and the Nuxt storefront:

```powershell
docker compose up -d
```

Compose ports:

- Storefront: `http://localhost:9100`
- API: `http://localhost:9200`
- PostgreSQL: `localhost:9400`
- Redis: `localhost:9500`

The admin console is not a service in the root Compose file; run it manually from `go-backend/web/admin/`.

Optional database/Redis tools are behind the `tools` profile:

```powershell
docker compose --profile tools up -d adminer redis-commander
```

Optional tools:

- Adminer: `http://localhost:9600`
- Redis Commander: `http://localhost:9700`

The root Compose file is for local development only.

## Production Notes

Production deployment assets exist, but they must be reviewed against the current environment before use:

- Compose: `compose.prod.yml`
- Environment template: `deployment/production.env.example`
- Image workflow: `.github/workflows/publish-images.yml`
- VPS runbook: `docs/ops/hostinger-vps-docker-runbook.md`
- Current readiness boundary: `docs/ops/production-readiness-status.md`

Do not share images, volumes, environment variables, databases, Redis instances, or Compose project names with unrelated applications.

Some deployment identifiers may still carry legacy names until the deployment assets are renamed. Treat those identifiers as operational configuration, not as product branding.

Build and deploy the Nuxt storefront through the Linux Docker image path. Do not upload a Windows-built `nuxt-i18n/.output`; native dependencies such as `sharp` are built for the target platform.

## Validation

Backend:

```powershell
cd go-backend
go test ./...
```

Storefront:

```powershell
cd nuxt-i18n
npm run check-locales
npm run scripts:typecheck
npm run check:production-artifacts
npm run build
```

Admin:

```powershell
cd go-backend/web/admin
npm run typecheck
npm run build
```

If a frontend build fails immediately after checkout, run `npm install` in that app directory first.

## Important Boundaries

- Brand/customer-specific names belong in runtime settings, uploaded media, localized content, and deployment environment, not in generic platform docs or new reusable code.
- Storefront locale definitions are fixed through `shared/storefront-locales.json`; Nuxt, Go, and Admin must stay aligned through `npm run check-locales`.
- Admin-editable localized content must use canonical storefront locale codes and controlled locale selectors.
- Payment and refund state should be changed through verified payment provider callbacks or controlled service methods, not by direct handler/repository writes.
- Admin order status must not manually write payment-owned states such as `paid` or `refunded`.
- Handlers should stay thin: parse requests, call services, and return responses.
- Business logic belongs in `go-backend/internal/service`; persistence details belong in `go-backend/internal/repository`.
- Avoid adding legacy CMS or storefront compatibility paths unless they are explicit migration tools.

## Documentation Map

- Project docs index: `docs/README.md`
- Backend guide: `go-backend/README.md`
- Backend API notes: `go-backend/API.md`
- Storefront i18n status: `nuxt-i18n/docs/notes/I18N-CURRENT-STATUS.md`
- Storefront locale registry: `go-backend/docs/STOREFRONT_LOCALE_REGISTRY.md`
- Admin app guide: `go-backend/web/admin/README.md`
- Kubernetes manifests and notes: `go-backend/k8s/`
- Historical reports: `docs/archive/`

Historical reports are context only. Prefer the current code, this README, and area-specific README files when documents conflict.

## Current Status

What is real today:

- Go backend with customer and admin APIs.
- Nuxt storefront app.
- Vue admin console.
- PostgreSQL and Redis local infrastructure.
- Fixed 20-locale storefront registry with cross-stack checks.
- Focused tests and build/typecheck commands for key packages.

What should not be assumed without verification:

- Production readiness.
- Complete CI/CD coverage.
- Full Kubernetes deployment readiness.
- Benchmark numbers or performance claims.
- Any architecture claim that is not exercised by the current code path.

Last updated: 2026-08-09.
