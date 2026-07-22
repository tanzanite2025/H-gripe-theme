-- Seed the first verified production catalog item.
-- Prices come from the independent-site pricing workbook. Inventory was not
-- provided, so the product remains visible while every variant stays at zero stock.

INSERT INTO product_types (name, slug, sort_order, is_enabled)
VALUES ('Carbon Rim', 'carbon_rim', 10, TRUE)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO product_spec_definitions (
    product_type_id, "group", name, slug, field_type, unit,
    is_required, is_filterable, is_visible, is_variant_option,
    sort_order, options, validation
)
SELECT
    id, 'Options', 'Listed Weight', 'listed_weight', 'select', NULL,
    TRUE, TRUE, TRUE, TRUE,
    10, '["370 g","460 g"]', NULL
FROM product_types
WHERE slug = 'carbon_rim'
ON CONFLICT (product_type_id, slug) DO NOTHING;

INSERT INTO product_spec_definitions (
    product_type_id, "group", name, slug, field_type, unit,
    is_required, is_filterable, is_visible, is_variant_option,
    sort_order, options, validation
)
SELECT
    id, 'Options', 'Pack Size', 'pack_size', 'select', NULL,
    TRUE, TRUE, TRUE, TRUE,
    20, '["1 piece","2 pieces"]', NULL
FROM product_types
WHERE slug = 'carbon_rim'
ON CONFLICT (product_type_id, slug) DO NOTHING;

INSERT INTO products (
    product_type_id, sku, name, slug, price, stock,
    status, locale, featured
)
SELECT
    id, 'G35-370G-1PC', 'G35 Carbon Rim', 'g35-carbon-rim', 111.14, 0,
    'active', 'en', FALSE
FROM product_types
WHERE slug = 'carbon_rim'
  AND NOT EXISTS (
      SELECT 1
      FROM products
      WHERE sku = 'G35-370G-1PC'
         OR (slug = 'g35-carbon-rim' AND locale = 'en')
  )
ON CONFLICT DO NOTHING;

INSERT INTO product_variants (
    product_id, sku, title, option_values, price, sale_price, stock,
    weight_grams, is_default, is_active, sort_order
)
-- The 370 g and 460 g values are verified option labels, not confirmed
-- shipment weights, so weight_grams intentionally remains zero.
SELECT
    p.id,
    seed.sku,
    seed.title,
    seed.option_values,
    seed.price,
    NULL,
    0,
    0,
    seed.is_default,
    TRUE,
    seed.sort_order
FROM products p
CROSS JOIN (
    VALUES
        ('G35-370G-1PC', '370 g / 1 piece',  '{"listed_weight":"370 g","pack_size":"1 piece"}',  111.14::DECIMAL(10, 2), TRUE,  10),
        ('G35-370G-2PC', '370 g / 2 pieces', '{"listed_weight":"370 g","pack_size":"2 pieces"}', 222.28::DECIMAL(10, 2), FALSE, 20),
        ('G35-460G-1PC', '460 g / 1 piece',  '{"listed_weight":"460 g","pack_size":"1 piece"}',   78.73::DECIMAL(10, 2), FALSE, 30),
        ('G35-460G-2PC', '460 g / 2 pieces', '{"listed_weight":"460 g","pack_size":"2 pieces"}', 157.45::DECIMAL(10, 2), FALSE, 40)
) AS seed(sku, title, option_values, price, is_default, sort_order)
WHERE p.sku = 'G35-370G-1PC'
  AND p.slug = 'g35-carbon-rim'
  AND p.locale = 'en'
  AND p.product_type_id = (
      SELECT id FROM product_types WHERE slug = 'carbon_rim'
  )
  AND p.deleted_at IS NULL
ON CONFLICT DO NOTHING;
