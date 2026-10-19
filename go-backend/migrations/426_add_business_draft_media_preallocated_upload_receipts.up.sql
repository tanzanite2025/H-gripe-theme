-- Only new, explicitly selected editor uploads receive a receipt. Historical
-- objects and uploads made directly in the media library are never enrolled.
CREATE TABLE IF NOT EXISTS business_draft_media_upload_receipts (
    id UUID PRIMARY KEY,
    asset_id BIGINT UNIQUE REFERENCES media_assets(id) ON DELETE CASCADE,
    uploader_id BIGINT NOT NULL CHECK (uploader_id > 0),
    upload_origin TEXT NOT NULL CHECK (upload_origin IN (
        'administrator_catalog_upload', 'administrator_content_upload',
        'administrator_website_profile_upload', 'administrator_workbench_feed_upload')),
    original_object_key TEXT NOT NULL UNIQUE,
    object_keys_json JSONB NOT NULL CHECK (jsonb_typeof(object_keys_json) = 'array' AND jsonb_array_length(object_keys_json) BETWEEN 1 AND 13),
    content_sha256 TEXT NOT NULL CONSTRAINT business_draft_media_receipt_content_digest_valid CHECK (content_sha256 ~ '^[a-f0-9]{64}$'),
    size BIGINT NOT NULL CHECK (size > 0),
    mime_type TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('uploading', 'ready', 'bound')),
    expires_at TIMESTAMPTZ NOT NULL,
    next_expiry_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT business_draft_media_receipt_status_asset_identity_valid CHECK ((status = 'uploading' AND asset_id IS NULL) OR (status IN ('ready', 'bound') AND asset_id IS NOT NULL)),
    CONSTRAINT business_draft_media_receipt_original_namespace_valid CHECK (original_object_key LIKE 'business-draft-media/' || id::text || '/%'),
    CHECK (object_keys_json->>0 = original_object_key)
);
CREATE INDEX IF NOT EXISTS idx_business_draft_media_receipt_expiry ON business_draft_media_upload_receipts(status, expires_at);
