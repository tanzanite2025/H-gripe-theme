-- Keep historical numeric uploader IDs, URLs, storage keys and objects intact.
-- Their administrator/customer namespace cannot be inferred from the number.
ALTER TABLE media_assets
    ADD COLUMN IF NOT EXISTS uploader_actor_type VARCHAR(32) NOT NULL DEFAULT 'legacy_account',
    ADD COLUMN IF NOT EXISTS uploader_visitor_session_hash VARCHAR(64) NOT NULL DEFAULT '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'media_assets'::regclass
          AND conname = 'media_assets_upload_actor_identity_valid'
    ) THEN
        ALTER TABLE media_assets
            ADD CONSTRAINT media_assets_upload_actor_identity_valid CHECK (
                (uploader_actor_type = 'legacy_account' AND uploader_visitor_session_hash = '')
                OR (uploader_actor_type IN ('administrator', 'customer')
                    AND uploader_id IS NOT NULL AND uploader_id > 0
                    AND uploader_visitor_session_hash = '')
                OR (uploader_actor_type = 'visitor'
                    AND uploader_id IS NOT NULL AND uploader_id = 0
                    AND uploader_visitor_session_hash ~ '^[0-9a-f]{64}$')
            );
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_media_assets_uploader_actor_account
    ON media_assets (uploader_actor_type, uploader_id);
CREATE INDEX IF NOT EXISTS idx_media_assets_uploader_visitor_session
    ON media_assets (uploader_visitor_session_hash);
