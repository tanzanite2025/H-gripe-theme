-- Preserve unrecovered loyalty points when their cash value is capped by a
-- smaller refund amount. The account debt is posted only after refund success.
ALTER TABLE refunds
    ADD COLUMN IF NOT EXISTS loyalty_points_cash_recovered INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS loyalty_points_debt INTEGER NOT NULL DEFAULT 0,
    ADD CONSTRAINT chk_refunds_loyalty_points_cash_recovered_non_negative
        CHECK (loyalty_points_cash_recovered >= 0),
    ADD CONSTRAINT chk_refunds_loyalty_points_debt_non_negative
        CHECK (loyalty_points_debt >= 0);

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
