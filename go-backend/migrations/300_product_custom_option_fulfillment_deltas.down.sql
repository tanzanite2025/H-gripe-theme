ALTER TABLE product_custom_option_policies
    DROP CONSTRAINT IF EXISTS chk_product_custom_option_policy_weight_deltas;

ALTER TABLE product_custom_option_policies
    DROP COLUMN IF EXISTS packaging_weight_delta_grams,
    DROP COLUMN IF EXISTS weight_delta_grams;
