-- Transactional order amounts are stored only in exact minor units. The
-- former major-unit columns allowed float-based calculations and are removed
-- before launch; read models must derive display values at their boundary.
ALTER TABLE orders
    DROP COLUMN IF EXISTS payment_amount,
    DROP COLUMN IF EXISTS subtotal_amount,
    DROP COLUMN IF EXISTS shipping_fee,
    DROP COLUMN IF EXISTS tax_amount,
    DROP COLUMN IF EXISTS discount_amount,
    DROP COLUMN IF EXISTS total_amount,
    DROP COLUMN IF EXISTS points_value;

ALTER TABLE order_items
    DROP COLUMN IF EXISTS price,
    DROP COLUMN IF EXISTS subtotal,
    DROP COLUMN IF EXISTS tax_amount,
    DROP COLUMN IF EXISTS discount,
    DROP COLUMN IF EXISTS total;
