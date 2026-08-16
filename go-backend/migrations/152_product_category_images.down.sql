DROP INDEX IF EXISTS idx_product_categories_image_media_asset_id;

ALTER TABLE product_categories
    DROP CONSTRAINT IF EXISTS fk_product_categories_image_media_asset;

ALTER TABLE product_categories
    DROP COLUMN IF EXISTS image_media_asset_id,
    DROP COLUMN IF EXISTS image_url;
