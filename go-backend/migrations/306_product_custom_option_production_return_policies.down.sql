ALTER TABLE product_custom_option_policies
    DROP CONSTRAINT IF EXISTS chk_product_custom_option_policy_production_return_policies;

ALTER TABLE product_custom_option_policies
    DROP COLUMN IF EXISTS return_policy,
    DROP COLUMN IF EXISTS cancellation_policy,
    DROP COLUMN IF EXISTS requires_production,
    DROP COLUMN IF EXISTS production_lead_time_days;
