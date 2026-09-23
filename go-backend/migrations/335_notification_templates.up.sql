CREATE TABLE IF NOT EXISTS email_templates (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    locale VARCHAR(16) NOT NULL DEFAULT 'en',
    category VARCHAR(32) NOT NULL DEFAULT 'order',
    name VARCHAR(128) NOT NULL,
    subject_template VARCHAR(255) NOT NULL,
    body_html TEXT NOT NULL DEFAULT '',
    body_text TEXT NOT NULL DEFAULT '',
    allowed_variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    required_variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_email_templates_code_locale UNIQUE (code, locale),
    CONSTRAINT chk_email_templates_version_positive CHECK (version > 0),
    CONSTRAINT chk_email_templates_body_present CHECK (length(body_html) > 0 OR length(body_text) > 0)
);

CREATE INDEX IF NOT EXISTS idx_email_templates_code_enabled
    ON email_templates(code, is_enabled);

CREATE TABLE IF NOT EXISTS email_template_versions (
    id BIGSERIAL PRIMARY KEY,
    template_id BIGINT NOT NULL REFERENCES email_templates(id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL,
    locale VARCHAR(16) NOT NULL,
    version INTEGER NOT NULL,
    name VARCHAR(128) NOT NULL,
    subject_template VARCHAR(255) NOT NULL,
    body_html TEXT NOT NULL DEFAULT '',
    body_text TEXT NOT NULL DEFAULT '',
    allowed_variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    required_variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    changed_by_user_id BIGINT,
    change_reason VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_email_template_versions_template_version UNIQUE (template_id, version),
    CONSTRAINT chk_email_template_versions_version_positive CHECK (version > 0),
    CONSTRAINT chk_email_template_versions_body_present CHECK (length(body_html) > 0 OR length(body_text) > 0)
);

CREATE INDEX IF NOT EXISTS idx_email_template_versions_lookup
    ON email_template_versions(template_id, version DESC);
