ALTER TABLE referral_program_configs
    DROP CONSTRAINT IF EXISTS referral_program_configs_benefit_valid;

ALTER TABLE referral_program_configs
    ADD COLUMN IF NOT EXISTS referee_benefit_max_amount_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS coupon_stackable BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE referral_program_configs
    ADD CONSTRAINT referral_program_configs_benefit_valid
    CHECK (referee_benefit_type = 'points' AND referee_benefit_value > 0);

ALTER TABLE referral_reward_logs
    ADD COLUMN IF NOT EXISTS coupon_id BIGINT;

ALTER TABLE referral_reward_logs
    DROP CONSTRAINT IF EXISTS referral_reward_logs_type_valid,
    DROP CONSTRAINT IF EXISTS referral_reward_logs_target_valid;

ALTER TABLE referral_reward_logs
    ADD CONSTRAINT referral_reward_logs_type_valid
    CHECK (reward_type IN ('points', 'coupon')),
    ADD CONSTRAINT referral_reward_logs_target_valid
    CHECK (
        (reward_type = 'points' AND points_amount > 0)
        OR (reward_type = 'coupon' AND coupon_id IS NOT NULL)
    );
