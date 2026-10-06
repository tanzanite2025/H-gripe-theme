ALTER TABLE shipping_carrier_services
    ADD COLUMN IF NOT EXISTS fpx_channel_id BIGINT,
    ADD COLUMN IF NOT EXISTS yanwen_published_channel_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_shipping_carrier_services_fpx_channel_id
    ON shipping_carrier_services (fpx_channel_id);

CREATE INDEX IF NOT EXISTS idx_shipping_carrier_services_yanwen_channel_id
    ON shipping_carrier_services (yanwen_published_channel_id);

-- Existing 4PX and Yanwen routes were previously identified by carrier code
-- plus service code. Backfill only an unambiguous production collection row;
-- unmatched rows remain visible for repair but are closed by the ID-based
-- quote/public projections until an operator selects a collection record.
UPDATE shipping_carrier_services AS service
SET fpx_channel_id = channel.id
FROM carriers AS carrier, shipping_fpx_channels AS channel
WHERE service.fpx_channel_id IS NULL
  AND service.yanwen_published_channel_id IS NULL
  AND UPPER(TRIM(carrier.code)) IN ('4PX', 'FPX')
  AND carrier.id = service.carrier_id
  AND channel.environment = 'production'
  AND channel.deleted_at IS NULL
  AND UPPER(TRIM(channel.service_code)) = UPPER(TRIM(service.service_code));

UPDATE shipping_carrier_services AS service
SET yanwen_published_channel_id = channel.id
FROM carriers AS carrier, shipping_yanwen_published_channels AS channel
WHERE service.fpx_channel_id IS NULL
  AND service.yanwen_published_channel_id IS NULL
  AND UPPER(TRIM(carrier.code)) = 'YANWEN'
  AND carrier.id = service.carrier_id
  AND channel.environment = 'production'
  AND channel.deleted_at IS NULL
  AND UPPER(TRIM(channel.product_code)) = UPPER(
        REGEXP_REPLACE(TRIM(service.service_code), '^YANWEN:', '', 'i')
    );

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_shipping_carrier_services_fpx_channel'
          AND conrelid = 'shipping_carrier_services'::regclass
    ) THEN
        ALTER TABLE shipping_carrier_services
            ADD CONSTRAINT fk_shipping_carrier_services_fpx_channel
            FOREIGN KEY (fpx_channel_id)
            REFERENCES shipping_fpx_channels(id)
            ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_shipping_carrier_services_yanwen_channel'
          AND conrelid = 'shipping_carrier_services'::regclass
    ) THEN
        ALTER TABLE shipping_carrier_services
            ADD CONSTRAINT fk_shipping_carrier_services_yanwen_channel
            FOREIGN KEY (yanwen_published_channel_id)
            REFERENCES shipping_yanwen_published_channels(id)
            ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_shipping_carrier_services_single_published_collection'
          AND conrelid = 'shipping_carrier_services'::regclass
    ) THEN
        ALTER TABLE shipping_carrier_services
            ADD CONSTRAINT chk_shipping_carrier_services_single_published_collection
            CHECK (NOT (fpx_channel_id IS NOT NULL AND yanwen_published_channel_id IS NOT NULL));
    END IF;
END $$;
