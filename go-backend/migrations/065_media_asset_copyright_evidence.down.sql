DROP INDEX IF EXISTS idx_media_assets_content_sha256;

ALTER TABLE media_assets
    DROP COLUMN IF EXISTS copyright_claim_json;

ALTER TABLE media_assets
    DROP COLUMN IF EXISTS content_sha256;
