-- Payment-risk monetary facts are stored in the currency's smallest unit.
-- Portfolio reports aggregate by currency and never sum major-unit floats.

ALTER TABLE payment_risk_events
    ADD COLUMN IF NOT EXISTS amount_minor BIGINT;

UPDATE payment_risk_events
SET amount_minor = CASE
    WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(COALESCE(amount, 0))::BIGINT
    ELSE ROUND(COALESCE(amount, 0) * 100)::BIGINT
END
WHERE amount_minor IS NULL;

ALTER TABLE payment_risk_events
    ALTER COLUMN amount_minor SET NOT NULL,
    ALTER COLUMN amount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_payment_risk_events_amount_minor_non_negative
        CHECK (amount_minor >= 0);

ALTER TABLE stripe_disputes
    ADD COLUMN IF NOT EXISTS amount_minor BIGINT;

UPDATE stripe_disputes
SET amount_minor = CASE
    WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(COALESCE(amount, 0))::BIGINT
    ELSE ROUND(COALESCE(amount, 0) * 100)::BIGINT
END
WHERE amount_minor IS NULL;

ALTER TABLE stripe_disputes
    ALTER COLUMN amount_minor SET NOT NULL,
    ALTER COLUMN amount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_stripe_disputes_amount_minor_non_negative
        CHECK (amount_minor >= 0);

ALTER TABLE payment_risk_checkout_decisions
    ADD COLUMN IF NOT EXISTS amount_minor BIGINT;

UPDATE payment_risk_checkout_decisions
SET amount_minor = CASE
    WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(COALESCE(amount, 0))::BIGINT
    ELSE ROUND(COALESCE(amount, 0) * 100)::BIGINT
END
WHERE amount_minor IS NULL;

ALTER TABLE payment_risk_checkout_decisions
    ALTER COLUMN amount_minor SET NOT NULL,
    ALTER COLUMN amount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_payment_risk_checkout_decisions_amount_minor_non_negative
        CHECK (amount_minor >= 0);

ALTER TABLE payment_risk_snapshots
    ADD COLUMN IF NOT EXISTS successful_payment_amount_minor_by_currency JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN IF NOT EXISTS dispute_amount_minor_by_currency JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN IF NOT EXISTS refund_amount_minor_by_currency JSONB NOT NULL DEFAULT '{}'::JSONB;

