ALTER TABLE media_assets
    ADD COLUMN IF NOT EXISTS content_sha256 CHAR(64);

ALTER TABLE media_assets
    ADD COLUMN IF NOT EXISTS copyright_claim_json TEXT;

CREATE INDEX IF NOT EXISTS idx_media_assets_content_sha256
    ON media_assets(content_sha256);
