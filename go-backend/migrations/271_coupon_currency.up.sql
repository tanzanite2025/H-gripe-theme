-- Coupon monetary fields are entered in an explicit currency. Existing
-- coupons used the primary (USD) currency, so backfill and default to USD.
ALTER TABLE coupons
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

UPDATE coupons
SET currency = 'USD'
WHERE currency IS NULL OR BTRIM(currency) = '';

