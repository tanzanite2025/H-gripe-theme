ALTER TABLE payment_refund_executions
    ADD COLUMN IF NOT EXISTS merchant_order_number VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS provider_transaction_id VARCHAR(255) NOT NULL DEFAULT '';

ALTER TABLE payment_refund_executions
    ALTER COLUMN merchant_order_number DROP DEFAULT,
    ALTER COLUMN provider_transaction_id DROP DEFAULT;

CREATE INDEX IF NOT EXISTS idx_payment_refund_executions_merchant_order
    ON payment_refund_executions(merchant_order_number);

CREATE INDEX IF NOT EXISTS idx_payment_refund_executions_provider_transaction
    ON payment_refund_executions(provider, provider_transaction_id);
