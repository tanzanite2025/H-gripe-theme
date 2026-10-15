-- Ownership metadata is independent from URLs, object keys and business references.
-- Unknown historical files remain unclassified; no inference from current uses.
ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS library_partition VARCHAR(32) NOT NULL DEFAULT 'unclassified';
ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS upload_origin VARCHAR(64) NOT NULL DEFAULT 'legacy_unknown';
ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS classification_method VARCHAR(32) NOT NULL DEFAULT 'historical_unknown';
ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS customer_service_media_purpose VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS upload_purpose VARCHAR(64) NOT NULL DEFAULT '';
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='media_assets'::regclass AND conname='media_assets_business_library_classification_valid') THEN
        ALTER TABLE media_assets ADD CONSTRAINT media_assets_business_library_classification_valid CHECK (
            library_partition IN ('catalog','content','shared','customer_service','customer_submissions','unclassified')
            AND classification_method IN ('historical_unknown','recorded_upload_context')
            AND upload_origin IN ('legacy_unknown','administrator_general_upload','administrator_catalog_upload','administrator_content_upload',
                'administrator_website_profile_upload','administrator_shared_upload','administrator_customer_service_reusable_upload',
                'customer_service_conversation_upload','administrator_workbench_feed_upload')
            AND ((library_partition='customer_service' AND customer_service_media_purpose IN ('conversation_attachment','reusable_reply'))
                OR (library_partition<>'customer_service' AND customer_service_media_purpose=''))
        );
    END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_media_assets_library_partition_created ON media_assets(library_partition,created_at DESC,id DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_media_assets_upload_origin ON media_assets(upload_origin) WHERE deleted_at IS NULL;
