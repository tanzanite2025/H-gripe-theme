-- A rollback restores the first active selection; multiple customer choices
-- cannot be represented by the former single-template schema.
ALTER TABLE products ADD COLUMN IF NOT EXISTS shipping_template_id BIGINT REFERENCES shipping_templates(id) ON DELETE SET NULL;
ALTER TABLE product_variants ADD COLUMN IF NOT EXISTS shipping_template_id BIGINT REFERENCES shipping_templates(id) ON DELETE SET NULL;

UPDATE products p SET shipping_template_id = (
    SELECT b.shipping_template_id FROM product_shipping_template_bindings b
    WHERE b.product_id = p.id AND b.variant_id IS NULL AND b.enabled AND b.deleted_at IS NULL
    ORDER BY b.priority, b.id LIMIT 1
);
UPDATE product_variants pv SET shipping_template_id = (
    SELECT b.shipping_template_id FROM product_shipping_template_bindings b
    WHERE b.product_id = pv.product_id AND b.variant_id = pv.id AND b.enabled AND b.deleted_at IS NULL
    ORDER BY b.priority, b.id LIMIT 1
);

DROP TABLE IF EXISTS product_shipping_template_bindings;

ALTER TABLE shipping_carrier_services
    ADD COLUMN IF NOT EXISTS template_id BIGINT REFERENCES shipping_templates(id) ON DELETE SET NULL;

ALTER TABLE shipping_templates
    DROP CONSTRAINT IF EXISTS fk_shipping_templates_route_group,
    DROP COLUMN IF EXISTS route_group_id,
    DROP COLUMN IF EXISTS customer_billing_mode,
    DROP COLUMN IF EXISTS customer_first_weight_grams,
    DROP COLUMN IF EXISTS customer_additional_weight_grams,
    DROP COLUMN IF EXISTS customer_min_charge_weight_grams,
    DROP COLUMN IF EXISTS customer_volumetric_divisor,
    DROP COLUMN IF EXISTS customer_fuel_surcharge_percent_decimal,
    DROP COLUMN IF EXISTS customer_remote_surcharge_minor,
    DROP COLUMN IF EXISTS customer_remote_postal_codes,
    DROP COLUMN IF EXISTS customer_eta_min_days,
    DROP COLUMN IF EXISTS customer_eta_max_days;

DROP TABLE IF EXISTS shipping_route_group_carrier_services;
DROP TABLE IF EXISTS shipping_route_groups;
