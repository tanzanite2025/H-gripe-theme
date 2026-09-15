-- Stage three completion: the explicit role and candidate item tables are now
-- the only product specification contract. The project is pre-operation, so
-- legacy columns can be removed without a compatibility window.

ALTER TABLE product_spec_definitions
    DROP COLUMN IF EXISTS is_variant_option,
    DROP COLUMN IF EXISTS options;
