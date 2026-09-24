CREATE TABLE IF NOT EXISTS email_provider_configs (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    driver VARCHAR(32) NOT NULL DEFAULT 'smtp',
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL DEFAULT 587,
    username VARCHAR(255) NOT NULL DEFAULT '',
    password_encrypted TEXT NOT NULL DEFAULT '',
    from_name VARCHAR(128) NOT NULL,
    from_email VARCHAR(255) NOT NULL,
    reply_to VARCHAR(255) NOT NULL DEFAULT '',
    encryption_type VARCHAR(16) NOT NULL DEFAULT 'starttls',
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    last_tested_at TIMESTAMPTZ,
    last_test_status VARCHAR(32) NOT NULL DEFAULT '',
    last_test_error VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_email_provider_port CHECK (port BETWEEN 1 AND 65535),
    CONSTRAINT chk_email_provider_driver CHECK (driver = 'smtp'),
    CONSTRAINT chk_email_provider_encryption CHECK (encryption_type IN ('none', 'starttls', 'tls'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_email_provider_default
    ON email_provider_configs (is_default)
    WHERE is_default = TRUE;

CREATE INDEX IF NOT EXISTS idx_email_provider_active_default
    ON email_provider_configs (is_active, is_default);
