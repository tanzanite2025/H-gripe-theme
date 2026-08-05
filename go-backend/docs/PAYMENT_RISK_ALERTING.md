# Payment Risk Alerting

## Purpose

Payment-risk monitoring evaluates provider-neutral operational indicators over
a rolling window. This alerting path notifies an external operations system
only when a provider's evaluated level changes.

It does not create refunds or change payment providers. Checkout protection is
limited to audited, time-bounded manual controls such as `force_3ds` and
`pause_payment`; it does not perform automatic provider failover or permanent
global checkout shutdown.

## Current Payment Protection Boundary

The current delivery includes:

- provider-level risk snapshots and rolling-window alerts;
- persisted risk events and idempotent Outbox delivery;
- time-bounded, audited manual `force_3ds` and `pause_payment` controls;
- webhook-driven refund recommendation queue for operator review;
- manual dispute evidence preparation and submission.

The current delivery does not automatically refund payments, switch payment
providers, or close checkout permanently. `pause_payment` is evaluated before
new payment starts in two concrete places: public order creation
(`POST /api/v1/orders`) and Stripe PaymentIntent creation
(`POST /api/v1/payment/stripe/payment-intents`). PayPal provider-order
creation (`POST /api/v1/payment/paypal/orders`) uses the same guard before
invoking the PayPal SDK. Alipay and WeChat Pay provider-order creation
(`POST /api/v1/payment/alipay/orders` and
`POST /api/v1/payment/wechat/orders`) also use the same guard before invoking
their provider SDKs. The control affects only new starts in the matching
scope and must have an expiry time; it does not cancel existing provider
sessions, alter provider confirmation/webhook completion, or block admin
refund execution.
The admin panel treats these controls as high-risk operations: create and
revoke flows require an explicit typed confirmation and are recorded through
the protection-control audit log. The Admin API also requires `confirm=true`
for both creation and revoke calls, mirroring the manual refund execution
boundary.

Manual control duration is policy-bound by action. The default configuration
allows ordinary controls such as `force_3ds` for up to 168 hours, scoped
`pause_payment` controls for up to 24 hours, and global `pause_payment` for up
to 2 hours. The service validates this server-side and the admin UI exposes
the active limit while creating a control.

`GET /api/v1/payment/methods` also evaluates the same manual protection state
as a read model. It returns compatibility fields such as `provider`,
`available`, and `unavailable_reason` so storefronts can gray out a temporarily
paused method before the customer starts payment. This is advisory UI state;
the payment-start guard remains the enforcement point.
Storefronts should consume this through a small payment-method availability
adapter and pass normalized options into the checkout selector, rather than
duplicating response-shape handling inside UI components. Existing provider
sessions that were already created are not cancelled by this advisory state.

PayPal, Alipay, and WeChat Pay checkout are explicit payment paths, not
routing fallbacks. The storefront creates a provider session only after the
customer has selected that method on the local order. Every provider endpoint
requires the local order to belong to the authenticated customer and to use the
matching selected payment method. Confirmation verifies that the provider
response still references the local order number before marking the order paid.

Storefront PayPal flow is split into three explicit steps:

- create the local order with `payment_method = paypal`;
- create the provider order and redirect the buyer to PayPal approval;
- on `/checkout/paypal/return`, capture the PayPal order by `token` and local
  `order_number`.

Storefront Alipay flow is split into three explicit steps:

- create the local order with `payment_method = alipay`;
- create the provider payment through `POST /api/v1/payment/alipay/orders` and
  redirect the buyer to Alipay;
- on `/checkout/alipay/return`, query/confirm the provider trade through
  `POST /api/v1/payment/alipay/orders/:order_number/confirm`.

Storefront WeChat Pay flow is split into three explicit steps:

- create the local order with `payment_method = wechat`;
- create the Native QR payment through `POST /api/v1/payment/wechat/orders`;
- on `/checkout/wechat/pay`, display the QR code and poll
  `POST /api/v1/payment/wechat/orders/:order_number/confirm` until the provider
  reports `SUCCESS`.

Local cart cleanup is deferred for asynchronous gateway methods such as Stripe
PayPal, Alipay, and WeChat Pay. The storefront clears the cart only after the
selected provider confirms success. Cancellation, incomplete return, or QR
timeout leaves the local order unpaid and does not trigger automatic refunding,
channel switching, or a retry through a different provider.

Refund recommendations are suggestions and audit records by default. Accepting
a recommendation only records the decision. A separate operator action can
create a local `pending` refund draft linked to that recommendation, but it
still does not call a payment provider. A second explicit operator action can
execute that pending refund through the configured payment gateway with an
idempotency key and execution audit row. Automatic refunding, provider routing,
and broader circuit breakers are later, separately gated capabilities and must
not be described as enabled production behavior.

## Future Provider Routing Boundary

Provider failover should be added only after there is a dedicated routing
layer, not by having risk monitoring silently change checkout behavior. The
future shape should be a separate, audited control plane that stores candidate
providers, route priority, supported currencies/countries, provider health, and
time-bounded operator approvals. Risk monitoring may recommend a routing
change, but the live route should change only through an explicit control with
rollback and audit history.

Hard checkout closure should follow the same rule. The current `pause_payment`
control is a scoped, temporary block on new payment starts. A later global
circuit breaker should be a distinct action with a short maximum duration,
operator confirmation, customer-facing messaging, and exclusion paths for
webhook processing, refunds, and admin evidence work.

## Refund Recommendation Queue

Provider risk webhooks now create operator-facing refund recommendations:

- Stripe Early Fraud Warning events create high-priority recommendations to
  review whether a manual refund should be created before dispute escalation.
- Stripe and PayPal dispute events create recommendations to review the refund,
  evidence-response, or customer-contact path.
- Repeated provider webhook delivery is idempotent by
  `provider + source_kind + external_reference`.
- If a later dispute event indicates the case is no longer actionable, the
  pending recommendation is cancelled instead of remaining open.
- The queue stores provider, risk-event reference, order/transaction linkage
  when resolvable, recommended amount, currency, reason, priority, review
  deadline, and operator decision metadata.

Implemented endpoints:

- `GET /api/admin/payment/risk/refund-recommendations`
- `PATCH /api/admin/payment/risk/refund-recommendations/:id`
- `POST /api/admin/payment/risk/refund-recommendations/:id/pending-refund`
- `POST /api/admin/payment/refunds/:id/execute`

Supported decision statuses are `pending`, `accepted`, `dismissed`, and
`cancelled`. `accepted` only records that an operator agrees with the
recommendation. The `pending-refund` endpoint requires refund permission and
`confirm=true`; it creates only a local `refunds.status = pending` record,
stores `linked_refund_id`, `refund_created_by_id`, and `refund_created_at` on
the recommendation, and does not contact the payment gateway.

Provider refund execution is a separate manual step. The `refunds/:id/execute`
endpoint requires refund permission and `confirm=true`, writes or reuses a
`payment_refund_executions` row, sends a stable idempotency key to supported
gateways, and only marks the local refund `completed` after the gateway returns
a provider refund id.

These admin actions also write global audit-log entries with operator, IP,
user agent, request path, status, failure reason, and safe operational
metadata:

- updating a refund recommendation decision;
- creating a local pending refund from a recommendation;
- creating a direct local refund draft;
- executing a pending provider refund;
- recomputing risk monitoring snapshots;
- creating or revoking manual `force_3ds` / `pause_payment` controls;
- updating the checkout currency policy;
- creating, updating, or deleting admin payment methods;
- creating manual visitor-risk decisions from daily risk facts;
- creating or updating payment reviews;
- submitting or staging Stripe dispute evidence.

The audit-log payload stores identifiers, statuses, amount/currency metadata,
presence flags, and text lengths. It must not store refund reasons, dispute
narratives, customer-communication content, payment-review notes, provider
secrets, payment-method settings JSON/text, visitor-risk decision reason text,
raw IPs, raw user agents, full visitor hash values, or typed confirmation text.

## 3DS Timing

Application risk checks run before the server creates a payment intent. The
payment provider performs 3DS during payment confirmation, before the payment
can complete. A successful provider webhook is the source for the final
payment state; refunds, early-fraud handling, and dispute workflows run after
that state change.

The application may use IP, session, account, country, address, velocity, and
other non-sensitive signals before payment confirmation. The client IP is
resolved through the configured trusted proxy chain. Edge-provided country
headers are accepted only from configured trusted proxy peers; untrusted
headers are ignored. IP or country alone must not be treated as proof of
fraud.

## Delivery Model

1. The monitoring worker computes a provider snapshot.
2. The snapshot and the provider's current level state are written in one
   database transaction.
3. When alert delivery is enabled and the level changed, the same transaction
   writes a `payment.risk_level_changed` Outbox event.
4. The existing Outbox worker delivers the event to the configured webhook.

The external receiver must use `event_key` as its idempotency key.

## Alert Semantics

- A transition from `normal` to `warning` or `critical` emits an alert.
- A transition between `warning` and `critical` emits an alert.
- A recovery from `warning` or `critical` to `normal` emits an alert.
- Repeated snapshots at the same level do not emit duplicate alerts.
- While alerting is disabled, snapshots and provider state still update, but
  no alert event is created.

The reported levels are internal operational policy. They are not payment
provider program thresholds or guarantees.

## Webhook Configuration

All of the following are required before alert events are generated:

```dotenv
PAYMENT_RISK_MONITORING_ALERT_ENABLED=true
PAYMENT_RISK_ALERT_OUTBOX_WEBHOOK_URL=https://alerts.example.internal/payment-risk
PAYMENT_RISK_ALERT_OUTBOX_WEBHOOK_TOKEN=replace-with-a-rotatable-secret
```

The dispatcher posts JSON with:

- `Authorization: Bearer <token>` when a token is configured.
- `X-Outbox-Event-Key` for receiver-side idempotency.
- `X-Outbox-Event-Type: payment.risk_level_changed`.

Do not enable `PAYMENT_RISK_MONITORING_ALERT_ENABLED` until the webhook URL is
configured and the Outbox dispatch worker is running.

## Payload Boundary

The event payload is intentionally limited to provider-level operational data:

- provider and previous/current level;
- rolling-window payment, dispute, early-fraud-warning, and refund counts;
- the three calculated rates;
- reasons, recommended action, and computation time.

It excludes customer data, order IDs, transaction IDs, payment identifiers,
and raw provider webhook payloads.
