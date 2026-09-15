ALTER TABLE ticket_messages
    ALTER COLUMN user_id SET NOT NULL;

DROP INDEX IF EXISTS idx_ticket_messages_ticket_created_id;
