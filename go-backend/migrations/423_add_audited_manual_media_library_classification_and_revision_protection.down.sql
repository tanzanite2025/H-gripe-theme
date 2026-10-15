DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM media_assets WHERE classification_method='manual_reviewed' OR classification_revision<>0) THEN
        RAISE EXCEPTION 'Cannot discard manually reviewed media ownership or classification revisions';
    END IF;
END $$;
ALTER TABLE media_assets DROP CONSTRAINT IF EXISTS media_assets_manual_classification_revision_valid;
ALTER TABLE media_assets DROP CONSTRAINT IF EXISTS media_assets_business_library_classification_valid;
ALTER TABLE media_assets ADD CONSTRAINT media_assets_business_library_classification_valid CHECK (
    library_partition IN ('catalog','content','shared','customer_service','customer_submissions','unclassified')
    AND classification_method IN ('historical_unknown','recorded_upload_context')
    AND upload_origin IN ('legacy_unknown','administrator_general_upload','administrator_catalog_upload','administrator_content_upload',
        'administrator_website_profile_upload','administrator_shared_upload','administrator_customer_service_reusable_upload',
        'customer_service_conversation_upload','administrator_workbench_feed_upload')
    AND ((library_partition='customer_service' AND customer_service_media_purpose IN ('conversation_attachment','reusable_reply'))
        OR (library_partition<>'customer_service' AND customer_service_media_purpose=''))
);
ALTER TABLE media_assets DROP COLUMN IF EXISTS classification_revision;
