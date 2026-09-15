DROP INDEX IF EXISTS idx_orders_fulfillment_hold;

ALTER TABLE orders
    DROP COLUMN IF EXISTS fulfillment_hold;
