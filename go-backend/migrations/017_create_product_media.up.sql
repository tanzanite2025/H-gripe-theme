-- Product media is the long-term product-owned media line.
-- Galleries remain independent showcase/marketing collections.

CREATE TABLE IF NOT EXISTS media_assets (
    id BIGSERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255),
    url VARCHAR(800) NOT NULL,
    storage_key VARCHAR(500),
    mime_type VARCHAR(120),
    media_type VARCHAR(32) NOT NULL DEFAULT 'image',
    size BIGINT DEFAULT 0,
    width INT DEFAULT 0,
    height INT DEFAULT 0,
    duration_seconds INT DEFAULT 0,
    alt VARCHAR(255),
    caption TEXT,
    uploader_id BIGINT,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    visibility VARCHAR(32) NOT NULL DEFAULT 'public',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_media_assets_url ON media_assets(url);
CREATE INDEX IF NOT EXISTS idx_media_assets_media_type ON media_assets(media_type);
CREATE INDEX IF NOT EXISTS idx_media_assets_uploader_id ON media_assets(uploader_id);
CREATE INDEX IF NOT EXISTS idx_media_assets_deleted_at ON media_assets(deleted_at);

CREATE TABLE IF NOT EXISTS product_media (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    variant_id BIGINT,
    media_asset_id BIGINT,
    media_type VARCHAR(32) NOT NULL DEFAULT 'image',
    role VARCHAR(32) NOT NULL DEFAULT 'gallery',
    url VARCHAR(800) NOT NULL,
    thumbnail_url VARCHAR(800),
    poster_url VARCHAR(800),
    alt VARCHAR(255),
    title VARCHAR(255),
    locale VARCHAR(16),
    sort_order INT NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_product_media_product
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    CONSTRAINT fk_product_media_variant
        FOREIGN KEY (variant_id) REFERENCES product_variants(id) ON DELETE SET NULL,
    CONSTRAINT fk_product_media_asset
        FOREIGN KEY (media_asset_id) REFERENCES media_assets(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_product_media_product_id ON product_media(product_id);
CREATE INDEX IF NOT EXISTS idx_product_media_variant_id ON product_media(variant_id);
CREATE INDEX IF NOT EXISTS idx_product_media_asset_id ON product_media(media_asset_id);
CREATE INDEX IF NOT EXISTS idx_product_media_type_role ON product_media(media_type, role);
CREATE INDEX IF NOT EXISTS idx_product_media_sort ON product_media(product_id, sort_order, id);
CREATE INDEX IF NOT EXISTS idx_product_media_deleted_at ON product_media(deleted_at);

-- Seed the sample catalog image onto the new product-owned media line.
-- This is intentionally not a legacy image-table backfill.
INSERT INTO product_media (
    product_id,
    media_type,
    role,
    url,
    alt,
    sort_order,
    is_primary,
    is_visible
)
SELECT
    p.id,
    'image',
    'primary',
    '/company/aboutus/appearance/tanzanite-carbon-rim-finish1.webp',
    'Carbon rim finish reference',
    0,
    TRUE,
    TRUE
FROM products p
WHERE p.sku = 'G35-370G-1PC'
  AND p.slug = 'g35-carbon-rim'
  AND p.locale = 'en'
  AND p.product_type_id = (
      SELECT id FROM product_types WHERE slug = 'carbon_rim'
  )
  AND p.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM product_media pm
      WHERE pm.product_id = p.id
        AND pm.url = '/company/aboutus/appearance/tanzanite-carbon-rim-finish1.webp'
        AND pm.deleted_at IS NULL
  );
