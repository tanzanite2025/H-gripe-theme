-- Rollback restores schema compatibility for the historical migrations. The
-- retired columns are intentionally empty and are not reintroduced in code.
ALTER TABLE refunds
    ADD COLUMN IF NOT EXISTS loyalty_points_returned INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS loyalty_points_cash_recovered INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS loyalty_cash_deduction_amount_minor BIGINT NOT NULL DEFAULT 0;

DROP INDEX IF EXISTS uq_loyalty_transactions_refund_settlement;
CREATE UNIQUE INDEX uq_loyalty_transactions_refund_settlement
    ON loyalty_transactions(user_id, type, source, source_id)
    WHERE source IN (
        'refund_loyalty_clawback',
        'refund_loyalty_clawback_reversal',
        'refund_loyalty_points_return',
        'refund_loyalty_cash_recovery_debt'
    )
    AND source_id > 0;
