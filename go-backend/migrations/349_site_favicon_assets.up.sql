CREATE TABLE IF NOT EXISTS site_favicon_assets (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    filename TEXT NOT NULL,
    original_filename TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    mime_type VARCHAR(120) NOT NULL DEFAULT 'image/webp',
    size BIGINT NOT NULL DEFAULT 0,
    content_sha256 CHAR(64) NOT NULL DEFAULT '',
    width INTEGER NOT NULL DEFAULT 512 CHECK (width = 512),
    height INTEGER NOT NULL DEFAULT 512 CHECK (height = 512),
    uploader_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_site_favicon_assets_storage_key
    ON site_favicon_assets (storage_key);

-- The legacy URL is intentionally not copied into the new table. A missing
-- object must never become a managed favicon just because its setting survived.
UPDATE settings
SET value = '', updated_at = NOW()
WHERE key = 'site_favicon'
  AND value LIKE '%/uploads/%';
