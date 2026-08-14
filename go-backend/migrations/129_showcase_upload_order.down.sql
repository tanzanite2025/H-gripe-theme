ALTER TABLE showcases
    DROP CONSTRAINT IF EXISTS fk_showcases_order;

DROP INDEX IF EXISTS idx_showcases_order_id;

ALTER TABLE showcases
    DROP COLUMN IF EXISTS order_id;
