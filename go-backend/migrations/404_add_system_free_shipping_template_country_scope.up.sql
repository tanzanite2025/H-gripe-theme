-- Keep the built-in free-shipping policy on the generic shipping template
-- aggregate. It does not reference 4PX, Yanwen, or any carrier service.
ALTER TABLE shipping_templates
    ADD COLUMN IF NOT EXISTS template_kind VARCHAR(40) NOT NULL DEFAULT 'carrier',
    ADD COLUMN IF NOT EXISTS is_system_managed BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS free_shipping_countries TEXT NOT NULL DEFAULT '[]';

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_templates_single_system_free_shipping
    ON shipping_templates (template_kind)
    WHERE template_kind = 'system_free_shipping'
      AND is_system_managed = TRUE
      AND deleted_at IS NULL;

-- The empty country list is intentionally safe: the row is available in the
-- admin template tab, but cannot make a product orderable until an operator
-- explicitly selects its supported countries.
INSERT INTO shipping_templates (
    name,
    type,
    currency,
    free_shipping,
    free_threshold_minor,
    default_fee_minor,
    description,
    enabled,
    template_kind,
    is_system_managed,
    free_shipping_countries
)
SELECT
    '系统免邮模板',
    'free_shipping',
    'USD',
    TRUE,
    0,
    0,
    '系统级免邮策略；仅对明确选择的国家开放下单，不绑定任何承运商线路。',
    TRUE,
    'system_free_shipping',
    TRUE,
    '[]'
WHERE NOT EXISTS (
    SELECT 1
    FROM shipping_templates
    WHERE template_kind = 'system_free_shipping'
      AND is_system_managed = TRUE
      AND deleted_at IS NULL
);
