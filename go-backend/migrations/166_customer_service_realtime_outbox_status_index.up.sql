-- The customer-service Relay refreshes aggregate Outbox status metrics on a
-- short interval. Keep that operational query scoped to its own event type
-- instead of scanning historical Outbox rows from unrelated consumers.
CREATE INDEX IF NOT EXISTS idx_outbox_events_customer_service_realtime_status
    ON outbox_events(status)
    WHERE event_type = 'customer_service.realtime'
      AND deleted_at IS NULL;
