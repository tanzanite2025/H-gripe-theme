CREATE TABLE IF NOT EXISTS quick_buy_flows (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(120) NOT NULL UNIQUE,
    name VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    entry_surface VARCHAR(80) NOT NULL DEFAULT 'dock',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_quick_buy_flows_slug
        CHECK (slug ~ '^[a-z0-9]+([_-][a-z0-9]+)*$')
);

CREATE INDEX IF NOT EXISTS idx_quick_buy_flows_surface_enabled
    ON quick_buy_flows(entry_surface, is_enabled, sort_order);

CREATE TABLE IF NOT EXISTS quick_buy_flow_versions (
    id BIGSERIAL PRIMARY KEY,
    flow_id BIGINT NOT NULL REFERENCES quick_buy_flows(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(24) NOT NULL DEFAULT 'draft',
    published_at TIMESTAMPTZ,
    published_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_quick_buy_flow_versions_status
        CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT ck_quick_buy_flow_versions_window
        CHECK (ends_at IS NULL OR starts_at IS NULL OR ends_at > starts_at)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_flow_versions_flow_number
    ON quick_buy_flow_versions(flow_id, version_number);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_flow_versions_one_published
    ON quick_buy_flow_versions(flow_id)
    WHERE status = 'published';

CREATE INDEX IF NOT EXISTS idx_quick_buy_flow_versions_status_window
    ON quick_buy_flow_versions(status, starts_at, ends_at);

CREATE TABLE IF NOT EXISTS quick_buy_steps (
    id BIGSERIAL PRIMARY KEY,
    flow_version_id BIGINT NOT NULL REFERENCES quick_buy_flow_versions(id) ON DELETE CASCADE,
    step_key VARCHAR(120) NOT NULL,
    name VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    help_text TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 100,
    selection_mode VARCHAR(24) NOT NULL DEFAULT 'single',
    is_required BOOLEAN NOT NULL DEFAULT TRUE,
    min_select INTEGER NOT NULL DEFAULT 0,
    max_select INTEGER NOT NULL DEFAULT 1,
    default_quantity INTEGER NOT NULL DEFAULT 1,
    allow_skip BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_quick_buy_steps_key
        CHECK (step_key ~ '^[a-z0-9]+([_-][a-z0-9]+)*$'),
    CONSTRAINT ck_quick_buy_steps_selection_mode
        CHECK (selection_mode IN ('single', 'multiple', 'quantity', 'auto')),
    CONSTRAINT ck_quick_buy_steps_select_bounds
        CHECK (min_select >= 0 AND max_select >= 0 AND default_quantity >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_steps_version_key
    ON quick_buy_steps(flow_version_id, step_key);

CREATE INDEX IF NOT EXISTS idx_quick_buy_steps_version_order
    ON quick_buy_steps(flow_version_id, sort_order, id);

CREATE TABLE IF NOT EXISTS quick_buy_step_product_specification_templates (
    id BIGSERIAL PRIMARY KEY,
    step_id BIGINT NOT NULL REFERENCES quick_buy_steps(id) ON DELETE CASCADE,
    product_specification_template_id BIGINT NOT NULL REFERENCES product_specification_templates(id) ON DELETE RESTRICT,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_step_product_specification_templates_unique
    ON quick_buy_step_product_specification_templates(step_id, product_specification_template_id);

CREATE INDEX IF NOT EXISTS idx_quick_buy_step_product_specification_templates_order
    ON quick_buy_step_product_specification_templates(step_id, sort_order, id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_step_product_specification_templates_one_primary
    ON quick_buy_step_product_specification_templates(step_id)
    WHERE is_primary = TRUE;

CREATE TABLE IF NOT EXISTS quick_buy_step_filters (
    id BIGSERIAL PRIMARY KEY,
    step_id BIGINT NOT NULL REFERENCES quick_buy_steps(id) ON DELETE CASCADE,
    filter_type VARCHAR(40) NOT NULL,
    spec_definition_id BIGINT REFERENCES product_spec_definitions(id) ON DELETE RESTRICT,
    operator VARCHAR(24) NOT NULL DEFAULT 'eq',
    value JSONB NOT NULL DEFAULT 'null'::jsonb,
    sort_order INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_quick_buy_step_filters_step_order
    ON quick_buy_step_filters(step_id, sort_order, id);

CREATE TABLE IF NOT EXISTS quick_buy_compatibility_rules (
    id BIGSERIAL PRIMARY KEY,
    flow_version_id BIGINT NOT NULL REFERENCES quick_buy_flow_versions(id) ON DELETE CASCADE,
    rule_key VARCHAR(120) NOT NULL,
    rule_type VARCHAR(40) NOT NULL,
    source_step_key VARCHAR(120) NOT NULL DEFAULT '',
    source_spec_key VARCHAR(120) NOT NULL DEFAULT '',
    target_step_key VARCHAR(120) NOT NULL DEFAULT '',
    target_spec_key VARCHAR(120) NOT NULL DEFAULT '',
    rule JSONB NOT NULL DEFAULT '{}'::jsonb,
    severity VARCHAR(16) NOT NULL DEFAULT 'error',
    message_key VARCHAR(160) NOT NULL DEFAULT '',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_quick_buy_compatibility_rules_severity
        CHECK (severity IN ('error', 'warning', 'info'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_compatibility_rules_version_key
    ON quick_buy_compatibility_rules(flow_version_id, rule_key);

CREATE INDEX IF NOT EXISTS idx_quick_buy_compatibility_rules_enabled
    ON quick_buy_compatibility_rules(flow_version_id, is_enabled, sort_order);

CREATE TABLE IF NOT EXISTS quick_buy_sessions (
    id BIGSERIAL PRIMARY KEY,
    session_token VARCHAR(96) NOT NULL UNIQUE,
    flow_id BIGINT NOT NULL REFERENCES quick_buy_flows(id) ON DELETE RESTRICT,
    flow_version_id BIGINT NOT NULL REFERENCES quick_buy_flow_versions(id) ON DELETE RESTRICT,
    locale VARCHAR(32) NOT NULL DEFAULT 'en',
    market_country VARCHAR(8) NOT NULL DEFAULT '',
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    anonymous_id VARCHAR(128) NOT NULL DEFAULT '',
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'active',
    validation_status VARCHAR(24) NOT NULL DEFAULT 'valid',
    subtotal_snapshot NUMERIC(12,2) NOT NULL DEFAULT 0,
    weight_snapshot_g INTEGER NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_quick_buy_sessions_status
        CHECK (status IN ('active', 'completed', 'added_to_cart', 'ordered', 'abandoned', 'expired')),
    CONSTRAINT ck_quick_buy_sessions_validation_status
        CHECK (validation_status IN ('valid', 'warning', 'invalid'))
);

CREATE INDEX IF NOT EXISTS idx_quick_buy_sessions_flow_version
    ON quick_buy_sessions(flow_version_id, status);

CREATE INDEX IF NOT EXISTS idx_quick_buy_sessions_user
    ON quick_buy_sessions(user_id, status, updated_at);

CREATE INDEX IF NOT EXISTS idx_quick_buy_sessions_anonymous
    ON quick_buy_sessions(anonymous_id, status, updated_at);

CREATE TABLE IF NOT EXISTS quick_buy_session_items (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES quick_buy_sessions(id) ON DELETE CASCADE,
    step_id BIGINT NOT NULL REFERENCES quick_buy_steps(id) ON DELETE RESTRICT,
    step_key VARCHAR(120) NOT NULL,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    variant_id BIGINT REFERENCES product_variants(id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price_snapshot NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency_snapshot VARCHAR(3) NOT NULL DEFAULT 'USD',
    weight_snapshot_g INTEGER NOT NULL DEFAULT 0,
    product_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    variant_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    sort_order INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_quick_buy_session_items_quantity
        CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_quick_buy_session_items_session_order
    ON quick_buy_session_items(session_id, sort_order, id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_session_items_single_step
    ON quick_buy_session_items(session_id, step_key, product_id, COALESCE(variant_id, 0));
