CREATE TABLE IF NOT EXISTS product_information_templates (
    id BIGSERIAL PRIMARY KEY,
    kind VARCHAR(32) NOT NULL CHECK (kind IN ('after_sales', 'packaging')),
    name VARCHAR(160) NOT NULL,
    slug VARCHAR(160) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    locale VARCHAR(32) NOT NULL DEFAULT 'en',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_information_templates_kind_slug_locale
    ON product_information_templates(kind, slug, locale)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_product_information_templates_kind_enabled
    ON product_information_templates(kind, is_enabled, sort_order);

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS after_sales_template_id BIGINT NULL
        REFERENCES product_information_templates(id) ON DELETE SET NULL;

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS packaging_template_id BIGINT NULL
        REFERENCES product_information_templates(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_products_after_sales_template_id
    ON products(after_sales_template_id);

CREATE INDEX IF NOT EXISTS idx_products_packaging_template_id
    ON products(packaging_template_id);

INSERT INTO product_information_templates (kind, name, slug, content, locale, is_enabled, sort_order)
SELECT 'after_sales', 'Standard After-sales', 'standard-after-sales',
       '<h3>After-sales service</h3><p>Please contact our support team with your order number if you need help after purchase.</p>',
       'en', TRUE, 10
WHERE NOT EXISTS (
    SELECT 1 FROM product_information_templates
    WHERE kind = 'after_sales' AND slug = 'standard-after-sales' AND locale = 'en' AND deleted_at IS NULL
);

INSERT INTO product_information_templates (kind, name, slug, content, locale, is_enabled, sort_order)
SELECT 'packaging', 'Standard Packaging', 'standard-packaging',
       '<h3>Packaging</h3><p>Your product is packed with protective materials before dispatch.</p>',
       'en', TRUE, 10
WHERE NOT EXISTS (
    SELECT 1 FROM product_information_templates
    WHERE kind = 'packaging' AND slug = 'standard-packaging' AND locale = 'en' AND deleted_at IS NULL
);
