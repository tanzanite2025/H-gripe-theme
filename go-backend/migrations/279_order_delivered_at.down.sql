DROP INDEX IF EXISTS idx_orders_delivered_at;
ALTER TABLE orders DROP COLUMN IF EXISTS delivered_at;
