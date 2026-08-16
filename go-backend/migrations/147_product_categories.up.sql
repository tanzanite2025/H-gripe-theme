CREATE TABLE IF NOT EXISTS product_categories (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT NULL,
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    depth INTEGER NOT NULL DEFAULT 1,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_product_categories_parent
        FOREIGN KEY (parent_id)
        REFERENCES product_categories(id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_product_categories_depth
        CHECK (depth >= 1 AND depth <= 5),
    CONSTRAINT chk_product_categories_not_self_parent
        CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_categories_slug ON product_categories(slug);
CREATE INDEX IF NOT EXISTS idx_product_categories_parent_id ON product_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_product_categories_depth ON product_categories(depth);
CREATE INDEX IF NOT EXISTS idx_product_categories_enabled ON product_categories(is_enabled);
CREATE INDEX IF NOT EXISTS idx_product_categories_sort ON product_categories(parent_id, sort_order, name, id);
