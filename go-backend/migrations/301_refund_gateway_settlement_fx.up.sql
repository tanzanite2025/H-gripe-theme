ALTER TABLE refunds
    ADD COLUMN IF NOT EXISTS settlement_amount_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS settlement_currency VARCHAR(3) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS settlement_balance_transaction_id VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS fx_gain_loss_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fx_gain_loss_currency VARCHAR(3) NOT NULL DEFAULT '';

ALTER TABLE payment_refund_executions
    ADD COLUMN IF NOT EXISTS settlement_amount_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS settlement_currency VARCHAR(3) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS settlement_balance_transaction_id VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS fx_gain_loss_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fx_gain_loss_currency VARCHAR(3) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_refunds_settlement_balance_transaction
    ON refunds(settlement_balance_transaction_id)
    WHERE settlement_balance_transaction_id <> '';

CREATE INDEX IF NOT EXISTS idx_payment_refund_executions_settlement_balance_transaction
    ON payment_refund_executions(settlement_balance_transaction_id)
    WHERE settlement_balance_transaction_id <> '';
