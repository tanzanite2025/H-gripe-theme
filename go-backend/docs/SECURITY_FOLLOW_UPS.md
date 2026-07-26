# Backend Security Follow-Ups

Last updated: 2026-07-26

This file tracks security work that is not fully closed by code-only mitigations. The items below are mandatory before treating the related public flows as fully hardened.

## Current mitigations

- User-owned resource reads and deletes enforce owner/admin checks.
- Public responses are being trimmed so private, disabled, unpublished, or sensitive facts are not exposed accidentally.
- High-risk public endpoints use generic responses and rate limits where immediate product changes are not yet in place.
- SMS/email verification delivery uses Turnstile, Redis-backed IP and destination windows, daily destination quotas, a global sliding-window budget, and a short circuit breaker.
- The current product wiring covers email challenge delivery for newsletter and warranty flows. No SMS provider or SMS OTP endpoint is enabled; any future SMS route must use the same `Guard` path with `channel=sms` before provider dispatch.
- Checkout/card-like payment attempts use Redis failure windows, risk scoring, and a two-failure delay policy. Prometheus metrics and alert rules cover payment failures, risk delays, verification volume, and global-budget exhaustion.
- Production Compose separates PostgreSQL into an internal `db` network and Redis into an internal `cache` network. PostgreSQL is reachable only by the backend `api` and one-shot `migrate` containers; no business container publishes a host port.

These mitigations reduce immediate exposure. They do not fully close public flows that intentionally rely on possession of an identifier or email address as proof.

## Product-security implementation status

| Area | Current shape | Required target | Acceptance criteria |
| --- | --- | --- | --- |
| Public shipment tracking | Implemented: `GET /api/v1/shipping/track/:tracking_number` now requires authentication and checks tracking shipment -> order -> current user, with admin/order-view bypass. | Keep the raw tracking number out of unauthenticated access. Preserve generic not-found behavior for non-owned shipments. | Raw tracking number alone cannot read shipment history. Tests cover raw unauthenticated access, wrong owner, valid owner, and admin access. |
| Newsletter subscription flows | Implemented: new subscriptions remain `pending`; confirmation, unsubscribe, resubscribe, and status actions use signed, expiring, single-use email challenges. Email-only requests do not mutate state. | Configure SMTP and `STOREFRONT_BASE_URL` in each deployed environment. Add an end-to-end test that proves the email link reaches the intended frontend/API flow. | A subscription is not active before confirmation. Tokens expire and replay safely. Responses remain enumeration-safe. |
| Warranty order verification and claims | Implemented: order-number/email verification requests are generic; warranty claim creation requires a matching signed, expiring, single-use email challenge. | Configure SMTP and `STOREFRONT_BASE_URL`, then add an end-to-end test from request email through storefront token verification and claim submission. | No claim is created without the matching challenge. Wrong order/email, expired, and replayed tokens are rejected. |

## Abuse-control boundary

The current controls are suitable for a VPS deployment, but they are layered abuse controls rather than a promise that carding is impossible:

- Carding protection currently includes session/user failure windows, a two-failure response delay, country/VPN/UA risk signals, provider failure metrics, and Prometheus alerts. It does not replace Stripe Radar or another provider-native decision, 3DS/SCA, velocity rules, BIN/IP/device intelligence, or manual review.
- The browser request signature is an anti-abuse and scraping signal. Its key is delivered to the browser and must not be treated as a secret, authentication factor, or payment authorization boundary.
- Do not expose PostgreSQL, Redis, `/metrics`, or internal service ports on the VPS public interface. Monitoring must scrape the API over a private Docker network or a private tunnel.
- The production Compose file is the only supported VPS topology. The root development Compose file must not be used for deployment.

## Release blocking follow-ups

These items remain mandatory product-security follow-ups even though the current code has mitigations:

1. Public tracking lookup must use authenticated ownership, a short-lived signed token, or an order-number plus email verification challenge. A raw tracking number must not be treated as durable authorization.
2. Newsletter subscribe, unsubscribe, resubscribe, and status flows must remain behind double opt-in or signed, expiring email tokens. CAPTCHA and rate limits alone are not proof of mailbox ownership.
3. Warranty order-number plus email verification must remain behind email confirmation or a second proof factor. CAPTCHA and matching fields alone are not a complete claim-authorization factor.

## Mandatory release/security scans

These checks are mandatory before a production image is released. The existing local security script may be used as a starting point, but release readiness requires the exact built image and repository history to be scanned.

| Area | Required target | Acceptance criteria |
| --- | --- | --- |
| Container image vulnerability scanning | Scan the final backend runtime image with both Trivy and Grype using the exact tag or digest that will be deployed. File-system or Dockerfile-only scans are not enough for release approval. | Reports are generated and retained for the release. Critical/high findings fail the release unless a documented exception includes exploitability analysis, owner, expiry, and remediation ticket. The scan command records the image tag or digest. |
| Historical secrets scanning | Run gitleaks against the full Git history, not only the current working tree. The scan must include all branches/tags or the agreed release history boundary. | The report is retained. Any confirmed secret is revoked or rotated, not only removed from files. False positives are documented with stable allowlist rules. Release approval records the gitleaks version, command, and commit range. |

## VPS release evidence

For each release, retain:

- the exact image tag or digest deployed;
- Trivy and Grype reports for that exact image;
- the gitleaks report and scanned history boundary;
- the Compose config output showing no business-service host ports;
- a connectivity check proving PostgreSQL is reachable from `api`/`migrate` only and not from `web`, `storefront`, `admin`, or the public host interface;
- Prometheus alert rules and the notification receiver used for payment and verification alerts.

## Operating rule

- Generic responses and rate limits are temporary mitigations for these product-strategy risks, not final closure.
- Any public launch or security readiness review must treat unresolved items in this file as blocking for the corresponding public flow.
- Production release approval must include Trivy, Grype, and gitleaks evidence for the exact image/history being released.
