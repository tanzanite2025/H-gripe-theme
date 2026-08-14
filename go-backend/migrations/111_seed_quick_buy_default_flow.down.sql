DELETE FROM quick_buy_flows AS flow
WHERE flow.slug = 'quick-build'
  AND NOT EXISTS (
      SELECT 1
      FROM quick_buy_flow_versions AS version
      WHERE version.flow_id = flow.id
        AND version.status <> 'draft'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM quick_buy_flow_versions AS version
      JOIN quick_buy_steps AS step
          ON step.flow_version_id = version.id
      WHERE version.flow_id = flow.id
        AND (
            step.step_key NOT IN ('product-search', 'specifications', 'quantity')
            OR EXISTS (
                SELECT 1
                FROM quick_buy_step_product_types AS product_type
                WHERE product_type.step_id = step.id
            )
        )
  );
