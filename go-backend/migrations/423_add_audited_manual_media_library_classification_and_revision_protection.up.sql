-- Management reviews never rewrite upload provenance, URLs or references.
ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS classification_revision BIGINT NOT NULL DEFAULT 0;
ALTER TABLE media_assets DROP CONSTRAINT IF EXISTS media_assets_business_library_classification_valid;
ALTER TABLE media_assets ADD CONSTRAINT media_assets_business_library_classification_valid CHECK (
    library_partition IN ('catalog','content','shared','customer_service','customer_submissions','unclassified')
    AND classification_method IN ('historical_unknown','recorded_upload_context','manual_reviewed')
    AND upload_origin IN ('legacy_unknown','administrator_general_upload','administrator_catalog_upload','administrator_content_upload',
        'administrator_website_profile_upload','administrator_shared_upload','administrator_customer_service_reusable_upload',
        'customer_service_conversation_upload','administrator_workbench_feed_upload')
    AND ((library_partition='customer_service' AND customer_service_media_purpose IN ('conversation_attachment','reusable_reply'))
        OR (library_partition<>'customer_service' AND customer_service_media_purpose=''))
);
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='media_assets'::regclass AND conname='media_assets_manual_classification_revision_valid') THEN
        ALTER TABLE media_assets ADD CONSTRAINT media_assets_manual_classification_revision_valid CHECK (
            classification_revision >= 0
            AND ((classification_method='manual_reviewed' AND classification_revision > 0)
                OR (classification_method<>'manual_reviewed' AND classification_revision=0))
        );
    END IF;
END $$;
