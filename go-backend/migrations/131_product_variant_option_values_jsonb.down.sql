DROP INDEX IF EXISTS idx_product_variants_option_values_gin;
DROP INDEX IF EXISTS idx_product_variant_options;

ALTER TABLE product_variants
    DROP CONSTRAINT IF EXISTS chk_product_variants_option_values_object;

ALTER TABLE product_variants
    ALTER COLUMN option_values DROP DEFAULT,
    ALTER COLUMN option_values TYPE TEXT USING option_values::text,
    ALTER COLUMN option_values SET DEFAULT '{}',
    ALTER COLUMN option_values SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variant_options
    ON product_variants(product_id, option_values)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_product_variants_option_values_gin
    ON product_variants
    USING GIN ((option_values::jsonb) jsonb_ops)
    WHERE deleted_at IS NULL;
