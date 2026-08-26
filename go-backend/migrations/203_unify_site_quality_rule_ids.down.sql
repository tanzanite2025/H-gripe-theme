DROP INDEX IF EXISTS idx_preflight_content_link_issues_rule_id;
DROP INDEX IF EXISTS idx_site_quality_findings_rule_id;

ALTER TABLE preflight_content_link_issues
    DROP COLUMN IF EXISTS provider_audit_id,
    DROP COLUMN IF EXISTS rule_id;

ALTER TABLE site_quality_findings
    DROP COLUMN IF EXISTS provider_audit_id,
    DROP COLUMN IF EXISTS rule_id;
