ALTER TABLE payment_risk_snapshots
    DROP COLUMN IF EXISTS refund_amount_minor_by_currency,
    DROP COLUMN IF EXISTS dispute_amount_minor_by_currency,
    DROP COLUMN IF EXISTS successful_payment_amount_minor_by_currency;

ALTER TABLE payment_risk_checkout_decisions
    DROP CONSTRAINT IF EXISTS chk_payment_risk_checkout_decisions_amount_minor_non_negative,
    DROP COLUMN IF EXISTS amount_minor;

ALTER TABLE stripe_disputes
    DROP CONSTRAINT IF EXISTS chk_stripe_disputes_amount_minor_non_negative,
    DROP COLUMN IF EXISTS amount_minor;

ALTER TABLE payment_risk_events
    DROP CONSTRAINT IF EXISTS chk_payment_risk_events_amount_minor_non_negative,
    DROP COLUMN IF EXISTS amount_minor;
