ALTER TABLE product_specification_templates
    ADD COLUMN IF NOT EXISTS image_media_asset_id BIGINT,
    ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_product_specification_templates_image_media_asset_id
    ON product_specification_templates(image_media_asset_id);
