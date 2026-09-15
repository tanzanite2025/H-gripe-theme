-- Persist order monetary facts as integer minor units. These columns are the
-- canonical transaction snapshot; the legacy major-unit columns remain only
-- until downstream read models finish migrating.
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS payment_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS subtotal_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS shipping_fee_minor BIGINT,
    ADD COLUMN IF NOT EXISTS tax_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS discount_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS total_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS points_value_minor BIGINT;

UPDATE orders
SET subtotal_amount_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(subtotal_amount)::BIGINT ELSE ROUND(subtotal_amount * 100)::BIGINT END,
    shipping_fee_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(shipping_fee)::BIGINT ELSE ROUND(shipping_fee * 100)::BIGINT END,
    tax_amount_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(tax_amount)::BIGINT ELSE ROUND(tax_amount * 100)::BIGINT END,
    discount_amount_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(discount_amount)::BIGINT ELSE ROUND(discount_amount * 100)::BIGINT END,
    total_amount_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(total_amount)::BIGINT ELSE ROUND(total_amount * 100)::BIGINT END,
    points_value_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(points_value)::BIGINT ELSE ROUND(points_value * 100)::BIGINT END
WHERE subtotal_amount_minor IS NULL
   OR shipping_fee_minor IS NULL
   OR tax_amount_minor IS NULL
   OR discount_amount_minor IS NULL
   OR total_amount_minor IS NULL
   OR points_value_minor IS NULL;

UPDATE orders
SET payment_amount_minor = CASE WHEN UPPER(COALESCE(payment_currency, currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(payment_amount)::BIGINT ELSE ROUND(payment_amount * 100)::BIGINT END
WHERE payment_amount_minor IS NULL;

ALTER TABLE orders
    ALTER COLUMN payment_amount_minor SET NOT NULL,
    ALTER COLUMN subtotal_amount_minor SET NOT NULL,
    ALTER COLUMN shipping_fee_minor SET NOT NULL,
    ALTER COLUMN tax_amount_minor SET NOT NULL,
    ALTER COLUMN discount_amount_minor SET NOT NULL,
    ALTER COLUMN total_amount_minor SET NOT NULL,
    ALTER COLUMN points_value_minor SET NOT NULL,
    ADD CONSTRAINT chk_orders_payment_amount_minor_non_negative CHECK (payment_amount_minor >= 0),
    ADD CONSTRAINT chk_orders_subtotal_amount_minor_non_negative CHECK (subtotal_amount_minor >= 0),
    ADD CONSTRAINT chk_orders_shipping_fee_minor_non_negative CHECK (shipping_fee_minor >= 0),
    ADD CONSTRAINT chk_orders_tax_amount_minor_non_negative CHECK (tax_amount_minor >= 0),
    ADD CONSTRAINT chk_orders_discount_amount_minor_non_negative CHECK (discount_amount_minor >= 0),
    ADD CONSTRAINT chk_orders_total_amount_minor_non_negative CHECK (total_amount_minor >= 0),
    ADD CONSTRAINT chk_orders_points_value_minor_non_negative CHECK (points_value_minor >= 0);

ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS price_minor BIGINT,
    ADD COLUMN IF NOT EXISTS subtotal_minor BIGINT,
    ADD COLUMN IF NOT EXISTS tax_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS discount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS total_minor BIGINT;

UPDATE order_items
SET price_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(price)::BIGINT ELSE ROUND(price * 100)::BIGINT END,
    subtotal_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(subtotal)::BIGINT ELSE ROUND(subtotal * 100)::BIGINT END,
    tax_amount_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(tax_amount)::BIGINT ELSE ROUND(tax_amount * 100)::BIGINT END,
    discount_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(discount)::BIGINT ELSE ROUND(discount * 100)::BIGINT END,
    total_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(total)::BIGINT ELSE ROUND(total * 100)::BIGINT END
WHERE price_minor IS NULL
   OR subtotal_minor IS NULL
   OR tax_amount_minor IS NULL
   OR discount_minor IS NULL
   OR total_minor IS NULL;

ALTER TABLE order_items
    ALTER COLUMN price_minor SET NOT NULL,
    ALTER COLUMN subtotal_minor SET NOT NULL,
    ALTER COLUMN tax_amount_minor SET NOT NULL,
    ALTER COLUMN discount_minor SET NOT NULL,
    ALTER COLUMN total_minor SET NOT NULL,
    ADD CONSTRAINT chk_order_items_price_minor_non_negative CHECK (price_minor >= 0),
    ADD CONSTRAINT chk_order_items_subtotal_minor_non_negative CHECK (subtotal_minor >= 0),
    ADD CONSTRAINT chk_order_items_tax_amount_minor_non_negative CHECK (tax_amount_minor >= 0),
    ADD CONSTRAINT chk_order_items_discount_minor_non_negative CHECK (discount_minor >= 0),
    ADD CONSTRAINT chk_order_items_total_minor_non_negative CHECK (total_minor >= 0);
