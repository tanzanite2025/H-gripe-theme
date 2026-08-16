UPDATE selection_assistant_flow_versions AS version
SET
    status = 'published',
    published_at = COALESCE(version.published_at, NOW())
FROM selection_assistant_flows AS flow
WHERE version.flow_id = flow.id
  AND flow.slug = 'wheelset-fit-helper'
  AND version.version_number = 1
  AND version.status = 'draft'
  AND NOT EXISTS (
      SELECT 1
      FROM selection_assistant_flow_versions AS published
      WHERE published.flow_id = version.flow_id
        AND published.status = 'published'
  );
