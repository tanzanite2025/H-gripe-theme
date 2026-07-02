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

- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`

## Run Backend

```powershell
cd go-backend
Copy-Item .env.example .env -ErrorAction SilentlyContinue
Copy-Item config/config.example.yaml config/config.yaml -ErrorAction SilentlyContinue
go run ./cmd/server
```

Default local address:

- API: `http://localhost:9000`
- Health: `http://localhost:9000/health`
- Ready: `http://localhost:9000/ready`

## Smoke Test

```powershell
Invoke-RestMethod http://localhost:9000/health
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
