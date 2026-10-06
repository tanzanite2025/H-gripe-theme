DELETE FROM shipping_templates
WHERE template_kind = 'system_free_shipping'
  AND is_system_managed = TRUE;

DROP INDEX IF EXISTS idx_shipping_templates_single_system_free_shipping;

ALTER TABLE shipping_templates
    DROP COLUMN IF EXISTS free_shipping_countries,
    DROP COLUMN IF EXISTS is_system_managed,
    DROP COLUMN IF EXISTS template_kind;
