-- Published route identity and customer-facing billing are separate domains.
-- Provider collections may describe 4PX/Yanwen capabilities, but never set
-- the amount Checkout charges a customer.

CREATE TABLE IF NOT EXISTS shipping_route_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(160) NOT NULL,
    description TEXT,
    countries TEXT NOT NULL DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_shipping_route_groups_enabled
    ON shipping_route_groups (enabled);
CREATE INDEX IF NOT EXISTS idx_shipping_route_groups_deleted_at
    ON shipping_route_groups (deleted_at);

CREATE TABLE IF NOT EXISTS shipping_route_group_carrier_services (
    route_group_id BIGINT NOT NULL REFERENCES shipping_route_groups(id) ON DELETE CASCADE,
    carrier_service_id BIGINT NOT NULL REFERENCES shipping_carrier_services(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (route_group_id, carrier_service_id)
);

ALTER TABLE shipping_templates
    ADD COLUMN IF NOT EXISTS route_group_id BIGINT,
    ADD COLUMN IF NOT EXISTS customer_billing_mode VARCHAR(40) NOT NULL DEFAULT 'actual_weight',
    ADD COLUMN IF NOT EXISTS customer_first_weight_grams INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS customer_additional_weight_grams INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS customer_min_charge_weight_grams INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS customer_volumetric_divisor INTEGER NOT NULL DEFAULT 6000,
    ADD COLUMN IF NOT EXISTS customer_fuel_surcharge_percent_decimal NUMERIC(30,15) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS customer_remote_surcharge_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS customer_remote_postal_codes TEXT NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS customer_eta_min_days INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS customer_eta_max_days INTEGER NOT NULL DEFAULT 0;

-- Seed one explicit aggregate for each existing pricing template that had
-- attached provider routes. New code uses route_group_id; this only preserves
-- the reviewed route selection while separating its ownership boundary.
DO $$
DECLARE
    source_template RECORD;
    aggregate_id BIGINT;
BEGIN
    FOR source_template IN
        SELECT st.id, st.name, st.enabled FROM shipping_templates st
        WHERE st.deleted_at IS NULL AND st.route_group_id IS NULL
          AND st.template_kind <> 'system_free_shipping'
          AND EXISTS (SELECT 1 FROM shipping_carrier_services scs WHERE scs.template_id = st.id AND scs.deleted_at IS NULL)
    LOOP
        INSERT INTO shipping_route_groups (name, description, enabled)
        VALUES (LEFT(source_template.name, 140) || ' route aggregate', 'Migrated route identity; customer prices require separate configuration', source_template.enabled)
        RETURNING id INTO aggregate_id;
        UPDATE shipping_templates SET route_group_id = aggregate_id WHERE id = source_template.id;
    END LOOP;
END $$;

INSERT INTO shipping_route_group_carrier_services (route_group_id, carrier_service_id, sort_order)
SELECT st.route_group_id, scs.id, scs.sort_order
FROM shipping_carrier_services scs JOIN shipping_templates st ON st.id = scs.template_id
WHERE st.route_group_id IS NOT NULL AND scs.deleted_at IS NULL
ON CONFLICT (route_group_id, carrier_service_id) DO UPDATE SET sort_order = EXCLUDED.sort_order;

-- Carrier identities cannot own or select customer billing templates.
ALTER TABLE shipping_carrier_services DROP COLUMN IF EXISTS template_id;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_shipping_templates_route_group') THEN
        ALTER TABLE shipping_templates
            ADD CONSTRAINT fk_shipping_templates_route_group
            FOREIGN KEY (route_group_id) REFERENCES shipping_route_groups(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_shipping_templates_route_group_id
    ON shipping_templates (route_group_id);

CREATE TABLE IF NOT EXISTS product_shipping_template_bindings (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_id BIGINT REFERENCES product_variants(id) ON DELETE CASCADE,
    shipping_template_id BIGINT NOT NULL REFERENCES shipping_templates(id) ON DELETE RESTRICT,
    priority INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_shipping_template_bindings_product_scope
    ON product_shipping_template_bindings (product_id, shipping_template_id)
    WHERE variant_id IS NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_shipping_template_bindings_variant_scope
    ON product_shipping_template_bindings (product_id, variant_id, shipping_template_id)
    WHERE variant_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_product_shipping_template_bindings_product
    ON product_shipping_template_bindings (product_id, enabled, priority DESC);
CREATE INDEX IF NOT EXISTS idx_product_shipping_template_bindings_variant
    ON product_shipping_template_bindings (variant_id, enabled, priority DESC);

-- Existing single selections become the initial canonical multi-selection.
INSERT INTO product_shipping_template_bindings (product_id, shipping_template_id, priority)
SELECT p.id, p.shipping_template_id, 0
FROM products p
WHERE p.shipping_template_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM product_shipping_template_bindings b
      WHERE b.product_id = p.id AND b.variant_id IS NULL
        AND b.shipping_template_id = p.shipping_template_id AND b.deleted_at IS NULL
  );

INSERT INTO product_shipping_template_bindings (product_id, variant_id, shipping_template_id, priority)
SELECT pv.product_id, pv.id, pv.shipping_template_id, 0
FROM product_variants pv
WHERE pv.shipping_template_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM product_shipping_template_bindings b
      WHERE b.variant_id = pv.id AND b.shipping_template_id = pv.shipping_template_id
        AND b.deleted_at IS NULL
  );

-- Runtime readers and writers use bindings exclusively. Remove the obsolete
-- columns so startup AutoMigrate cannot leave a second mutable selection.
ALTER TABLE products DROP COLUMN IF EXISTS shipping_template_id;
ALTER TABLE product_variants DROP COLUMN IF EXISTS shipping_template_id;
