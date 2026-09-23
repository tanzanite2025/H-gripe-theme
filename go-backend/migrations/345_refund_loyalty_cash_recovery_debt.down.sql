DROP INDEX IF EXISTS uq_loyalty_transactions_refund_settlement;

CREATE UNIQUE INDEX uq_loyalty_transactions_refund_settlement
    ON loyalty_transactions(user_id, type, source, source_id)
    WHERE source IN (
        'refund_loyalty_clawback',
        'refund_loyalty_clawback_reversal',
        'refund_loyalty_points_return'
    )
    AND source_id > 0;

ALTER TABLE refunds
    DROP CONSTRAINT IF EXISTS chk_refunds_loyalty_points_cash_recovered_non_negative,
    DROP CONSTRAINT IF EXISTS chk_refunds_loyalty_points_debt_non_negative,
    DROP COLUMN IF EXISTS loyalty_points_cash_recovered,
    DROP COLUMN IF EXISTS loyalty_points_debt;
