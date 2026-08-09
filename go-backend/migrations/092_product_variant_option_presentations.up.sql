-- Product templates define how a variant option is rendered.
-- Product-specific option values keep stable keys separate from labels and swatch media.

ALTER TABLE product_spec_definitions
    ADD COLUMN IF NOT EXISTS presentation VARCHAR(32) NOT NULL DEFAULT 'text';

CREATE TABLE IF NOT EXISTS product_variant_option_values (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    spec_definition_id BIGINT NOT NULL REFERENCES product_spec_definitions(id) ON DELETE CASCADE,
    value_key VARCHAR(160) NOT NULL,
    label VARCHAR(160) NOT NULL,
    color_hex VARCHAR(20),
    swatch_media_asset_id BIGINT REFERENCES media_assets(id) ON DELETE SET NULL,
    swatch_url VARCHAR(800),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_product_variant_option_value_key
        UNIQUE (product_id, spec_definition_id, value_key)
);

CREATE INDEX IF NOT EXISTS idx_product_variant_option_values_product
    ON product_variant_option_values(product_id, spec_definition_id, sort_order, id);

CREATE INDEX IF NOT EXISTS idx_product_variant_option_values_asset
    ON product_variant_option_values(swatch_media_asset_id);

ALTER TABLE product_media
    ADD COLUMN IF NOT EXISTS variant_option_value_id BIGINT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_product_media_variant_option_value'
    ) THEN
        ALTER TABLE product_media
            ADD CONSTRAINT fk_product_media_variant_option_value
            FOREIGN KEY (variant_option_value_id)
            REFERENCES product_variant_option_values(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_product_media_variant_option_value_id
    ON product_media(variant_option_value_id);
