ALTER TABLE product_specification_templates
    ADD COLUMN IF NOT EXISTS is_system_managed BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE product_specification_templates
SET is_system_managed = TRUE
WHERE slug IN (
    'rim',
    'carbon_frame',
    'wheelset',
    'handlebar',
    'hub',
    'spoke'
);

CREATE INDEX IF NOT EXISTS idx_product_specification_templates_system_managed
    ON product_specification_templates (is_system_managed, sort_order, id);
