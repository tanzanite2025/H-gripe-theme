DROP INDEX IF EXISTS idx_product_types_image_media_asset_id;

ALTER TABLE product_types
    DROP COLUMN IF EXISTS image_media_asset_id,
    DROP COLUMN IF EXISTS image_url;
