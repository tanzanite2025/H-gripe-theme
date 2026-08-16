ALTER TABLE products
    DROP CONSTRAINT IF EXISTS fk_products_customs_classification_profile;

DROP INDEX IF EXISTS idx_products_customs_classification_profile_id;

ALTER TABLE products
    DROP COLUMN IF EXISTS customs_classification_profile_id;
