CREATE TABLE IF NOT EXISTS shipping_templates (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'weight',
    free_shipping BOOLEAN NOT NULL DEFAULT FALSE,
    free_threshold NUMERIC(12, 2) NOT NULL DEFAULT 0,
    default_fee NUMERIC(12, 2) NOT NULL DEFAULT 0,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE shipping_templates ADD COLUMN IF NOT EXISTS name TEXT;
ALTER TABLE shipping_templates ADD COLUMN IF NOT EXISTS type TEXT DEFAULT 'weight';
ALTER TABLE shipping_templates ADD COLUMN IF NOT EXISTS free_shipping BOOLEAN DEFAULT FALSE;
ALTER TABLE shipping_templates ADD COLUMN IF NOT EXISTS free_threshold NUMERIC(12, 2) DEFAULT 0;
ALTER TABLE shipping_templates ADD COLUMN IF NOT EXISTS default_fee NUMERIC(12, 2) DEFAULT 0;
ALTER TABLE shipping_templates ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE shipping_templates ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE shipping_templates ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_templates ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_templates ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_shipping_templates_deleted_at ON shipping_templates (deleted_at);
CREATE INDEX IF NOT EXISTS idx_shipping_templates_enabled ON shipping_templates (enabled);

CREATE TABLE IF NOT EXISTS shipping_rules (
    id BIGSERIAL PRIMARY KEY,
    template_id BIGINT NOT NULL REFERENCES shipping_templates(id) ON DELETE CASCADE,
    region TEXT NOT NULL DEFAULT '',
    min_value NUMERIC(12, 3) NOT NULL DEFAULT 0,
    max_value NUMERIC(12, 3) NOT NULL DEFAULT 0,
    fee NUMERIC(12, 2) NOT NULL DEFAULT 0,
    additional NUMERIC(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE shipping_rules ADD COLUMN IF NOT EXISTS template_id BIGINT;
ALTER TABLE shipping_rules ADD COLUMN IF NOT EXISTS region TEXT DEFAULT '';
ALTER TABLE shipping_rules ADD COLUMN IF NOT EXISTS min_value NUMERIC(12, 3) DEFAULT 0;
ALTER TABLE shipping_rules ADD COLUMN IF NOT EXISTS max_value NUMERIC(12, 3) DEFAULT 0;
ALTER TABLE shipping_rules ADD COLUMN IF NOT EXISTS fee NUMERIC(12, 2) DEFAULT 0;
ALTER TABLE shipping_rules ADD COLUMN IF NOT EXISTS additional NUMERIC(12, 2) DEFAULT 0;
ALTER TABLE shipping_rules ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_shipping_rules_template_id ON shipping_rules (template_id);
CREATE INDEX IF NOT EXISTS idx_shipping_rules_region ON shipping_rules (region);

CREATE TABLE IF NOT EXISTS shipping_zones (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    countries TEXT NOT NULL DEFAULT '[]',
    states TEXT NOT NULL DEFAULT '[]',
    postal_codes TEXT NOT NULL DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE shipping_zones ADD COLUMN IF NOT EXISTS name TEXT;
ALTER TABLE shipping_zones ADD COLUMN IF NOT EXISTS countries TEXT DEFAULT '[]';
ALTER TABLE shipping_zones ADD COLUMN IF NOT EXISTS states TEXT DEFAULT '[]';
ALTER TABLE shipping_zones ADD COLUMN IF NOT EXISTS postal_codes TEXT DEFAULT '[]';
ALTER TABLE shipping_zones ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE shipping_zones ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_zones ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_zones ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_shipping_zones_deleted_at ON shipping_zones (deleted_at);
CREATE INDEX IF NOT EXISTS idx_shipping_zones_enabled ON shipping_zones (enabled);

CREATE TABLE IF NOT EXISTS carriers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    tracking_url TEXT,
    contact TEXT,
    phone TEXT,
    email TEXT,
    service_area TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE carriers ADD COLUMN IF NOT EXISTS name TEXT;
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS code TEXT;
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS tracking_url TEXT;
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS contact TEXT;
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS phone TEXT;
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS email TEXT;
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS service_area TEXT;
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS sort_order INTEGER DEFAULT 0;
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_carriers_code ON carriers (code);
CREATE INDEX IF NOT EXISTS idx_carriers_deleted_at ON carriers (deleted_at);
CREATE INDEX IF NOT EXISTS idx_carriers_enabled ON carriers (enabled);

CREATE TABLE IF NOT EXISTS tracking_events (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    tracking_number TEXT,
    provider_carrier_code TEXT,
    status TEXT,
    location TEXT,
    description TEXT,
    event_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE tracking_events ADD COLUMN IF NOT EXISTS order_id BIGINT;
ALTER TABLE tracking_events ADD COLUMN IF NOT EXISTS tracking_number TEXT;
ALTER TABLE tracking_events ADD COLUMN IF NOT EXISTS provider_carrier_code TEXT;
ALTER TABLE tracking_events ADD COLUMN IF NOT EXISTS status TEXT;
ALTER TABLE tracking_events ADD COLUMN IF NOT EXISTS location TEXT;
ALTER TABLE tracking_events ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE tracking_events ADD COLUMN IF NOT EXISTS event_time TIMESTAMPTZ;
ALTER TABLE tracking_events ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_tracking_events_order_id ON tracking_events (order_id);
CREATE INDEX IF NOT EXISTS idx_tracking_events_tracking_number ON tracking_events (tracking_number);

CREATE TABLE IF NOT EXISTS shipping_packaging_rules (
    id BIGSERIAL PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL,
    description TEXT,
    box_weight NUMERIC(10, 3) NOT NULL DEFAULT 0,
    box_length NUMERIC(10, 2) NOT NULL DEFAULT 0,
    box_width NUMERIC(10, 2) NOT NULL DEFAULT 0,
    box_height NUMERIC(10, 2) NOT NULL DEFAULT 0,
    max_weight NUMERIC(10, 3) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE shipping_packaging_rules ADD COLUMN IF NOT EXISTS rule_name VARCHAR(100);
ALTER TABLE shipping_packaging_rules ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE shipping_packaging_rules ADD COLUMN IF NOT EXISTS box_weight NUMERIC(10, 3) DEFAULT 0;
ALTER TABLE shipping_packaging_rules ADD COLUMN IF NOT EXISTS box_length NUMERIC(10, 2) DEFAULT 0;
ALTER TABLE shipping_packaging_rules ADD COLUMN IF NOT EXISTS box_width NUMERIC(10, 2) DEFAULT 0;
ALTER TABLE shipping_packaging_rules ADD COLUMN IF NOT EXISTS box_height NUMERIC(10, 2) DEFAULT 0;
ALTER TABLE shipping_packaging_rules ADD COLUMN IF NOT EXISTS max_weight NUMERIC(10, 3) DEFAULT 0;
ALTER TABLE shipping_packaging_rules ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;
ALTER TABLE shipping_packaging_rules ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_packaging_rules ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_shipping_packaging_rules_is_active ON shipping_packaging_rules (is_active);

CREATE TABLE IF NOT EXISTS shipping_packaging_rule_applies (
    id BIGSERIAL PRIMARY KEY,
    rule_id BIGINT NOT NULL REFERENCES shipping_packaging_rules(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE shipping_packaging_rule_applies ADD COLUMN IF NOT EXISTS rule_id BIGINT;
ALTER TABLE shipping_packaging_rule_applies ADD COLUMN IF NOT EXISTS product_id BIGINT;
ALTER TABLE shipping_packaging_rule_applies ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_shipping_packaging_rule_applies_rule_id ON shipping_packaging_rule_applies (rule_id);
CREATE INDEX IF NOT EXISTS idx_shipping_packaging_rule_applies_product_id ON shipping_packaging_rule_applies (product_id);

CREATE TABLE IF NOT EXISTS shipping_template_bindings (
    id BIGSERIAL PRIMARY KEY,
    template_id BIGINT NOT NULL REFERENCES shipping_templates(id) ON DELETE CASCADE,
    scope VARCHAR(30) NOT NULL,
    product_type_id BIGINT,
    product_id BIGINT,
    variant_id BIGINT,
    priority INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE shipping_template_bindings ADD COLUMN IF NOT EXISTS template_id BIGINT;
ALTER TABLE shipping_template_bindings ADD COLUMN IF NOT EXISTS scope VARCHAR(30);
ALTER TABLE shipping_template_bindings ADD COLUMN IF NOT EXISTS product_type_id BIGINT;
ALTER TABLE shipping_template_bindings ADD COLUMN IF NOT EXISTS product_id BIGINT;
ALTER TABLE shipping_template_bindings ADD COLUMN IF NOT EXISTS variant_id BIGINT;
ALTER TABLE shipping_template_bindings ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 0;
ALTER TABLE shipping_template_bindings ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE shipping_template_bindings ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_template_bindings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_template_bindings ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_shipping_template_bindings_template_id ON shipping_template_bindings (template_id);
CREATE INDEX IF NOT EXISTS idx_shipping_template_bindings_scope ON shipping_template_bindings (scope);
CREATE INDEX IF NOT EXISTS idx_shipping_template_bindings_product_type_id ON shipping_template_bindings (product_type_id);
CREATE INDEX IF NOT EXISTS idx_shipping_template_bindings_product_id ON shipping_template_bindings (product_id);
CREATE INDEX IF NOT EXISTS idx_shipping_template_bindings_variant_id ON shipping_template_bindings (variant_id);
CREATE INDEX IF NOT EXISTS idx_shipping_template_bindings_enabled ON shipping_template_bindings (enabled);
CREATE INDEX IF NOT EXISTS idx_shipping_template_bindings_deleted_at ON shipping_template_bindings (deleted_at);
