LOCK TABLE ticket_messages IN ACCESS EXCLUSIVE MODE;
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM ticket_messages
        WHERE client_message_id IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Cannot remove customer-service message idempotency keys while persisted messages depend on retry identity';
    END IF;
END $$;

DROP INDEX IF EXISTS idx_ticket_messages_ticket_client_message_id;
ALTER TABLE ticket_messages
    DROP COLUMN client_message_payload_hash,
    DROP COLUMN client_message_id;
