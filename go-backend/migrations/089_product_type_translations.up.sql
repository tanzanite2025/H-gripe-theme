CREATE TABLE IF NOT EXISTS product_type_translations (
    id BIGSERIAL PRIMARY KEY,
    product_type_id BIGINT NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    locale VARCHAR(32) NOT NULL,
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_type_translations_type_locale
    ON product_type_translations(product_type_id, locale);

CREATE INDEX IF NOT EXISTS idx_product_type_translations_locale
    ON product_type_translations(locale);
