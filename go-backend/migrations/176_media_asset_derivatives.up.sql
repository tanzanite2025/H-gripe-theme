CREATE TABLE IF NOT EXISTS media_asset_derivatives (
    id BIGSERIAL PRIMARY KEY,
    media_asset_id BIGINT NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
    preset VARCHAR(40) NOT NULL,
    url VARCHAR(800) NOT NULL,
    storage_key VARCHAR(800) NOT NULL,
    mime_type VARCHAR(120) NOT NULL DEFAULT '',
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    size BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_media_asset_derivatives_asset_preset
    ON media_asset_derivatives(media_asset_id, preset)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_media_asset_derivatives_url
    ON media_asset_derivatives(url)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_media_asset_derivatives_asset_id
    ON media_asset_derivatives(media_asset_id);

CREATE INDEX IF NOT EXISTS idx_media_asset_derivatives_storage_key
    ON media_asset_derivatives(storage_key);

CREATE INDEX IF NOT EXISTS idx_media_asset_derivatives_deleted_at
    ON media_asset_derivatives(deleted_at);

ALTER TABLE product_media
    ADD COLUMN IF NOT EXISTS image_variants JSONB NOT NULL DEFAULT '{}'::jsonb;
