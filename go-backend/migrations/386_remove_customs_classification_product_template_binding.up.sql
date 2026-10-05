-- Customs classification profiles are independent reusable master data.
-- Products select a profile directly; a product specification template must not
-- constrain which customs profile can be used.
ALTER TABLE customs_classification_profiles
    DROP CONSTRAINT IF EXISTS customs_classification_profiles_product_specification_template_id_fkey,
    DROP CONSTRAINT IF EXISTS customs_classification_profiles_product_specification_template_,
    DROP CONSTRAINT IF EXISTS customs_classification_profiles_product_type_id_fkey;

DROP INDEX IF EXISTS idx_customs_classification_profiles_product_specification_template_id;
DROP INDEX IF EXISTS idx_customs_classification_profiles_product_specification_templ;
DROP INDEX IF EXISTS idx_customs_classification_profiles_product_type_id;

ALTER TABLE customs_classification_profiles
    DROP COLUMN IF EXISTS product_specification_template_id;
