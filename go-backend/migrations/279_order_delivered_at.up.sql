ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_orders_delivered_at
    ON orders(delivered_at)
    WHERE delivered_at IS NOT NULL;
