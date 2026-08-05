# Order Completion Loyalty Reward Design

## Purpose

This document defines how customer loyalty points should be awarded after an order is completed.

The goal is to close the loyalty loop without coupling payment, order fulfillment, refunds, and marketing rules into one fragile path. A bug in the loyalty reward module must not block payment verification, order completion, shipment updates, or refund recording.

## Current State

The system already has the following pieces:

- Orders have lifecycle status fields: `status`, `payment_status`, `shipping_status`, `paid_at`, `completed_at`, and refund-related states.
- Checkout can spend points on an order and writes a `loyalty_transactions` row with `type = spend`, `source = order`, and `source_id = order_id`.
- Order cancellation can refund spent points.
- Payment success marks the order as paid/processing and emits an `order.paid` outbox event.
- Loyalty accounts are stored in `user_loyalty`.
- Loyalty ledger entries are stored in `loyalty_transactions`.
- Member levels are stored in `member_levels` and include `points_multiplier`.
- Versioned loyalty configuration exists in `loyalty_program_configs`, but it currently focuses on redemption, check-in, and referral rules.

The missing piece is an isolated, idempotent reward flow for:

```text
paid order -> completed order -> award earned points -> update member level
```

## Design Principles

- Award points only after `orders.status = completed`.
- Do not award points when payment succeeds.
- Do not let loyalty reward failures block order completion.
- Keep payment services unaware of loyalty reward calculation.
- Keep order status transitions separate from marketing settlement logic.
- Make all reward actions idempotent at database level.
- Support retry and repair if an async worker fails.
- Preserve historical rule versions used to calculate rewards.

## Non Goals

This design does not implement:

- Real-time exchange-rate conversion.
- Per-product reward rules.
- Product-category reward exclusions.
- Manual admin point adjustment redesign.
- A full refund clawback policy for every partial refund edge case.

Those can be added later after the basic completion reward pipeline is stable.

## Trigger Point

The canonical trigger is:

```text
order.status changes to completed
```

Reward eligibility must require:

```text
orders.status = completed
orders.payment_status = paid
orders.user_id > 0
no existing earn reward for this order
```

The initial trigger will usually come from the admin order status update path:

```text
processing/shipped -> completed
```

Later, if shipping webhooks automatically close orders after delivery, that path must emit the same completion event and reuse the same reward processor.

## Architecture

Use an outbox event to separate order completion from loyalty settlement.

```text
OrderService.UpdateOrderStatus(order_id, completed)
  -> transaction updates orders.status and completed_at
  -> transaction creates outbox_events row: order.completed
  -> returns success

Outbox worker
  -> handles order.completed
  -> runs independent marketing settlement processors:
       1. Order loyalty reward processor
       2. Referral completion processor
```

If the reward processor fails, the outbox event retries. The order remains completed.

## Proposed Modules

### Order Service

Responsibilities:

- Validate order status transition.
- Update order status.
- Create `order.completed` outbox event.

Must not:

- Calculate loyalty reward points.
- Update member level directly.
- Complete referral rewards directly.

### Marketing Settlement Service

Suggested name:

```text
MarketingSettlementService
```

Responsibilities:

- Receive `order.completed` event payload.
- Load the completed order.
- Run independent processors.
- Allow one processor to fail without corrupting another.

Recommended processors:

```text
OrderCompletionPointsProcessor
ReferralCompletionProcessor
```

### Order Completion Points Processor

Responsibilities:

- Validate reward eligibility.
- Load active loyalty reward config.
- Determine pre-reward member level.
- Calculate points.
- Insert one loyalty ledger entry.
- Update `user_loyalty`.
- Update `user_loyalty.member_level_id` after points are added.

Must be idempotent.

### Referral Completion Processor

Responsibilities:

- Check if the completed order qualifies as the referred user's first completed paid order.
- Call existing referral completion logic.
- Use its own idempotency guard.

This processor should not be mixed into order reward calculation.

## Configuration

Do not reuse `exchange_rate_points`. That setting means:

```text
points required to redeem 1 currency unit
```

Order reward needs separate fields. Proposed additions to `loyalty_program_configs`:

```sql
order_reward_enabled BOOLEAN NOT NULL DEFAULT TRUE
order_reward_points_per_currency NUMERIC(12, 4) NOT NULL DEFAULT 1
order_reward_use_member_multiplier BOOLEAN NOT NULL DEFAULT TRUE
```

Meaning:

```text
order_reward_enabled
  Whether completed paid orders earn points.

order_reward_points_per_currency
  Base points per 1 unit of configured currency.
  Example: 1 USD = 1 point.

order_reward_use_member_multiplier
  Whether member_levels.points_multiplier affects reward points.
```

The admin panel can show the calculation basis as fixed text:

```text
Reward basis: paid product amount after discounts, excluding shipping and tax.
```

## Reward Amount Basis

Use product amount after discounts, excluding shipping and tax.

Recommended formula:

```text
gross_product_amount = orders.subtotal_amount
product_discount = max(orders.discount_amount - orders.shipping_fee_discount - orders.tax_discount, 0)
rewardable_amount = max(gross_product_amount - product_discount, 0)
```

Current order data does not split discount by type. Until split fields exist, use this conservative formula:

```text
rewardable_amount = max(orders.subtotal_amount - orders.discount_amount, 0)
```

Do not use `orders.total_amount`, because it may include shipping and tax.

Do not award points for:

- Shipping fee.
- Tax amount.
- Fully discounted product amount.
- Gift card value itself if gift cards later become purchasable products.

## Reward Calculation

Use the member level before this order reward is added.

```text
base_points = floor(rewardable_amount * order_reward_points_per_currency)
multiplier = current_member_level.points_multiplier
earned_points = floor(base_points * multiplier)
```

If multiplier is disabled:

```text
earned_points = base_points
```

If calculated points are `0`, do not write a ledger entry.

## Member Level Update

After reward points are successfully added:

```text
new_total_points = user_loyalty.total_points + earned_points
new_level = member_levels range matching new_total_points
user_loyalty.member_level_id = new_level.id
```

This update belongs inside the same transaction as the reward ledger insert.

## Idempotency

Database idempotency is mandatory.

Add a unique index so the same order can receive the completion reward only once:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS idx_loyalty_order_completion_reward_once
    ON loyalty_transactions(source, source_id, type)
    WHERE source = 'order' AND type = 'earn' AND deleted_at IS NULL;
```

Processor behavior:

```text
if unique conflict:
  treat as success
```

Do not rely only on application-level "find before insert" checks.

## Outbox Event

Add event type:

```text
order.completed
```

Recommended event key:

```text
order.completed:{order_id}
```

Recommended payload:

```json
{
  "order_id": 123,
  "order_number": "ORD202607290001",
  "user_id": 456,
  "completed_at": "2026-07-29T10:00:00Z"
}
```

The event key prevents duplicated completion events for the same order.

## Failure Isolation

Order completion should not fail because loyalty reward failed.

Order completion transaction:

```text
update order
insert outbox event
commit
```

Reward worker transaction:

```text
lock order
validate eligibility
lock user_loyalty
insert reward ledger
update user loyalty
commit
```

Failure outcomes:

- Order update failure: order remains unchanged.
- Outbox insert failure: order completion transaction rolls back, because the event is part of the completion contract.
- Reward processing failure: outbox retries; order stays completed.
- Duplicate reward attempt: unique index turns it into a no-op success.

## Refund and Clawback

A completed order can later be fully or partially refunded. The reward design must reserve a clean clawback path.

Initial minimum rule:

```text
When a completed order becomes fully refunded, reverse the order reward if it exists.
```

Suggested reversal ledger:

```text
type = spend or reverse
source = order_refund
source_id = refund_id
description = Reversed reward points for refunded order #{order_number}
```

Important edge case:

If the user has already spent the earned reward points, immediate clawback may fail due to insufficient available points.

Possible policies:

- Allow negative available balance.
- Record pending clawback debt.
- Block refund completion until points are available.
- Do not claw back points after settlement.

Recommended first policy:

```text
Record pending clawback debt instead of blocking refund.
```

This keeps refund processing isolated from loyalty balance problems.

## Data Integrity Checks

Before implementation, verify:

- `loyalty_transactions` has `program_config_id`.
- `loyalty_transactions` can support a partial unique index in production Postgres.
- `user_loyalty` rows exist or can be created under row lock.
- `member_levels` default six levels exist in all environments.
- `order.completed` outbox event can be created inside the order status transaction.
- Outbox scheduler is enabled in DEV and production workers.

## Rollout Plan

1. Add schema migration:
   - Add reward config fields.
   - Add unique index for one order reward.
   - Optionally add an order reward settlement table if richer audit is needed.

2. Update domain/config service:
   - Add fields to `ProgramConfig`.
   - Add admin/public response fields.
   - Validate reward config separately from redemption config.

3. Add outbox event:
   - Add `EventTypeOrderCompleted`.
   - Emit on successful transition to `completed`.

4. Add reward processor:
   - Load order.
   - Validate eligibility.
   - Calculate points.
   - Insert ledger.
   - Update loyalty summary and member level.

5. Add referral processor:
   - Trigger existing referral completion only for qualifying first completed order.

6. Add refund hook:
   - For full refund, enqueue or process reward reversal.
   - Keep refund success independent from loyalty reversal failure.

7. Add admin visibility:
   - Show order reward rule in loyalty settings.
   - Show reward ledger entries in loyalty transaction table.

## Test Plan

Required backend tests:

- Completing an unpaid order does not award points.
- Completing a paid order awards points once.
- Re-processing the same `order.completed` event does not duplicate points.
- Reward uses `subtotal_amount - discount_amount`, not `total_amount`.
- Member multiplier is based on pre-reward member level.
- Member level updates after reward.
- Reward config disabled means no points are awarded.
- Missing member level falls back safely or returns retryable error.
- Full refund reverses or schedules reversal of previously awarded points.
- Reward failure does not roll back order completion.

Required migration checks:

- Existing `loyalty_transactions` rows do not violate the new unique index.
- Existing DEV database can migrate cleanly.
- Empty database baseline plus SQL migrations produces the same schema.

## Open Decisions

These should be decided before coding:

- Should partial refunds claw back proportional reward points?
- Should negative loyalty balance be allowed?
- Should gift card purchases earn points?
- Should manually completed orders award points immediately, or only after a delay?
- Should reward points expire, and if so after how many days?

## Recommended First Implementation

Implement the minimal safe version:

- Award points only for `payment_status = paid` and `status = completed`.
- Use `rewardable_amount = max(subtotal_amount - discount_amount, 0)`.
- Use active loyalty config at processing time.
- Store `program_config_id` on the reward ledger.
- Use pre-reward member level multiplier.
- Update member level after reward.
- Add partial unique index for idempotency.
- Emit and handle `order.completed` through outbox.
- Treat duplicate reward as success.
- Do not block order completion if reward processing fails.
