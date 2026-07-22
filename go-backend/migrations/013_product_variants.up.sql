-- Product variants become the source of truth for purchasable SKU, price, stock and shipping weight.
-- Products remain the catalog shell; products.price/stock are summary fields for old read paths.

ALTER TABLE product_spec_definitions
    ADD COLUMN IF NOT EXISTS is_variant_option BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS product_variants (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku VARCHAR(120) UNIQUE NOT NULL,
    title VARCHAR(160),
    option_values TEXT NOT NULL DEFAULT '{}',
    price DECIMAL(10, 2) NOT NULL,
    sale_price DECIMAL(10, 2),
    stock INT NOT NULL DEFAULT 0,
    weight_grams INT DEFAULT 0,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX IF NOT EXISTS idx_product_variants_deleted_at ON product_variants(deleted_at);
CREATE INDEX IF NOT EXISTS idx_product_variants_active ON product_variants(product_id, is_active);
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variant_options
    ON product_variants(product_id, option_values)
    WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variants_one_default
    ON product_variants(product_id)
    WHERE is_default = TRUE AND deleted_at IS NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_product_variants_price_positive'
    ) THEN
        ALTER TABLE product_variants
            ADD CONSTRAINT chk_product_variants_price_positive CHECK (price > 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_product_variants_sale_price_non_negative'
    ) THEN
        ALTER TABLE product_variants
            ADD CONSTRAINT chk_product_variants_sale_price_non_negative CHECK (sale_price IS NULL OR sale_price >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_product_variants_stock_non_negative'
    ) THEN
        ALTER TABLE product_variants
            ADD CONSTRAINT chk_product_variants_stock_non_negative CHECK (stock >= 0);
    END IF;
END $$;

ALTER TABLE cart_items ADD COLUMN IF NOT EXISTS variant_id BIGINT;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS variant_id BIGINT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_cart_items_variant'
    ) THEN
        ALTER TABLE cart_items
            ADD CONSTRAINT fk_cart_items_variant
            FOREIGN KEY (variant_id) REFERENCES product_variants(id);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_order_items_variant'
    ) THEN
        ALTER TABLE order_items
            ADD CONSTRAINT fk_order_items_variant
            FOREIGN KEY (variant_id) REFERENCES product_variants(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_cart_items_variant_id ON cart_items(variant_id);
CREATE INDEX IF NOT EXISTS idx_order_items_variant_id ON order_items(variant_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_product_variant ON cart_items(cart_id, product_id, variant_id);

INSERT INTO product_variants (
    product_id, sku, title, option_values, price, sale_price, stock, weight_grams,
    is_default, is_active, sort_order, created_at, updated_at
)
SELECT
    p.id,
    p.sku,
    'Default',
    '{}',
    p.price,
    p.sale_price,
    COALESCE(p.stock, 0),
    0,
    TRUE,
    TRUE,
    0,
    NOW(),
    NOW()
FROM products p
WHERE NOT EXISTS (
    SELECT 1 FROM product_variants pv WHERE pv.product_id = p.id
)
ON CONFLICT (sku) DO NOTHING;

UPDATE cart_items ci
SET variant_id = pv.id
FROM product_variants pv
WHERE ci.variant_id IS NULL
  AND ci.product_id = pv.product_id
  AND pv.is_default = TRUE
  AND pv.deleted_at IS NULL;

UPDATE order_items oi
SET variant_id = pv.id
FROM product_variants pv
WHERE oi.variant_id IS NULL
  AND oi.product_id = pv.product_id
  AND pv.is_default = TRUE
  AND pv.deleted_at IS NULL;

ALTER TABLE cart_items ALTER COLUMN variant_id SET NOT NULL;
ALTER TABLE order_items ALTER COLUMN variant_id SET NOT NULL;
