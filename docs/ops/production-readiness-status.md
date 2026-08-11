# Production Readiness Status

Last updated: 2026-08-11.

This document is the current handover boundary for the deployed production
stack. A healthy deployment means the application, database, cache, migrations,
network boundary, and public HTTP routes are running. It does not mean that
every external provider or customer-facing workflow has been configured and
live-tested.

## Verified In Production

- The release image set is pinned to a published Git commit SHA.
- PostgreSQL, Redis, API, storefront, admin, and web services are healthy.
- Database migrations complete successfully.
- `PAYMENT_CONFIG_MASTER_KEY` is present in the API runtime and enables
  encrypted admin payment settings.
- The VPS Compose boundary has no business-service host ports.
- PostgreSQL and Redis remain on private Docker networks.
- Storefront HTML cache purge and public health/API/admin routes respond
  successfully.

## Not Completed Yet

These items must not be described as live production integrations until their
credentials are supplied and the acceptance checks below are completed.

| Area | Current state | Required next step |
| --- | --- | --- |
| Stripe | Provider credentials are not configured in production. | Save the real production API key, publishable key, and webhook secret in the admin payment settings; configure the exact HTTPS webhook URL; complete a small live or provider-approved test payment and refund. |
| PayPal | Provider credentials are not configured in production. | Save the real production client ID, secret, and webhook ID; configure the webhook; complete approval, capture, webhook, refund, and dispute-evidence tests. |
| Alipay / WeChat Pay | No production credential acceptance has been recorded. | Configure the provider-specific keys/certificates and complete notify/signature/settlement tests before enabling the methods. |
| SMTP | The deployed environment uses placeholder or disabled SMTP values. | Configure a real SMTP account and verify newsletter, warranty, and other email challenge delivery. |
| Turnstile | Production enforcement is currently disabled because real site/secret keys were not supplied. | Configure real Turnstile keys, enable enforcement, and test the protected verification flows. |
| Google Merchant Center | OAuth variables are intentionally empty, so startup is not blocked. | Configure client ID, client secret, redirect URL, post-connect URL, and token encryption key as one complete set before enabling the channel. |
| Multi-merchant provider login | Not implemented. The current product is a single-merchant encrypted-credential flow, not Stripe Connect or PayPal partner onboarding. | Treat provider OAuth/Connect onboarding as a separate product phase requiring tenant ownership, callback/token lifecycle, disconnect/reconnect, and provider approval. |
| PayPal dispute file evidence | Structured evidence and the commercial invoice PDF path exist. Official carrier POD PDF and other external document attachments are not fully integrated. | Add a carrier/document adapter and submit verified POD/customer-communication files through PayPal `documents`; do not claim those files are attached before this is implemented. |
| BIN limiter | BIN-level card testing protection applies only where the backend receives card details. PayPal hosted checkout is intentionally outside this boundary. | Keep PayPal protected by IP/session/account/order/payer/provider-failure controls unless a separate PayPal card-fields compliance project is approved. |
| Historical FX refund guard | Code is now wired to capture an immutable order-time FX snapshot in `orders.fx_snapshot` and copy it to `refunds.fx_snapshot`, but the non-USD refund path still needs end-to-end acceptance. | Keep this as a follow-up until a non-USD order, refund creation, and refund execution run has been verified against the stored snapshot and cap. |

## Payment Configuration Boundary

`PAYMENT_CONFIG_MASTER_KEY` is generated and owned by the deployment, not issued
by Stripe or PayPal. It only unlocks encrypted storage for credentials; it does
not create, bind, or authenticate a provider account.

The current production setup requires an operator to:

1. Create or select the merchant application in the provider's official
   dashboard.
2. Enter the provider credentials in the admin payment settings panel.
3. Copy the runtime callback URL into the provider dashboard.
4. Run callback, payment, webhook, and refund acceptance checks.

The admin panel can show whether an encrypted gateway configuration exists, but
it must not display raw secrets.

## Release Language

Use these terms precisely:

- **Production runtime deployed:** the stack and public routes are healthy.
- **Provider configured:** required provider credentials and callback settings
  are present.
- **Provider production-accepted:** a real or provider-approved payment,
  callback/webhook, and refund test passed.
- **Fully live:** all enabled providers, email/security challenges, and
  customer-facing external integrations have passed their acceptance checks.

The current release has reached the first state. It must not be reported as the
last three states without the corresponding external configuration and tests.
