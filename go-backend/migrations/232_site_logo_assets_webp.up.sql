DO $$
DECLARE
    constraint_name text;
BEGIN
    FOR constraint_name IN
        SELECT tc.constraint_name
        FROM information_schema.table_constraints tc
        JOIN information_schema.check_constraints cc
          ON tc.constraint_catalog = cc.constraint_catalog
         AND tc.constraint_schema = cc.constraint_schema
         AND tc.constraint_name = cc.constraint_name
        WHERE tc.table_schema = current_schema()
          AND tc.table_name = 'site_logo_assets'
          AND tc.constraint_type = 'CHECK'
          AND (
              cc.check_clause ILIKE '%width%'
              OR cc.check_clause ILIKE '%height%'
          )
    LOOP
        EXECUTE format('ALTER TABLE site_logo_assets DROP CONSTRAINT IF EXISTS %I', constraint_name);
    END LOOP;
END $$;

ALTER TABLE site_logo_assets
    ALTER COLUMN mime_type SET DEFAULT 'image/webp',
    ALTER COLUMN width SET DEFAULT 512,
    ALTER COLUMN height SET DEFAULT 512;
