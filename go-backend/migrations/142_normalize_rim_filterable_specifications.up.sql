-- RIM keeps the complete reusable field skeleton, but only customer-facing
-- categorical/size fields participate in storefront and QuickBuy filtering.
UPDATE product_spec_definitions definition
SET is_filterable = CASE definition.slug
        WHEN 'material' THEN TRUE
        WHEN 'brake_type' THEN TRUE
        WHEN 'tire_type' THEN TRUE
        WHEN 'wheel_size' THEN TRUE
        WHEN 'rim_depth' THEN FALSE
        WHEN 'inner_width' THEN FALSE
        WHEN 'outer_width' THEN FALSE
        WHEN 'spoke_holes' THEN FALSE
        WHEN 'erd' THEN FALSE
        ELSE definition.is_filterable
    END,
    updated_at = NOW()
FROM product_specification_templates product_specification_template
WHERE definition.product_specification_template_id = product_specification_template.id
  AND product_specification_template.slug = 'rim'
  AND definition.slug IN (
      'material',
      'brake_type',
      'tire_type',
      'wheel_size',
      'rim_depth',
      'inner_width',
      'outer_width',
      'spoke_holes',
      'erd'
  );
