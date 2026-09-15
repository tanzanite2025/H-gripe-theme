CREATE TABLE IF NOT EXISTS product_display_price_snapshots (
    id BIGSERIAL PRIMARY KEY,
    scope_key VARCHAR(64) NOT NULL UNIQUE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_id BIGINT REFERENCES product_variants(id) ON DELETE CASCADE,
    source_currency VARCHAR(3) NOT NULL,
    source_price_minor BIGINT NOT NULL,
    source_sale_price_minor BIGINT,
    display_prices JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_product_display_price_snapshot_source_price_non_negative CHECK (source_price_minor >= 0),
    CONSTRAINT chk_product_display_price_snapshot_source_sale_price_non_negative CHECK (source_sale_price_minor IS NULL OR source_sale_price_minor >= 0),
    CONSTRAINT chk_product_display_price_snapshot_scope CHECK (
        (variant_id IS NULL AND scope_key = 'product:' || product_id::text)
        OR (variant_id IS NOT NULL AND scope_key = 'variant:' || variant_id::text)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_display_price_snapshots_product_scope
    ON product_display_price_snapshots(product_id)
    WHERE variant_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_display_price_snapshots_variant_scope
    ON product_display_price_snapshots(variant_id)
    WHERE variant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_product_display_price_snapshots_product_id
    ON product_display_price_snapshots(product_id);
