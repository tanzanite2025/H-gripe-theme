ALTER TABLE products
  ADD COLUMN generic_product_schema JSONB;

ALTER TABLE products
  ADD CONSTRAINT products_generic_product_schema_check CHECK (
    (
      product_type = ''
      AND (
        generic_product_schema IS NULL
        OR (
          jsonb_typeof(generic_product_schema) = 'object'
          AND (generic_product_schema->>'schema_version') IS NOT DISTINCT FROM '1'
        )
      )
    )
    OR (
      product_type <> ''
      AND generic_product_schema IS NULL
    )
  );

COMMENT ON COLUMN products.generic_product_schema IS
  'Optional versioned product-level Schema.org facts for ordinary products; dedicated product types use their own specification columns.';
