DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM product_types AS product_type
        LEFT JOIN media_assets AS media_asset
            ON media_asset.id = product_type.image_media_asset_id
        WHERE product_type.image_media_asset_id IS NOT NULL
          AND media_asset.id IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot add product type image asset foreign key: orphan references exist';
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM product_types
        WHERE (image_media_asset_id IS NULL AND length(btrim(COALESCE(image_url, ''))) > 0)
           OR (image_media_asset_id IS NOT NULL AND length(btrim(COALESCE(image_url, ''))) = 0)
    ) THEN
        RAISE EXCEPTION 'cannot add product type image reference check: inconsistent image references exist';
    END IF;
END $$;

ALTER TABLE product_types
    ADD CONSTRAINT fk_product_types_image_media_asset
    FOREIGN KEY (image_media_asset_id)
    REFERENCES media_assets(id)
    ON DELETE RESTRICT,
    ADD CONSTRAINT ck_product_types_image_reference_pair
    CHECK (
        (image_media_asset_id IS NULL AND image_url = '')
        OR (image_media_asset_id IS NOT NULL AND length(btrim(image_url)) > 0)
    );
