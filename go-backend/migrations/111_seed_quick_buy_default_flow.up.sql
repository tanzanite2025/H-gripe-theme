INSERT INTO quick_buy_flows (
    slug,
    name,
    description,
    entry_surface,
    is_enabled,
    sort_order
)
VALUES (
    'quick-build',
    'QUICK Build',
    'Default QUICK build flow',
    'dock',
    TRUE,
    100
)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO quick_buy_flow_versions (
    flow_id,
    version_number,
    status
)
SELECT
    flow.id,
    COALESCE(MAX(version.version_number), 0) + 1,
    'draft'
FROM quick_buy_flows AS flow
LEFT JOIN quick_buy_flow_versions AS version
    ON version.flow_id = flow.id
WHERE flow.slug = 'quick-build'
  AND NOT EXISTS (
      SELECT 1
      FROM quick_buy_flow_versions AS draft
      WHERE draft.flow_id = flow.id
        AND draft.status = 'draft'
  )
GROUP BY flow.id;

WITH default_version AS (
    SELECT version.id
    FROM quick_buy_flow_versions AS version
    JOIN quick_buy_flows AS flow
        ON flow.id = version.flow_id
    WHERE flow.slug = 'quick-build'
      AND version.status = 'draft'
    ORDER BY version.version_number DESC, version.id DESC
    LIMIT 1
),
default_steps(step_key, name, description, help_text, sort_order) AS (
    VALUES
        ('product-search', 'Step 1', '', '', 10),
        ('specifications', 'Step 2', '', '', 20),
        ('quantity', 'Step 3', '', '', 30)
)
INSERT INTO quick_buy_steps (
    flow_version_id,
    step_key,
    name,
    description,
    help_text,
    sort_order,
    selection_mode,
    is_required,
    min_select,
    max_select,
    default_quantity,
    allow_skip
)
SELECT
    default_version.id,
    default_steps.step_key,
    default_steps.name,
    default_steps.description,
    default_steps.help_text,
    default_steps.sort_order,
    'single',
    TRUE,
    0,
    1,
    1,
    FALSE
FROM default_version
CROSS JOIN default_steps
ON CONFLICT (flow_version_id, step_key) DO NOTHING;
