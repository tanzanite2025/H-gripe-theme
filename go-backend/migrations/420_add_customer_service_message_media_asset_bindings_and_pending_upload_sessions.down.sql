DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM ticket_messages WHERE attachment_reference_format='media_asset_bindings')
        OR EXISTS (SELECT 1 FROM media_upload_sessions)
        OR EXISTS (SELECT 1 FROM ticket_message_media_attachments) THEN
        RAISE EXCEPTION 'Cannot erase customer-service attachment ownership or pending uploads; keep the additive schema when rolling back application code';
    END IF;
END $$;
DROP TABLE IF EXISTS ticket_message_media_attachments;
DROP TABLE IF EXISTS media_upload_sessions;
ALTER TABLE ticket_messages DROP CONSTRAINT IF EXISTS ticket_messages_attachment_reference_format_valid;
ALTER TABLE ticket_messages DROP COLUMN IF EXISTS attachment_reference_format;
