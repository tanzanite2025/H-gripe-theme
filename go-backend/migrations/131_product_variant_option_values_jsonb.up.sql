-- Promote variant option filters to a first-class JSONB storage contract.
-- 130 kept the legacy TEXT column usable through an expression GIN index; from
-- this migration forward product_variants.option_values is JSONB directly.

DROP INDEX IF EXISTS idx_product_variants_option_values_gin;
DROP INDEX IF EXISTS idx_product_variant_options;

UPDATE product_variants
SET option_values = '{}'
WHERE option_values IS NULL OR btrim(option_values::text) = '';

ALTER TABLE product_variants
    ALTER COLUMN option_values DROP DEFAULT,
    ALTER COLUMN option_values TYPE jsonb USING option_values::jsonb,
    ALTER COLUMN option_values SET DEFAULT '{}'::jsonb,
    ALTER COLUMN option_values SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_product_variants_option_values_object'
    ) THEN
        ALTER TABLE product_variants
            ADD CONSTRAINT chk_product_variants_option_values_object
            CHECK (jsonb_typeof(option_values) = 'object');
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variant_options
    ON product_variants(product_id, option_values)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_product_variants_option_values_gin
    ON product_variants
    USING GIN (option_values jsonb_ops)
    WHERE deleted_at IS NULL;
