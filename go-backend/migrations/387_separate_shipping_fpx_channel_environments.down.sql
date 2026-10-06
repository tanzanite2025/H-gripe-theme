DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM shipping_fpx_channels
        WHERE environment = 'test'
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot remove 4PX channel environments while active test channels exist';
    END IF;
END $$;

DROP INDEX IF EXISTS idx_shipping_fpx_channel_environment_service_code;

ALTER TABLE shipping_fpx_channels
    DROP COLUMN IF EXISTS environment;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_fpx_channel_service_code
    ON shipping_fpx_channels (service_code)
    WHERE deleted_at IS NULL;
