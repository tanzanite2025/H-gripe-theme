DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM shipping_yanwen_published_channels
        WHERE environment = 'fat'
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot remove Yanwen channel environments while active FAT channels exist';
    END IF;
END $$;

DROP INDEX IF EXISTS idx_shipping_yanwen_channel_environment_product_code;

ALTER TABLE shipping_yanwen_published_channels
    DROP COLUMN IF EXISTS environment;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_yanwen_channel_product_code
    ON shipping_yanwen_published_channels (product_code)
    WHERE deleted_at IS NULL;
