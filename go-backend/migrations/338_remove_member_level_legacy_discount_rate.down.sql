-- Restore the historical member-level discount column only for an explicit
-- rollback.  The canonical decimal value is preserved at numeric precision.
ALTER TABLE member_levels
    ADD COLUMN IF NOT EXISTS discount_rate NUMERIC;

UPDATE member_levels
SET discount_rate = discount_rate_decimal::NUMERIC
WHERE discount_rate IS NULL;

ALTER TABLE member_levels
    ALTER COLUMN discount_rate SET NOT NULL,
    ALTER COLUMN discount_rate SET DEFAULT 0,
    DROP CONSTRAINT IF EXISTS chk_member_levels_discount_rate_decimal_range,
    DROP COLUMN IF EXISTS discount_rate_decimal;
