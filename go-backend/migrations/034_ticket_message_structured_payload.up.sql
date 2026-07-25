ALTER TABLE ticket_messages
  ADD COLUMN IF NOT EXISTS message_type VARCHAR(40) NOT NULL DEFAULT 'text';

ALTER TABLE ticket_messages
  ADD COLUMN IF NOT EXISTS metadata TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_ticket_messages_message_type
  ON ticket_messages(message_type);
