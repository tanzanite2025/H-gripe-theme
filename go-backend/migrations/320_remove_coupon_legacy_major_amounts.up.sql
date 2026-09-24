-- Coupon monetary facts are now represented only by exact minor units (or the
-- decimal percentage rate). The old FLOAT major-unit columns are removed
-- before launch so no service can accidentally perform a second conversion.
ALTER TABLE coupons
    DROP COLUMN IF EXISTS value,
    DROP COLUMN IF EXISTS min_amount,
    DROP COLUMN IF EXISTS max_discount;

ALTER TABLE coupon_usage
    DROP COLUMN IF EXISTS discount;
