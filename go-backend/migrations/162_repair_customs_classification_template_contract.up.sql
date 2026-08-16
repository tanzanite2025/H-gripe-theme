-- Repair databases created before the product specification template contract rename.
-- Renaming the column keeps the existing foreign key and index definitions intact; their
-- identifiers are updated below to match the current schema contract.
DO $$
BEGIN
    IF to_regclass('public.customs_classification_profiles') IS NOT NULL
       AND EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'customs_classification_profiles'
             AND column_name = 'product_type_id'
       )
       AND NOT EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'customs_classification_profiles'
             AND column_name = 'product_specification_template_id'
       ) THEN
        ALTER TABLE customs_classification_profiles
            RENAME COLUMN product_type_id TO product_specification_template_id;
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('public.idx_customs_classification_profiles_product_type_id') IS NOT NULL
       AND to_regclass('public.idx_customs_classification_profiles_product_specification_templ') IS NULL THEN
        ALTER INDEX idx_customs_classification_profiles_product_type_id
            RENAME TO idx_customs_classification_profiles_product_specification_templ;
    END IF;

    IF to_regclass('public.customs_classification_profiles') IS NOT NULL
       AND EXISTS (
           SELECT 1
           FROM pg_constraint
           WHERE conrelid = 'public.customs_classification_profiles'::regclass
             AND conname = 'customs_classification_profiles_product_type_id_fkey'
       )
       AND NOT EXISTS (
           SELECT 1
           FROM pg_constraint
           WHERE conrelid = 'public.customs_classification_profiles'::regclass
             AND conname = 'customs_classification_profiles_product_specification_template_'
       ) THEN
        ALTER TABLE customs_classification_profiles
            RENAME CONSTRAINT customs_classification_profiles_product_type_id_fkey
            TO customs_classification_profiles_product_specification_template_;
    END IF;
END $$;
