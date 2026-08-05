ALTER TABLE payment_risk_events
    ADD COLUMN IF NOT EXISTS provider_payment_id VARCHAR(255) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_payment_risk_events_provider_payment_id
    ON payment_risk_events(provider_payment_id);
