ALTER TABLE preflight_content_link_runs
    ADD COLUMN IF NOT EXISTS rule_id VARCHAR(128) NOT NULL DEFAULT 'link_descriptive_text',
    ADD COLUMN IF NOT EXISTS provider_audit_id VARCHAR(128) NOT NULL DEFAULT 'link-text';

UPDATE preflight_content_link_runs
SET rule_id = 'link_descriptive_text'
WHERE rule_id IN ('', 'link-text', 'descriptive_link_text');

UPDATE preflight_content_link_runs
SET provider_audit_id = 'link-text'
WHERE provider_audit_id IN ('', 'link_descriptive_text', 'descriptive_link_text');

CREATE INDEX IF NOT EXISTS idx_preflight_content_link_runs_rule_id
    ON preflight_content_link_runs(rule_id, checked_at DESC);
