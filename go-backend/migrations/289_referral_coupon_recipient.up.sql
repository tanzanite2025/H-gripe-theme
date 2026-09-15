ALTER TABLE coupons
    ADD COLUMN IF NOT EXISTS referral_recipient_user_id BIGINT REFERENCES users(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_coupons_referral_recipient_user
    ON coupons(referral_recipient_user_id)
    WHERE referral_recipient_user_id IS NOT NULL;
