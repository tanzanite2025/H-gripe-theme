CREATE TABLE IF NOT EXISTS product_quality_requirement_rules (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    variant_id BIGINT REFERENCES product_variants(id) ON DELETE RESTRICT,
    requirement_type VARCHAR(64) NOT NULL,
    spoke_tension_qc_required BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    rule_version VARCHAR(64) NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_by BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT product_quality_requirement_type_check
        CHECK (requirement_type = 'spoke_tension_qc'),
    CONSTRAINT product_quality_requirement_status_check
        CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_product_quality_requirement_product
    ON product_quality_requirement_rules(product_id);
CREATE INDEX IF NOT EXISTS idx_product_quality_requirement_variant
    ON product_quality_requirement_rules(variant_id);
CREATE INDEX IF NOT EXISTS idx_product_quality_requirement_status
    ON product_quality_requirement_rules(status);

CREATE UNIQUE INDEX IF NOT EXISTS uq_product_quality_requirement_product_default
    ON product_quality_requirement_rules(product_id, requirement_type)
    WHERE variant_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_product_quality_requirement_variant
    ON product_quality_requirement_rules(product_id, variant_id, requirement_type)
    WHERE variant_id IS NOT NULL;
