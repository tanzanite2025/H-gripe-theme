UPDATE quick_buy_flow_versions AS version
SET
    status = 'draft',
    published_at = NULL
FROM quick_buy_flows AS flow
WHERE version.flow_id = flow.id
  AND flow.slug = 'quick-build'
  AND version.status = 'published'
  AND NOT EXISTS (
      SELECT 1
      FROM quick_buy_flow_versions AS other
      WHERE other.flow_id = version.flow_id
        AND other.id <> version.id
        AND other.status = 'published'
  );
