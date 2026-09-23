-- Shipping money is persisted only in canonical minor-unit columns.  The
-- min_value/max_value columns are intentionally retained because weight and
-- quantity templates use them as dimensional thresholds (not amounts).
ALTER TABLE shipping_templates
    DROP COLUMN IF EXISTS free_threshold,
    DROP COLUMN IF EXISTS default_fee;

ALTER TABLE shipping_rules
    DROP COLUMN IF EXISTS fee,
    DROP COLUMN IF EXISTS additional;

-- Price rules use *_minor for thresholds. Clear the old major projection from
-- the shared dimensional columns; weight/quantity rules keep their values.
UPDATE shipping_rules r
SET min_value = 0,
    max_value = 0
FROM shipping_templates t
WHERE t.id = r.template_id
  AND LOWER(COALESCE(t.type, '')) IN ('price', 'amount');
