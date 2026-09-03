-- Preserve tracking history and make repeated provider pushes idempotent.
WITH ranked_tracking_events AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY order_id, tracking_number, event_time, status
            ORDER BY created_at DESC NULLS LAST, id DESC
        ) AS duplicate_rank
    FROM tracking_events
)
DELETE FROM tracking_events
WHERE id IN (
    SELECT id
    FROM ranked_tracking_events
    WHERE duplicate_rank > 1
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tracking_events_identity
    ON tracking_events (order_id, tracking_number, event_time, status);
