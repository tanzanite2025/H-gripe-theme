ALTER TABLE ticket_messages
    ADD COLUMN IF NOT EXISTS client_message_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS client_message_payload_hash VARCHAR(64) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_ticket_messages_ticket_client_message_id
    ON ticket_messages (ticket_id, client_message_id);
