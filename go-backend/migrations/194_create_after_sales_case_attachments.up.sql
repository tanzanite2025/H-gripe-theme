CREATE TABLE IF NOT EXISTS after_sales_case_attachments (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT NOT NULL REFERENCES after_sales_cases(id) ON DELETE CASCADE,
    kind VARCHAR(16) NOT NULL,
    storage_url TEXT NOT NULL,
    filename VARCHAR(255) NOT NULL DEFAULT '',
    content_type VARCHAR(128) NOT NULL DEFAULT '',
    size_bytes BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_after_sales_case_attachments_kind
        CHECK (kind IN ('image', 'video')),
    CONSTRAINT ck_after_sales_case_attachments_size
        CHECK (size_bytes >= 0)
);

CREATE INDEX IF NOT EXISTS idx_after_sales_case_attachments_case_id
    ON after_sales_case_attachments(case_id, id);
