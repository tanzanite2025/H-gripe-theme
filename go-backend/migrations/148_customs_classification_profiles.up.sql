-- Repair older production databases that still carry the pre-rename product
-- type contract before this migration creates its foreign key.
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

CREATE TABLE IF NOT EXISTS customs_classification_profiles (
    id BIGSERIAL PRIMARY KEY,
    product_specification_template_id BIGINT REFERENCES product_specification_templates(id) ON DELETE SET NULL,
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(140) NOT NULL UNIQUE,
    component_kind VARCHAR(64) NOT NULL DEFAULT '',
    material VARCHAR(64) NOT NULL DEFAULT '',
    hs_code VARCHAR(12) NOT NULL,
    cn_code VARCHAR(12) NOT NULL DEFAULT '',
    country_of_origin VARCHAR(2) NOT NULL DEFAULT '',
    customs_description VARCHAR(255) NOT NULL DEFAULT '',
    source VARCHAR(32) NOT NULL DEFAULT '',
    source_code VARCHAR(64) NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_customs_classification_status CHECK (status IN ('draft', 'active', 'paused')),
    CONSTRAINT ck_customs_classification_hs_code CHECK (hs_code ~ '^[0-9]{6}$'),
    CONSTRAINT ck_customs_classification_cn_code CHECK (cn_code = '' OR cn_code ~ '^[0-9]{8}$'),
    CONSTRAINT ck_customs_classification_origin CHECK (country_of_origin = '' OR country_of_origin ~ '^[A-Z]{2}$')
);

CREATE INDEX IF NOT EXISTS idx_customs_classification_profiles_product_specification_template_id
    ON customs_classification_profiles(product_specification_template_id);

CREATE INDEX IF NOT EXISTS idx_customs_classification_profiles_status
    ON customs_classification_profiles(status);

CREATE INDEX IF NOT EXISTS idx_customs_classification_profiles_component_material
    ON customs_classification_profiles(component_kind, material);
