ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS chk_orders_payment_amount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_orders_subtotal_amount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_orders_shipping_fee_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_orders_tax_amount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_orders_discount_amount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_orders_total_amount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_orders_points_value_minor_non_negative;
ALTER TABLE orders
    DROP COLUMN IF EXISTS payment_amount_minor,
    DROP COLUMN IF EXISTS subtotal_amount_minor,
    DROP COLUMN IF EXISTS shipping_fee_minor,
    DROP COLUMN IF EXISTS tax_amount_minor,
    DROP COLUMN IF EXISTS discount_amount_minor,
    DROP COLUMN IF EXISTS total_amount_minor,
    DROP COLUMN IF EXISTS points_value_minor;

ALTER TABLE order_items
    DROP CONSTRAINT IF EXISTS chk_order_items_price_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_order_items_subtotal_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_order_items_tax_amount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_order_items_discount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_order_items_total_minor_non_negative;
ALTER TABLE order_items
    DROP COLUMN IF EXISTS price_minor,
    DROP COLUMN IF EXISTS subtotal_minor,
    DROP COLUMN IF EXISTS tax_amount_minor,
    DROP COLUMN IF EXISTS discount_minor,
    DROP COLUMN IF EXISTS total_minor;
