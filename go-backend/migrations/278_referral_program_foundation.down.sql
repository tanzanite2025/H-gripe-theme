DROP TRIGGER IF EXISTS trg_referral_transitions_append_only ON referral_transitions;
DROP FUNCTION IF EXISTS prevent_referral_transition_mutation();
DROP TABLE IF EXISTS referral_transitions;
DROP TABLE IF EXISTS referral_reward_logs;
DROP TABLE IF EXISTS referral_records;
DROP TABLE IF EXISTS user_referral_identities;
DROP TABLE IF EXISTS referral_program_configs;
