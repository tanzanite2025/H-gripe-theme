UPDATE quick_buy_steps AS step
SET
    name = CASE step.step_key
        WHEN 'product-search' THEN 'Search products'
        WHEN 'specifications' THEN 'Choose specifications'
        WHEN 'quantity' THEN 'Confirm quantity'
        ELSE step.name
    END,
    description = CASE step.step_key
        WHEN 'product-search' THEN 'Search or filter products'
        WHEN 'specifications' THEN 'Choose product specifications'
        WHEN 'quantity' THEN 'Confirm product information'
        ELSE step.description
    END,
    help_text = CASE step.step_key
        WHEN 'product-search' THEN 'Search or filter products'
        WHEN 'specifications' THEN 'Select product specifications and quantity to continue'
        WHEN 'quantity' THEN 'Confirm product information'
        ELSE step.help_text
    END,
    updated_at = NOW()
FROM quick_buy_flow_versions AS version
JOIN quick_buy_flows AS flow
    ON flow.id = version.flow_id
WHERE step.flow_version_id = version.id
  AND flow.slug = 'quick-build'
  AND version.status = 'draft'
  AND step.step_key IN ('product-search', 'specifications', 'quantity')
  AND step.name IN ('Step 1', 'Step 2', 'Step 3');
