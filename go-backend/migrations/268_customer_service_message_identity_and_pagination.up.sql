ALTER TABLE ticket_messages
    ALTER COLUMN user_id DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ticket_messages_ticket_created_id
    ON ticket_messages (ticket_id, created_at, id);
