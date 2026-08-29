-- Durable application-wide IP/CIDR blocking rules.
--
-- The table intentionally keeps the rule source separate from the matching
-- engine. Visitor-profile actions, commercial crawler detection, and future
-- risk automation can all create rules without duplicating enforcement code.

CREATE TABLE IF NOT EXISTS global_ip_block_rules (
    id BIGSERIAL PRIMARY KEY,
    cidr VARCHAR(120) NOT NULL,
    source VARCHAR(64) NOT NULL DEFAULT 'manual',
    source_reference VARCHAR(160) NOT NULL DEFAULT '',
    reason VARCHAR(500) NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_by BIGINT NULL,
    disabled_by BIGINT NULL,
    disabled_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_global_ip_block_rules_cidr_live
    ON global_ip_block_rules (cidr)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_global_ip_block_rules_active_live
    ON global_ip_block_rules (enabled, expires_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_global_ip_block_rules_source_live
    ON global_ip_block_rules (source, source_reference)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_global_ip_block_rules_created_at_live
    ON global_ip_block_rules (created_at DESC)
    WHERE deleted_at IS NULL;

ALTER TABLE visitor_profiles
    ADD COLUMN IF NOT EXISTS ip_address VARCHAR(45) NULL;

CREATE INDEX IF NOT EXISTS idx_visitor_profiles_ip_address_live
    ON visitor_profiles (ip_address)
    WHERE deleted_at IS NULL AND ip_address IS NOT NULL AND ip_address <> '';
