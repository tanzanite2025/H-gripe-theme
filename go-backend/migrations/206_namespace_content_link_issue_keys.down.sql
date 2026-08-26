UPDATE preflight_content_link_issues
SET issue_key = 'content-link:' ||
    substring(issue_key FROM char_length('content-link:' || rule_id || ':') + 1)
WHERE issue_key LIKE 'content-link:%:%';

DROP INDEX IF EXISTS idx_preflight_content_link_issues_rule_key;

ALTER TABLE preflight_content_link_issues
    ALTER COLUMN issue_key TYPE VARCHAR(128);
