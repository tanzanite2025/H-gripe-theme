CREATE TABLE IF NOT EXISTS customer_service_retention_policies (
    id BIGSERIAL PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    interval_seconds INTEGER NOT NULL DEFAULT 86400,
    minimum_retention_days INTEGER NOT NULL DEFAULT 730,
    recovery_window_days INTEGER NOT NULL DEFAULT 30,
    batch_limit INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_customer_service_retention_interval CHECK (interval_seconds BETWEEN 60 AND 604800),
    CONSTRAINT chk_customer_service_retention_minimum_days CHECK (minimum_retention_days BETWEEN 1 AND 36500),
    CONSTRAINT chk_customer_service_retention_recovery_days CHECK (recovery_window_days BETWEEN 1 AND 3650),
    CONSTRAINT chk_customer_service_retention_batch_limit CHECK (batch_limit BETWEEN 1 AND 1000)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_customer_service_retention_policy_singleton
    ON customer_service_retention_policies ((TRUE));
