ALTER TABLE shipping_fpx_channels
    ADD COLUMN IF NOT EXISTS environment VARCHAR(16) NOT NULL DEFAULT 'production';

DROP INDEX IF EXISTS idx_shipping_fpx_channel_service_code;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_fpx_channel_environment_service_code
    ON shipping_fpx_channels (environment, service_code)
    WHERE deleted_at IS NULL;
