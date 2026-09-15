ALTER TABLE payment_refund_executions
    DROP CONSTRAINT IF EXISTS chk_payment_refund_executions_amount_minor_non_negative,
    DROP COLUMN IF EXISTS amount_minor;
