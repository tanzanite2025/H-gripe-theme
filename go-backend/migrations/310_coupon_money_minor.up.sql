ALTER TABLE coupons
    ADD COLUMN IF NOT EXISTS value_minor BIGINT,
    ADD COLUMN IF NOT EXISTS min_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS max_discount_minor BIGINT;

UPDATE coupons
SET min_amount_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(COALESCE(min_amount, 0))::BIGINT
        ELSE ROUND(COALESCE(min_amount, 0) * 100)::BIGINT
    END,
    max_discount_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(COALESCE(max_discount, 0))::BIGINT
        ELSE ROUND(COALESCE(max_discount, 0) * 100)::BIGINT
    END,
    value_minor = CASE
        WHEN LOWER(COALESCE(type, '')) = 'fixed' AND UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(COALESCE(value, 0))::BIGINT
        WHEN LOWER(COALESCE(type, '')) = 'fixed'
            THEN ROUND(COALESCE(value, 0) * 100)::BIGINT
        ELSE 0
    END
WHERE min_amount_minor IS NULL
   OR max_discount_minor IS NULL
   OR value_minor IS NULL;

ALTER TABLE coupons
    ALTER COLUMN value_minor SET NOT NULL,
    ALTER COLUMN value_minor SET DEFAULT 0,
    ALTER COLUMN min_amount_minor SET NOT NULL,
    ALTER COLUMN min_amount_minor SET DEFAULT 0,
    ALTER COLUMN max_discount_minor SET NOT NULL,
    ALTER COLUMN max_discount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_coupons_money_minor_non_negative
        CHECK (value_minor >= 0 AND min_amount_minor >= 0 AND max_discount_minor >= 0);

ALTER TABLE coupon_usage
    ADD COLUMN IF NOT EXISTS discount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

UPDATE coupon_usage u
SET currency = UPPER(COALESCE(o.currency, 'USD'))
FROM orders o
WHERE o.id = u.order_id;

UPDATE coupon_usage u
SET discount_minor = CASE
        WHEN UPPER(COALESCE(u.currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(COALESCE(u.discount, 0))::BIGINT
        ELSE ROUND(COALESCE(u.discount, 0) * 100)::BIGINT
    END
WHERE u.discount_minor IS NULL;

ALTER TABLE coupon_usage
    ALTER COLUMN discount_minor SET NOT NULL,
    ALTER COLUMN discount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_coupon_usage_discount_minor_non_negative
        CHECK (discount_minor >= 0);
