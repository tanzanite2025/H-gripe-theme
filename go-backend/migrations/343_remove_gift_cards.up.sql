-- Gift cards are retired. Remove their settings, redemption ledger entries,
-- and value-bearing tables. Historical migrations remain untouched so already
-- deployed databases retain a valid migration chain.

-- These are the only legacy settings created for gift-card redemption by
-- migration 043. Delete by exact key across every locale/group so an old
-- override cannot leave the retired feature configurable.
DELETE FROM settings
WHERE key IN (
  'tz_redeem_enabled',
  'tz_redeem_exchange_rate',
  'tz_redeem_min_points',
  'tz_redeem_max_value_per_day',
  'tz_redeem_card_expiry_days',
  'tz_redeem_preset_values'
);

-- A legacy gift-card redemption is represented by a negative spend entry.
-- Removing that entry must return its spent points while preserving earned
-- points and member-tier history. The source value is the stable discriminator
-- written by migration 044; no ordinary order/check-in/referral ledger rows
-- are included.
WITH retired_gift_card_spend AS (
  SELECT
    user_id,
    COALESCE(SUM(-points) FILTER (WHERE type = 'spend' AND points < 0), 0)::INTEGER AS points_to_restore
  FROM loyalty_transactions
  WHERE LOWER(COALESCE(source, '')) = 'giftcard'
  GROUP BY user_id
)
UPDATE user_loyalty AS ul
SET
  available_points = GREATEST(0, ul.available_points + retired.points_to_restore),
  used_points = GREATEST(0, ul.used_points - retired.points_to_restore),
  updated_at = CURRENT_TIMESTAMP
FROM retired_gift_card_spend AS retired
WHERE ul.user_id = retired.user_id
  AND retired.points_to_restore > 0;

-- referral_reward_logs has an optional RESTRICT foreign key to the unified
-- ledger. It cannot normally point at a gift-card spend, but clear any such
-- unexpected historical reference before deleting the retired ledger rows.
UPDATE referral_reward_logs AS reward
SET loyalty_transaction_id = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE loyalty_transaction_id IN (
  SELECT id
  FROM loyalty_transactions
  WHERE LOWER(COALESCE(source, '')) = 'giftcard'
);

-- gift_card_transactions references gift_card_redemptions, and redemptions
-- retain a RESTRICT foreign key to loyalty_transactions. Drop those dependent
-- tables before deleting the retired ledger rows; PostgreSQL otherwise rejects
-- a redemption's loyalty_transaction_id while the parent row is removed.
DROP TABLE IF EXISTS gift_card_transactions;
DROP TABLE IF EXISTS gift_card_redemptions;

DELETE FROM loyalty_transactions
WHERE LOWER(COALESCE(source, '')) = 'giftcard';

-- gift cards can only be dropped after their redemption records are gone.
DROP TABLE IF EXISTS gift_cards;
DROP TABLE IF EXISTS loyalty_program_redeem_options;

ALTER TABLE loyalty_program_configs
  DROP COLUMN IF EXISTS min_redeem_points,
  DROP COLUMN IF EXISTS max_value_per_day_minor,
  DROP COLUMN IF EXISTS card_expiry_days;

ALTER TABLE refunds
  DROP COLUMN IF EXISTS gift_card_refund_amount_minor;
