CREATE TABLE IF NOT EXISTS site_quality_runs (
    id BIGSERIAL PRIMARY KEY,
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
    CONSTRAINT ck_site_quality_runs_strategy
        CHECK (strategy IN ('mobile', 'desktop')),
    CONSTRAINT ck_site_quality_runs_status
        CHECK (status IN ('success', 'failed')),
    CONSTRAINT ck_site_quality_runs_scores
        CHECK (
            (performance_score IS NULL OR performance_score BETWEEN 0 AND 100)
            AND (accessibility_score IS NULL OR accessibility_score BETWEEN 0 AND 100)
            AND (best_practices_score IS NULL OR best_practices_score BETWEEN 0 AND 100)
            AND (seo_score IS NULL OR seo_score BETWEEN 0 AND 100)
        )
);

CREATE INDEX IF NOT EXISTS idx_site_quality_runs_created
    ON site_quality_runs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_site_quality_runs_target_strategy_created
    ON site_quality_runs (target_url, strategy, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_site_quality_runs_status_created
    ON site_quality_runs (status, created_at DESC);
