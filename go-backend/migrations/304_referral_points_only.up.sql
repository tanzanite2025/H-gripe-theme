-- Referral rewards are points only. Monetary order qualification remains
-- separate from the points amounts and is not part of the reward contract.
ALTER TABLE referral_program_configs
    DROP COLUMN IF EXISTS referee_benefit_max_amount_minor,
    DROP COLUMN IF EXISTS coupon_stackable;

UPDATE referral_program_configs
SET referee_benefit_type = 'points'
WHERE referee_benefit_type <> 'points';

UPDATE referral_program_configs
SET referee_benefit_value = 50
WHERE referee_benefit_value <= 0;

ALTER TABLE referral_program_configs
    DROP CONSTRAINT IF EXISTS referral_program_configs_benefit_valid;

ALTER TABLE referral_program_configs
    ADD CONSTRAINT referral_program_configs_benefit_valid
    CHECK (referee_benefit_type = 'points' AND referee_benefit_value > 0);

ALTER TABLE referral_reward_logs
    DROP CONSTRAINT IF EXISTS referral_reward_logs_type_valid,
    DROP CONSTRAINT IF EXISTS referral_reward_logs_target_valid;

DELETE FROM referral_reward_logs
WHERE reward_type <> 'points';

ALTER TABLE referral_reward_logs
    DROP COLUMN IF EXISTS coupon_id;

ALTER TABLE referral_reward_logs
    ADD CONSTRAINT referral_reward_logs_type_valid
    CHECK (reward_type = 'points'),
    ADD CONSTRAINT referral_reward_logs_target_valid
    CHECK (reward_type = 'points' AND points_amount > 0);
