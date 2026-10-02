# Backend Deployment Notes

These notes describe deployment preparation for the Go backend. They are not a production-readiness certification.

## Required Services

- PostgreSQL
- Redis
- Object storage if upload/storage features are enabled
- Payment provider credentials if payment flows are enabled

## Configuration

Start from the examples:

```powershell
cd go-backend
Copy-Item .env.example .env -ErrorAction SilentlyContinue
Copy-Item config/config.example.yaml config/config.yaml -ErrorAction SilentlyContinue
```

Before deploying, set real values for:

- Database host, username, password, and database name
- Redis host and credentials
- JWT and cookie secrets
- CORS origins
- Upload/storage provider settings
- Payment provider settings
- Log level and server mode

Do not commit real secrets.

## Admin Account Recovery

Production admin passwords are not recoverable because `users.password` stores bcrypt hashes only. Use the audited `adminctl` command to create or reset a backoffice account with a newly generated password.

See `docs/ADMIN_ACCOUNT_RECOVERY.md` for the runbook.

## Build

```powershell
cd go-backend
go test ./...
go build ./cmd/server
go build ./cmd/adminctl
```

<a id="schwalbe-migrations-357-370-preflight"></a>
<a id="schwalbe-migrations-357-361-preflight"></a>
<a id="schwalbe-migrations-357-360-preflight"></a>
<a id="schwalbe-migrations-357-359-preflight"></a>
<a id="schwalbe-migrations-357-358-preflight"></a>

## Schwalbe Catalog Migrations 357-370 Preflight

Migration 357 may remove only an empty legacy catalog; migration 358 recreates the standalone candidate catalog for the all-spectrum selector and template autofill; migration 359 seeds the official Schwalbe sitemap snapshot; migration 360 adds the database uniqueness boundary for Article No. on sales products; migration 361 creates the route-bound, backend-editable FAQ page for the Schwalbe selector and seeds the English/Chinese bead-label explanations. Migration 362 retires the old Schwalbe tire-width/inner-rim-width guidance table, migration 369 adds reviewed Hookless facts, and migration 370 removes any stale table left by the superseded 362 migration. The selector has no broad tire-width compatibility endpoint. Inspect each environment before applying 357. First check whether the table exists:

```sql
SELECT to_regclass('public.schwalbe_tire_specifications') AS legacy_table;
```

If it exists, check its row count:

```sql
SELECT count(*) FROM public.schwalbe_tire_specifications;
```

Migration 357 stops if the table contains any rows; it does not clear or convert them. If the count is nonzero, stop the deployment and preserve/inspect those catalog rows before deciding how to proceed. Do not map candidate rows to sales Products merely to satisfy the migration. After 357 succeeds, migration 358 recreates the candidate table without review or approval fields. Migration 359 applies the checked-in official snapshot as an upsert seed and must run after 358. Migration 360 fails closed if existing sales Products already contain duplicate Article No. values; resolve those duplicates before retrying and do not delete catalog or product data to bypass the check. Allow the migration runner enough statement and lock timeout for the seed transaction.

In release mode keep `DB_AUTO_MIGRATE=false`, run the migration command before starting or rolling the API, then deploy the API after the migration succeeds. From `go-backend`, run `go run ./cmd/server migrate`; it applies all pending migrations rather than only one numbered migration. After the migration and API rollout, the API startup sync and the Admin FAQ “同步前台路由” action reconcile the current Nuxt route manifest. Purge storefront HTML cache once after the rollout if old cached pages are still serving.

After the migration completes, verify:

```sql
SELECT version, dirty FROM schema_migrations;
SELECT count(*) FROM public.schwalbe_tire_specifications;
SELECT indexname FROM pg_indexes WHERE tablename = 'product_spec_values' AND indexname = 'uq_schwalbe_product_article_no';
SELECT page_id, route_path, locale, status FROM faq_pages WHERE page_id = 'guides-schwalbe-tire-selector';
SELECT locale, count(*) FROM faqs WHERE page_id = 'guides-schwalbe-tire-selector' AND status = 'published' GROUP BY locale;
SELECT to_regclass('public.schwalbe_tire_hookless_compatibility') AS hookless_compatibility_table;
SELECT to_regclass('public.schwalbe_tire_rim_width_combination_rules') AS retired_combination_table;
```

The selector has no broad Schwalbe combination endpoint. The calculator uses the DT Swiss chart and physical calculations documented in the tire-rim helper.

## Runtime Checks

Expose these internal checks to your load balancer or platform health probes:

- `/health`
- `/ready`
- `/liveness`

Metrics are exposed at:

- `/metrics`

## Docker Compose

For local or staging-like environments, use the root compose file:

```powershell
docker compose up -d
```

The root compose file is mainly a development convenience. Review environment variables, volumes, secrets, ports, and persistence before using it outside local development.

## Kubernetes

Kubernetes manifests live under `k8s/`. Treat them as deployment templates that need environment-specific review before use.

## Pre-Launch Checklist

- `go test ./...` passes.
- Database migrations are reviewed and reversible where practical.
- Logs do not print secrets or sensitive customer data.
- Cookie, CSRF, CORS, and WebSocket origin settings are configured for the real domains.
- Upload endpoints enforce type, size, and storage limits.
- Payment callbacks are verified and idempotent.
- Backups and restore drills exist for PostgreSQL and object storage.
