-- Product type and specification templates.
-- Backend is the source of truth for product-specific fields used by admin forms and storefront rendering.

CREATE TABLE IF NOT EXISTS product_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(120) UNIQUE NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_spec_definitions (
    id BIGSERIAL PRIMARY KEY,
    product_type_id BIGINT NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    "group" VARCHAR(80) NOT NULL DEFAULT 'specs',
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    field_type VARCHAR(32) NOT NULL DEFAULT 'text',
    unit VARCHAR(32),
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    is_filterable BOOLEAN NOT NULL DEFAULT FALSE,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    options TEXT,
    validation TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_product_type_spec_slug UNIQUE (product_type_id, slug)
);

ALTER TABLE products ADD COLUMN IF NOT EXISTS product_type_id BIGINT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_products_product_type'
    ) THEN
        ALTER TABLE products
            ADD CONSTRAINT fk_products_product_type
            FOREIGN KEY (product_type_id) REFERENCES product_types(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_products_product_type_id ON products(product_type_id);

CREATE TABLE IF NOT EXISTS product_spec_values (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    spec_definition_id BIGINT NOT NULL REFERENCES product_spec_definitions(id) ON DELETE CASCADE,
    value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_product_spec_value UNIQUE (product_id, spec_definition_id)
);

CREATE INDEX IF NOT EXISTS idx_product_spec_values_product_id ON product_spec_values(product_id);
CREATE INDEX IF NOT EXISTS idx_product_spec_values_definition_id ON product_spec_values(spec_definition_id);
CREATE INDEX IF NOT EXISTS idx_product_spec_values_value ON product_spec_values(value);
