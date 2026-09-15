ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS pricing_snapshot JSONB NOT NULL DEFAULT '{}';

CREATE OR REPLACE FUNCTION prevent_order_pricing_snapshot_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.pricing_snapshot <> '{}'::jsonb
       AND NEW.pricing_snapshot IS DISTINCT FROM OLD.pricing_snapshot THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'order pricing snapshot is immutable once populated';
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trigger_prevent_order_pricing_snapshot_mutation ON orders;

CREATE TRIGGER trigger_prevent_order_pricing_snapshot_mutation
BEFORE UPDATE OF pricing_snapshot ON orders
FOR EACH ROW
EXECUTE FUNCTION prevent_order_pricing_snapshot_mutation();
