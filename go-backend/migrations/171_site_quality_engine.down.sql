DROP INDEX IF EXISTS idx_site_quality_findings_target_strategy_v2;
DROP INDEX IF EXISTS idx_site_quality_runs_target_strategy_created_v2;
DROP INDEX IF EXISTS uq_site_quality_findings_target_strategy_audit;

ALTER TABLE IF EXISTS site_quality_findings
    DROP CONSTRAINT IF EXISTS fk_site_quality_findings_target;
ALTER TABLE IF EXISTS site_quality_runs
    DROP CONSTRAINT IF EXISTS fk_site_quality_runs_job;
ALTER TABLE IF EXISTS site_quality_runs
    DROP CONSTRAINT IF EXISTS fk_site_quality_runs_target;

ALTER TABLE IF EXISTS site_quality_findings
    ADD CONSTRAINT uq_site_quality_findings_url_strategy_audit
        UNIQUE (target_url, strategy, audit_id);

DROP TABLE IF EXISTS site_quality_evaluations;
DROP TABLE IF EXISTS site_quality_provider_slots;
DROP TABLE IF EXISTS site_quality_jobs;
DROP TABLE IF EXISTS site_quality_targets;
