ALTER TABLE product_specification_templates
    DROP CONSTRAINT IF EXISTS ck_product_specification_templates_image_reference_pair,
    DROP CONSTRAINT IF EXISTS fk_product_specification_templates_image_media_asset;

DROP INDEX IF EXISTS idx_product_specification_templates_image_media_asset_id;

ALTER TABLE product_specification_templates
    DROP COLUMN IF EXISTS image_media_asset_id,
    DROP COLUMN IF EXISTS image_url;
