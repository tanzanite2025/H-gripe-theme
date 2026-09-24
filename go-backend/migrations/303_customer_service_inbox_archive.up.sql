-- Per-recipient inbox visibility is independent from the global conversation
-- lifecycle. NULL means visible in the recipient's active inbox.
ALTER TABLE customer_service_inbox_states
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_customer_service_inbox_states_recipient_archived
    ON customer_service_inbox_states(recipient_user_id, archived_at, updated_at DESC, ticket_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_customer_service_inbox_states_ticket_archived
    ON customer_service_inbox_states(ticket_id, archived_at, recipient_user_id)
    WHERE deleted_at IS NULL;
