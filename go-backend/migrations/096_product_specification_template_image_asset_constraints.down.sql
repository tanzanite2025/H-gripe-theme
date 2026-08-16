ALTER TABLE product_specification_templates
    DROP CONSTRAINT IF EXISTS ck_product_specification_templates_image_reference_pair,
    DROP CONSTRAINT IF EXISTS fk_product_specification_templates_image_media_asset;
