-- Orders may be fulfilled in multiple packages. Keep one row per order and
-- tracking number instead of enforcing a single tracking row per order.
DROP INDEX IF EXISTS idx_shipping_tracking_shipments_order;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_order_number
    ON shipping_tracking_shipments (order_id, tracking_number)
    WHERE deleted_at IS NULL;
