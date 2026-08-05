ALTER TABLE IF EXISTS refunds
  ADD COLUMN IF NOT EXISTS requested_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS discount_clawback_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS calculation_snapshot TEXT NOT NULL DEFAULT '';

UPDATE refunds
SET requested_amount = amount
WHERE requested_amount = 0;
