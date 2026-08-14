UPDATE quick_buy_steps AS step
SET
    name = CASE step.step_key
        WHEN 'product-search' THEN 'Step 1'
        WHEN 'specifications' THEN 'Step 2'
        WHEN 'quantity' THEN 'Step 3'
        ELSE step.name
    END,
    description = '',
    help_text = '',
    updated_at = NOW()
FROM quick_buy_flow_versions AS version
JOIN quick_buy_flows AS flow
    ON flow.id = version.flow_id
WHERE step.flow_version_id = version.id
  AND flow.slug = 'quick-build'
  AND version.status = 'draft'
  AND step.step_key IN ('product-search', 'specifications', 'quantity')
  AND step.name IN ('Search products', 'Choose specifications', 'Confirm quantity');
