DROP INDEX IF EXISTS idx_orders_signature_required;

ALTER TABLE orders
    DROP COLUMN IF EXISTS signature_required;
