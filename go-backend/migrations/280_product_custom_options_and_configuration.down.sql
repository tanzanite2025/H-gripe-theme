DROP TRIGGER IF EXISTS trigger_prevent_order_item_configuration_snapshot_mutation ON order_items;
DROP FUNCTION IF EXISTS prevent_order_item_configuration_snapshot_mutation();

ALTER TABLE order_items
    DROP COLUMN IF EXISTS configuration_snapshot;

DROP INDEX IF EXISTS uq_cart_items_identity;
CREATE INDEX IF NOT EXISTS idx_cart_items_product_variant
    ON cart_items(cart_id, product_id, variant_id);
ALTER TABLE cart_items
    DROP COLUMN IF EXISTS configuration_hash,
    DROP COLUMN IF EXISTS configuration;

DROP INDEX IF EXISTS idx_product_custom_option_policies_component_variant;
DROP TABLE IF EXISTS product_custom_option_policies;

DROP INDEX IF EXISTS idx_product_variant_option_values_template_item;
ALTER TABLE product_variant_option_values
    DROP COLUMN IF EXISTS source_template_revision,
    DROP COLUMN IF EXISTS template_option_item_id;

ALTER TABLE product_specification_templates
    DROP COLUMN IF EXISTS revision;

ALTER TABLE product_spec_definitions
    DROP CONSTRAINT IF EXISTS chk_product_spec_definitions_selection_bounds,
    DROP CONSTRAINT IF EXISTS chk_product_spec_definitions_selection_mode,
    DROP CONSTRAINT IF EXISTS chk_product_spec_definitions_role;

ALTER TABLE product_spec_definitions
    DROP COLUMN IF EXISTS max_selections,
    DROP COLUMN IF EXISTS min_selections,
    DROP COLUMN IF EXISTS selection_mode,
    DROP COLUMN IF EXISTS role;
