ALTER TABLE product_types
    ADD COLUMN IF NOT EXISTS image_media_asset_id BIGINT,
    ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_product_types_image_media_asset_id
    ON product_types(image_media_asset_id);
