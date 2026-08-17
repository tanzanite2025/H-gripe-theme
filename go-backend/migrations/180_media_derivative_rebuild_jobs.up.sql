CREATE TABLE IF NOT EXISTS media_derivative_rebuild_jobs (
    id BIGSERIAL PRIMARY KEY,
    reason VARCHAR(80) NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL,
    cursor_asset_id BIGINT NOT NULL DEFAULT 0,
    scanned_assets INTEGER NOT NULL DEFAULT 0,
    generated_assets INTEGER NOT NULL DEFAULT 0,
    generated_derivatives INTEGER NOT NULL DEFAULT 0,
    failed_assets INTEGER NOT NULL DEFAULT 0,
    updated_product_media_rows BIGINT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    locked_at TIMESTAMPTZ,
    locked_by VARCHAR(128) NOT NULL DEFAULT '',
    lease_generation BIGINT NOT NULL DEFAULT 0,
    lease_expires_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_media_derivative_rebuild_jobs_status
        CHECK (status IN ('pending', 'running', 'succeeded'))
);

CREATE INDEX IF NOT EXISTS idx_media_derivative_rebuild_jobs_status_id
    ON media_derivative_rebuild_jobs(status, id);

CREATE INDEX IF NOT EXISTS idx_media_derivative_rebuild_jobs_lease
    ON media_derivative_rebuild_jobs(status, lease_expires_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_media_derivative_rebuild_jobs_single_pending
    ON media_derivative_rebuild_jobs((status))
    WHERE status = 'pending';
