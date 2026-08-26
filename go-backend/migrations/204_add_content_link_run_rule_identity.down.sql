DROP INDEX IF EXISTS idx_preflight_content_link_runs_rule_id;

ALTER TABLE preflight_content_link_runs
    DROP COLUMN IF EXISTS provider_audit_id,
    DROP COLUMN IF EXISTS rule_id;
