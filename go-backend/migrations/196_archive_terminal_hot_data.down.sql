INSERT INTO site_quality_runs (
    id,
    target_id,
    job_id,
    target_url,
    canonical_url,
    final_url,
    strategy,
    status,
    initiated_by_user_id,
    provider,
    lighthouse_version,
    environment_json,
    release_id,
    performance_score,
    accessibility_score,
    best_practices_score,
    seo_score,
    first_contentful_paint_ms,
    largest_contentful_paint_ms,
    interaction_to_next_paint_ms,
    cumulative_layout_shift,
    total_blocking_time_ms,
    speed_index_ms,
    issues_json,
    raw_response_json,
    error_message,
    created_at,
    updated_at
)
SELECT
    id,
    target_id,
    job_id,
    target_url,
    canonical_url,
    final_url,
    strategy,
    status,
    initiated_by_user_id,
    provider,
    lighthouse_version,
    environment_json,
    release_id,
    performance_score,
    accessibility_score,
    best_practices_score,
    seo_score,
    first_contentful_paint_ms,
    largest_contentful_paint_ms,
    interaction_to_next_paint_ms,
    cumulative_layout_shift,
    total_blocking_time_ms,
    speed_index_ms,
    issues_json,
    raw_response_json,
    error_message,
    created_at,
    updated_at
FROM site_quality_runs_archive
ON CONFLICT (id) DO NOTHING;

INSERT INTO after_sales_case_events (
    id,
    case_id,
    from_status,
    to_status,
    resolution,
    updated_by,
    created_at
)
SELECT
    id,
    case_id,
    from_status,
    to_status,
    resolution,
    updated_by,
    created_at
FROM after_sales_case_events_archive
ON CONFLICT (id) DO NOTHING;

ALTER TABLE site_quality_findings
    ADD CONSTRAINT fk_site_quality_findings_latest_run
    FOREIGN KEY (latest_run_id)
    REFERENCES site_quality_runs(id)
    ON DELETE RESTRICT;

ALTER TABLE site_quality_finding_events
    ADD CONSTRAINT fk_site_quality_finding_events_run
    FOREIGN KEY (run_id)
    REFERENCES site_quality_runs(id)
    ON DELETE SET NULL;

DROP TABLE IF EXISTS after_sales_case_events_archive;
DROP TABLE IF EXISTS site_quality_runs_archive;

DROP INDEX IF EXISTS idx_after_sales_cases_closed_at;
ALTER TABLE after_sales_cases
    DROP COLUMN IF EXISTS closed_at;
