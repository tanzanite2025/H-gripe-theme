CREATE TABLE IF NOT EXISTS visual_showcase_items (
    id BIGSERIAL PRIMARY KEY,
    showcase_key VARCHAR(120) NOT NULL,
    locale VARCHAR(32) NOT NULL,
    image_url TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL DEFAULT '',
    storage_key TEXT NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL,
    caption TEXT NOT NULL DEFAULT '',
    alt_text VARCHAR(500) NOT NULL,
    desktop_order INTEGER NOT NULL DEFAULT 0,
    mobile_pair_index INTEGER NOT NULL DEFAULT 0,
    target_url TEXT NOT NULL DEFAULT '',
    target_label VARCHAR(255) NOT NULL DEFAULT '',
    layout_variant VARCHAR(32) NOT NULL DEFAULT 'standard',
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    published_from TIMESTAMPTZ NULL,
    published_until TIMESTAMPTZ NULL,
    width INTEGER NOT NULL DEFAULT 900,
    height INTEGER NOT NULL DEFAULT 1200,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_visual_showcase_items_scope
    ON visual_showcase_items (showcase_key, locale, desktop_order);

CREATE INDEX IF NOT EXISTS idx_visual_showcase_items_published
    ON visual_showcase_items (showcase_key, locale, is_published);

CREATE INDEX IF NOT EXISTS idx_visual_showcase_items_storage_key
    ON visual_showcase_items (storage_key);
