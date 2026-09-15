-- Stage three: sparse variant-level applicability and price overrides.
CREATE TABLE IF NOT EXISTS product_option_group_variant_rules (
    id BIGSERIAL PRIMARY KEY,
    variant_id BIGINT NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    spec_definition_id BIGINT NOT NULL REFERENCES product_spec_definitions(id) ON DELETE CASCADE,
    is_applicable BOOLEAN NOT NULL DEFAULT TRUE,
    min_selections_override INTEGER,
    max_selections_override INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_product_option_group_variant_rule UNIQUE (variant_id, spec_definition_id),
    CONSTRAINT ck_product_option_group_variant_rule_bounds CHECK (
        (min_selections_override IS NULL OR min_selections_override >= 0)
        AND (max_selections_override IS NULL OR max_selections_override >= 0)
        AND (min_selections_override IS NULL OR max_selections_override IS NULL OR max_selections_override >= min_selections_override)
    )
);

CREATE TABLE IF NOT EXISTS product_option_value_variant_rules (
    id BIGSERIAL PRIMARY KEY,
    variant_id BIGINT NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    product_variant_option_value_id BIGINT NOT NULL REFERENCES product_variant_option_values(id) ON DELETE CASCADE,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    price_delta_minor_override BIGINT,
    unavailable_reason VARCHAR(240),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_product_option_value_variant_rule UNIQUE (variant_id, product_variant_option_value_id),
    CONSTRAINT ck_product_option_value_variant_rule_price CHECK (price_delta_minor_override IS NULL OR price_delta_minor_override >= 0)
);

CREATE INDEX IF NOT EXISTS idx_product_option_group_variant_rules_variant
    ON product_option_group_variant_rules(variant_id);
CREATE INDEX IF NOT EXISTS idx_product_option_value_variant_rules_variant
    ON product_option_value_variant_rules(variant_id);
