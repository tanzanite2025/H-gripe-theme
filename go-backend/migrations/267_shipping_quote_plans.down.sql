DROP INDEX IF EXISTS idx_orders_shipping_quote_id;

ALTER TABLE orders
    DROP COLUMN IF EXISTS shipping_plan_snapshot,
    DROP COLUMN IF EXISTS shipping_quote_plan_id,
    DROP COLUMN IF EXISTS shipping_quote_id;

DROP TABLE IF EXISTS shipping_quote_snapshots;
