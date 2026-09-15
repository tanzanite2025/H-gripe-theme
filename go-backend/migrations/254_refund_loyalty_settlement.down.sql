DROP INDEX IF EXISTS uq_loyalty_transactions_refund_settlement;
DROP INDEX IF EXISTS idx_refunds_loyalty_settlement_prepared;

ALTER TABLE refunds
    DROP COLUMN IF EXISTS loyalty_cash_deduction_amount,
    DROP COLUMN IF EXISTS loyalty_points_returned,
    DROP COLUMN IF EXISTS loyalty_points_clawback,
    DROP COLUMN IF EXISTS loyalty_settlement_prepared;
