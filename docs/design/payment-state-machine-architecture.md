# 支付域状态机与故障排查契约

Last updated: 2026-09-13

## Status

Implementation contract (first version). This document defines the business
states, legal transitions, concurrency rules, money invariants, and webhook
acknowledgement semantics for the payment domain. Code changes and contract
tests must conform to this document. When the implementation and this document
disagree, the disagreement is a release blocker until one side is explicitly
updated.

This document complements
[`payment-channel-domain-architecture.md`](./payment-channel-domain-architecture.md):
that document defines channel and admin-domain boundaries; this one defines
runtime facts and transitions.

## 1. Scope and source of truth

The payment domain is an append-oriented financial record plus a set of
operational projections:

```text
Provider webhook / frontend capture
              |
              v
     signature + merchant verification
              |
              v
       Webhook Inbox claim (idempotent)
              |
              v
   transaction + order transition (one DB tx)
              |
              +--> refund / dispute / review records
              |
              +--> Outbox (email, fulfilment, conversion, audit)
```

The following rules are normative:

1. A provider fact is accepted only after signature and merchant identity
   verification.
2. A financial fact is durable before a success response is returned to the
   provider or frontend.
3. Every side effect that can be retried has an idempotency identity and is
   safe to execute more than once.
4. A state transition is performed inside a database transaction and is
   serialized on the affected order/transaction/refund row.
5. An order's financial state, fulfilment hold, and provider dispute state are
   separate dimensions. A dispute must not destroy the state needed to recover
   the order when the dispute is won or closed.

## 2. Aggregate ownership and locking

| Aggregate | Durable fact | Lock owner | External side effects |
|---|---|---|---|
| Order | payment/fulfilment projection | order row (`FOR UPDATE`) | Outbox only |
| Transaction | provider charge/attempt | transaction row by provider transaction ID | none in the DB transaction |
| Refund | requested or provider-confirmed refund | refund row and transaction row | provider refund call is outside DB tx; result returns through webhook/reconciliation |
| Dispute | provider chargeback lifecycle | order row, then dispute row | evidence submission job/outbox |
| PaymentReview | human/automatic decision | order row when decision changes fulfilment | outbox/audit after commit |
| Webhook Inbox | delivery claim and processing result | unique provider+event identity | none |
| Outbox | durable intent to notify/fulfil | event row | worker performs retryable side effect |

The canonical lock order is:

```text
Webhook Inbox claim -> Order FOR UPDATE -> Transaction/Refund/Dispute FOR UPDATE
```

No code path may lock a transaction first and then decide the order state using
a stale, unlocked order object. If a path has no order ID yet, it may resolve
the provider identity first, but it must lock and reload the order before any
financial decision.

## 3. State dimensions

### 3.1 Order

`orders.status` is the operational projection. `orders.payment_status` is the
financial projection and must not be inferred from the operational status.

#### Order operational states

| State | Meaning | Allowed transitions |
|---|---|---|
| `pending` | created, awaiting payment | `cancelled`, `payment_expired`, `paid` |
| `paid` | payment accepted, fulfilment may be started | `processing`, `needs_review`, `disputed`, `refunded`, `cancelled` |
| `processing` | fulfilment in progress | `shipped`, `cancelled`, `disputed`, `refunded` |
| `needs_review` | payment is held for an unresolved review | `processing`, `disputed`, `refunded` |
| `shipped` | shipment handed to carrier | `completed`, `disputed`, `refunded` |
| `completed` | fulfilment complete | `disputed`, `refunded` |
| `cancelled` | order cancelled before a valid payment | terminal for normal checkout; a later payment becomes a review case |
| `payment_expired` | payment window expired | terminal for normal checkout; a later payment becomes a review case |
| `disputed` | legacy/projection state only; see §7 | must retain and restore the previous operational state |
| `refunded` | the order's refundable payment has been fully refunded | terminal for the order payment projection |

#### Payment states

```text
unpaid -> paid -> refunded
   \-> expired
```

`payment_status=paid` is set only after a completed transaction has passed
currency and amount validation. `payment_status=refunded` is set only after
completed refunds cover the refundable amount. A duplicate-paid refund does not
make the original order unpaid or refunded.

`fulfillment_hold=true` is an independent guard. It is set for an active
dispute or unresolved liability review and may be cleared only when all active
holds have been resolved.

### 3.2 Transaction

The provider transaction ID is globally unique and is the idempotency identity
for a charge.

| State | Meaning | Allowed transitions |
|---|---|---|
| `pending` | attempt created, no final provider result | `processing`, `requires_action`, `completed`, `failed`, `expired` |
| `processing` | provider is still processing | `completed`, `failed`, `expired` |
| `requires_action` | customer authentication/confirmation required | `processing`, `completed`, `failed`, `expired` |
| `completed` | verified provider charge | `refunded` only when refunds cover the transaction |
| `duplicate_paid` | verified second charge after the order was already paid | remains `duplicate_paid`; creates at most one linked full-amount pending refund |
| `failed` | provider declined/failed | terminal |
| `expired` | attempt expired | terminal |
| `refunded` | original transaction fully refunded | terminal |

Receiving the same provider transaction ID again is an idempotent delivery,
not a duplicate charge. A different provider transaction ID received after the
order's locked `payment_status` is `paid` is a real duplicate charge and must
be recorded as `duplicate_paid`.

### 3.3 Refund

| State | Meaning | Transition |
|---|---|---|
| `pending` | local refund intent is durable; provider execution is outstanding | `completed`, `failed` |
| `completed` | provider refund amount and identity were verified | terminal |
| `failed` | provider rejected or permanently failed the refund | may be retried by a new execution attempt, never silently discarded |

The refund row stores both the requested/local amount and the net provider
amount, the transaction link, currency/FX snapshot, provider refund ID, and raw
response. A duplicate-paid refund is a normal `pending` refund with reason
`duplicate_paid_payment`; creation is idempotent on transaction + reason.

### 3.4 Dispute

Provider-specific statuses are normalized into these operational classes:

```text
active: needs_response, warning_needs_response, under_review,
        WAITING_FOR_SELLER_RESPONSE, INQUIRY, etc.
terminal-success: won, closed, resolved, cancelled, denied, rejected, withdrawn
terminal-loss: lost, refunded
```

The raw provider status, payload, evidence timestamps, and normalized class are
all retained. Repeated provider events update the same dispute by provider
dispute ID.

### 3.5 PaymentReview

`pending -> approved | rejected | cancelled` is the only finalization path.
Each `reason` is a policy, not a label. At minimum the implementation must
define side effects for:

| Reason | Approved | Rejected/cancelled |
|---|---|---|
| `high_value_liability_shift_not_transferred` | clear the liability hold only if no other active review/dispute remains | keep fulfilment blocked and retain audit trail |
| `payment_succeeded_after_cancellation` | create/execute the configured refund path; never silently ship a cancelled order | keep refund/exception work visible |
| `payment_succeeded_after_expiration` | same as above, with expiration retained as the original fact | keep exception visible |
| `payment_succeeded_after_refund` | reconcile the late charge against the refund; do not mark the order paid again | keep exception visible |
| Stripe automatic review reasons | release only the matching automatic review and re-evaluate holds | retain the provider decision |

An update endpoint returning success without changing the review or its defined
side effect is a contract violation.

## 4. Verified payment transition algorithm

All frontend capture confirmations and provider payment webhooks use the same
service-level algorithm:

1. Verify signature, timestamp/replay window where applicable, and merchant
   identity.
2. Claim the Webhook Inbox identity (`provider + event ID`; if a provider has
   no event ID, use a documented request key). Already `processed` events return
   success without re-running side effects; `failed` events may be reclaimed.
3. Begin one DB transaction and lock the order by order number/ID.
4. Reload the transaction by provider transaction ID under the same transaction.
5. If that transaction is already `completed`, return idempotent success. Never
   create a refund merely because the order projection is already paid.
6. Re-read the locked order's current `payment_status` and operational status.
   The first transaction to commit the order payment wins. A second, different
   transaction ID that observes `payment_status=paid` is `duplicate_paid`.
7. Resolve the expected payable currency and amount from the immutable payment
   snapshot (§5). Validate before changing any financial state.
8. For a normal payment, persist the transaction as `completed`, set the order
   payment projection to `paid`, apply any review/hold, and enqueue order-paid
   and notification Outbox events in the same DB transaction.
9. For a late payment on `cancelled`, `payment_expired`, or `refunded`, persist
   the completed transaction and a pending exception/review or refund according
   to the reason policy. Do not silently revive the order.
10. For a duplicate payment, persist `duplicate_paid`, create exactly one linked
    full-amount pending refund, and do not emit a second order-paid event.
11. Commit. Only then acknowledge the Inbox event/provider request.

The critical invariant is:

```text
duplicate_paid iff
  current locked order.payment_status == paid
  AND provider transaction ID is different from the completed transaction(s)
```

This prevents the frontend Capture/Webhook millisecond race from producing a
ghost refund for the transaction that actually won the payment transition.

## 5. Money and currency contract

The order's `Currency` and `TotalAmount` are the customer-facing order facts.
They are not automatically the amount/currency sent to every provider.

Checkout must persist an immutable payable snapshot containing, at minimum:

```text
order currency + order amount
provider/channel currency + provider amount
FX rate/source + captured-at timestamp
minor-unit rounding mode
```

Callback validation uses the provider snapshot for that provider/channel:

```text
provider currency == snapshot.provider_currency
abs(provider amount - snapshot.provider_amount) < minor_unit_tolerance
```

Comparing a CNY provider amount directly with a USD order total is invalid. If a
historical order has no provider payable snapshot, the callback must go to a
visible manual-review/reconciliation path; it must not be accepted using an
unsafe cross-currency comparison.

Provider defaults and normalization:

- Alipay domestic notifications commonly omit `currency`; an empty value is
  normalized to `CNY` after signature and merchant checks.
- WeChat notifications default an empty amount currency to `CNY`.
- Stripe and PayPal currency codes are normalized to uppercase ISO codes.
- Unknown, unsupported, or conflicting currencies fail closed and are retained
  in the Inbox/error audit for replay after correction.

All comparisons use decimal/minor-unit arithmetic and the currency's precision;
binary floating-point equality is not a business rule.

## 6. Webhook Inbox, provider parsing, and acknowledgement

Payment and refund notifications are different event schemas and must never be
forced through the same decrypted structure.

```text
verify -> classify event -> parse provider-specific payment/refund resource
       -> normalize -> invoke domain service -> persist -> acknowledge
```

### Payment events

- Success/terminal-payment events invoke the verified payment algorithm (§4).
- Non-success payment states are acknowledged only after they are either
  durably recorded as an attempt state or explicitly classified as an ignored,
  non-financial provider event.

### Refund events

- WeChat refund notifications use the refund resource schema and RSA-V3
  decryption, not the collection transaction schema.
- Alipay refund notifications use the refund-specific notification fields and
  preserve non-success provider outcomes as a local `failed`/exception fact;
  they must not be silently acknowledged as if the refund succeeded.
- PayPal `PAYMENT.CAPTURE.REFUNDED` resources map to a refund by provider
  refund ID, capture ID, order/invoice metadata, and amount/currency.
- A repeated refund event is idempotent by provider refund ID. A provider refund
  ID may not be attached to two different local refunds.

The provider response is success/HTTP 2xx only after the local result is
durable. Transient internal failures return a retryable failure response and
mark the Inbox event `failed`; permanent validation failures are retained with
an actionable error and are not silently dropped.

## 7. Dispute and fulfilment hold lifecycle

Dispute ingestion runs in one transaction:

1. Resolve the order/transaction and lock the order.
2. Upsert the provider dispute by its unique provider ID.
3. If the dispute is active, set `fulfillment_hold=true` and record the
   pre-dispute operational status (or an equivalent immutable transition event).
4. Do not overwrite `shipped`, `completed`, or another operational status in a
   way that loses the recovery target. If the legacy `status=disputed`
   projection is retained, it must carry `dispute_previous_status`.
5. On `won`, `closed`, `resolved`, `cancelled`, or equivalent terminal-success
   status, lock the order again and clear only this dispute's hold. Restore the
   saved pre-dispute status only when no other active dispute, payment review,
   or manual hold remains and no newer fulfilment transition supersedes it.
6. On `lost`/`refunded`, preserve the dispute fact and reconcile the refund;
   never restore shipment automatically.

The restore operation is compare-and-swap guarded: a stale dispute event cannot
move a newer order state backwards.

## 8. Outbox and audit guarantees

The following events are written transactionally with the state transition:

- `order.paid` exactly once per completed original payment transaction;
- `order.payment_exception` for late/ambiguous payments;
- `payment.refund_pending`, `payment.refund_completed`, and
  `payment.refund_failed`;
- `payment.dispute_opened` and `payment.dispute_resolved`;
- `payment.review_created` and `payment.review_resolved`.

Outbox workers are at-least-once. Consumers must use `event_key` idempotency.
Provider API calls are never held open inside the database transaction; their
result is reconciled by execution records and verified provider webhooks.

## 9. Incident runbook

For any payment incident, collect these identifiers first:

```text
order_number, order_id, provider, provider event/request ID,
provider transaction ID, local transaction ID, refund ID, review ID, dispute ID
```

Then inspect, in order:

1. Inbox row: claimed/processing/processed/failed and last error.
2. Order row: `status`, `payment_status`, `fulfillment_hold`, timestamps.
3. All transactions for the order, ordered by creation time and provider ID.
4. Refunds and provider refund IDs for each transaction.
5. Active reviews/disputes and their reasons/statuses.
6. Outbox events and delivery attempts.

Never repair a financial incident by directly editing only `orders.status`.
Create or reconcile the missing transaction/refund/dispute/review fact first,
then let the projection transition run under its normal lock and idempotency
rules.

## 10. WH-01 to WH-06 acceptance matrix

| ID | Required invariant/test |
|---|---|
| WH-01 | Concurrent Capture and Webhook for one charge produce one `completed` transaction, no duplicate refund, and one order-paid event. A different second charge produces one `duplicate_paid` transaction and one pending refund. |
| WH-02 | A signed domestic Alipay success notification with no `currency` is normalized to `CNY` and marks the matching order paid. |
| WH-03 | An order in USD with a CNY provider payable snapshot validates against the CNY snapshot amount; direct CNY-vs-USD comparison is rejected or routed to review when no snapshot exists. |
| WH-04 | Opening a dispute sets a hold without losing the pre-dispute shipping state. A won/closed dispute clears its own hold and restores that state only when no other hold exists. |
| WH-05 | WeChat/Alipay refund notifications use refund parsers, persist success/failure outcomes, and are idempotent by provider refund ID. |
| WH-06 | Approving/rejecting every supported review reason changes the review and executes its documented side effect; no reason may return success as a no-op. |

## 11. Implementation order

1. Add/complete contract tests for the acceptance matrix above.
2. Make the payable snapshot and provider-currency policy explicit for new
   orders; add a safe reconciliation path for historical rows.
3. Harden the shared verified-payment transaction algorithm and lock/reload
   behavior.
4. Separate refund webhook parsing and acknowledgement from payment parsing.
5. Add dispute previous-state/hold restoration with compare-and-swap guards.
6. Complete review reason policies and Outbox/audit events.
7. Add dashboards and the runbook queries before enabling new provider traffic.

