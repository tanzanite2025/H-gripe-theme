ALTER TABLE products
    ADD COLUMN IF NOT EXISTS fulfillment_mode VARCHAR(20) NOT NULL DEFAULT 'stock';

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS fulfillment_mode VARCHAR(20) NOT NULL DEFAULT 'stock',
    ADD COLUMN IF NOT EXISTS production_status VARCHAR(20) NOT NULL DEFAULT 'not_applicable',
    ADD COLUMN IF NOT EXISTS production_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS production_completed_at TIMESTAMPTZ;

ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS fulfillment_mode VARCHAR(20) NOT NULL DEFAULT 'stock';

CREATE INDEX IF NOT EXISTS idx_products_fulfillment_mode
    ON products(fulfillment_mode);
CREATE INDEX IF NOT EXISTS idx_orders_fulfillment_mode
    ON orders(fulfillment_mode);
CREATE INDEX IF NOT EXISTS idx_orders_production_status
    ON orders(production_status);
CREATE INDEX IF NOT EXISTS idx_order_items_fulfillment_mode
    ON order_items(fulfillment_mode);
