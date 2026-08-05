-- Time-bounded manual payment protection controls.
-- The first supported action is force_3ds. Checkout blocking and provider
-- routing will use this foundation in later migrations.

CREATE TABLE IF NOT EXISTS payment_protection_controls (
    id BIGSERIAL PRIMARY KEY,
    action VARCHAR(48) NOT NULL,
    scope_type VARCHAR(32) NOT NULL,
    scope_value VARCHAR(128) NOT NULL DEFAULT '',
    reason TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_by BIGINT NOT NULL,
    updated_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_protection_controls_active
    ON payment_protection_controls(enabled, expires_at);
CREATE INDEX IF NOT EXISTS idx_payment_protection_controls_scope
    ON payment_protection_controls(scope_type, scope_value);
