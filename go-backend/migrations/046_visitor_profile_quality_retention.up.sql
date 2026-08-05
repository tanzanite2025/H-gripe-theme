-- Add profile quality, status, and retention fields so visitor_profiles can
-- represent business-worthy profiles without becoming a raw traffic log.

ALTER TABLE visitor_profiles
    ADD COLUMN IF NOT EXISTS profile_quality_score INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS profile_status VARCHAR(24) NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS last_meaningful_action VARCHAR(64) NULL,
    ADD COLUMN IF NOT EXISTS first_meaningful_seen_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS last_meaningful_seen_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS retention_until TIMESTAMPTZ NULL;

UPDATE visitor_profiles
SET
    profile_status = COALESCE(NULLIF(profile_status, ''), 'active'),
    profile_quality_score = GREATEST(
        profile_quality_score,
        CASE
            WHEN user_id IS NOT NULL THEN 20
            WHEN email IS NOT NULL AND email <> '' THEN 12
            WHEN customer_service_visitor_hash IS NOT NULL AND customer_service_visitor_hash <> '' THEN 12
            WHEN cart_session_id IS NOT NULL AND cart_session_id <> '' THEN 8
            ELSE 0
        END
    ),
    last_meaningful_action = COALESCE(
        NULLIF(last_meaningful_action, ''),
        CASE
            WHEN user_id IS NOT NULL THEN 'account'
            WHEN email IS NOT NULL AND email <> '' THEN 'email_capture'
            WHEN customer_service_visitor_hash IS NOT NULL AND customer_service_visitor_hash <> '' THEN 'customer_service'
            WHEN cart_session_id IS NOT NULL AND cart_session_id <> '' THEN 'cart_action'
            ELSE NULL
        END
    ),
    first_meaningful_seen_at = COALESCE(first_meaningful_seen_at, created_at, last_seen_at, NOW()),
    last_meaningful_seen_at = COALESCE(last_meaningful_seen_at, last_seen_at, updated_at, NOW()),
    retention_until = CASE
        WHEN user_id IS NOT NULL THEN NULL
        ELSE COALESCE(retention_until, COALESCE(last_seen_at, updated_at, NOW()) + INTERVAL '180 days')
    END
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_visitor_profiles_status_live
    ON visitor_profiles (profile_status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_visitor_profiles_quality_live
    ON visitor_profiles (profile_quality_score DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_visitor_profiles_last_meaningful_live
    ON visitor_profiles (last_meaningful_seen_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_visitor_profiles_retention_until_live
    ON visitor_profiles (retention_until)
    WHERE deleted_at IS NULL AND retention_until IS NOT NULL;
