# Backend Security Follow-Ups

## Production deployment versus full integration

The VPS runtime has been deployed and its service health, migration status,
public routes, and Docker network boundary were verified. This is not a claim
that every external integration is live. The current unresolved production
items are tracked in
[`docs/ops/production-readiness-status.md`](../../docs/ops/production-readiness-status.md).
In particular, real payment-provider credentials, SMTP, Turnstile enforcement,
Google Merchant OAuth, provider callback/payment/refund acceptance tests, and
the remaining PayPal document adapters are still required before the related
flows can be called fully live.

Last updated: 2026-08-11

This file tracks security work that is not fully closed by code-only mitigations. The items below are mandatory before treating the related public flows as fully hardened.

## Current mitigations

- User-owned resource reads and deletes enforce owner/admin checks.
- Public responses are being trimmed so private, disabled, unpublished, or sensitive facts are not exposed accidentally.
- High-risk public endpoints use generic responses and rate limits where immediate product changes are not yet in place.
- SMS/email verification delivery uses Turnstile, Redis-backed IP and destination windows, daily destination quotas, a global sliding-window budget, and a short circuit breaker.
- The current product wiring covers email challenge delivery for newsletter and warranty flows. No SMS provider or SMS OTP endpoint is enabled; any future SMS route must use the same `Guard` path with `channel=sms` before provider dispatch.
- Checkout/card-like payment attempts use Redis failure windows, risk scoring, and a two-failure delay policy. Prometheus metrics and alert rules cover payment failures, risk delays, verification volume, and global-budget exhaustion.
- BIN-level card testing limits are not yet implemented. The current controls
  do not inspect raw card PAN/BIN data, and PayPal hosted checkout is
  intentionally outside any BIN-level boundary.
- Production Compose separates PostgreSQL into an internal `db` network and Redis into an internal `cache` network. PostgreSQL is reachable only by the backend `api` and one-shot `migrate` containers; no business container publishes a host port.

These mitigations reduce immediate exposure. They do not fully close public flows that intentionally rely on possession of an identifier or email address as proof.

## Product-security implementation status

| Area | Current shape | Required target | Acceptance criteria |
| --- | --- | --- | --- |
| Public shipment tracking | Implemented and route-tested: `GET /api/v1/shipping/track/:tracking_number` requires authentication and checks tracking shipment -> order -> current user, with admin/order-view bypass. | Keep the raw tracking number out of unauthenticated access. Preserve generic not-found behavior for non-owned shipments. | Raw tracking number alone cannot read shipment history. Tests cover raw unauthenticated access, wrong owner, valid owner, and admin access. |
| Newsletter subscription flows | Implemented and route-tested: new subscriptions remain `pending`; confirmation, unsubscribe, resubscribe, and status actions use signed, expiring, single-use email challenges. Email-only requests do not mutate state. | Configure SMTP and `STOREFRONT_BASE_URL` in each deployed environment. | A subscription is not active before confirmation. Tokens expire and replay safely. Responses remain enumeration-safe. |
| Warranty order verification and claims | Implemented and route-tested: order-number/email verification requests are generic; the storefront token verification route unlocks claim submission; claim creation requires a matching signed, expiring, single-use email challenge. | Configure SMTP and `STOREFRONT_BASE_URL` in each deployed environment. Keep the challenge and second-factor flow enabled. | The route test covers request email, storefront token verification, multipart claim submission, and token replay rejection. No claim is created without the matching challenge. |

## Abuse-control boundary

The current controls are suitable for a VPS deployment, but they are layered abuse controls rather than a promise that carding is impossible:

- Carding protection currently includes session/user failure windows, a two-failure response delay, country/VPN/UA risk signals, provider failure metrics, and Prometheus alerts. It does not replace Stripe Radar or another provider-native decision, 3DS/SCA, velocity rules, BIN/IP/device intelligence, or manual review.
- The browser request signature is an anti-abuse and scraping signal. Its key is delivered to the browser and must not be treated as a secret, authentication factor, or payment authorization boundary.
- Do not expose PostgreSQL, Redis, `/metrics`, or internal service ports on the VPS public interface. Monitoring must scrape the API over a private Docker network or a private tunnel.
- The production Compose file is the only supported VPS topology. The root development Compose file must not be used for deployment.

## Release blocking follow-ups

These items remain mandatory product-security follow-ups even though the current code has mitigations:

1. Newsletter production release requires verified `SMTP_*` and `STOREFRONT_BASE_URL` configuration in each deployed environment; subscribe, unsubscribe, resubscribe, and status flows must remain behind double opt-in or signed, expiring email tokens.
2. Warranty production release requires verified `SMTP_*` and `STOREFRONT_BASE_URL` configuration in each deployed environment; order-number plus email verification must remain behind email confirmation or a second proof factor. CAPTCHA and matching fields alone are not a complete claim-authorization factor.
3. Production payment enablement requires real provider credentials, exact HTTPS callback configuration, and a provider-approved payment/webhook/refund acceptance run. `PAYMENT_CONFIG_MASTER_KEY` only enables encrypted credential storage; it does not bind a provider account.
4. The historical FX refund guard is implemented in code with `orders.fx_snapshot` and `refunds.fx_snapshot`, but it is still a release follow-up until a non-USD order/refund execution run proves the stored snapshot and cap. Do not use mutable catalog display prices as the refund source of truth.

## Mandatory release/security scans

These checks are mandatory before a production image is released. The existing local security script may be used as a starting point, but release readiness requires the exact built image and repository history to be scanned.

The release workflow in `.github/workflows/publish-images.yml` already enforces the container-image control: it builds each `linux/amd64` release image, scans the exact local tag with pinned Trivy `0.58.2` and Grype `0.116.0` containers, uploads both reports, and blocks the push when either scanner reports a high or critical finding. The workflow implementation does not replace release evidence; reports for the exact release tag or digest must still be retained.

Historical secret scanning is also enforced in `.github/workflows/go-backend-ci.yml` and `.github/workflows/publish-images.yml`: both check out full history, run the pinned gitleaks `v8.24.2` container with `.gitleaks.toml` and `--log-opts="--all --tags"`, upload the SARIF report, and fail the job when scanning fails or returns findings. The current allowlist is limited to documented examples and test-only WeChat key material; any real secret still requires revocation or rotation.

| Area | Required target | Acceptance criteria |
| --- | --- | --- |
| Container image vulnerability scanning | Scan the final backend runtime image with both Trivy and Grype using the exact tag or digest that will be deployed. File-system or Dockerfile-only scans are not enough for release approval. | Reports are generated and retained for the release. Critical/high findings fail the release unless a documented exception includes exploitability analysis, owner, expiry, and remediation ticket. The scan command records the image tag or digest. |
| Historical secrets scanning | Run gitleaks against the full Git history, not only the current working tree. The scan must include all branches/tags or the agreed release history boundary. | The report is retained. Any confirmed secret is revoked or rotated, not only removed from files. False positives are documented with stable allowlist rules. Release approval records the gitleaks version, command, and commit range. |

## VPS release evidence

Use `deployment/verify-vps-release-boundary.sh` from the production checkout to
create the boundary evidence. Run it once before deployment for the static
Compose check, then run it with `CHECK_CONNECTIVITY=true` after all services are
healthy for runtime network and host-port checks. The verifier reads resolved
network names from `docker compose config --format json` and writes a
timestamped, sanitized report directory under `release-evidence/`.

For each release, retain:

- the exact image tag or digest deployed;
- Trivy and Grype reports for that exact image;
- the gitleaks report and scanned history boundary;
- the sanitized Compose JSON showing no business-service host ports and the resolved network names;
- a connectivity check proving PostgreSQL is reachable from `api`/`migrate` only and not from `web`, `storefront`, `admin`, or the public host interface;
- Prometheus alert rules and the notification receiver used for payment and verification alerts.

## Operating rule

- Generic responses and rate limits are temporary mitigations for these product-strategy risks, not final closure.
- Any public launch or security readiness review must treat unresolved items in this file as blocking for the corresponding public flow.
- Production release approval must include Trivy, Grype, and gitleaks evidence for the exact image/history being released.
