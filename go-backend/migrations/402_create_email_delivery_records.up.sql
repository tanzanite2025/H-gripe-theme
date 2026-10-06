-- Durable metadata for transactional email delivery history. Rendered HTML,
-- plain text, and template variables remain outside this table by design.
CREATE TABLE IF NOT EXISTS email_delivery_records (
    id BIGSERIAL PRIMARY KEY,
    outbox_event_id BIGINT NOT NULL,
    event_key VARCHAR(255) NOT NULL,
    event_type VARCHAR(80) NOT NULL,
    template_code VARCHAR(64) NOT NULL,
    template_locale VARCHAR(16) NOT NULL DEFAULT 'en',
    template_version INTEGER NOT NULL DEFAULT 1,
    recipient_email VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL DEFAULT '',
    reference_type VARCHAR(32) NOT NULL DEFAULT '',
    reference_number VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'sending',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    provider_code VARCHAR(64) NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    first_attempt_at TIMESTAMPTZ NOT NULL,
    last_attempt_at TIMESTAMPTZ NOT NULL,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_email_delivery_records_event_key UNIQUE (event_key),
    CONSTRAINT ck_email_delivery_records_status CHECK (status IN ('sending', 'sent', 'failed', 'unknown'))
);

CREATE INDEX IF NOT EXISTS idx_email_delivery_records_created_at
    ON email_delivery_records(created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_email_delivery_records_status_created_at
    ON email_delivery_records(status, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_email_delivery_records_template_created_at
    ON email_delivery_records(template_code, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_email_delivery_records_event_type_created_at
    ON email_delivery_records(event_type, created_at DESC, id DESC);
