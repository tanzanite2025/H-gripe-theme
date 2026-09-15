ALTER TABLE payment_refund_executions
    ADD COLUMN IF NOT EXISTS amount_minor BIGINT;

UPDATE payment_refund_executions e
SET amount_minor = CASE
    WHEN UPPER(COALESCE(NULLIF(e.currency, ''), 'USD')) IN ('JPY', 'KRW', 'CLP') THEN ROUND(e.amount)::BIGINT
    ELSE ROUND(e.amount * 100)::BIGINT
END
WHERE amount_minor IS NULL;

ALTER TABLE payment_refund_executions
    ALTER COLUMN amount_minor SET NOT NULL,
    ADD CONSTRAINT chk_payment_refund_executions_amount_minor_non_negative CHECK (amount_minor >= 0);
