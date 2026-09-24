-- Stage five: explicit product-local custom option dependency relations.
CREATE TABLE IF NOT EXISTS product_option_value_relations (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    source_option_value_id BIGINT NOT NULL REFERENCES product_variant_option_values(id) ON DELETE CASCADE,
    target_option_value_id BIGINT NOT NULL REFERENCES product_variant_option_values(id) ON DELETE CASCADE,
    relation_type VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_product_option_value_relation_key UNIQUE (product_id, source_option_value_id, target_option_value_id, relation_type),
    CONSTRAINT chk_product_option_value_relation_type CHECK (relation_type IN ('requires', 'conflicts')),
    CONSTRAINT chk_product_option_value_relation_distinct CHECK (source_option_value_id <> target_option_value_id)
);

CREATE INDEX IF NOT EXISTS idx_product_option_value_relations_product
    ON product_option_value_relations(product_id);
CREATE INDEX IF NOT EXISTS idx_product_option_value_relations_source
    ON product_option_value_relations(source_option_value_id);
CREATE INDEX IF NOT EXISTS idx_product_option_value_relations_target
    ON product_option_value_relations(target_option_value_id);
