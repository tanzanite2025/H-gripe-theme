DROP INDEX IF EXISTS idx_coupons_referral_recipient_user;

ALTER TABLE coupons
    DROP COLUMN IF EXISTS referral_recipient_user_id;
