ALTER TABLE shipping_yanwen_published_channels
    ADD COLUMN IF NOT EXISTS environment VARCHAR(16) NOT NULL DEFAULT 'production';

DROP INDEX IF EXISTS idx_shipping_yanwen_channel_product_code;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_yanwen_channel_environment_product_code
    ON shipping_yanwen_published_channels (environment, product_code)
    WHERE deleted_at IS NULL;
