DROP INDEX IF EXISTS idx_payment_risk_events_provider_payment_id;

ALTER TABLE payment_risk_events
    DROP COLUMN IF EXISTS provider_payment_id;
