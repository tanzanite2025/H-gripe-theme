-- Order lines carry the currency captured by the checkout pricing snapshot.
-- This is a transactional fact and must not be inferred from the current
-- product or variant catalog row during refunds or fulfillment.
ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

UPDATE order_items oi
SET currency = COALESCE(NULLIF(UPPER(o.currency), ''), 'USD')
FROM orders o
WHERE o.id = oi.order_id;

CREATE INDEX IF NOT EXISTS idx_order_items_currency ON order_items(currency);
