# Storefront Theme Hostinger VPS Docker Runbook

This runbook applies only to the `h-gripe-theme` storefront, admin, and Go API. The ERP application remains a separate Compose project with separate images, databases, volumes, and production secrets.

## Production Boundary

The Hostinger VPS has one shared public gateway project named `shared-edge`. The storefront joins that gateway network as a separate Compose project named `h-gripe-theme`.

Current Hostinger VPS target:

- Virtual Machine ID: `1834903`
- Hostname: `srv1834903.hstgr.cloud`
- IPv4: `2.25.85.201`
- OS: Ubuntu 24.04 LTS

Do not guess the machine target during deployment. If the MCP tools do not already have a VM ID, first call `VPS_getVirtualMachinesV1`, confirm the expected single production VPS, then call project-level tools with that ID.

```text
Cloudflare
  -> Hostinger firewall: 80/443
  -> shared-edge (shared Caddy)
      -> erp.tanzanite.site -> erp-web:8080
      -> learn.gripe       -> theme-web:8080
      -> admin.learn.gripe -> theme-web:8080

h-gripe-theme project
  -> web -> storefront
         -> admin
         -> api -> PostgreSQL
                -> Redis
                -> uploads volume
```

The shared gateway is infrastructure. It is not the ERP application and it is not the storefront application.

## Current Gateway State

The neutral gateway name is `shared-edge`. The storefront joins that shared
edge network through the `web` service only.

1. Keep the shared gateway running.
2. Keep `learn.gripe`, `www.learn.gripe`, and `admin.learn.gripe` routed to `theme-web`.
3. Keep the ERP route separate.
4. Do not recreate the old `tanzanite-edge` project or reintroduce old public hostnames into the active boundary.

Browser API requests stay same-origin through `web`. Nuxt server-side requests to `/api/**` use Nitro's internal proxy and go directly to `api:9000`, so SSR does not loop through Cloudflare or the public gateway.

## Production Files

| File | Purpose |
| --- | --- |
| `compose.prod.yml` | Hostinger production Compose project |
| `deployment/production.env.example` | Production environment template |
| `deployment/docker/web.Dockerfile` | Internal Nginx entry image |
| `deployment/nginx/theme-web.conf` | Same-origin API, storefront, admin, and upload routing |
| `deployment/edge/learn-gripe.caddy` | Route fragment for the shared Caddy gateway |
| `docs/ops/learn-gripe-cutover.md` | Prepared migration procedure for the replacement storefront domain |
| `deployment/verify-vps-release-boundary.sh` | Static and runtime Compose/network release evidence |
| `.github/workflows/publish-images.yml` | GHCR image publishing |
| `deploy.sh` | Recommended SSH deployment path |

The root `docker-compose.yml` remains a local development convenience and must not be deployed through Hostinger Docker Manager.

## Project Isolation

The production project must keep these boundaries:

1. Project name: `h-gripe-theme`.
2. Images: `ghcr.io/tanzanite2025/tanzanite-theme-*`.
3. Volumes: `h-gripe-theme-postgres-data`, `h-gripe-theme-redis-data`, and `h-gripe-theme-uploads`.
4. Data networks: project-owned internal `db` and `cache` networks.
5. Application network: project-owned `app` network for service-to-service traffic and required outbound integrations.
6. Shared network: external `shared-edge`, joined only by `web`.
7. Edge alias: `theme-web`.
8. No `container_name` and no host `ports` in the business stack.
9. No ERP environment variables, volumes, image tags, or database credentials.
10. `TRUSTED_PROXIES` contains only private Docker CIDRs. The shared Caddy gateway remains responsible for strict Cloudflare proxy trust.

The Compose file now defaults to the `h-gripe-theme` project and volume names. Existing
VPS resources named `tanzanite-theme-*` are not renamed automatically. During a
volume migration, set `COMPOSE_PROJECT_NAME`, `POSTGRES_DATA_VOLUME_NAME`,
`REDIS_DATA_VOLUME_NAME`, and `UPLOADS_VOLUME_NAME` in the untracked production env
to the old names until the backups and data copy have been verified. Do not start
the new boundary against empty volumes before that migration is complete.

Database and cache reachability is intentional:

- `db` is internal and contains only `db`, `api`, and the one-shot `migrate` service. `storefront`, `admin`, and `web` cannot resolve or connect to PostgreSQL through Docker networking.
- `cache` is internal and contains `redis`, `api`, and `storefront` because the SSR HTML cache is Redis-backed. Redis has a password and is never published to the host.
- `app` is not marked `internal` because the API and storefront need outbound SMTP, Turnstile, payment, and registry-related traffic. No service in the business stack publishes a host port.
- `edge` is the only public gateway path. Only `web` joins it; the API, database, Redis, storefront, and admin are not directly reachable from Cloudflare or the VPS interface.

## Publish Images

Every commit pushed to `master` produces one complete deployable release. GitHub Actions validates the backend and both frontends, then publishes:

- `ghcr.io/tanzanite2025/tanzanite-theme-api`
- `ghcr.io/tanzanite2025/tanzanite-theme-storefront`
- `ghcr.io/tanzanite2025/tanzanite-theme-admin`
- `ghcr.io/tanzanite2025/tanzanite-theme-web`

Production images are published with both a mutable `master` tag and an immutable full tag `sha-<40-character-commit>`.

- Hostinger Docker Manager / MCP Project Update uses `IMAGE_TAG=master` so a normal Project Update pulls the latest tested `master` images.
- `deploy.sh` overrides `IMAGE_TAG` at runtime with `sha-<full-commit>` for immutable SSH releases and rollback.

Publishing on every `master` commit is intentional: both deployment paths require the branch head to have a matching four-image release. The GHCR packages must be public or the VPS must have read-only registry credentials.

`.github/workflows/go-backend-ci.yml` and `.github/workflows/ci.yml` are validation only. They must not publish production images or deploy Kubernetes. `.github/workflows/publish-images.yml` is the only production image publisher.

As of 2026-07-25, `publish-images.yml` triggers on every `master` push, including documentation-only changes. That behavior is safe but noisy. If release frequency becomes confusing, add explicit path filters so production images are published only when runtime, deployment, or workflow files change, then update this runbook at the same time.

## Create The Hostinger Project

Create `deployment/production.env` from the example and replace every `CHANGE_ME` value. Keep the real file outside Git. Keep `IMAGE_TAG=master` when the project is managed through Hostinger Docker Manager or MCP Project Update.

The recommended release path is a repository clone on the VPS:

```bash
./deploy.sh
```

The script fetches `origin/master`, resolves the full commit SHA, waits for all four matching GHCR images, validates Compose, and recreates the project only after every image is available. Run it immediately after a push or after GitHub Actions completes.

In Hostinger Docker Manager create:

```text
Project name: h-gripe-theme
Compose source: compose.prod.yml
Environment: deployment/production.env values, including IMAGE_TAG=master
```

Hostinger Docker Manager cannot derive the Git commit tag itself. When using the Manager or MCP Project Update, keep `IMAGE_TAG=master`; the publish workflow moves that tag after validation. Set `IMAGE_TAG=sha-<full-commit>` only for a deliberate pinned release or rollback.

Before deployment, confirm the external Docker network `shared-edge` exists.

Keep `TRUSTED_PROXIES` at the private Docker ranges from the example unless the Docker network design is intentionally changed. Do not add `0.0.0.0/0` or `::/0`.

Expected services:

- `db`
- `redis`
- `api`
- `storefront`
- `admin`
- `web`

All six services must become Healthy. Only the `shared-edge` gateway may publish host ports.

## Release Boundary Evidence

Run the repository-provided verifier from the same checkout and with the same
`deployment/production.env` used for the release. It uses the resolved network
names from `docker compose config --format json`; it does not infer names from
the repository directory.

Before deployment, record the static Compose boundary:

```bash
./deployment/verify-vps-release-boundary.sh
```

After the services are healthy, record runtime network and port evidence:

```bash
CHECK_CONNECTIVITY=true ./deployment/verify-vps-release-boundary.sh
```

Each run writes a timestamped directory under `release-evidence/`. The
Compose JSON is sanitized before it is retained. Copy the report directory to
the release evidence store and keep it with the exact image tag or digest.

## Add Shared Gateway Routes

Merge `deployment/edge/learn-gripe.caddy` into the existing `shared-edge` Caddyfile without changing the ERP route.

Required routes:

```text
learn.gripe, www.learn.gripe -> theme-web:8080
admin.learn.gripe            -> theme-web:8080
```

Updating the shared gateway is a separate infrastructure operation. Do not copy the storefront Compose into the ERP project and do not replace the gateway project.

## DNS State

`learn.gripe` is the canonical public domain. Keep DNS, TLS, and proxy settings aligned with that hostname and its `www` and `admin` aliases.

Keep the ERP record set separate from the storefront boundary.

## Deliberately Disabled Integrations

Do not add payment provider secrets to the production environment yet. The current generic payment webhook endpoint is not a complete provider-native adapter: Stripe payload mapping is incomplete, and the PayPal and Alipay verification paths are not production-ready. With no provider secrets in the API container, these callbacks fail closed and cannot update payment state.

Outbound SMTP is wired for newsletter and warranty email challenges and is present in the production template. Keep it configured with a dedicated sender, provider-side rate limits, and credentials that are rotated independently from JWT, Redis, and database secrets.

## Verification

Verify before enabling public traffic:

```text
GET /                          -> storefront 200
GET /healthz                   -> theme-web 200
GET /api/v1/settings/site      -> Go API response
GET admin.learn.gripe/         -> admin login page
GET /uploads/<known-file>      -> static file response
WebSocket Upgrade              -> reaches customer-service authentication
```

Also verify:

- Customer login, refresh, logout, and CSRF behavior.
- Admin login and token refresh.
- Product list and product detail.
- Checkout quote and order creation.
- Disabled payment provider callbacks fail closed and do not change order state.
- Upload persistence after recreating the API and Web containers.
- PostgreSQL and Redis are not reachable on the VPS public address.
- From a temporary diagnostic container on `app`, confirm PostgreSQL connections fail because it is not on `db`; from `api`, confirm PostgreSQL and Redis connections succeed.
- Confirm `/metrics` is not routed by the public Nginx/Caddy gateway. Scrape it only from a private monitoring container or private tunnel.

## Abuse Controls And Monitoring

This VPS topology is appropriate for the current application size and security model when deployed with the production Compose file. It includes:

- Turnstile before newsletter/warranty verification delivery.
- Redis IP, destination, daily, and global sliding-window delivery budgets with a short circuit breaker.
- Redis payment-failure windows, risk scoring, and a two-failure delay response for card-like checkout attempts.
- Prometheus metrics for request volume, verification delivery/rejections, payment outcomes, and risk delays.

Carding protection is not a complete provider-native fraud program. Before meaningful payment volume, configure Stripe Radar or the equivalent provider controls, 3DS/SCA where applicable, webhook signature verification, provider decline-rate monitoring, and a manual review path. The two-second delay is deliberately an abuse-cost control, not authorization.

The repository includes Prometheus alert rules for:

- payment failure rate above 20% for five minutes with sustained traffic;
- unusual verification delivery volume;
- global verification budget exhaustion;
- elevated risk-based payment delays.

The VPS Compose stack does not expose Prometheus or Alertmanager publicly. Run them on a private monitoring network or use a private tunnel, then route critical alerts to the chosen Telegram, DingTalk, Slack, or email receiver. Never publish `/metrics` to the internet.

## Release And Rollback

Normal release:

1. Push the tested commit to `master`.
2. Wait for the publish-images workflow to finish.
3. For MCP deployment, call `VPS_getVirtualMachinesV1`, confirm VM `1834903`, call `VPS_getProjectListV1`, confirm the existing `h-gripe-theme` project, then call `VPS_updateProjectV1` for that project. For SSH deployment, run `./deploy.sh` on the VPS for an immutable SHA release.
4. Verify all containers and smoke tests.

For a Hostinger Docker Manager release, keep `IMAGE_TAG=master` and run Project Update.

Do not use `VPS_createNewProjectV1` for routine releases. It is only for initial project creation or deliberate project replacement. A normal release must update the existing `h-gripe-theme` Compose project so PostgreSQL, Redis, uploads, networks, and environment stay attached to the same production boundary.

If a local helper command is unavailable, for example Windows blocks the bundled `rg.exe`, use an equivalent read-only PowerShell command such as `Get-ChildItem ... | Select-String ...`. Treat local tooling failures as local diagnostics, not as proof that the production deployment target is wrong.

Rollback:

1. Run `DEPLOY_REF=<previous-full-commit> ./deploy.sh` using a previously published deployment commit.
2. Re-run health, login, upload, and checkout smoke tests.

Database migrations need their own rollback plan. Image rollback cannot reverse an incompatible schema migration.

## Backups

Before accepting production data, establish:

- Daily PostgreSQL logical backups.
- Daily backup of `h-gripe-theme-uploads`.
- Off-VPS copy of both backup sets.
- Monthly restore exercise.

Hostinger snapshots are disaster recovery aids, not substitutes for database and upload backups.
