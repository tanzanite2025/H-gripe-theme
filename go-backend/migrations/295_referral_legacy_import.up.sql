-- Preserve the old referrals rows while importing them into the v2 lifecycle
-- model. The legacy key is intentionally retained only as an internal audit
-- pointer so the import is rerunnable and never duplicates a record.
ALTER TABLE referral_records
    ADD COLUMN IF NOT EXISTS legacy_referral_id BIGINT;

CREATE UNIQUE INDEX IF NOT EXISTS uq_referral_records_legacy_referral
    ON referral_records(legacy_referral_id);

INSERT INTO user_referral_identities (user_id, referral_code, is_active)
SELECT DISTINCT ON (legacy.referrer_id)
    legacy.referrer_id,
    CASE
        WHEN UPPER(BTRIM(COALESCE(legacy.referral_code, ''))) ~ '^[A-HJ-NP-Z2-9]{6,16}$'
            THEN UPPER(BTRIM(legacy.referral_code))
        ELSE UPPER(SUBSTRING(TRANSLATE(MD5('legacy-referrer:' || legacy.referrer_id::TEXT), '01', '23') FROM 1 FOR 10))
    END,
    TRUE
FROM referrals legacy
JOIN users referrer ON referrer.id = legacy.referrer_id
WHERE legacy.deleted_at IS NULL
ORDER BY legacy.referrer_id, legacy.created_at ASC NULLS LAST, legacy.id ASC
ON CONFLICT (user_id) DO NOTHING;

WITH active_config AS (
    SELECT id, currency, attribution_ttl_days, vesting_period_days
    FROM referral_program_configs
    WHERE status = 'active'
    ORDER BY version DESC
    LIMIT 1
)
INSERT INTO referral_records (
    legacy_referral_id,
    referral_identity_id,
    program_config_id,
    referrer_id,
    referee_id,
    referral_code_snapshot,
    attribution_source,
    currency,
    order_id,
    order_amount_minor,
    status,
    record_version,
    expires_at,
    ordered_at,
    delivered_at,
    vesting_until,
    settled_at,
    revoked_at,
    revoke_reason,
    risk_flags,
    created_at,
    updated_at
)
SELECT
    legacy.id,
    identity.id,
    config.id,
    legacy.referrer_id,
    CASE WHEN referee.id IS NULL THEN NULL ELSE legacy.referred_id END,
    identity.referral_code,
    CASE WHEN BTRIM(COALESCE(legacy.referral_code, '')) = '' THEN 'manual_input' ELSE 'link' END,
    UPPER(COALESCE(NULLIF(o.currency, ''), config.currency)),
    CASE WHEN o.id IS NULL THEN NULL ELSE o.id END,
    COALESCE(
        NULLIF(o.total_amount_minor, 0),
        CASE
            WHEN UPPER(COALESCE(NULLIF(o.currency, ''), config.currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
                THEN ROUND(COALESCE(o.total_amount, 0))::BIGINT
            ELSE ROUND(COALESCE(o.total_amount, 0) * 100)::BIGINT
        END,
        0
    ),
    CASE LOWER(COALESCE(legacy.status, 'pending'))
        WHEN 'completed' THEN 'settled'
        WHEN 'settled' THEN 'settled'
        WHEN 'vesting' THEN 'vesting'
        WHEN 'ordered' THEN 'ordered'
        WHEN 'expired' THEN 'expired'
        WHEN 'revoked' THEN 'revoked'
        WHEN 'reversed' THEN 'reversed'
        ELSE 'pending'
    END,
    1,
    COALESCE(legacy.created_at, CURRENT_TIMESTAMP) + (config.attribution_ttl_days * INTERVAL '1 day'),
    COALESCE(o.paid_at, legacy.completed_at),
    o.delivered_at,
    CASE
        WHEN LOWER(COALESCE(legacy.status, 'pending')) IN ('completed', 'settled', 'vesting', 'reversed')
            THEN COALESCE(o.delivered_at, legacy.completed_at, legacy.created_at, CURRENT_TIMESTAMP) + (config.vesting_period_days * INTERVAL '1 day')
        ELSE NULL
    END,
    CASE WHEN LOWER(COALESCE(legacy.status, 'pending')) IN ('completed', 'settled', 'reversed') THEN COALESCE(legacy.completed_at, legacy.updated_at, legacy.created_at) ELSE NULL END,
    CASE WHEN LOWER(COALESCE(legacy.status, 'pending')) = 'revoked' THEN COALESCE(legacy.updated_at, legacy.created_at) ELSE NULL END,
    CASE WHEN LOWER(COALESCE(legacy.status, 'pending')) IN ('revoked', 'reversed') THEN 'imported from legacy referrals' ELSE '' END,
    '[]'::jsonb,
    COALESCE(legacy.created_at, CURRENT_TIMESTAMP),
    COALESCE(legacy.updated_at, legacy.created_at, CURRENT_TIMESTAMP)
FROM referrals legacy
JOIN users referrer ON referrer.id = legacy.referrer_id
JOIN user_referral_identities identity ON identity.user_id = legacy.referrer_id
CROSS JOIN active_config config
LEFT JOIN users referee ON referee.id = legacy.referred_id
LEFT JOIN orders o ON o.id = legacy.completed_order_id
WHERE legacy.deleted_at IS NULL
ON CONFLICT DO NOTHING;

INSERT INTO referral_reward_logs (
    referral_record_id,
    program_config_id,
    recipient_user_id,
    recipient_role,
    reward_type,
    points_amount,
    idempotency_key,
    status,
    rule_snapshot,
    released_at,
    created_at,
    updated_at
)
SELECT
    record.id,
    record.program_config_id,
    record.referrer_id,
    'referrer',
    'points',
    GREATEST(COALESCE(NULLIF(legacy.referrer_points, 0), legacy.points_earned, 0), 0),
    'legacy-referral:' || legacy.id::TEXT || ':referrer',
    CASE WHEN record.status IN ('settled', 'reversed') THEN 'released' ELSE 'locked' END,
    jsonb_build_object('source', 'legacy_referrals', 'legacy_referral_id', legacy.id),
    CASE WHEN record.status IN ('settled', 'reversed') THEN COALESCE(legacy.completed_at, legacy.updated_at, legacy.created_at) ELSE NULL END,
    record.created_at,
    record.updated_at
FROM referral_records record
JOIN referrals legacy ON legacy.id = record.legacy_referral_id
WHERE GREATEST(COALESCE(NULLIF(legacy.referrer_points, 0), legacy.points_earned, 0), 0) > 0
ON CONFLICT (idempotency_key) DO NOTHING;

INSERT INTO referral_reward_logs (
    referral_record_id,
    program_config_id,
    recipient_user_id,
    recipient_role,
    reward_type,
    points_amount,
    idempotency_key,
    status,
    rule_snapshot,
    released_at,
    created_at,
    updated_at
)
SELECT
    record.id,
    record.program_config_id,
    record.referee_id,
    'referee',
    'points',
    GREATEST(COALESCE(legacy.referred_points, 0), 0),
    'legacy-referral:' || legacy.id::TEXT || ':referee',
    CASE WHEN record.status IN ('settled', 'reversed') THEN 'released' ELSE 'locked' END,
    jsonb_build_object('source', 'legacy_referrals', 'legacy_referral_id', legacy.id),
    CASE WHEN record.status IN ('settled', 'reversed') THEN COALESCE(legacy.completed_at, legacy.updated_at, legacy.created_at) ELSE NULL END,
    record.created_at,
    record.updated_at
FROM referral_records record
JOIN referrals legacy ON legacy.id = record.legacy_referral_id
WHERE record.referee_id IS NOT NULL
  AND legacy.referred_points > 0
ON CONFLICT (idempotency_key) DO NOTHING;
