-- Code rollback can retain these additive columns. Schema rollback must not
-- erase visitor ownership or merge newly distinguished account namespaces.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM media_assets
        WHERE uploader_actor_type <> 'legacy_account'
           OR uploader_visitor_session_hash <> ''
    ) THEN
        RAISE EXCEPTION 'Cannot remove media upload actor identity while typed uploads exist; retain the additive schema when rolling back application code';
    END IF;
END $$;

DROP INDEX IF EXISTS idx_media_assets_uploader_visitor_session;
DROP INDEX IF EXISTS idx_media_assets_uploader_actor_account;
ALTER TABLE media_assets
    DROP CONSTRAINT IF EXISTS media_assets_upload_actor_identity_valid,
    DROP COLUMN IF EXISTS uploader_visitor_session_hash,
    DROP COLUMN IF EXISTS uploader_actor_type;
