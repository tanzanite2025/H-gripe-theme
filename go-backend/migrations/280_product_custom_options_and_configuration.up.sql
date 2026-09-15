-- Stage two: explicit product option roles, custom-option pricing policy,
-- and immutable cart/order configuration snapshots.

ALTER TABLE product_spec_definitions
    ADD COLUMN IF NOT EXISTS role VARCHAR(24) NOT NULL DEFAULT 'attribute',
    ADD COLUMN IF NOT EXISTS selection_mode VARCHAR(16) NOT NULL DEFAULT 'single',
    ADD COLUMN IF NOT EXISTS min_selections INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS max_selections INTEGER;

UPDATE product_spec_definitions
   SET role = 'variant'
 WHERE is_variant_option = TRUE
   AND (role IS NULL OR role = 'attribute');

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'chk_product_spec_definitions_role'
    ) THEN
        ALTER TABLE product_spec_definitions
            ADD CONSTRAINT chk_product_spec_definitions_role
            CHECK (role IN ('attribute', 'variant', 'custom_option'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'chk_product_spec_definitions_selection_mode'
    ) THEN
        ALTER TABLE product_spec_definitions
            ADD CONSTRAINT chk_product_spec_definitions_selection_mode
            CHECK (selection_mode IN ('single', 'multiple'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'chk_product_spec_definitions_selection_bounds'
    ) THEN
        ALTER TABLE product_spec_definitions
            ADD CONSTRAINT chk_product_spec_definitions_selection_bounds
            CHECK (min_selections >= 0 AND (max_selections IS NULL OR max_selections >= min_selections));
    END IF;
END $$;

ALTER TABLE product_specification_templates
    ADD COLUMN IF NOT EXISTS revision INTEGER NOT NULL DEFAULT 1;

ALTER TABLE product_variant_option_values
    ADD COLUMN IF NOT EXISTS template_option_item_id BIGINT,
    ADD COLUMN IF NOT EXISTS source_template_revision INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_product_variant_option_values_template_item
    ON product_variant_option_values(template_option_item_id);

CREATE TABLE IF NOT EXISTS product_custom_option_policies (
    id BIGSERIAL PRIMARY KEY,
    product_variant_option_value_id BIGINT NOT NULL UNIQUE
        REFERENCES product_variant_option_values(id) ON DELETE CASCADE,
    price_delta_minor BIGINT NOT NULL DEFAULT 0,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    inventory_policy VARCHAR(24) NOT NULL DEFAULT 'none',
    component_variant_id BIGINT REFERENCES product_variants(id) ON DELETE RESTRICT,
    component_quantity INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_product_custom_option_policy_price_delta_non_negative
        CHECK (price_delta_minor >= 0),
    CONSTRAINT chk_product_custom_option_policy_inventory_policy
        CHECK (inventory_policy IN ('none', 'component')),
    CONSTRAINT chk_product_custom_option_policy_component_quantity_non_negative
        CHECK (component_quantity >= 0)
);

CREATE INDEX IF NOT EXISTS idx_product_custom_option_policies_component_variant
    ON product_custom_option_policies(component_variant_id);

ALTER TABLE cart_items
    ADD COLUMN IF NOT EXISTS configuration JSONB NOT NULL
        DEFAULT '{"schema_version":1,"selections":[]}'::jsonb,
    ADD COLUMN IF NOT EXISTS configuration_hash CHAR(64) NOT NULL
        DEFAULT '49aaaf6402be0bdf1a77fbed39f489bd4b6f3e6364e8aca69b5c49db4899a9a1';

UPDATE cart_items
   SET configuration = '{"schema_version":1,"selections":[]}'::jsonb
 WHERE configuration IS NULL;

UPDATE cart_items
   SET configuration_hash = '49aaaf6402be0bdf1a77fbed39f489bd4b6f3e6364e8aca69b5c49db4899a9a1'
 WHERE configuration_hash IS NULL OR btrim(configuration_hash) = '';

-- Existing rows predate configuration identity. Merge rows that become
-- identical before installing the stronger unique index.
DO $$
DECLARE
    duplicate_item RECORD;
BEGIN
    FOR duplicate_item IN
        SELECT duplicate.id AS duplicate_id,
               canonical.id AS canonical_id
          FROM cart_items AS duplicate
          JOIN LATERAL (
              SELECT candidate.id
                FROM cart_items AS candidate
               WHERE candidate.cart_id = duplicate.cart_id
                 AND candidate.product_id = duplicate.product_id
                 AND candidate.variant_id = duplicate.variant_id
                 AND candidate.configuration_hash = duplicate.configuration_hash
               ORDER BY candidate.id
               LIMIT 1
          ) AS canonical ON TRUE
         WHERE duplicate.id <> canonical.id
         ORDER BY duplicate.id
    LOOP
        UPDATE cart_items AS target
           SET quantity = target.quantity + source.quantity,
               price = source.price,
               currency = source.currency,
               updated_at = CURRENT_TIMESTAMP
          FROM cart_items AS source
         WHERE target.id = duplicate_item.canonical_id
           AND source.id = duplicate_item.duplicate_id;
        DELETE FROM cart_items WHERE id = duplicate_item.duplicate_id;
    END LOOP;
END $$;

DROP INDEX IF EXISTS idx_cart_product_variant;
DROP INDEX IF EXISTS idx_cart_items_product_variant;
CREATE UNIQUE INDEX IF NOT EXISTS uq_cart_items_identity
    ON cart_items(cart_id, product_id, variant_id, configuration_hash);

ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS configuration_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE OR REPLACE FUNCTION prevent_order_item_configuration_snapshot_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.configuration_snapshot <> '{}'::jsonb
       AND NEW.configuration_snapshot IS DISTINCT FROM OLD.configuration_snapshot THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'order item configuration snapshot is immutable once populated';
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trigger_prevent_order_item_configuration_snapshot_mutation ON order_items;

CREATE TRIGGER trigger_prevent_order_item_configuration_snapshot_mutation
BEFORE UPDATE OF configuration_snapshot ON order_items
FOR EACH ROW
EXECUTE FUNCTION prevent_order_item_configuration_snapshot_mutation();
