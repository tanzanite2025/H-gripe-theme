ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS payment_currency VARCHAR(3),
    ADD COLUMN IF NOT EXISTS payment_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS dispute_previous_status VARCHAR(32),
    ADD COLUMN IF NOT EXISTS dispute_previous_hold BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_orders_payment_currency
    ON orders(payment_currency);
