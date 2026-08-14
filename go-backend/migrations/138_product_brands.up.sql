CREATE TABLE IF NOT EXISTS product_brands (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(160) NOT NULL,
    slug VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    website_url TEXT NOT NULL DEFAULT '',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_brands_slug
    ON product_brands (slug);
CREATE INDEX IF NOT EXISTS idx_product_brands_enabled_order
    ON product_brands (is_enabled, sort_order, id);

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS brand_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_products_brand_id
    ON products (brand_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_products_brand'
    ) THEN
        ALTER TABLE products
            ADD CONSTRAINT fk_products_brand
            FOREIGN KEY (brand_id) REFERENCES product_brands (id)
            ON DELETE RESTRICT;
    END IF;
END $$;
