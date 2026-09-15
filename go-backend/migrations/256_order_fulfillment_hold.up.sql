ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS fulfillment_hold BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_orders_fulfillment_hold
    ON orders(fulfillment_hold);
