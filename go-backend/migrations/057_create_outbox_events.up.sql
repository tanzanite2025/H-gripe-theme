CREATE TABLE IF NOT EXISTS outbox_events (
    id BIGSERIAL PRIMARY KEY,
    event_key VARCHAR(160) NOT NULL,
    event_type VARCHAR(80) NOT NULL,
    aggregate_type VARCHAR(80) NOT NULL,
    aggregate_id VARCHAR(80) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 10,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ NULL,
    locked_by VARCHAR(128) NOT NULL DEFAULT '',
    processed_at TIMESTAMPTZ NULL,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_outbox_events_event_key
    ON outbox_events(event_key);

CREATE INDEX IF NOT EXISTS idx_outbox_events_ready
    ON outbox_events(status, available_at, attempts, id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_outbox_events_processing_lock
    ON outbox_events(status, locked_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_outbox_events_aggregate
    ON outbox_events(aggregate_type, aggregate_id)
    WHERE deleted_at IS NULL;
