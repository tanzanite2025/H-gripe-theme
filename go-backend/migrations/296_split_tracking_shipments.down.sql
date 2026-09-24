DROP INDEX IF EXISTS idx_shipping_tracking_shipments_order_number;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_order
    ON shipping_tracking_shipments (order_id)
    WHERE deleted_at IS NULL;
