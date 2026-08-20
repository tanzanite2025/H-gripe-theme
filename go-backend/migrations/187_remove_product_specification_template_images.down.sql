ALTER TABLE product_specification_templates
    ADD COLUMN IF NOT EXISTS image_media_asset_id BIGINT,
    ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_product_specification_templates_image_media_asset_id
    ON product_specification_templates(image_media_asset_id);

ALTER TABLE product_specification_templates
    ADD CONSTRAINT fk_product_specification_templates_image_media_asset
    FOREIGN KEY (image_media_asset_id)
    REFERENCES media_assets(id)
    ON DELETE RESTRICT,
    ADD CONSTRAINT ck_product_specification_templates_image_reference_pair
    CHECK (
        (image_media_asset_id IS NULL AND image_url = '')
        OR (image_media_asset_id IS NOT NULL AND length(btrim(image_url)) > 0)
    );
