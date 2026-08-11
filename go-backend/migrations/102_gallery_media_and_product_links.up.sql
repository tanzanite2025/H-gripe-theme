-- Gallery images use the shared media library, while gallery product links
-- keep product navigation out of storefront code.

ALTER TABLE gallery_images
    ADD COLUMN IF NOT EXISTS media_asset_id BIGINT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_gallery_images_media_asset'
    ) THEN
        ALTER TABLE gallery_images
            ADD CONSTRAINT fk_gallery_images_media_asset
            FOREIGN KEY (media_asset_id)
            REFERENCES media_assets(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_gallery_images_media_asset_id
    ON gallery_images(media_asset_id);

CREATE TABLE IF NOT EXISTS gallery_product_links (
    id BIGSERIAL PRIMARY KEY,
    gallery_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_gallery_product_links_gallery
        FOREIGN KEY (gallery_id) REFERENCES galleries(id) ON DELETE CASCADE,
    CONSTRAINT fk_gallery_product_links_product
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_gallery_product_links_unique
    ON gallery_product_links(gallery_id, product_id);
CREATE INDEX IF NOT EXISTS idx_gallery_product_links_gallery_sort
    ON gallery_product_links(gallery_id, sort_order, id);
CREATE INDEX IF NOT EXISTS idx_gallery_product_links_product_id
    ON gallery_product_links(product_id);
