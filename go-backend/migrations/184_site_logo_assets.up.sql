CREATE TABLE IF NOT EXISTS site_logo_assets (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    filename TEXT NOT NULL,
    original_filename TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    mime_type VARCHAR(120) NOT NULL DEFAULT 'image/svg+xml',
    size BIGINT NOT NULL DEFAULT 0,
    content_sha256 CHAR(64) NOT NULL DEFAULT '',
    width INTEGER NOT NULL DEFAULT 48 CHECK (width = 48),
    height INTEGER NOT NULL DEFAULT 48 CHECK (height = 48),
    uploader_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_site_logo_assets_storage_key
    ON site_logo_assets (storage_key);
