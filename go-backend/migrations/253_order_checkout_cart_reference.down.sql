ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS fk_orders_checkout_cart;

DROP INDEX IF EXISTS idx_orders_checkout_cart_id;

ALTER TABLE orders
    DROP COLUMN IF EXISTS checkout_cart_id;
