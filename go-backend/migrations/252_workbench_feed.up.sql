CREATE TABLE IF NOT EXISTS workbench_feed_entries (
    id BIGSERIAL PRIMARY KEY,
    entry_number VARCHAR(32) NOT NULL,
    published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locale VARCHAR(16) NOT NULL DEFAULT 'en',
    content TEXT NOT NULL,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(16) NOT NULL DEFAULT 'draft',
    created_by BIGINT NOT NULL,
    updated_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT ck_workbench_feed_entries_status CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT ck_workbench_feed_entries_content_length CHECK (char_length(content) BETWEEN 1 AND 300)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_workbench_feed_entries_entry_number
    ON workbench_feed_entries(entry_number) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_workbench_feed_entries_public_timeline
    ON workbench_feed_entries(status, locale, published_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS workbench_feed_media (
    id BIGSERIAL PRIMARY KEY,
    entry_id BIGINT NOT NULL REFERENCES workbench_feed_entries(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    file_size_kb INTEGER NOT NULL DEFAULT 0,
    caption TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT ck_workbench_feed_media_dimensions CHECK (width > 0 AND height > 0),
    CONSTRAINT ck_workbench_feed_media_size CHECK (file_size_kb >= 0)
);

CREATE INDEX IF NOT EXISTS idx_workbench_feed_media_entry_order
    ON workbench_feed_media(entry_id, sort_order, id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS workbench_feed_tagged_products (
    id BIGSERIAL PRIMARY KEY,
    entry_id BIGINT NOT NULL REFERENCES workbench_feed_entries(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    variant_id BIGINT NULL,
    product_slug VARCHAR(255) NOT NULL DEFAULT '',
    display_title VARCHAR(255) NOT NULL,
    price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    direct_action VARCHAR(32) NOT NULL DEFAULT 'detail_drawer',
    available BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT ck_workbench_feed_tagged_products_action CHECK (direct_action = 'detail_drawer'),
    CONSTRAINT ck_workbench_feed_tagged_products_price CHECK (price >= 0)
);

CREATE INDEX IF NOT EXISTS idx_workbench_feed_tagged_products_entry_order
    ON workbench_feed_tagged_products(entry_id, sort_order, id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_workbench_feed_tagged_products_product
    ON workbench_feed_tagged_products(product_id, variant_id) WHERE deleted_at IS NULL;
