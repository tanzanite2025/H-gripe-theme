ALTER TABLE orders
    DROP COLUMN IF EXISTS production_completed_at,
    DROP COLUMN IF EXISTS production_started_at,
    DROP COLUMN IF EXISTS production_status,
    DROP COLUMN IF EXISTS fulfillment_mode;

ALTER TABLE order_items
    DROP COLUMN IF EXISTS fulfillment_mode;

ALTER TABLE products
    DROP COLUMN IF EXISTS fulfillment_mode;
