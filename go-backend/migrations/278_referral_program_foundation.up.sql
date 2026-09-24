-- Referral v2 is additive. The legacy referrals table remains readable while
-- the new program runs in disabled/shadow mode.

CREATE TABLE IF NOT EXISTS referral_program_configs (
    id BIGSERIAL PRIMARY KEY,
    version INTEGER NOT NULL UNIQUE,
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    min_order_amount_minor BIGINT NOT NULL DEFAULT 20000,
    referrer_reward_points INTEGER NOT NULL DEFAULT 1000,
    referee_benefit_type VARCHAR(24) NOT NULL DEFAULT 'points',
    referee_benefit_value BIGINT NOT NULL DEFAULT 0,
    vesting_period_days INTEGER NOT NULL DEFAULT 30,
    undelivered_fallback_days INTEGER NOT NULL DEFAULT 45,
    attribution_ttl_days INTEGER NOT NULL DEFAULT 30,
    monthly_cap_per_referrer INTEGER NOT NULL DEFAULT 10,
    anti_fraud_mode VARCHAR(16) NOT NULL DEFAULT 'monitor',
    source_loyalty_program_config_id BIGINT REFERENCES loyalty_program_configs(id) ON DELETE RESTRICT,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT referral_program_configs_status_valid CHECK (status IN ('active', 'archived')),
    CONSTRAINT referral_program_configs_currency_valid CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT referral_program_configs_amounts_valid CHECK (
        min_order_amount_minor >= 0
        AND referrer_reward_points >= 0
        AND referee_benefit_value >= 0
    ),
    CONSTRAINT referral_program_configs_benefit_valid CHECK (
        referee_benefit_type = 'points'
    ),
    CONSTRAINT referral_program_configs_windows_valid CHECK (
        vesting_period_days > 0
        AND undelivered_fallback_days >= vesting_period_days
        AND attribution_ttl_days BETWEEN 1 AND 90
        AND monthly_cap_per_referrer > 0
    ),
    CONSTRAINT referral_program_configs_fraud_mode_valid CHECK (anti_fraud_mode IN ('monitor', 'strict'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_referral_program_configs_active
    ON referral_program_configs(status)
    WHERE status = 'active';

CREATE TABLE IF NOT EXISTS user_referral_identities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    referral_code VARCHAR(16) NOT NULL,
    custom_slug VARCHAR(64),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_referral_identities_user_unique UNIQUE (user_id),
    CONSTRAINT user_referral_identities_code_format CHECK (referral_code ~ '^[A-HJ-NP-Z2-9]{6,16}$'),
    CONSTRAINT user_referral_identities_slug_nonblank CHECK (custom_slug IS NULL OR BTRIM(custom_slug) <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_referral_identities_code_ci
    ON user_referral_identities(UPPER(referral_code));
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_referral_identities_slug_ci
    ON user_referral_identities(LOWER(custom_slug))
    WHERE custom_slug IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_user_referral_identities_active
    ON user_referral_identities(is_active);

CREATE TABLE IF NOT EXISTS referral_records (
    id BIGSERIAL PRIMARY KEY,
    referral_identity_id BIGINT NOT NULL REFERENCES user_referral_identities(id) ON DELETE RESTRICT,
    program_config_id BIGINT NOT NULL REFERENCES referral_program_configs(id) ON DELETE RESTRICT,
    referrer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    referee_id BIGINT REFERENCES users(id) ON DELETE RESTRICT,
    referral_code_snapshot VARCHAR(16) NOT NULL,
    attribution_source VARCHAR(24) NOT NULL,
    referee_email_hash VARCHAR(64) NOT NULL DEFAULT '',
    client_ip_hash VARCHAR(64) NOT NULL DEFAULT '',
    device_fingerprint_hash VARCHAR(64) NOT NULL DEFAULT '',
    shipping_address_hash VARCHAR(64) NOT NULL DEFAULT '',
    shipping_phone_hash VARCHAR(64) NOT NULL DEFAULT '',
    payment_fingerprint_hash VARCHAR(64) NOT NULL DEFAULT '',
    hash_key_version INTEGER NOT NULL DEFAULT 1,
    order_id BIGINT REFERENCES orders(id) ON DELETE RESTRICT,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    order_amount_minor BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    record_version INTEGER NOT NULL DEFAULT 1,
    expires_at TIMESTAMPTZ NOT NULL,
    ordered_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    vesting_until TIMESTAMPTZ,
    settled_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    reversed_at TIMESTAMPTZ,
    revoke_reason TEXT NOT NULL DEFAULT '',
    reverse_reason TEXT NOT NULL DEFAULT '',
    risk_flags JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT referral_records_not_self CHECK (referee_id IS NULL OR referrer_id <> referee_id),
    CONSTRAINT referral_records_status_valid CHECK (
        status IN ('pending', 'ordered', 'vesting', 'settled', 'expired', 'revoked', 'reversed')
    ),
    CONSTRAINT referral_records_attribution_source_valid CHECK (attribution_source IN ('link', 'manual_input')),
    CONSTRAINT referral_records_amount_valid CHECK (order_amount_minor >= 0),
    CONSTRAINT referral_records_version_valid CHECK (record_version > 0 AND hash_key_version > 0),
    CONSTRAINT referral_records_currency_valid CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT referral_records_risk_flags_array CHECK (jsonb_typeof(risk_flags) = 'array')
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_referral_records_referee
    ON referral_records(referee_id)
    WHERE referee_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_referral_records_order
    ON referral_records(order_id)
    WHERE order_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_referral_records_referrer_created
    ON referral_records(referrer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_referral_records_status_vesting
    ON referral_records(status, vesting_until)
    WHERE status = 'vesting';
CREATE INDEX IF NOT EXISTS idx_referral_records_pending_expiry
    ON referral_records(status, expires_at)
    WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS referral_reward_logs (
    id BIGSERIAL PRIMARY KEY,
    referral_record_id BIGINT NOT NULL REFERENCES referral_records(id) ON DELETE RESTRICT,
    program_config_id BIGINT NOT NULL REFERENCES referral_program_configs(id) ON DELETE RESTRICT,
    recipient_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    recipient_role VARCHAR(16) NOT NULL,
    reward_type VARCHAR(16) NOT NULL,
    points_amount INTEGER NOT NULL DEFAULT 0,
    coupon_id BIGINT REFERENCES coupons(id) ON DELETE RESTRICT,
    loyalty_transaction_id BIGINT REFERENCES loyalty_transactions(id) ON DELETE RESTRICT,
    idempotency_key VARCHAR(160) NOT NULL UNIQUE,
    status VARCHAR(16) NOT NULL DEFAULT 'locked',
    rule_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    released_at TIMESTAMPTZ,
    forfeited_at TIMESTAMPTZ,
    reversed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT referral_reward_logs_recipient_valid CHECK (recipient_role IN ('referrer', 'referee')),
    CONSTRAINT referral_reward_logs_type_valid CHECK (reward_type IN ('points', 'coupon')),
    CONSTRAINT referral_reward_logs_status_valid CHECK (status IN ('locked', 'released', 'forfeited', 'reversed')),
    CONSTRAINT referral_reward_logs_amount_valid CHECK (points_amount >= 0),
    CONSTRAINT referral_reward_logs_target_valid CHECK (
        (reward_type = 'points' AND points_amount > 0)
        OR (reward_type = 'coupon' AND coupon_id IS NOT NULL)
    ),
    CONSTRAINT referral_reward_logs_rule_snapshot_object CHECK (jsonb_typeof(rule_snapshot) = 'object'),
    CONSTRAINT referral_reward_logs_once UNIQUE (referral_record_id, recipient_role, reward_type)
);

CREATE INDEX IF NOT EXISTS idx_referral_reward_logs_recipient
    ON referral_reward_logs(recipient_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_referral_reward_logs_status
    ON referral_reward_logs(status);

CREATE TABLE IF NOT EXISTS referral_transitions (
    id BIGSERIAL PRIMARY KEY,
    referral_record_id BIGINT NOT NULL REFERENCES referral_records(id) ON DELETE RESTRICT,
    from_status VARCHAR(16) NOT NULL,
    to_status VARCHAR(16) NOT NULL,
    trigger VARCHAR(40) NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    actor_type VARCHAR(24) NOT NULL,
    actor_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    event_key VARCHAR(160),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT referral_transitions_state_changed CHECK (from_status <> to_status),
    CONSTRAINT referral_transitions_metadata_object CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE INDEX IF NOT EXISTS idx_referral_transitions_record_created
    ON referral_transitions(referral_record_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uq_referral_transitions_event
    ON referral_transitions(referral_record_id, event_key)
    WHERE event_key IS NOT NULL;

CREATE OR REPLACE FUNCTION prevent_referral_transition_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'referral transitions are append-only';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_referral_transitions_append_only ON referral_transitions;
CREATE TRIGGER trg_referral_transitions_append_only
    BEFORE UPDATE OR DELETE ON referral_transitions
    FOR EACH ROW EXECUTE FUNCTION prevent_referral_transition_mutation();

-- Preserve the currently configured legacy point values as a disabled v1
-- snapshot. Enabling referral v2 always requires an explicit admin publish.
INSERT INTO referral_program_configs (
    version,
    status,
    enabled,
    currency,
    min_order_amount_minor,
    referrer_reward_points,
    referee_benefit_type,
    referee_benefit_value,
    vesting_period_days,
    undelivered_fallback_days,
    attribution_ttl_days,
    monthly_cap_per_referrer,
    anti_fraud_mode,
    source_loyalty_program_config_id
)
SELECT
    1,
    'active',
    FALSE,
    'USD',
    20000,
    active.referral_referrer_points,
    'points',
    active.referral_referee_points,
    30,
    45,
    30,
    10,
    'monitor',
    active.id
FROM loyalty_program_configs active
WHERE active.status = 'active'
  AND NOT EXISTS (SELECT 1 FROM referral_program_configs)
ORDER BY active.version DESC
LIMIT 1;

INSERT INTO referral_program_configs (
    version,
    status,
    enabled,
    currency,
    min_order_amount_minor,
    referrer_reward_points,
    referee_benefit_type,
    referee_benefit_value,
    vesting_period_days,
    undelivered_fallback_days,
    attribution_ttl_days,
    monthly_cap_per_referrer,
    anti_fraud_mode
)
SELECT 1, 'active', FALSE, 'USD', 20000, 1000, 'points', 50, 30, 45, 30, 10, 'monitor'
WHERE NOT EXISTS (SELECT 1 FROM referral_program_configs);
