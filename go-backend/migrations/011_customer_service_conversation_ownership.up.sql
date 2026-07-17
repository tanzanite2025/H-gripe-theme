-- 011_customer_service_conversation_ownership.sql
-- Store public customer-service conversation ownership on the ticket row.

ALTER TABLE tickets
    ADD COLUMN IF NOT EXISTS customer_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS conversation_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS visitor_session_hash VARCHAR(64);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tickets_conversation_id
    ON tickets(conversation_id)
    WHERE conversation_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tickets_customer_user_id
    ON tickets(customer_user_id);

CREATE INDEX IF NOT EXISTS idx_tickets_visitor_session_hash
    ON tickets(visitor_session_hash);

CREATE INDEX IF NOT EXISTS idx_tickets_customer_service_owner_user
    ON tickets(category, customer_user_id)
    WHERE category = 'customer_service' AND customer_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tickets_customer_service_owner_visitor
    ON tickets(category, visitor_session_hash)
    WHERE category = 'customer_service' AND visitor_session_hash IS NOT NULL;
