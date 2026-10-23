-- Migration 001 and the public API use short_description / meta_description.
-- Earlier GORM writers could create duplicate abbreviated columns in development
-- databases. Preserve exact content and reject conflicting pairs atomically.
DO $$
DECLARE
    description_columns RECORD;
    has_canonical_column BOOLEAN;
    has_legacy_column BOOLEAN;
    conflicting_products BIGINT;
BEGIN
    LOCK TABLE products IN ACCESS EXCLUSIVE MODE;
    FOR description_columns IN
        SELECT * FROM (VALUES
            ('short_description', 'short_desc'),
            ('meta_description', 'meta_desc')
        ) AS description_column_pairs(canonical_name, legacy_name)
    LOOP
        SELECT EXISTS (
            SELECT 1 FROM pg_attribute
            WHERE attrelid = 'products'::regclass AND attname = description_columns.canonical_name
              AND attnum > 0 AND NOT attisdropped
        ), EXISTS (
            SELECT 1 FROM pg_attribute
            WHERE attrelid = 'products'::regclass AND attname = description_columns.legacy_name
              AND attnum > 0 AND NOT attisdropped
        ) INTO has_canonical_column, has_legacy_column;

        IF NOT has_canonical_column AND NOT has_legacy_column THEN
            RAISE EXCEPTION 'products has no recognized % column', description_columns.canonical_name;
        END IF;
        IF EXISTS (
            SELECT 1 FROM pg_attribute
            WHERE attrelid = 'products'::regclass
              AND attname IN (description_columns.canonical_name, description_columns.legacy_name)
              AND attnum > 0 AND NOT attisdropped
              AND atttypid NOT IN ('text'::regtype, 'character varying'::regtype)
        ) THEN
            RAISE EXCEPTION 'product % columns must have a recognized text type', description_columns.canonical_name;
        END IF;
        IF NOT has_legacy_column THEN
            CONTINUE;
        END IF;
        IF NOT has_canonical_column THEN
            EXECUTE format('ALTER TABLE products RENAME COLUMN %I TO %I', description_columns.legacy_name, description_columns.canonical_name);
            CONTINUE;
        END IF;

        EXECUTE format(
            'SELECT COUNT(*) FROM products WHERE COALESCE(%I, '''') <> '''' AND COALESCE(%I, '''') <> '''' AND %I IS DISTINCT FROM %I',
            description_columns.canonical_name, description_columns.legacy_name,
            description_columns.canonical_name, description_columns.legacy_name
        ) INTO conflicting_products;
        IF conflicting_products > 0 THEN
            RAISE EXCEPTION 'product % columns contain conflicting content', description_columns.canonical_name
                USING DETAIL = format('%s product rows require explicit content reconciliation before migration', conflicting_products);
        END IF;

        EXECUTE format(
            'UPDATE products SET %I = %I WHERE COALESCE(%I, '''') = '''' AND COALESCE(%I, '''') <> ''''',
            description_columns.canonical_name, description_columns.legacy_name,
            description_columns.canonical_name, description_columns.legacy_name
        );
        -- RESTRICT is intentional: dependent views/extensions must block cleanup.
        EXECUTE format('ALTER TABLE products DROP COLUMN %I RESTRICT', description_columns.legacy_name);
    END LOOP;
END $$;
