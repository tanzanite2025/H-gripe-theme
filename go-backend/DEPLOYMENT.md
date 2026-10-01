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

<a id="schwalbe-migrations-357-362-preflight"></a>
<!-- Keep previous anchors for existing runbooks and bookmarks. -->
<a id="schwalbe-migrations-357-361-preflight"></a>
<a id="schwalbe-migrations-357-360-preflight"></a>
<a id="schwalbe-migrations-357-359-preflight"></a>
<a id="schwalbe-migrations-357-358-preflight"></a>

## Schwalbe Migrations 357-362 Preflight

Migration 357 may remove only an empty legacy catalog; migration 358 recreates the standalone candidate catalog for the all-spectrum selector and template autofill; migration 359 seeds the official Schwalbe sitemap snapshot; migration 360 adds the database uniqueness boundary for Article No. on sales products; migration 361 creates the route-bound, backend-editable FAQ page for the Schwalbe selector and seeds the English/Chinese bead-label explanations; migration 362 creates the official tire-width/inner-rim-width possible-combination rules table and seeds 13 rows from the 05/2024 matrix; migration 363 repairs historical FAQ route paths and adds manifest-driven route reconciliation metadata without deleting FAQ content. The 362 guidance is not a model-specific compatibility certification and its down migration is intentionally forward-only. Inspect each environment before applying 357. First check whether the table exists:

```sql
SELECT to_regclass('public.schwalbe_tire_specifications') AS legacy_table;
```

If it exists, check its row count:

```sql
SELECT count(*) FROM public.schwalbe_tire_specifications;
```

Migration 357 stops if the table contains any rows; it does not clear or convert them. If the count is nonzero, stop the deployment and preserve/inspect those catalog rows before deciding how to proceed. Do not map candidate rows to sales Products merely to satisfy the migration. After 357 succeeds, migration 358 recreates the candidate table without review or approval fields. Migration 359 applies the checked-in official snapshot as an upsert seed and must run after 358. Migration 360 fails closed if existing sales Products already contain duplicate Article No. values; resolve those duplicates before retrying and do not delete catalog or product data to bypass the check. Migration 362 upserts the 13 official possible-combination ranges and must not be rolled back by deleting its table or rows; use a corrective forward migration if the source data changes. Allow the migration runner enough statement and lock timeout for the seed transaction.

In release mode keep `DB_AUTO_MIGRATE=false`, run the migration command before starting or rolling the API, then deploy the API after the migration succeeds. From `go-backend`, run `go run ./cmd/server migrate`; it applies all pending migrations, including 363, rather than only that migration. After migration 363 and the API rollout, the API startup sync and the Admin FAQ “同步前台路由” action reconcile the current Nuxt route manifest. Purge storefront HTML cache once after the rollout if old cached pages are still serving.

After the migration completes, verify:

```sql
SELECT version, dirty FROM schema_migrations;
SELECT count(*) FROM public.schwalbe_tire_specifications;
SELECT indexname FROM pg_indexes WHERE tablename = 'product_spec_values' AND indexname = 'uq_schwalbe_product_article_no';
SELECT page_id, route_path, locale, status FROM faq_pages WHERE page_id = 'guides-schwalbe-tire-selector';
SELECT locale, count(*) FROM faqs WHERE page_id = 'guides-schwalbe-tire-selector' AND status = 'published' GROUP BY locale;
SELECT count(*) AS rule_count,
       min(source_version) AS min_source_version,
       max(source_version) AS max_source_version,
       min(source_checked_at) AS first_checked_at,
       max(source_checked_at) AS last_checked_at
FROM public.schwalbe_tire_rim_width_combination_rules;
SELECT count(*) AS duplicate_rule_count
FROM (
    SELECT tire_width_min_mm, tire_width_max_mm, inner_rim_width_min_mm, inner_rim_width_max_mm
    FROM public.schwalbe_tire_rim_width_combination_rules
    GROUP BY 1, 2, 3, 4
    HAVING count(*) > 1
) AS duplicate_rules;
SELECT indexname
FROM pg_indexes
WHERE schemaname = 'public'
  AND tablename = 'schwalbe_tire_rim_width_combination_rules'
ORDER BY indexname;
```

For the current 362 baseline, expect `version=362`, `dirty=false`, 13 rules with `source_version=05/2024` and `source_checked_at=2026-09-28`, zero duplicate rules, and the two range indexes plus the unique constraint index. After the API starts, verify the public endpoint returns the expected ordered data:

```powershell
$response = Invoke-RestMethod 'http://localhost:9200/api/v1/products/schwalbe-tire-rim-width-combination-rules'
if ($response.code -ne 0 -or @($response.data).Count -ne 13) {
    throw 'Expected 13 Schwalbe tire/rim-width rules from the API.'
}
$response.data | Select-Object tire_width_min_mm, tire_width_max_mm, inner_rim_width_min_mm, inner_rim_width_max_mm, source_version, source_checked_at
```

The first range should be tire width 20–21 mm with inner rim width 15–17 mm; the last should be tire width 114–132 mm with inner rim width 72–100 mm. These ranges describe possible combinations only; they do not certify a specific tire model or replace rim-manufacturer and frame-clearance checks.

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
