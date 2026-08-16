-- Per-staff customer-service read cursors. These replace the global
-- ticket_messages.is_read flag for backoffice inbox/badge semantics.
CREATE TABLE IF NOT EXISTS customer_service_inbox_states (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    recipient_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_message_id BIGINT NOT NULL DEFAULT 0,
    unread_count INTEGER NOT NULL DEFAULT 0,
    assignment_version INTEGER NOT NULL DEFAULT 1,
    last_read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_customer_service_inbox_state_recipient_ticket
        UNIQUE (recipient_user_id, ticket_id),
    CONSTRAINT ck_customer_service_inbox_state_unread_count
        CHECK (unread_count >= 0),
    CONSTRAINT ck_customer_service_inbox_state_assignment_version
        CHECK (assignment_version > 0)
);

CREATE INDEX IF NOT EXISTS idx_customer_service_inbox_states_recipient_unread
    ON customer_service_inbox_states(recipient_user_id, unread_count DESC, ticket_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_customer_service_inbox_states_ticket
    ON customer_service_inbox_states(ticket_id, recipient_user_id)
    WHERE deleted_at IS NULL;

-- Seed states only for assigned support users. Admin/manager states are
-- created lazily when they open/read a conversation.
INSERT INTO customer_service_inbox_states (
    ticket_id,
    recipient_user_id,
    last_read_message_id,
    unread_count,
    assignment_version,
    created_at,
    updated_at
)
SELECT
    tickets.id,
    tickets.assigned_to,
    0,
    COALESCE((
        SELECT COUNT(*)
        FROM ticket_messages AS messages
        WHERE messages.ticket_id = tickets.id
          AND messages.is_staff = FALSE
    ), 0),
    1,
    NOW(),
    NOW()
FROM tickets
JOIN users AS recipients ON recipients.id = tickets.assigned_to
WHERE tickets.category = 'customer_service'
  AND tickets.assigned_to IS NOT NULL
  AND tickets.assigned_to > 0
  AND recipients.status = 'active'
  AND recipients.role IN ('admin', 'manager', 'support')
ON CONFLICT (recipient_user_id, ticket_id) DO NOTHING;
