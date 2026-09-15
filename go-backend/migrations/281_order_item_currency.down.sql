DROP INDEX IF EXISTS idx_order_items_currency;
ALTER TABLE order_items DROP COLUMN IF EXISTS currency;
