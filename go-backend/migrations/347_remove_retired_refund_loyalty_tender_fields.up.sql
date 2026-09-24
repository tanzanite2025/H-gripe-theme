-- Points are earned loyalty credit, never a checkout tender. Remove the
-- retired refund columns that represented returning or cash-converting spent
-- points; keep loyalty_points_debt for unrecovered earned-point clawbacks.
ALTER TABLE refunds
    DROP COLUMN IF EXISTS loyalty_points_returned,
    DROP COLUMN IF EXISTS loyalty_points_cash_recovered,
    DROP COLUMN IF EXISTS loyalty_cash_deduction_amount_minor;

DROP INDEX IF EXISTS uq_loyalty_transactions_refund_settlement;
CREATE UNIQUE INDEX uq_loyalty_transactions_refund_settlement
    ON loyalty_transactions(user_id, type, source, source_id)
    WHERE source IN (
        'refund_loyalty_clawback',
        'refund_loyalty_clawback_reversal',
        'refund_loyalty_cash_recovery_debt'
    )
    AND source_id > 0;
