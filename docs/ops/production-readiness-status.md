# Production Readiness Status

Last updated: 2026-09-21.

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
- Backend regression baseline passes with `go test ./... -count=1`; transactional
  email code paths, after-sales return shipment fields, Sitemap catalog entries,
  and media derivative capacity have automated coverage.

## Not Completed Yet

These items must not be described as live production integrations until their
credentials are supplied and the acceptance checks below are completed.

| Area | Current state | Required next step |
| --- | --- | --- |
| Stripe | Provider credentials are not configured in production. | Save the real production API key, publishable key, and webhook secret in the admin payment settings; configure the exact HTTPS webhook URL; complete a small live or provider-approved test payment and refund. |
| PayPal | Provider credentials are not configured in production. | Save the real production client ID, secret, and webhook ID; configure the webhook; complete approval, capture, webhook, refund, and dispute-evidence tests. |
| Alipay / WeChat Pay | No production credential acceptance has been recorded. | Configure provider credentials and complete notify/signature/settlement tests before enabling the methods. For WeChat Pay API v3 callbacks, prefer the official platform public key plus its matching key ID; the static platform certificate is a legacy fallback with expiry and no automatic rotation in this application. |
| SMTP / transactional email | Provider configuration, encrypted credentials, template management, Outbox rules, and SMTP test-send code are implemented; production credentials and customer-facing acceptance are not recorded. | Configure a real SMTP account, verify email challenge delivery (newsletter, warranty), and complete transactional email template acceptance tests (order confirmation, shipping, after-sales status notifications) per `../design/transactional-email-notification-status.md`. |
| Turnstile | Production enforcement is currently disabled because real site/secret keys were not supplied. | Configure real Turnstile keys, enable enforcement, and test the protected verification flows. |
| Google Merchant Center | OAuth variables are intentionally empty, so startup is not blocked. | Configure client ID, client secret, redirect URL, post-connect URL, and token encryption key as one complete set before enabling the channel. |
| Multi-merchant provider login | Not implemented. The current product is a single-merchant encrypted-credential flow, not Stripe Connect or PayPal partner onboarding. | Treat provider OAuth/Connect onboarding as a separate product phase requiring tenant ownership, callback/token lifecycle, disconnect/reconnect, and provider approval. |
| PayPal dispute file evidence | Structured evidence, commercial invoice PDF generation, and carrier POD PDF upload/PayPal `documents` submission are implemented. A dispute still requires an operator to attach and verify the carrier-issued PDF before submission. | Upload a carrier-issued PDF to the mutable `signed_pod` evidence item, verify the generated HTTPS document entry in the evidence preview, then submit the dispute package. |
| Stripe / PayPal webhook event subscriptions | Runtime readiness exposes the concrete event names, the 14-item Stripe operator checklist (`review.opened` / `review.closed` are one checklist item), and the official dashboard URL. The application does not currently call provider APIs to read back and compare the events enabled on the configured webhook endpoint; `webhook_events_verified` therefore remains `false` and is not proof of dashboard configuration. | In the Stripe/PayPal dashboard, subscribe the full checklist from `payment.RequiredWebhookEventChecklist` (including both review event names), record the manual check in the release/operations record, and run a signed webhook test for payment, refund, and dispute events. |
| BIN limiter | BIN-level card testing protection applies only where the backend receives card details. PayPal hosted checkout is intentionally outside this boundary. | Keep PayPal protected by IP/session/account/order/payer/provider-failure controls unless a separate PayPal card-fields compliance project is approved. |
| Historical FX refund guard | Code captures immutable order-time snapshots and copies them to refunds. The order detail page provides an audited backfill form for historical non-USD orders; valid snapshots cannot be overwritten. Refund API errors direct operators to that form. | Follow the "Historical Non-USD Refund Recovery" procedure below, then retry the refund creation or execution. |

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

### Webhook Event Verification Boundary

This is a limitation of the current application, not a claim that Stripe or
PayPal APIs can never return webhook subscription settings. Provider APIs can
expose the event types configured on a webhook endpoint, but the application
does not currently make those read-back requests and compare the result with
its required event list. Runtime readiness only returns the locally defined
checklist and the provider dashboard link; it does not inspect the merchant's
dashboard, and `webhook_events_verified: false` means "not checked by this
application," not "known to be missing."

Until provider-specific read-back verification is implemented and accepted,
an operator must compare the endpoint's configured events with the checklist
in the Stripe or PayPal dashboard and record that manual check in the release
or operations record. A successful signed test webhook verifies delivery and
signature handling for that test event only; it does not prove that every
required event type is subscribed. Automatic verification would require
provider-specific API calls, suitable API permissions, and the identifiers for
the exact webhook endpoint being checked.

### Historical Non-USD Refund Recovery

When refund creation or execution reports `historical refund FX snapshot is
missing`, do not retry with a guessed or current market rate. The refund is
blocked intentionally because the historical snapshot is part of the refund
amount/cap calculation.

1. Open the order in Admin > Orders and use the **Historical FX Snapshot
   Backfill** section on the order detail overview. This section is available
   for non-USD orders to staff with order-edit permission. Staff without that
   permission must escalate the case to Finance or an administrator.
2. Finance must verify the order-time rate from an archived exchange-rate
   record, provider statement, or other approved payment/settlement evidence.
   Record the evidence source; never substitute today's rate.
3. Enter the base and order currencies, rate, source, and snapshot capture
   time. The rate convention is `1 base currency = N order currency` (for
   example, `1 USD = 0.92 EUR`). Use the historical snapshot/checkout time,
   usually the order creation time, not the time of this repair.
4. Save the snapshot, then retry the normal refund creation or execution flow.
   The endpoint is `POST /api/admin/payment/orders/:order_id/fx-snapshot`; the
   write is row-locked and audited. It rejects USD orders, currency mismatch,
   invalid rates, and attempts to replace an already-valid snapshot.

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
