ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS signature_required BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_orders_signature_required
    ON orders(signature_required);
