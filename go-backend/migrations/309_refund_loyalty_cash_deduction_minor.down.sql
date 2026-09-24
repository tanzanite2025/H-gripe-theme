ALTER TABLE refunds
    DROP CONSTRAINT IF EXISTS chk_refunds_loyalty_cash_deduction_minor_non_negative,
    DROP COLUMN IF EXISTS loyalty_cash_deduction_amount_minor;
