DO $$
BEGIN
    IF to_regclass('public.product_types') IS NOT NULL
       AND to_regclass('public.product_specification_templates') IS NULL THEN
        ALTER TABLE product_types RENAME TO product_specification_templates;
    END IF;

    IF to_regclass('public.product_type_translations') IS NOT NULL
       AND to_regclass('public.product_specification_template_translations') IS NULL THEN
        ALTER TABLE product_type_translations RENAME TO product_specification_template_translations;
    END IF;

    IF to_regclass('public.quick_buy_step_product_types') IS NOT NULL
       AND to_regclass('public.quick_buy_step_product_specification_templates') IS NULL THEN
        ALTER TABLE quick_buy_step_product_types RENAME TO quick_buy_step_product_specification_templates;
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('public.product_spec_definitions') IS NOT NULL
       AND EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'product_spec_definitions'
             AND column_name = 'product_type_id'
       )
       AND NOT EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'product_spec_definitions'
             AND column_name = 'product_specification_template_id'
       ) THEN
        ALTER TABLE product_spec_definitions
            RENAME COLUMN product_type_id TO product_specification_template_id;
    END IF;

    IF to_regclass('public.products') IS NOT NULL
       AND EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'products'
             AND column_name = 'product_type_id'
       )
       AND NOT EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'products'
             AND column_name = 'product_specification_template_id'
       ) THEN
        ALTER TABLE products
            RENAME COLUMN product_type_id TO product_specification_template_id;
    END IF;

    IF to_regclass('public.product_specification_template_translations') IS NOT NULL
       AND EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'product_specification_template_translations'
             AND column_name = 'product_type_id'
       )
       AND NOT EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'product_specification_template_translations'
             AND column_name = 'product_specification_template_id'
       ) THEN
        ALTER TABLE product_specification_template_translations
            RENAME COLUMN product_type_id TO product_specification_template_id;
    END IF;

    IF to_regclass('public.quick_buy_step_product_specification_templates') IS NOT NULL
       AND EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'quick_buy_step_product_specification_templates'
             AND column_name = 'product_type_id'
       )
       AND NOT EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'quick_buy_step_product_specification_templates'
             AND column_name = 'product_specification_template_id'
       ) THEN
        ALTER TABLE quick_buy_step_product_specification_templates
            RENAME COLUMN product_type_id TO product_specification_template_id;
    END IF;
END $$;

ALTER INDEX IF EXISTS idx_product_type_spec_slug
    RENAME TO idx_product_specification_template_spec_slug;
ALTER INDEX IF EXISTS idx_product_type_translations_type_locale
    RENAME TO idx_product_specification_template_translations_type_locale;
ALTER INDEX IF EXISTS idx_product_type_translations_locale
    RENAME TO idx_product_specification_template_translations_locale;
ALTER INDEX IF EXISTS idx_product_types_image_media_asset_id
    RENAME TO idx_product_specification_templates_image_media_asset_id;
ALTER INDEX IF EXISTS idx_product_types_system_managed
    RENAME TO idx_product_specification_templates_system_managed;
ALTER INDEX IF EXISTS idx_products_product_type_id
    RENAME TO idx_products_product_specification_template_id;
ALTER INDEX IF EXISTS idx_quick_buy_step_product_types_unique
    RENAME TO idx_quick_buy_step_product_specification_templates_unique;
ALTER INDEX IF EXISTS idx_quick_buy_step_product_types_order
    RENAME TO idx_quick_buy_step_product_specification_templates_order;
ALTER INDEX IF EXISTS idx_quick_buy_step_product_types_one_primary
    RENAME TO idx_quick_buy_step_product_specification_templates_one_primary;

DO $$
BEGIN
    IF to_regclass('public.product_specification_templates') IS NOT NULL THEN
        IF EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conrelid = 'public.product_specification_templates'::regclass
              AND conname = 'fk_product_types_image_media_asset'
        ) THEN
            ALTER TABLE product_specification_templates
                RENAME CONSTRAINT fk_product_types_image_media_asset
                TO fk_product_specification_templates_image_media_asset;
        END IF;
        IF EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conrelid = 'public.product_specification_templates'::regclass
              AND conname = 'ck_product_types_image_reference_pair'
        ) THEN
            ALTER TABLE product_specification_templates
                RENAME CONSTRAINT ck_product_types_image_reference_pair
                TO ck_product_specification_templates_image_reference_pair;
        END IF;
    END IF;

    IF to_regclass('public.products') IS NOT NULL THEN
        IF EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conrelid = 'public.products'::regclass
              AND conname = 'fk_products_product_type'
        ) THEN
            ALTER TABLE products
                RENAME CONSTRAINT fk_products_product_type
                TO fk_products_product_specification_template;
        END IF;
    END IF;
END $$;
