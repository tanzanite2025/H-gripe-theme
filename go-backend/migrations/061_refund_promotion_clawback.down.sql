ALTER TABLE IF EXISTS refunds
  DROP COLUMN IF EXISTS calculation_snapshot,
  DROP COLUMN IF EXISTS discount_clawback_amount,
  DROP COLUMN IF EXISTS requested_amount;
