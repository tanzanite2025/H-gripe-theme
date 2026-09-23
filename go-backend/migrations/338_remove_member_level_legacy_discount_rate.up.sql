-- Member-level discounts are transactional percentages and must retain their
-- exact decimal representation.  Backfill the canonical decimal column before
-- removing the legacy, unconstrained numeric field.
ALTER TABLE member_levels
    ADD COLUMN IF NOT EXISTS discount_rate_decimal NUMERIC(30,15);

UPDATE member_levels
SET discount_rate_decimal = ROUND(COALESCE(discount_rate, 0)::NUMERIC, 15)
WHERE discount_rate_decimal IS NULL;

ALTER TABLE member_levels
    ALTER COLUMN discount_rate_decimal SET NOT NULL,
    ALTER COLUMN discount_rate_decimal SET DEFAULT 0,
    DROP CONSTRAINT IF EXISTS chk_member_levels_discount_rate_decimal_range,
    ADD CONSTRAINT chk_member_levels_discount_rate_decimal_range
        CHECK (discount_rate_decimal >= 0 AND discount_rate_decimal <= 100),
    DROP COLUMN IF EXISTS discount_rate;
