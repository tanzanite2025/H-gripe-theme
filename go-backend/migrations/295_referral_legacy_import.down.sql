DELETE FROM referral_reward_logs
WHERE idempotency_key LIKE 'legacy-referral:%';

DELETE FROM referral_records
WHERE legacy_referral_id IS NOT NULL;

DROP INDEX IF EXISTS uq_referral_records_legacy_referral;

ALTER TABLE referral_records
    DROP COLUMN IF EXISTS legacy_referral_id;
