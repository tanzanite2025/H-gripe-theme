ALTER TABLE product_categories
    ADD COLUMN IF NOT EXISTS image_media_asset_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_product_categories_image_media_asset'
    ) THEN
        ALTER TABLE product_categories
            ADD CONSTRAINT fk_product_categories_image_media_asset
            FOREIGN KEY (image_media_asset_id)
            REFERENCES media_assets(id)
            ON DELETE SET NULL;
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_product_categories_image_media_asset_id
    ON product_categories(image_media_asset_id);
