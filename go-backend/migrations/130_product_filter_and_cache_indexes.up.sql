-- Product filter indexes for the current specification/variant data model.
-- product_variants.option_values stores JSON objects in a legacy TEXT column,
-- so the expression index keeps the storage contract while allowing JSONB
-- containment/key queries to use a GIN index.

CREATE INDEX IF NOT EXISTS idx_product_variants_option_values_gin
    ON product_variants
    USING GIN ((option_values::jsonb) jsonb_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_product_spec_values_filter_lookup
    ON product_spec_values (spec_definition_id, value, product_id);

CREATE INDEX IF NOT EXISTS idx_product_spec_definitions_filter_lookup
    ON product_spec_definitions (slug, is_filterable, id);
