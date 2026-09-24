DROP INDEX IF EXISTS idx_payment_refund_executions_settlement_balance_transaction;
DROP INDEX IF EXISTS idx_refunds_settlement_balance_transaction;

ALTER TABLE payment_refund_executions
    DROP COLUMN IF EXISTS fx_gain_loss_currency,
    DROP COLUMN IF EXISTS fx_gain_loss_minor,
    DROP COLUMN IF EXISTS settlement_balance_transaction_id,
    DROP COLUMN IF EXISTS settlement_currency,
    DROP COLUMN IF EXISTS settlement_amount_minor;

ALTER TABLE refunds
    DROP COLUMN IF EXISTS fx_gain_loss_currency,
    DROP COLUMN IF EXISTS fx_gain_loss_minor,
    DROP COLUMN IF EXISTS settlement_balance_transaction_id,
    DROP COLUMN IF EXISTS settlement_currency,
    DROP COLUMN IF EXISTS settlement_amount_minor;
