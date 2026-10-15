-- Additive read/write upgrade; historical attachment JSON is not rewritten.
ALTER TABLE ticket_messages ADD COLUMN IF NOT EXISTS attachment_reference_format VARCHAR(32) NOT NULL DEFAULT 'legacy_urls';
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='ticket_messages'::regclass AND conname='ticket_messages_attachment_reference_format_valid') THEN
        ALTER TABLE ticket_messages ADD CONSTRAINT ticket_messages_attachment_reference_format_valid
            CHECK (attachment_reference_format IN ('legacy_urls','media_asset_bindings') AND
                (attachment_reference_format='legacy_urls' OR COALESCE(attachments,'') IN ('','[]')));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS media_upload_sessions (
    id VARCHAR(36) PRIMARY KEY,
    business_context_type VARCHAR(40) NOT NULL CHECK (business_context_type='customer_service'),
    business_context_id BIGINT NOT NULL REFERENCES tickets(id) ON DELETE RESTRICT,
    upload_actor_type VARCHAR(32) NOT NULL,
    upload_actor_account_id BIGINT NOT NULL DEFAULT 0,
    upload_actor_visitor_session_hash VARCHAR(64) NOT NULL DEFAULT '',
    object_key TEXT NOT NULL UNIQUE,
    media_asset_id BIGINT UNIQUE REFERENCES media_assets(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL CHECK (status IN ('uploading','ready','bound','failed')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT media_upload_sessions_owner_identity_valid CHECK (
        (upload_actor_type IN ('administrator','customer') AND upload_actor_account_id > 0 AND upload_actor_visitor_session_hash='')
        OR (upload_actor_type='visitor' AND upload_actor_account_id=0 AND upload_actor_visitor_session_hash ~ '^[0-9a-f]{64}$')),
    CONSTRAINT media_upload_sessions_registered_asset_state_valid CHECK (
        (status IN ('ready','bound') AND media_asset_id IS NOT NULL) OR status IN ('uploading','failed')),
    CONSTRAINT media_upload_sessions_customer_service_original_namespace_valid CHECK (
        object_key LIKE 'customer-service/attachments/originals/%')
);
CREATE INDEX IF NOT EXISTS idx_media_upload_session_business_context ON media_upload_sessions(business_context_type,business_context_id);
CREATE INDEX IF NOT EXISTS idx_media_upload_sessions_status_expires_at ON media_upload_sessions(status,expires_at);

CREATE TABLE IF NOT EXISTS ticket_message_media_attachments (
    id BIGSERIAL PRIMARY KEY,
    ticket_message_id BIGINT NOT NULL REFERENCES ticket_messages(id) ON DELETE RESTRICT,
    media_asset_id BIGINT NOT NULL REFERENCES media_assets(id) ON DELETE RESTRICT,
    attachment_index INTEGER NOT NULL CHECK (attachment_index >= 0 AND attachment_index < 4),
    upload_session_id VARCHAR(36) UNIQUE REFERENCES media_upload_sessions(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ticket_message_id,attachment_index),
    UNIQUE(ticket_message_id,media_asset_id)
);
CREATE INDEX IF NOT EXISTS idx_ticket_message_media_attachments_media_asset ON ticket_message_media_attachments(media_asset_id);
