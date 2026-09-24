ALTER TABLE coupon_usage
    DROP CONSTRAINT IF EXISTS chk_coupon_usage_discount_minor_non_negative,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS discount_minor;

ALTER TABLE coupons
    DROP CONSTRAINT IF EXISTS chk_coupons_money_minor_non_negative,
    DROP COLUMN IF EXISTS max_discount_minor,
    DROP COLUMN IF EXISTS min_amount_minor,
    DROP COLUMN IF EXISTS value_minor;
