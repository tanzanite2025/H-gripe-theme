DROP INDEX IF EXISTS idx_customer_service_inbox_states_ticket_archived;
DROP INDEX IF EXISTS idx_customer_service_inbox_states_recipient_archived;
ALTER TABLE customer_service_inbox_states
    DROP COLUMN IF EXISTS archived_at;
