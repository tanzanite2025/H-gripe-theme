DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM media_assets WHERE library_partition<>'unclassified' OR upload_origin<>'legacy_unknown'
        OR classification_method<>'historical_unknown' OR customer_service_media_purpose<>'' OR upload_purpose<>'') THEN
        RAISE EXCEPTION 'Cannot discard recorded media library ownership or upload origins';
    END IF;
END $$;
DROP INDEX IF EXISTS idx_media_assets_library_partition_created;
DROP INDEX IF EXISTS idx_media_assets_upload_origin;
ALTER TABLE media_assets DROP CONSTRAINT IF EXISTS media_assets_business_library_classification_valid;
ALTER TABLE media_assets DROP COLUMN IF EXISTS upload_purpose;
ALTER TABLE media_assets DROP COLUMN IF EXISTS customer_service_media_purpose;
ALTER TABLE media_assets DROP COLUMN IF EXISTS classification_method;
ALTER TABLE media_assets DROP COLUMN IF EXISTS upload_origin;
ALTER TABLE media_assets DROP COLUMN IF EXISTS library_partition;
