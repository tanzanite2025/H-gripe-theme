-- Compact daily visitor risk facts. This table intentionally stores hashed
-- request identity and aggregate counters only; it is not a raw request log.

CREATE TABLE IF NOT EXISTS visitor_risk_daily_facts (
    id BIGSERIAL PRIMARY KEY,
    day DATE NOT NULL,
    ip_hash VARCHAR(64) NOT NULL,
    user_agent_hash VARCHAR(64) NOT NULL DEFAULT '',
    country_code VARCHAR(8),
    first_seen_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    request_count INTEGER NOT NULL DEFAULT 0,
    unique_path_count INTEGER NOT NULL DEFAULT 0,
    unique_anonymous_count INTEGER NOT NULL DEFAULT 0,
    unique_session_count INTEGER NOT NULL DEFAULT 0,
    invalid_request_count INTEGER NOT NULL DEFAULT 0,
    auth_failure_count INTEGER NOT NULL DEFAULT 0,
    checkout_failure_count INTEGER NOT NULL DEFAULT 0,
    bot_like_user_agent_count INTEGER NOT NULL DEFAULT 0,
    no_cookie_request_count INTEGER NOT NULL DEFAULT 0,
    meaningful_action_count INTEGER NOT NULL DEFAULT 0,
    risk_score INTEGER NOT NULL DEFAULT 0,
    risk_level VARCHAR(16) NOT NULL DEFAULT 'normal',
    sample_paths JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_visitor_risk_daily_fact
    ON visitor_risk_daily_facts(day, ip_hash, user_agent_hash);

CREATE INDEX IF NOT EXISTS idx_visitor_risk_daily_facts_day_score
    ON visitor_risk_daily_facts(day, risk_score DESC);

CREATE INDEX IF NOT EXISTS idx_visitor_risk_daily_facts_ip_day
    ON visitor_risk_daily_facts(ip_hash, day DESC);

CREATE INDEX IF NOT EXISTS idx_visitor_risk_daily_facts_level_day
    ON visitor_risk_daily_facts(risk_level, day DESC);
