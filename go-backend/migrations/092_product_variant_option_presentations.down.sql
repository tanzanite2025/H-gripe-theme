ALTER TABLE product_media
    DROP CONSTRAINT IF EXISTS fk_product_media_variant_option_value;

DROP INDEX IF EXISTS idx_product_media_variant_option_value_id;

ALTER TABLE product_media
    DROP COLUMN IF EXISTS variant_option_value_id;

DROP INDEX IF EXISTS idx_product_variant_option_values_asset;
DROP INDEX IF EXISTS idx_product_variant_option_values_product;

DROP TABLE IF EXISTS product_variant_option_values;

ALTER TABLE product_spec_definitions
    DROP COLUMN IF EXISTS presentation;
