DROP INDEX IF EXISTS idx_product_specification_templates_system_managed;

ALTER TABLE product_specification_templates
    DROP COLUMN IF EXISTS is_system_managed;
