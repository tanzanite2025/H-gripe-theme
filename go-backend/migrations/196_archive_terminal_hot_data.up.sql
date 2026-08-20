CREATE TABLE IF NOT EXISTS site_quality_runs_archive (
    id BIGINT PRIMARY KEY,
    target_id BIGINT,
    job_id BIGINT,
    target_url TEXT NOT NULL,
    canonical_url TEXT NOT NULL DEFAULT '',
    final_url TEXT NOT NULL DEFAULT '',
    strategy VARCHAR(16) NOT NULL,
    status VARCHAR(16) NOT NULL,
    initiated_by_user_id BIGINT NOT NULL DEFAULT 0,
    provider VARCHAR(32) NOT NULL DEFAULT 'lighthouse_runner',
    lighthouse_version VARCHAR(64) NOT NULL DEFAULT '',
    environment_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    release_id VARCHAR(128) NOT NULL DEFAULT '',
    performance_score SMALLINT,
    accessibility_score SMALLINT,
    best_practices_score SMALLINT,
    seo_score SMALLINT,
    first_contentful_paint_ms DOUBLE PRECISION,
    largest_contentful_paint_ms DOUBLE PRECISION,
    interaction_to_next_paint_ms DOUBLE PRECISION,
    cumulative_layout_shift DOUBLE PRECISION,
    total_blocking_time_ms DOUBLE PRECISION,
    speed_index_ms DOUBLE PRECISION,
    issues_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    raw_response_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_site_quality_runs_archive_strategy
        CHECK (strategy IN ('mobile', 'desktop')),
    CONSTRAINT ck_site_quality_runs_archive_status
        CHECK (status IN ('success', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_site_quality_runs_archive_created
    ON site_quality_runs_archive (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_site_quality_runs_archive_target_strategy_created
    ON site_quality_runs_archive (target_url, strategy, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_site_quality_runs_archive_status_created
    ON site_quality_runs_archive (status, created_at DESC);

CREATE TABLE IF NOT EXISTS after_sales_case_events_archive (
    id BIGINT PRIMARY KEY,
    case_id BIGINT NOT NULL,
    from_status VARCHAR(32) NOT NULL DEFAULT '',
    to_status VARCHAR(32) NOT NULL,
    resolution TEXT NOT NULL DEFAULT '',
    updated_by BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_after_sales_case_events_archive_case_id_created_at
    ON after_sales_case_events_archive(case_id, created_at, id);

ALTER TABLE after_sales_cases
    ADD COLUMN IF NOT EXISTS closed_at TIMESTAMPTZ;

UPDATE after_sales_cases
SET closed_at = updated_at
WHERE status IN ('completed', 'rejected', 'cancelled')
  AND closed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_after_sales_cases_closed_at
    ON after_sales_cases(closed_at);

ALTER TABLE site_quality_findings
    DROP CONSTRAINT IF EXISTS fk_site_quality_findings_latest_run;

ALTER TABLE site_quality_finding_events
    DROP CONSTRAINT IF EXISTS fk_site_quality_finding_events_run;
