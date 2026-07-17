# Tanzanite Theme Hostinger VPS Docker Runbook

This runbook applies only to the `tanzanite-theme` storefront, admin, and Go API. The ERP application remains a separate Compose project with separate images, databases, volumes, and production secrets.

## Production Boundary

The Hostinger VPS has one shared public gateway project named `tanzanite-edge`. Tanzanite Theme joins that gateway network as a separate Compose project named `tanzanite-theme`.

```text
Cloudflare
  -> Hostinger firewall: 80/443
  -> tanzanite-edge (shared Caddy)
      -> erp.tanzanite.site -> erp-web:8080
      -> tanzanite.site     -> theme-web:8080
      -> admin.tanzanite.site -> theme-web:8080

tanzanite-theme project
  -> web -> storefront
         -> admin
         -> api -> PostgreSQL
                -> Redis
                -> uploads volume
```

The shared gateway is infrastructure. It is not the ERP application and it is not the Tanzanite Theme application.

Browser API requests stay same-origin through `web`. Nuxt server-side requests to `/api/**` use Nitro's internal proxy and go directly to `api:9000`, so SSR does not loop through Cloudflare or the public gateway.

## Production Files

| File | Purpose |
| --- | --- |
| `compose.prod.yml` | Hostinger production Compose project |
| `deployment/production.env.example` | Production environment template |
| `deployment/docker/web.Dockerfile` | Internal Nginx entry image |
| `deployment/nginx/theme-web.conf` | Same-origin API, storefront, admin, and upload routing |
| `deployment/edge/tanzanite-theme.caddy` | Route fragment for the shared Caddy gateway |
| `.github/workflows/publish-images.yml` | GHCR image publishing |
| `deploy.sh` | Recommended SSH deployment path |

The root `docker-compose.yml` remains a local development convenience and must not be deployed through Hostinger Docker Manager.

## Project Isolation

The production project must keep these boundaries:

1. Project name: `tanzanite-theme`.
2. Images: `ghcr.io/tanzanite2025/tanzanite-theme-*`.
3. Volumes: `tanzanite-theme-postgres-data`, `tanzanite-theme-redis-data`, and `tanzanite-theme-uploads`.
4. Internal networks: project-owned `data` and `app` networks.
5. Shared network: external `tanzanite-edge`, joined only by `web`.
6. Edge alias: `theme-web`.
7. No `container_name` and no host `ports` in the business stack.
8. No ERP environment variables, volumes, image tags, or database credentials.
9. `TRUSTED_PROXIES` contains only private Docker CIDRs. The shared Caddy gateway remains responsible for strict Cloudflare proxy trust.

## Publish Images

Every commit pushed to `master` produces one complete deployable release. GitHub Actions validates the backend and both frontends, then publishes:

- `ghcr.io/tanzanite2025/tanzanite-theme-api`
- `ghcr.io/tanzanite2025/tanzanite-theme-storefront`
- `ghcr.io/tanzanite2025/tanzanite-theme-admin`
- `ghcr.io/tanzanite2025/tanzanite-theme-web`

Production uses the immutable full tag `sha-<40-character-commit>` generated from the tested commit. Publishing on every `master` commit is intentional: `deploy.sh` always resolves `origin/master`, so the branch head must always have a matching four-image release. The GHCR packages must be public or the VPS must have read-only registry credentials.

`.github/workflows/go-backend-ci.yml` and `.github/workflows/ci.yml` are validation only. They must not publish production images or deploy Kubernetes. `.github/workflows/publish-images.yml` is the only production image publisher.

## Create The Hostinger Project

Create `deployment/production.env` from the example and replace every `CHANGE_ME` value. Keep the real file outside Git. The file does not contain `IMAGE_TAG`; `deploy.sh` derives it from Git.

The recommended release path is a repository clone on the VPS:

```bash
./deploy.sh
```

The script fetches `origin/master`, resolves the full commit SHA, waits for all four matching GHCR images, validates Compose, and recreates the project only after every image is available. Run it immediately after a push or after GitHub Actions completes.

In Hostinger Docker Manager create:

```text
Project name: tanzanite-theme
Compose source: compose.prod.yml
Environment: deployment/production.env values plus IMAGE_TAG=sha-<full-commit>
```

Hostinger Docker Manager cannot derive the Git commit tag itself. When using the Manager instead of `deploy.sh`, update `IMAGE_TAG` manually for every release.

Before deployment, confirm the external Docker network `tanzanite-edge` exists.

Keep `TRUSTED_PROXIES` at the private Docker ranges from the example unless the Docker network design is intentionally changed. Do not add `0.0.0.0/0` or `::/0`.

Expected services:

- `db`
- `redis`
- `api`
- `storefront`
- `admin`
- `web`

All six services must become Healthy. Only the shared `tanzanite-edge` gateway may publish host ports.

## Add Shared Gateway Routes

Merge `deployment/edge/tanzanite-theme.caddy` into the existing `tanzanite-edge` Caddyfile without changing the ERP route.

Required routes:

```text
tanzanite.site, www.tanzanite.site -> theme-web:8080
admin.tanzanite.site               -> theme-web:8080
```

Updating the shared gateway is a separate infrastructure operation. Do not copy the Tanzanite application Compose into the ERP project and do not replace the gateway project.

## Cloudflare Cutover

The existing root A and AAAA records still point to the old PHP host. Perform the cutover only after all Tanzanite containers are Healthy and the gateway route exists.

1. Leave `erp.tanzanite.site` unchanged.
2. Update the existing root A record to the Hostinger VPS IPv4.
3. Remove the two old root AAAA records. Add the VPS IPv6 only after direct IPv6 TLS validation.
4. Keep `www.tanzanite.site` as a CNAME to the root domain.
5. Add `admin.tanzanite.site` as a proxied record targeting the same VPS.
6. Temporarily use DNS only while Caddy obtains the first certificate if required.
7. Restore Proxied after direct HTTPS verification.
8. Verify `Full (strict)`, Minimum TLS 1.2, Always Use HTTPS, and WebSockets.
9. Bypass Cloudflare cache for `/api/*`, `/uploads/*`, payment webhooks, and WebSocket traffic.

Cloudflare changes for the root site must not modify the ERP A record or ERP host-specific rules.

## Deliberately Disabled Integrations

Do not add payment provider secrets to the production environment yet. The current generic payment webhook endpoint is not a complete provider-native adapter: Stripe payload mapping is incomplete, and the PayPal and Alipay verification paths are not production-ready. With no provider secrets in the API container, these callbacks fail closed and cannot update payment state.

Outbound SMTP is also not wired into the application dependency graph, so SMTP variables are intentionally absent from the production template. Add either integration only together with its service wiring, provider-specific contract tests, and deployment variables.

## Verification

Verify before enabling public traffic:

```text
GET /                          -> storefront 200
GET /healthz                   -> theme-web 200
GET /api/v1/settings/site      -> Go API response
GET admin.tanzanite.site/      -> admin login page
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

## Release And Rollback

Normal release:

1. Push the tested commit to `master`.
2. Run `./deploy.sh` on the VPS; it waits for the matching workflow images.
3. Verify all containers and smoke tests.

For a Hostinger Docker Manager release, wait for the workflow, set `IMAGE_TAG=sha-<full-commit>`, and run Project Update.

Rollback:

1. Run `DEPLOY_REF=<previous-full-commit> ./deploy.sh` using a previously published deployment commit.
2. Re-run health, login, upload, and checkout smoke tests.

Database migrations need their own rollback plan. Image rollback cannot reverse an incompatible schema migration.

## Backups

Before accepting production data, establish:

- Daily PostgreSQL logical backups.
- Daily backup of `tanzanite-theme-uploads`.
- Off-VPS copy of both backup sets.
- Monthly restore exercise.

Hostinger snapshots are disaster recovery aids, not substitutes for database and upload backups.
