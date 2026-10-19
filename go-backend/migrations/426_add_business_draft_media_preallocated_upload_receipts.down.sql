DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM business_draft_media_upload_receipts) THEN
        RAISE EXCEPTION 'Cannot discard business draft media upload receipts; finish or retire their workflows first';
    END IF;
END $$;
DROP TABLE IF EXISTS business_draft_media_upload_receipts;
