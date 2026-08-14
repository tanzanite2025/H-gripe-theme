UPDATE quick_buy_flow_versions AS version
SET
    status = 'published',
    published_at = COALESCE(version.published_at, NOW())
FROM quick_buy_flows AS flow
WHERE version.flow_id = flow.id
  AND flow.slug = 'quick-build'
  AND version.status = 'draft'
  AND NOT EXISTS (
      SELECT 1
      FROM quick_buy_flow_versions AS published
      WHERE published.flow_id = version.flow_id
        AND published.status = 'published'
  )
  AND (
      SELECT COUNT(DISTINCT step.step_key)
      FROM quick_buy_steps AS step
      WHERE step.flow_version_id = version.id
        AND step.step_key IN ('product-search', 'specifications', 'quantity')
  ) = 3;
