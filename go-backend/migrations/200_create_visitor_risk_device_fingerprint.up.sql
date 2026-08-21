ALTER TABLE visitor_risk_daily_facts
    ADD COLUMN IF NOT EXISTS device_fingerprint_hash VARCHAR(64) NOT NULL DEFAULT '';

DROP INDEX IF EXISTS uk_visitor_risk_daily_fact;
DROP INDEX IF EXISTS uk_visitor_risk_daily_fact_device;

CREATE UNIQUE INDEX IF NOT EXISTS uk_visitor_risk_daily_fact
    ON visitor_risk_daily_facts(day, ip_hash, user_agent_hash)
    WHERE device_fingerprint_hash = '';

CREATE UNIQUE INDEX IF NOT EXISTS uk_visitor_risk_daily_fact_device
    ON visitor_risk_daily_facts(day, device_fingerprint_hash)
    WHERE device_fingerprint_hash <> '';
