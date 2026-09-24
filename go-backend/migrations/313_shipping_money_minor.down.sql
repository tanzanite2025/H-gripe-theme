ALTER TABLE shipping_rules
    DROP CONSTRAINT IF EXISTS chk_shipping_rules_money_minor_non_negative;
ALTER TABLE shipping_rules
    DROP COLUMN IF EXISTS min_value_minor,
    DROP COLUMN IF EXISTS max_value_minor,
    DROP COLUMN IF EXISTS fee_minor,
    DROP COLUMN IF EXISTS additional_minor;

ALTER TABLE shipping_templates
    DROP CONSTRAINT IF EXISTS chk_shipping_templates_money_minor_non_negative;
ALTER TABLE shipping_templates
    DROP COLUMN IF EXISTS free_threshold_minor,
    DROP COLUMN IF EXISTS default_fee_minor;

