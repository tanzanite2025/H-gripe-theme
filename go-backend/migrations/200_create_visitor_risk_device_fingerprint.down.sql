DROP INDEX IF EXISTS uk_visitor_risk_daily_fact;
DROP INDEX IF EXISTS uk_visitor_risk_daily_fact_device;

CREATE UNIQUE INDEX IF NOT EXISTS uk_visitor_risk_daily_fact
    ON visitor_risk_daily_facts(day, ip_hash, user_agent_hash);

ALTER TABLE visitor_risk_daily_facts
    DROP COLUMN IF EXISTS device_fingerprint_hash;
