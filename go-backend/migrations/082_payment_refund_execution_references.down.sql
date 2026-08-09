DROP INDEX IF EXISTS idx_payment_refund_executions_provider_transaction;
DROP INDEX IF EXISTS idx_payment_refund_executions_merchant_order;

ALTER TABLE payment_refund_executions
    DROP COLUMN IF EXISTS provider_transaction_id,
    DROP COLUMN IF EXISTS merchant_order_number;
