-- Stage five: option-level production lead time and post-purchase policies.
ALTER TABLE product_custom_option_policies
    ADD COLUMN IF NOT EXISTS production_lead_time_days INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS requires_production BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS cancellation_policy VARCHAR(32) NOT NULL DEFAULT 'standard',
    ADD COLUMN IF NOT EXISTS return_policy VARCHAR(32) NOT NULL DEFAULT 'standard';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'chk_product_custom_option_policy_production_return_policies'
    ) THEN
        ALTER TABLE product_custom_option_policies
            ADD CONSTRAINT chk_product_custom_option_policy_production_return_policies
            CHECK (
                production_lead_time_days >= 0
                AND cancellation_policy IN ('standard', 'before_production', 'never')
                AND return_policy IN ('standard', 'not_allowed')
            );
    END IF;
END $$;
