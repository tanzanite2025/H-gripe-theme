-- Restore the legacy nullable binding only when explicitly rolling back the
-- decoupling migration.
ALTER TABLE customs_classification_profiles
    ADD COLUMN IF NOT EXISTS product_specification_template_id BIGINT;

DO $$
BEGIN
    IF to_regclass('public.product_specification_templates') IS NOT NULL
       AND NOT EXISTS (
           SELECT 1
           FROM pg_constraint
           WHERE conrelid = 'public.customs_classification_profiles'::regclass
             AND conname = 'customs_classification_profiles_product_specification_template_id_fkey'
       ) THEN
        ALTER TABLE customs_classification_profiles
            ADD CONSTRAINT customs_classification_profiles_product_specification_template_id_fkey
            FOREIGN KEY (product_specification_template_id)
            REFERENCES product_specification_templates(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_customs_classification_profiles_product_specification_template_id
    ON customs_classification_profiles(product_specification_template_id);
