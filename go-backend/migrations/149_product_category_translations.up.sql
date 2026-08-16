CREATE TABLE IF NOT EXISTS product_category_translations (
    id BIGSERIAL PRIMARY KEY,
    product_category_id BIGINT NOT NULL REFERENCES product_categories(id) ON DELETE CASCADE,
    locale VARCHAR(32) NOT NULL,
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_category_translations_category_locale
    ON product_category_translations(product_category_id, locale);

CREATE INDEX IF NOT EXISTS idx_product_category_translations_locale
    ON product_category_translations(locale);
