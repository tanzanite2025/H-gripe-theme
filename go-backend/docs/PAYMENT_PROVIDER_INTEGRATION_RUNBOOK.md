# Payment Provider Integration Runbook

This document records the long-term integration boundaries for Stripe, PayPal, Alipay, and WeChat Pay. It is written for implementation, deployment, and handover. Keep customer-facing copy brand-neutral.

The production stack can be healthy before provider integrations are fully
live. See `../../docs/ops/production-readiness-status.md` for the current
release boundary and the external configuration items that remain open.

## Architecture Boundary

Payment providers are explicit customer-selected checkout methods. The system must not silently move a customer from one provider to another after risk evaluation or provider failure.

Current supported providers:

- Stripe: card payment with PaymentIntent and 3DS policy support.
- PayPal: redirect approval flow, capture flow, and verified webhook finalization.
- Alipay: redirect payment flow with asynchronous notify verification.
- WeChat Pay: Native QR payment flow with asynchronous API v3 notify verification.

The following actions are intentionally not automatic:

- No automatic refund execution from risk thresholds alone.
- No automatic fallback to another payment channel.
- No global hard shutdown of checkout from monitoring alone.
- No marking an order as paid from an unsigned frontend callback alone.

Webhook or provider-side verified query remains the source of truth for paid state.

## Runtime Endpoints

Storefront checkout creates a local order first, then creates a provider order/session.

| Provider | Create Endpoint | Confirm/Capture Endpoint | Webhook/Notify Endpoint |
| --- | --- | --- | --- |
| Stripe | `POST /api/v1/payment/stripe/payment-intents` | Payment Element confirm on client | `POST /api/v1/payments/stripe/webhook` |
| PayPal | `POST /api/v1/payment/paypal/orders` | `POST /api/v1/payment/paypal/orders/:paypal_order_id/capture` | `POST /api/v1/payments/paypal/webhook` |
| Alipay | `POST /api/v1/payment/alipay/orders` | `POST /api/v1/payment/alipay/orders/:order_number/confirm` | `POST /api/v1/payments/alipay/webhook` |
| WeChat Pay | `POST /api/v1/payment/wechat/orders` | `POST /api/v1/payment/wechat/orders/:order_number/confirm` | `POST /api/v1/payments/wechat/webhook` |

Before the local order is created, the storefront sends the latest backend quote
as exact `expected_total_minor` (copied from `total_minor`) in `POST /api/v1/orders`.
The order service recomputes the quote inside the creation transaction. Any
minor-unit mismatch returns HTTP `409` with `code=order_total_changed`; no order,
stock deduction, loyalty spend, or provider payment is started. The storefront
refreshes the quote and asks the customer to review the updated total.

The API also passes the authenticated customer's explicit checkout cart ID into
order creation. The order transaction locks the cart row and cart-item rows,
rebuilds the quote from that locked snapshot, reserves stock, creates the local
order, and consumes exactly those cart rows before commit. This is the
concurrency boundary for card, PayPal, Alipay, WeChat Pay, and any future
asynchronous gateway. A browser return page or webhook must not be responsible
for consuming the cart.

If a second checkout request reaches the same cart after the first transaction
commits, it receives HTTP `409` with
`code=checkout_cart_already_consumed`. The request must not create another
order or deduct stock. The storefront reloads the backend cart state and tells
the customer to check the existing order; it does not issue a broad cart-clear
request that could remove products added during the first checkout.

If an unpaid order is cancelled or reaches payment expiration, the rollback
transaction restores the consumed order items to the recorded checkout cart
and reverses the other reservations. A paid order deliberately keeps the cart
consumed. If the cart row was deleted later, the order's nullable cart reference
is cleared by the foreign key and no restoration is attempted.

The confirm endpoints are for user experience and recovery. Provider notify/webhook must still be configured in production because browser redirects can be closed, blocked, delayed, or replayed.

Payment webhook payloads are capped at 1MiB at the application boundary before provider dispatch.

## Required Environment Variables

Environment variables are the fallback runtime source. Encrypted admin settings can override them when `PAYMENT_CONFIG_MASTER_KEY` is configured.

### Shared Runtime

- `SERVER_BASE_URL`: trusted backend public base URL used to build provider webhook/notify URLs.
- `PAYMENT_CONFIG_MASTER_KEY`: a deployment-owned random secret that enables encrypted admin payment settings. It is not issued by Stripe, PayPal, Alipay, or WeChat Pay.

Generate it once before starting the API:

```bash
openssl rand -hex 32
```

On PowerShell, use:

```powershell
$bytes = New-Object byte[] 32
[Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
[Convert]::ToHexString($bytes).ToLower()
```

Put the result in the backend environment as
`PAYMENT_CONFIG_MASTER_KEY=<generated-value>`. In the production Docker deployment,
the value belongs in `deployment/production.env`; the local `start-dev.ps1`
development flow supplies a development-only fallback. Restart the Go API after
changing it. Keep the value stable: losing or rotating it makes previously
encrypted admin payment credentials unreadable.

In production, `SERVER_BASE_URL` must be the externally reachable HTTPS backend origin. Do not rely on request `Host` headers for provider notify URLs behind a proxy or CDN.

The admin runtime readiness panel must show callback URLs derived from the same `SERVER_BASE_URL`. The URL copied into Stripe, PayPal, Alipay, or WeChat Pay dashboards should exactly match the runtime readiness callback URL and the URL generated during payment creation.

The admin callback reachability probe sends a minimal unsigned `POST` to the selected provider webhook URL. A provider signature failure response such as HTTP 400 or 401 is expected and means the request reached the application webhook boundary. HTTP 404/405 indicates a routing or proxy path problem. DNS failures, connection timeouts, TLS failures, or edge blocks mean the public callback URL is not reachable from the backend runtime.

### Stripe

- `STRIPE_API_KEY` or `STRIPE_SECRET_KEY`
- `STRIPE_PUBLISHABLE_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `STRIPE_ENVIRONMENT=production`
- Optional: `STRIPE_3DS_MODE=automatic|any|challenge`

### PayPal

- `PAYPAL_CLIENT_ID`
- `PAYPAL_SECRET`
- `PAYPAL_WEBHOOK_ID`
- `PAYPAL_ENVIRONMENT=production`

### Alipay

- `ALIPAY_APP_ID`
- `ALIPAY_PRIVATE_KEY`
- `ALIPAY_PUBLIC_KEY`
- `ALIPAY_ENVIRONMENT=production`

### WeChat Pay

- `WECHAT_MCH_ID`
- `WECHAT_APP_ID`
- `WECHAT_PRIVATE_KEY_PATH`
- `WECHAT_MERCHANT_SERIAL`
- `WECHAT_API_V3_KEY`
- Recommended callback verifier: `WECHAT_PAY_PLATFORM_PUBLIC_KEY` plus its matching `WECHAT_PAY_PLATFORM_PUBLIC_KEY_ID` from the WeChat Pay merchant platform.
- Legacy fallback only: `WECHAT_PAY_PLATFORM_CERTIFICATE`. The verifier uses the public-key pair whenever a public key is configured; it uses the certificate only when no public key is configured.
- `WECHAT_ENVIRONMENT=production`

The platform public key is the preferred API v3 notification verification mode; it replaces the static platform-certificate mode for this integration, but it is not an unmanaged, never-rotating key. Follow WeChat Pay's official key lifecycle notices and update the public key and matching ID together. This application does not automatically download or rotate either verifier material. A configured public key with a missing ID or invalid PEM fails verification and does not fall back to the certificate.

When migrating from a platform certificate, obtain the current **WeChat Pay platform public key** and its exact ID from the merchant platform, save both in the encrypted admin WeChat Pay settings, and confirm runtime readiness. Before enabling or resuming production callbacks, test both a payment-success notification and a refund-success notification with valid signed provider test payloads. Confirm each is accepted and processed once; also verify an invalid signature is rejected. Repeat the checks after every key rotation. Keep the certificate only as a temporary compatibility fallback while migrating, and track its expiry if it remains configured.

## Admin Encrypted Config Fields

When `PAYMENT_CONFIG_MASTER_KEY` is configured, operators can save provider secrets from the admin payment settings panel. Secrets are write-only and are not returned to the browser.

The admin connection flow is intentionally a single-merchant credential flow:

1. Log in to the provider's official developer dashboard and create or select the merchant application.
2. Copy the provider credentials into the admin payment settings page.
3. Copy the runtime readiness callback URL back into the provider webhook settings.
4. Run the callback probe and complete a sandbox payment before enabling production.

The current product is not a multi-merchant platform. It does not implement Stripe Connect OAuth,
PayPal partner referral/onboarding, or another provider account-login/token exchange. An official
dashboard link is provided in the admin UI to help the operator obtain the credentials; it is not
an OAuth account binding. Adding provider OAuth/Connect onboarding is a later phase and requires
tenant ownership, redirect/callback handling, token rotation, disconnect/reconnect behavior, and
provider partner/platform approval where applicable.

| Provider | Required Admin Fields |
| --- | --- |
| Stripe | `api_key`, `publishable_key`, `webhook_secret` |
| PayPal | `client_id`, `secret`, `webhook_id` |
| Alipay | `app_id`, `private_key`, `public_key` |
| WeChat Pay | `mch_id`, `app_id`, `private_key_path`, `merchant_serial`, `api_v3_key`, and preferably `platform_public_key` + matching `platform_public_key_id`; `platform_certificate` is a legacy fallback |

Use the runtime readiness panel before enabling a provider. Production readiness should be green only when both active payment creation and webhook verification credentials are present.

Saving encrypted admin credentials with `environment=production` requires the typed confirmation `PRODUCTION`. Deleting an encrypted gateway config requires a provider-scoped confirmation such as `DELETE STRIPE`, `DELETE PAYPAL`, `DELETE ALIPAY`, or `DELETE WECHAT`. These confirmations are enforced by the backend API, not only by the admin UI.

Generic admin settings endpoints must not write, read, or delete `payment_gateway_*` keys or the `payment_secret` group. Payment gateway secrets must go through the dedicated gateway config endpoints so encryption, confirmation, runtime readiness, and audit logging stay together.

When encrypted admin payment config exists and `PAYMENT_CONFIG_MASTER_KEY` is
configured, the runtime payment path uses that domain-managed config before
falling back to environment variables. This applies to provider order/session
creation, provider query/capture confirmation, webhook/notify verification,
runtime readiness display, and manual provider refund execution. Only a missing
domain-managed config falls back to environment variables; unreadable or
provider-mismatched encrypted config is treated as a configuration error.

PayPal currently uses the hosted checkout flow in this product. The backend creates and captures
PayPal orders and verifies PayPal webhooks, but it does not receive the buyer's complete card
number or CVV. Therefore the BIN-level card-testing limiter intentionally does not apply to
PayPal. This is a PCI/security boundary, not a payment-account configuration defect. PayPal
still requires compensating controls such as IP, session, account, order, payer identity, and
provider-failure velocity limits. A future PayPal card-fields/advanced-card-payment integration
would be a separate provider and compliance project, not a change to the current BIN limiter.

Manual provider refund execution passes a strict refund reference set into the
gateway adapter: merchant order number, provider transaction id, original
transaction currency, and original transaction amount. There is no legacy
fallback path because production orders have not been created yet.

Refunds are executed in the original transaction currency. The refund path now
captures an immutable order-time FX snapshot in `orders.fx_snapshot` and carries
it onto `refunds.fx_snapshot` so the historical base-currency cap stays fixed.
This still needs end-to-end acceptance before it should be described as fully
live, and it must not be sourced from mutable `products.display_prices` or
`product_variants.display_prices` JSON.

Provider-specific refund references are strict:

- PayPal refunds use the stored PayPal capture id only. A stored PayPal order id
  is not accepted as a refund reference.
- Alipay refunds use `trade_no` only. `out_trade_no` remains the merchant order
  number and is not used as the refund trade reference.
- WeChat Pay refunds use `transaction_id` only and send both refund amount and
  original transaction amount, as required by the provider API.

Refund records have two independent scopes and can coexist on the same order:

- An amount-only refund is a monetary adjustment or compensation. It has no
  `refund_line_items` rows and never restores inventory.
- An item-level refund identifies `order_item_id` and refunded quantity.
  Inventory restoration is evaluated per line item and only occurs when that
  line item has `restock=true`.
- The system does not enforce a whole-order exclusion between these refund
  types. A compensation refund may precede a returned-item refund, or the
  returned-item refund may precede the compensation. Transaction-level refund
  caps and per-order-item quantity caps still apply independently.

Successful provider finalization must write refundable platform transaction
references into the completed transaction row. PayPal must record the capture
id, Alipay must record `trade_no`, and WeChat Pay must record `transaction_id`.
Pending or in-progress attempts may use provider order numbers or merchant order
numbers for traceability, but those rows are not refundable until a verified
success writes the platform transaction id.

Duplicate successful payments have a separate ledger rule:

- The same provider transaction ID delivered more than once is an idempotent
  webhook retry. It reuses the existing transaction and refund state.
- A different verified provider transaction ID received after the order is
  already paid is a real duplicate charge. It is recorded as
  `transactions.status = duplicate_paid` with the original provider amount,
  currency, response, and transaction ID.
- The same transaction then receives one full-amount local
  `refunds.status = pending` record. The refund is linked to that duplicate
  transaction and is idempotent on later webhook retries.
- The webhook returns a success acknowledgement only after those local writes
  commit. This acknowledgement confirms durable local accounting, not
  completion of a provider refund.
- The existing explicit admin refund execution workflow remains responsible for
  calling Stripe, PayPal, Alipay, or WeChat Pay and moving the local refund to
  `completed`.

Refund loyalty settlement has one shared accounting boundary:

- `refunds.requested_amount` is the original refund request; `refunds.amount`
  is the net amount sent to the provider after coupon and loyalty deductions.
- Completed-order reward points are clawed back proportionally to the
  cumulative requested refund amount. Points spent as an order discount are
  returned by the same proportion.
- If the customer's available balance cannot cover the earned-point clawback,
  the missing points are converted with the active `ExchangeRatePoints` and
  deducted from the provider refund before the gateway call.
- The refund ID is the idempotency key for the clawback, reversal, and
  used-point-return ledger entries. Repeated provider notifications do not
  apply them twice.
- A provider call failure before a provider refund ID is returned releases the
  reservation. A provider response with a refund ID but a mismatched amount is
  retained as a failed execution for manual reconciliation; the reservation is
  not released because money may already have moved.
- A verified provider refund with no local pending refund must carry both the
  provider-confirmed net amount and the original requested amount. The service
  does not guess the original request from the provider net amount.

High-risk admin payment actions are written to the audit log:

- Saving encrypted gateway config records provider, environment, production flag, submitted/configured field names, status, failure stage, operator, IP, user agent, and request path.
- Deleting encrypted gateway config records provider, confirmation result, status, operator, IP, user agent, and request path.
- Running callback reachability probes records provider, callback URL, HTTP status, route reachability, expected signature failure state, operator, IP, user agent, and request path.
- Creating, updating, or deleting payment methods records method id, code, fee/min/max amounts, enabled flag, sort order, settings presence/length, old-value summary when available, operator, IP, user agent, and request path.
- Updating the storefront display-currency policy records display currency count/list, available catalog count/list, old/new summaries when available, operator, IP, user agent, and request path.
- Creating a local refund draft records order, transaction, requested/net amount, line-item count, restock count, status, operator, IP, user agent, and request path. It does not execute a provider refund.
- Executing a pending refund records refund, order, transaction, merchant order number, provider transaction id, provider, amount, currency, execution status, attempt count, provider-refund-id presence, operator, IP, user agent, and request path.
- Updating refund recommendations records recommendation, provider, source kind, requested status, linked refund id when present, and operator metadata. It does not call a provider.
- Submitting Stripe dispute evidence records dispute id, submit/stage mode, evidence-file presence flags, statement presence/length, result status, operator, IP, user agent, and request path.
- Creating or updating payment reviews records review id, status, linked order/transaction/dispute ids, payment-intent presence, assignee, operator, IP, user agent, and request path.
- Recomputing payment risk monitoring and creating/revoking manual payment protection controls record their operational inputs and outcome.

Audit payloads must never store raw credential values, webhook secrets, private keys, API v3 keys, payment-method settings JSON/text, typed confirmation text, refund reasons, dispute narrative statements, internal review notes, or customer communication content. Only field names, presence flags, lengths, identifiers, and operational metadata are safe to keep.

## Currency Boundary

Product, SKU, shipping, and other commercial configuration amounts use the backend primary pricing currency. The admin enters the primary amount directly; secondary display prices are filled or refreshed by a backend one-click action from cached ExchangeRate-API rates. Product and SKU price currency is the source of truth for local order currency. Storefront display currencies are only for price labels, exchange-rate hints, and admin pricing helpers.

The admin pricing-currency page owns the primary base currency and secondary display currency list. Exchange-rate sync reads that policy directly at runtime. The API management page only stores the provider enable flag/API key and triggers cache sync; it must not become a second currency-management surface.

Storefront payment buttons come from configured and enabled payment methods, plus runtime protection controls. If PayPal is configured and enabled, the storefront can show PayPal; if Stripe, WeChat Pay, or Alipay is configured and enabled, the storefront can show those buttons. Storefront display currency, IP-based market currency, and ExchangeRate-API conversion hints must not decide which payment buttons are exposed.

Binding a provider is the button policy, not a product pricing currency policy. The local product or variant price determines the order transaction currency. After the shopper selects PayPal, Stripe, WeChat Pay, Alipay, or another method, that provider decides whether the customer's card/account/funding currency can complete payment under its own account and regional rules. Provider order/session creation still uses the local order amount and order currency, and provider callbacks/webhooks must verify signature, idempotency, amount, and currency against that order before marking it paid.

Display-currency policy updates are audited because they affect storefront presentation. The audit record stores normalized ISO currency codes and counts only; it does not store raw request bodies.

## Risk Control Boundary

`pause_payment` controls can hide or block a provider, country, or payment method before provider SDK creation. This is the correct long-term boundary for operational intervention.

Examples:

- Pause `provider=stripe` during acquiring incidents.
- Pause `country=XX` during abuse spikes.
- Pause `payment_method=wechat_pay` during provider maintenance.

Do not implement automatic provider failover inside these controls. A customer who selected one provider should see a clear unavailable state and choose another method manually.

## Production Acceptance Checklist

Before enabling a provider in production:

- Runtime readiness shows required credentials configured.
- Public webhook URL is reachable over HTTPS.
- Provider dashboard points to the exact webhook URL shown by runtime readiness.
- Signature verification succeeds in sandbox or provider test mode.
- Provider ACK format is accepted:
  - PayPal receives a normal HTTP 2xx response after processing.
  - Alipay receives a plain text `success` body after successful notification verification and processing.
  - WeChat Pay receives HTTP 204 with no response body after successful API v3 notification verification and processing.
- A small live payment creates a local order, consumes the cart in the order transaction, creates a provider payment, receives provider notify/webhook, records a transaction, and marks the order paid. The browser return is not required for cart correctness.
- The completed transaction row stores a refundable platform reference: Stripe payment intent id, PayPal capture id, Alipay `trade_no`, or WeChat Pay `transaction_id`.
- Duplicate delivery of the same provider transaction ID is idempotent and does
  not create duplicate transactions.
- A different successful provider transaction ID after the order is paid is
  recorded as `duplicate_paid` and receives a full-amount pending refund record;
  it must not be silently acknowledged as an ordinary duplicate webhook.
- Payment creation and webhook/capture handling verify provider amount and currency against the local order before marking it paid.
- `pause_payment` for provider scope blocks active payment creation before provider SDK calls.

## Operational Notes

PayPal, Alipay, and WeChat Pay are asynchronous checkout methods. Their local
order transaction consumes the backend cart before provider creation. Provider
confirmation settles the order; it does not perform the cart-consumption
operation. Browser return pages may reload stale local UI state, but a closed
browser must not leave the backend cart available for a second checkout.

WeChat Native payment returns a QR code URL. The storefront generates the QR image locally in the browser and stores only the short payment session in `sessionStorage`.

Alipay browser return is not enough by itself. Treat provider notify verification as the final settlement signal.
