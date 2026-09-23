ALTER TABLE refunds
    ADD COLUMN IF NOT EXISTS loyalty_cash_deduction_amount_minor BIGINT;

UPDATE refunds
SET loyalty_cash_deduction_amount_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(loyalty_cash_deduction_amount)::BIGINT
        ELSE ROUND(loyalty_cash_deduction_amount * 100)::BIGINT
    END
WHERE loyalty_cash_deduction_amount_minor IS NULL;

ALTER TABLE refunds
    ALTER COLUMN loyalty_cash_deduction_amount_minor SET NOT NULL,
    ALTER COLUMN loyalty_cash_deduction_amount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_refunds_loyalty_cash_deduction_minor_non_negative
        CHECK (loyalty_cash_deduction_amount_minor >= 0);
