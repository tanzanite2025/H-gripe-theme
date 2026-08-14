# Backend Quick Start

Use this for local backend development. For the full monorepo flow, start from `../README.md`.

## Prerequisites

- Go matching `go.mod` (`1.25.1` at the time of writing)
- Docker Desktop, or local PostgreSQL and Redis
- PowerShell on Windows, or an equivalent shell

## Start Infrastructure

From the repository root:

```powershell
docker compose up -d postgres redis
```

This starts:

- PostgreSQL: `localhost:9400`
- Redis: `localhost:9510`

## Initialize or Migrate the Database

For a new or empty PostgreSQL database, run the backend migration entrypoint:

```powershell
go run ./cmd/server migrate
```

This applies the complete versioned SQL migration chain from `001` to the latest version. The migration files now contain the required historical baseline tables, so a raw `golang-migrate` invocation is also supported when needed:

```powershell
migrate -path ./migrations -database "$env:DATABASE_URL" up
```

Use the backend entrypoint for normal development and deployment workflows. Use the raw command for migration-only or rollback verification.

## Run Backend

```powershell
cd go-backend
Copy-Item .env.example .env -ErrorAction SilentlyContinue
Copy-Item config/config.example.yaml config/config.yaml -ErrorAction SilentlyContinue
go run ./cmd/server
```

Default local address:

- API: `http://localhost:9200`
- Health: `http://localhost:9200/health`
- Ready: `http://localhost:9200/ready`

## Smoke Test

```powershell
Invoke-RestMethod http://localhost:9200/health
```

Expected shape:

```json
{
  "status": "ok"
}
```

## Run Tests

```powershell
go test ./...
```

## Useful Links

- Backend README: `README.md`
- API notes: `API.md`
- API testing guide: `docs/API_TESTING_GUIDE.md`
- Admin console: `web/admin/README.md`
