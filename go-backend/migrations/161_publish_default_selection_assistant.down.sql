UPDATE selection_assistant_flow_versions AS version
SET
    status = 'draft',
    published_at = NULL,
    published_by = NULL
FROM selection_assistant_flows AS flow
WHERE version.flow_id = flow.id
  AND flow.slug = 'wheelset-fit-helper'
  AND version.version_number = 1
  AND version.status = 'published';
