-- NULL preserves unknown historical sizes. Accounting metadata grants no ownership.
ALTER TABLE media_upload_sessions ADD COLUMN IF NOT EXISTS expected_upload_bytes BIGINT
    CHECK (expected_upload_bytes IS NULL OR expected_upload_bytes >= 0);
