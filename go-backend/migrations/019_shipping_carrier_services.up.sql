CREATE TABLE IF NOT EXISTS shipping_carrier_services (
    id BIGSERIAL PRIMARY KEY,
    carrier_id BIGINT NOT NULL REFERENCES carriers(id) ON DELETE CASCADE,
    template_id BIGINT REFERENCES shipping_templates(id) ON DELETE SET NULL,
    service_code VARCHAR(80) NOT NULL,
    service_name VARCHAR(160) NOT NULL,
    route_name VARCHAR(160),
    countries TEXT NOT NULL DEFAULT '[]',
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    billing_mode VARCHAR(40) NOT NULL DEFAULT 'actual_weight',
    first_weight_grams INTEGER NOT NULL DEFAULT 0,
    additional_weight_grams INTEGER NOT NULL DEFAULT 0,
    min_charge_weight_grams INTEGER NOT NULL DEFAULT 0,
    volumetric_divisor INTEGER NOT NULL DEFAULT 6000,
    fuel_surcharge_percent NUMERIC(8, 3) NOT NULL DEFAULT 0,
    remote_surcharge NUMERIC(12, 2) NOT NULL DEFAULT 0,
    eta_min_days INTEGER NOT NULL DEFAULT 0,
    eta_max_days INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS carrier_id BIGINT;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS template_id BIGINT;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS service_code VARCHAR(80);
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS service_name VARCHAR(160);
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS route_name VARCHAR(160);
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS countries TEXT DEFAULT '[]';
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS currency VARCHAR(10) DEFAULT 'USD';
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS billing_mode VARCHAR(40) DEFAULT 'actual_weight';
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS first_weight_grams INTEGER DEFAULT 0;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS additional_weight_grams INTEGER DEFAULT 0;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS min_charge_weight_grams INTEGER DEFAULT 0;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS volumetric_divisor INTEGER DEFAULT 6000;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS fuel_surcharge_percent NUMERIC(8, 3) DEFAULT 0;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS remote_surcharge NUMERIC(12, 2) DEFAULT 0;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS eta_min_days INTEGER DEFAULT 0;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS eta_max_days INTEGER DEFAULT 0;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS sort_order INTEGER DEFAULT 0;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE shipping_carrier_services ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_carrier_services_carrier_code
    ON shipping_carrier_services (carrier_id, service_code)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shipping_carrier_services_carrier_id ON shipping_carrier_services (carrier_id);
CREATE INDEX IF NOT EXISTS idx_shipping_carrier_services_template_id ON shipping_carrier_services (template_id);
CREATE INDEX IF NOT EXISTS idx_shipping_carrier_services_enabled ON shipping_carrier_services (enabled);
CREATE INDEX IF NOT EXISTS idx_shipping_carrier_services_deleted_at ON shipping_carrier_services (deleted_at);
CREATE INDEX IF NOT EXISTS idx_shipping_carrier_services_sort_order ON shipping_carrier_services (sort_order);
