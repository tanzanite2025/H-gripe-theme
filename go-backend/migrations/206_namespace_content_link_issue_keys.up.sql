ALTER TABLE preflight_content_link_issues
    ALTER COLUMN issue_key TYPE VARCHAR(256);

UPDATE preflight_content_link_issues
SET issue_key = 'content-link:' || rule_id || ':' ||
    substring(issue_key FROM char_length('content-link:') + 1)
WHERE issue_key LIKE 'content-link:%'
  AND issue_key NOT LIKE 'content-link:%:%';

CREATE INDEX IF NOT EXISTS idx_preflight_content_link_issues_rule_key
    ON preflight_content_link_issues(rule_id, issue_key);
