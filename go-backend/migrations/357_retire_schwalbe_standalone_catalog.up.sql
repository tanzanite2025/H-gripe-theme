DO $$
DECLARE
    has_rows BOOLEAN;
BEGIN
    IF to_regclass('public.schwalbe_tire_specifications') IS NOT NULL THEN
        EXECUTE 'SELECT EXISTS (SELECT 1 FROM public.schwalbe_tire_specifications)' INTO has_rows;
        IF has_rows THEN
            RAISE EXCEPTION 'schwalbe_tire_specifications contains rows; migrate or inspect them before retiring the standalone catalog';
        END IF;
    END IF;
END $$;

DROP TABLE IF EXISTS schwalbe_tire_specifications;

UPDATE product_specification_templates
SET description = 'Schwalbe tire facts entered directly on each product through this system template.',
    updated_at = NOW()
WHERE slug = 'schwalbe_tire';
