# Storefront Theme Hostinger VPS Docker Runbook

This runbook applies only to the `commerce-platform` storefront, admin, and Go API. The ERP application remains a separate Compose project with separate images, databases, volumes, and production secrets.

## Production Boundary

The Hostinger VPS has one shared public gateway project named `shared-edge`. The storefront joins that gateway network as a separate Compose project named `commerce-platform`.

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
      -> erp.legacy.example -> erp-web:8080
      -> generated storefront domains -> theme-web:8080
      -> generated admin domains      -> theme-web:8080

commerce-platform project
  -> web -> storefront
         -> admin
         -> api (dedicated api_ingress path)
                -> PostgreSQL
                -> Redis
                -> uploads volume
```

The shared gateway is infrastructure. It is not the ERP application and it is not the storefront application.

## Current Gateway State

The neutral gateway name is `shared-edge`. The storefront joins that shared
edge network through the `web` service only.

1. Keep the shared gateway running.
2. Keep the enabled production domains from `ops_domain_bindings` routed to
   `theme-web`; do not maintain a second static hostname list in the gateway.
3. Keep the ERP route separate.
4. Do not recreate the old `commerce-platform-edge` project or reintroduce old public hostnames into the active boundary.

Browser API requests stay same-origin through `web`. Nuxt server-side requests to `/api/**` use Nitro's internal proxy and go directly to `api:9000`, so SSR does not loop through Cloudflare or the public gateway.

## Production Files

| File | Purpose |
| --- | --- |
| `compose.prod.yml` | Hostinger production Compose project |
| `deployment/production.env.example` | Production environment template |
| `deployment/docker/web.Dockerfile` | Internal Nginx entry image |
| `deployment/nginx/theme-web.conf` | Same-origin API, storefront, admin, and upload routing |
| `deployment/edge/commerce-platform.caddy` | Shared Caddy security fragment and generated-route import |
| `docs/archive/ops/learn-gripe-cutover-completed.md` | Completed `learn.gripe` domain cutover record |
| `deployment/verify-vps-release-boundary.sh` | Static and runtime Compose/network release evidence |
| `.github/workflows/publish-images.yml` | GHCR image publishing |
| `deploy.sh` | Recommended SSH deployment path |

The root `docker-compose.yml` remains a local development convenience and must not be deployed through Hostinger Docker Manager.

## Project Isolation

The production project must keep these boundaries:

1. Project name: `commerce-platform`.
2. Images: `${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/commerce-platform-*`.
3. Volumes: `commerce-platform-postgres-data`, `commerce-platform-redis-data`, and `commerce-platform-uploads`.
4. Data networks: project-owned internal `db` and `cache` networks.
5. Application network: project-owned `app` network for service-to-service traffic and required outbound integrations.
6. API ingress network: project-owned internal `api_ingress`, joined only by `api` and `web`.
7. Shared network: external `shared-edge`, joined only by `web`.
8. Edge alias: `theme-web`.
9. No `container_name` and no host `ports` in the business stack.
10. No ERP environment variables, volumes, image tags, or database credentials.
11. `web` is pinned to `172.30.0.10` on `api_ingress`; `TRUSTED_PROXIES` is exactly `172.30.0.10/32`. The shared Caddy gateway remains responsible for strict Cloudflare proxy trust.

The Compose file now defaults to the `commerce-platform` project and volume names. Existing
VPS resources named `commerce-platform-*` are not renamed automatically. During a
volume migration, set `COMPOSE_PROJECT_NAME`, `POSTGRES_DATA_VOLUME_NAME`,
`REDIS_DATA_VOLUME_NAME`, and `UPLOADS_VOLUME_NAME` in the untracked production env
to the old names until the backups and data copy have been verified. Do not start
the new boundary against empty volumes before that migration is complete.

Database and cache reachability is intentional:

- `db` is internal and contains only `db`, `api`, and the one-shot `migrate` and `edge-config` services. `storefront`, `admin`, and `web` cannot resolve or connect to PostgreSQL through Docker networking.
- `cache` is internal and contains `redis`, `api`, and `storefront` because the SSR HTML cache is Redis-backed. Redis has a password and is never published to the host.
- `app` is not marked `internal` because the API and storefront need outbound SMTP, Turnstile, payment, and registry-related traffic. No service in the business stack publishes a host port.
- `api_ingress` is internal and contains only `api` and `web`. Nginx resolves API upstream traffic through the `api-ingress` alias, and the Go API trusts only the pinned `web` address for forwarded client metadata. Storefront SSR traffic still reaches `api:9000` over `app` and is not treated as a trusted proxy.
- `edge` is the only public gateway path. Only `web` joins it; the API, database, Redis, storefront, and admin are not directly reachable from Cloudflare or the VPS interface.

## Publish Images

Every commit pushed to `master` produces one complete deployable release. GitHub Actions validates the backend and both frontends, then publishes:

- `${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/commerce-platform-api`
- `${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/commerce-platform-storefront`
- `${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/commerce-platform-admin`
- `${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/commerce-platform-web`

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

The script fetches `origin/master`, resolves the full commit SHA, waits for all four matching GHCR images, validates Compose, force-recreates and runs the database migration job, force-recreates the edge configuration job from the migrated database, validates the generated artifacts, and only then updates the shared gateway. Run it immediately after a push or after GitHub Actions completes.

In Hostinger Docker Manager create:

```text
Project name: commerce-platform
Compose source: compose.prod.yml
Environment: deployment/production.env values, including IMAGE_TAG=master
```

Hostinger Docker Manager cannot derive the Git commit tag itself. When using the Manager or MCP Project Update, keep `IMAGE_TAG=master`; the publish workflow moves that tag after validation. Set `IMAGE_TAG=sha-<full-commit>` only for a deliberate pinned release or rollback.

Before deployment, confirm the external Docker network `shared-edge` exists.

Keep `TRUSTED_PROXIES=172.30.0.10/32`. The production Compose file pins the
`web` container to that address on `api_ingress`; do not add RFC1918 supernets,
`0.0.0.0/0`, or `::/0`.

Expected services:

- `db`
- `redis`
- `migrate` (one-shot)
- `edge-config` (one-shot; must complete successfully)
- `api`
- `storefront`
- `admin`
- `web`

The long-running application services must become Healthy. `migrate` and
`edge-config` are one-shot services and must exit successfully. Only the
`shared-edge` gateway may publish host ports.

### Database Migration Release Gate

Every production release executes the `migrate` service before any application
container is allowed to start or update. It is the only release-mode schema
writer; API containers run with `DB_AUTO_MIGRATE=false` and must never be used
to apply a schema change opportunistically.

The `migrate` job uses the same API image as the release and waits for
PostgreSQL health. It is a direct `service_completed_successfully` dependency
of `edge-config`, `api`, `storefront`, `admin`, and `web`. A failed migration
therefore blocks the API, both frontends, the shared web entrypoint, and
generated edge configuration from rolling forward.

There are two supported release mechanisms, with one migration contract:

1. `deploy.sh` force-recreates and waits for `migrate` on every release before
   it starts application services. It also force-recreates `edge-config` after
   the migration completes.
2. Hostinger Docker Manager and MCP Project Update must update the complete
   `commerce-platform` Compose project, never an individual service. The
   project pulls the release API image and Compose runs the same one-shot
   migration gate before creating any dependent service.

An image publication builds all four production images for every `master`
commit, including the API image used by `migrate`. This ensures a normal
Hostinger Project Update receives a new release image and re-runs the job even
for a storefront-, admin-, or documentation-only release. Do not manually
start `api`, `storefront`, `admin`, or `web` to bypass a failed migration, and
do not use `go run`, `DB_AUTO_MIGRATE=true`, or an ad hoc SQL session as a
substitute for the release gate.

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

Merge `deployment/edge/commerce-platform.caddy` into the existing `shared-edge`
Caddyfile without changing the ERP route. The fragment imports the generated
route file at `/etc/caddy/generated/commerce-platform.caddy`.

`deploy.sh` writes the generated route to the configured
`EDGE_GATEWAY_ROUTE_FILE` only after database migration, edge rendering,
standalone Caddy validation, and generated Nginx/manifest checks pass. It then
validates and reloads the shared gateway. If rendering fails, the existing
gateway route is not touched; if gateway validation or reload fails, the
previous route is restored.

## DNS State

The canonical public domain and its aliases come from the enabled production
rows in `ops_domain_bindings`. Keep DNS, TLS, Caddy, Nginx, CORS, and public URL
environment values aligned with the generated manifest instead of copying a
hostname into multiple files.

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
GET <generated-admin-domain>/  -> admin login page
GET /uploads/<known-file>      -> static file response
WebSocket Upgrade              -> reaches customer-service authentication
```

### Customer-Service Realtime

Customer-service browser delivery is WebSocket-only. The production environment
must keep the durable relay paired with the SQL Outbox dispatcher on every API
replica:

```text
CUSTOMER_SERVICE_REALTIME_ENABLED=true
WORKER_OUTBOX_DISPATCH_ENABLED=true
WORKER_OUTBOX_DISPATCH_INTERVAL_SECONDS=2
```

All API replicas must share the project Redis deployment and retain the Stream
name `customer_service:{realtime}:v1` unless the hash-tagged name is changed as
one coordinated deployment. `WORKER_ENABLED` controls unrelated Asynq workers;
it does not enable the SQL Outbox scheduler.

The bundled Nginx `/api/` routes already forward `Upgrade`, `Connection`,
cookies, `Host`, and `X-Forwarded-Proto` to the API with a 120-second read
timeout. Preserve those directives in generated or hand-maintained edge
changes. The API emits WebSocket pings every 50 seconds, so any upstream idle
timeout must remain above that heartbeat interval.

Scrape `/metrics` privately and alert on non-zero
`commerce_platform_customer_service_realtime_outbox_events{status="dead_letter"}`;
also investigate sustained `failed`/`pending` counts, Relay read failures,
WebSocket capacity rejections, and outbound queue overflows. Before each
horizontal-scale release, perform the two-instance write/restart/reconnect/HTTP
reconciliation drill documented in
`go-backend/docs/CUSTOMER_SERVICE_RELIABLE_REALTIME_ARCHITECTURE.md`.

Also verify:

- Customer login, refresh, logout, and CSRF behavior.
- Admin login and token refresh.
- Product list and product detail.
- Checkout quote and order creation.
- Disabled payment provider callbacks fail closed and do not change order state.
- Upload persistence after recreating the API and Web containers.
- If object storage is used, keep the bucket private and verify anonymous
  access to `showcase/pending/<known-key>` returns `403` or `404`; the API
  service account must be the only writer and must be able to copy approved
  objects into `showcase/approved/`.
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
3. For MCP deployment, call `VPS_getVirtualMachinesV1`, confirm VM `1834903`, call `VPS_getProjectListV1`, confirm the existing `commerce-platform` project, then call `VPS_updateProjectV1` for that project. For SSH deployment, run `./deploy.sh` on the VPS for an immutable SHA release.
4. Verify all containers and smoke tests.

For a Hostinger Docker Manager release, keep `IMAGE_TAG=master` and run Project Update.

Do not use `VPS_createNewProjectV1` for routine releases. It is only for initial project creation or deliberate project replacement. A normal release must update the existing `commerce-platform` Compose project so PostgreSQL, Redis, uploads, networks, and environment stay attached to the same production boundary.

If a local helper command is unavailable, for example Windows blocks the bundled `rg.exe`, use an equivalent read-only PowerShell command such as `Get-ChildItem ... | Select-String ...`. Treat local tooling failures as local diagnostics, not as proof that the production deployment target is wrong.

Rollback:

The admin deployment workflow can execute the rollback after a production
workflow enters `rollback_required`, but only after the backend has been
configured with a dedicated SSH key and host verification file. The executor
does not accept a host, user, workdir, or shell command from the browser.

Required backend environment:

```text
OPS_DEPLOY_ROLLBACK_ENABLED=true
OPS_DEPLOY_ROLLBACK_SSH_HOST=srv1834903.hstgr.cloud
OPS_DEPLOY_ROLLBACK_SSH_PORT=22
OPS_DEPLOY_ROLLBACK_SSH_USER=deploy
OPS_DEPLOY_ROLLBACK_SSH_PRIVATE_KEY_PATH=/run/secrets/commerce-platform-deploy.key
OPS_DEPLOY_ROLLBACK_SSH_KNOWN_HOSTS_PATH=/run/secrets/commerce-platform-known_hosts
OPS_DEPLOY_ROLLBACK_SSH_WORKDIR=/srv/commerce-platform
OPS_DEPLOY_ROLLBACK_SSH_TIMEOUT_SECONDS=900
```

The configured host must match the production VPS binding (`srv1834903.hstgr.cloud`
or `2.25.85.201`). The private key and `known_hosts` file must be readable by
the backend process, and `known_hosts` must contain the exact SSH host key.
Keep both files outside Git and outside browser-visible settings.

Admin rollback flow:

1. Confirm the workflow is `rollback_required` and that its rollback point is a full 40-character Commit SHA.
2. Confirm the project name, VPS binding, target host, and impact in the confirmation dialog.
3. With `ops:deploy:rollback`, choose `执行回滚`. The backend runs the fixed command `cd -- <configured-workdir> && DEPLOY_REF=<full-sha> ./deploy.sh` over SSH.
4. The workflow re-syncs the Hostinger project, runs DNS/HTTP/HTTPS health checks, purges bound Cloudflare hosts, and records the remote operation ID and bounded command output.
5. If SSH, verification, or cache purge fails, the workflow remains `rollback_required` with the failed step and error evidence. A successful SSH command is not repeated when the next action only needs verification.

Manual fallback:

```bash
DEPLOY_REF=<previous-full-commit> ./deploy.sh
```

Use a previously published deployment commit and then re-run health, login,
upload, and checkout smoke tests.

Database migrations need their own rollback plan. Image rollback cannot reverse an incompatible schema migration.

## Backups

Before accepting production data, establish:

- Daily PostgreSQL logical backups.
- Daily backup of `commerce-platform-uploads`.
- Off-VPS copy of both backup sets.
- Monthly restore exercise.

Hostinger snapshots are disaster recovery aids, not substitutes for database and upload backups.
