CREATE TABLE IF NOT EXISTS product_specification_template_translations (
    id BIGSERIAL PRIMARY KEY,
    product_specification_template_id BIGINT NOT NULL REFERENCES product_specification_templates(id) ON DELETE CASCADE,
    locale VARCHAR(32) NOT NULL,
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_specification_template_translations_type_locale
    ON product_specification_template_translations(product_specification_template_id, locale);

CREATE INDEX IF NOT EXISTS idx_product_specification_template_translations_locale
    ON product_specification_template_translations(locale);
