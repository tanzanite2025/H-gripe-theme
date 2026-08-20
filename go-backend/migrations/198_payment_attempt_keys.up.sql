ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS attempt_key VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS provider_request_key VARCHAR(255) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_transactions_attempt_key
    ON transactions(order_id, payment_method, attempt_key);

CREATE UNIQUE INDEX IF NOT EXISTS uq_transactions_order_method_attempt
    ON transactions(order_id, payment_method, attempt_key)
    WHERE attempt_key <> '';
