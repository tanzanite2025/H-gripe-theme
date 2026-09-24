DROP INDEX IF EXISTS idx_referral_records_ip_subnet_created;

ALTER TABLE referral_records
    DROP COLUMN IF EXISTS client_ip_subnet_hash;
