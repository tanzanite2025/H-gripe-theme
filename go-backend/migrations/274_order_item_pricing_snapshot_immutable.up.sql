-- Historical rows start as an empty object and may be backfilled once. After
-- a line has an order-time pricing snapshot, it is financial evidence and
-- must not be silently changed by ordinary order-item updates.
CREATE OR REPLACE FUNCTION prevent_order_item_pricing_snapshot_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.pricing_snapshot <> '{}'::jsonb
       AND NEW.pricing_snapshot IS DISTINCT FROM OLD.pricing_snapshot THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'order item pricing snapshot is immutable once populated';
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trigger_prevent_order_item_pricing_snapshot_mutation ON order_items;

CREATE TRIGGER trigger_prevent_order_item_pricing_snapshot_mutation
BEFORE UPDATE OF pricing_snapshot ON order_items
FOR EACH ROW
EXECUTE FUNCTION prevent_order_item_pricing_snapshot_mutation();
