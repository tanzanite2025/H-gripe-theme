DROP TRIGGER IF EXISTS trigger_prevent_order_pricing_snapshot_mutation ON orders;
DROP FUNCTION IF EXISTS prevent_order_pricing_snapshot_mutation();
ALTER TABLE orders DROP COLUMN IF EXISTS pricing_snapshot;
