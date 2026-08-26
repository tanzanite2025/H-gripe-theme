ALTER TABLE site_quality_findings
    ADD COLUMN IF NOT EXISTS rule_id VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS provider_audit_id VARCHAR(128) NOT NULL DEFAULT '';

UPDATE site_quality_findings
SET rule_id = CASE
    WHEN audit_id IN ('link-text', 'descriptive_link_text') THEN 'link_descriptive_text'
    ELSE audit_id
END
WHERE rule_id = ''
   OR rule_id IN ('link-text', 'descriptive_link_text');

UPDATE site_quality_findings
SET provider_audit_id = 'link-text'
WHERE audit_id IN ('link-text', 'descriptive_link_text')
  AND (
      provider_audit_id = ''
      OR provider_audit_id IN ('link_descriptive_text', 'descriptive_link_text')
  );

UPDATE site_quality_findings
SET provider_audit_id = audit_id
WHERE provider_audit_id = ''
  AND audit_id NOT LIKE 'site-heading-%'
  AND audit_id NOT LIKE 'site-schema-%';

UPDATE site_quality_findings
SET latest_evidence = jsonb_set(
    jsonb_set(
        COALESCE(latest_evidence, '{}'::jsonb),
        '{rule_id}',
        to_jsonb(rule_id),
        true
    ),
    '{provider_audit_id}',
    to_jsonb(provider_audit_id),
    true
);

ALTER TABLE preflight_content_link_issues
    ADD COLUMN IF NOT EXISTS rule_id VARCHAR(128) NOT NULL DEFAULT 'link_descriptive_text',
    ADD COLUMN IF NOT EXISTS provider_audit_id VARCHAR(128) NOT NULL DEFAULT 'link-text';

UPDATE preflight_content_link_issues
SET rule_id = 'link_descriptive_text'
WHERE rule_id IN ('', 'link-text', 'descriptive_link_text');

UPDATE preflight_content_link_issues
SET provider_audit_id = 'link-text'
WHERE provider_audit_id IN ('', 'link_descriptive_text', 'descriptive_link_text');

UPDATE preflight_content_link_issues
SET latest_evidence = jsonb_set(
    jsonb_set(
        COALESCE(latest_evidence, '{}'::jsonb) - 'rule',
        '{rule_id}',
        to_jsonb(rule_id),
        true
    ),
    '{provider_audit_id}',
    to_jsonb(provider_audit_id),
    true
);

CREATE INDEX IF NOT EXISTS idx_site_quality_findings_rule_id
    ON site_quality_findings(rule_id);

CREATE INDEX IF NOT EXISTS idx_preflight_content_link_issues_rule_id
    ON preflight_content_link_issues(rule_id);
