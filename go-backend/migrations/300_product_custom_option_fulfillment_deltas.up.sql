-- Stage five: option-level physical fulfillment deltas. Values are stored in
-- grams and remain part of the materialized product policy, not the template
-- presentation metadata.
ALTER TABLE product_custom_option_policies
    ADD COLUMN IF NOT EXISTS weight_delta_grams INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS packaging_weight_delta_grams INTEGER NOT NULL DEFAULT 0;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'chk_product_custom_option_policy_weight_deltas'
    ) THEN
        ALTER TABLE product_custom_option_policies
            ADD CONSTRAINT chk_product_custom_option_policy_weight_deltas
            CHECK (weight_delta_grams >= 0 AND packaging_weight_delta_grams >= 0);
    END IF;
END $$;
