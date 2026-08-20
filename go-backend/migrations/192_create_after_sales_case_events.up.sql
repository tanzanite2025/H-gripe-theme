CREATE TABLE IF NOT EXISTS after_sales_case_events (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT NOT NULL REFERENCES after_sales_cases(id) ON DELETE CASCADE,
    from_status VARCHAR(32) NOT NULL DEFAULT '',
    to_status VARCHAR(32) NOT NULL,
    resolution TEXT NOT NULL DEFAULT '',
    updated_by BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_after_sales_case_events_case_id_created_at
    ON after_sales_case_events(case_id, created_at, id);
