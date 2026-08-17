ALTER TABLE product_media
    DROP COLUMN IF EXISTS image_variants;

DROP TABLE IF EXISTS media_asset_derivatives;
