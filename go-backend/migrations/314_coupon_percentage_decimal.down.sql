ALTER TABLE coupons
    DROP CONSTRAINT IF EXISTS chk_coupons_value_rate_decimal;

ALTER TABLE coupons
    DROP COLUMN IF EXISTS value_rate_decimal;

