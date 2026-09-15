-- Persist the exact loyalty effects attached to each refund.
-- The fields are internal accounting facts and are deliberately separate from
-- the gateway amount and the existing coupon discount clawback.

ALTER TABLE refunds
    ADD COLUMN IF NOT EXISTS loyalty_settlement_prepared BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS loyalty_points_clawback INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS loyalty_points_returned INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS loyalty_cash_deduction_amount NUMERIC(12,2) NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_refunds_loyalty_settlement_prepared
    ON refunds(loyalty_settlement_prepared);

CREATE UNIQUE INDEX IF NOT EXISTS uq_loyalty_transactions_refund_settlement
    ON loyalty_transactions(user_id, type, source, source_id)
    WHERE source IN (
        'refund_loyalty_clawback',
        'refund_loyalty_clawback_reversal',
        'refund_loyalty_points_return'
    )
    AND source_id > 0;
