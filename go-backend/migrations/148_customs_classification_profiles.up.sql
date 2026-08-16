CREATE TABLE IF NOT EXISTS customs_classification_profiles (
    id BIGSERIAL PRIMARY KEY,
    product_specification_template_id BIGINT REFERENCES product_specification_templates(id) ON DELETE SET NULL,
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(140) NOT NULL UNIQUE,
    component_kind VARCHAR(64) NOT NULL DEFAULT '',
    material VARCHAR(64) NOT NULL DEFAULT '',
    hs_code VARCHAR(12) NOT NULL,
    cn_code VARCHAR(12) NOT NULL DEFAULT '',
    country_of_origin VARCHAR(2) NOT NULL DEFAULT '',
    customs_description VARCHAR(255) NOT NULL DEFAULT '',
    source VARCHAR(32) NOT NULL DEFAULT '',
    source_code VARCHAR(64) NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_customs_classification_status CHECK (status IN ('draft', 'active', 'paused')),
    CONSTRAINT ck_customs_classification_hs_code CHECK (hs_code ~ '^[0-9]{6}$'),
    CONSTRAINT ck_customs_classification_cn_code CHECK (cn_code = '' OR cn_code ~ '^[0-9]{8}$'),
    CONSTRAINT ck_customs_classification_origin CHECK (country_of_origin = '' OR country_of_origin ~ '^[A-Z]{2}$')
);

CREATE INDEX IF NOT EXISTS idx_customs_classification_profiles_product_specification_template_id
    ON customs_classification_profiles(product_specification_template_id);

CREATE INDEX IF NOT EXISTS idx_customs_classification_profiles_status
    ON customs_classification_profiles(status);

CREATE INDEX IF NOT EXISTS idx_customs_classification_profiles_component_material
    ON customs_classification_profiles(component_kind, material);
