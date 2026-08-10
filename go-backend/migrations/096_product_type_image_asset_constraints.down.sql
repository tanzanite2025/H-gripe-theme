ALTER TABLE product_types
    DROP CONSTRAINT IF EXISTS ck_product_types_image_reference_pair,
    DROP CONSTRAINT IF EXISTS fk_product_types_image_media_asset;
