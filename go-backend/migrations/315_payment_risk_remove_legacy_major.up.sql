ALTER TABLE payment_risk_snapshots
    DROP COLUMN IF EXISTS successful_payment_amount,
    DROP COLUMN IF EXISTS dispute_amount,
    DROP COLUMN IF EXISTS refund_amount;

