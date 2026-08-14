ALTER TABLE product_types
    ADD COLUMN IF NOT EXISTS is_system_managed BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE product_types
SET is_system_managed = TRUE
WHERE slug IN (
    'rim',
    'carbon_frame',
    'wheelset',
    'handlebar',
    'hub',
    'spoke'
);

CREATE INDEX IF NOT EXISTS idx_product_types_system_managed
    ON product_types (is_system_managed, sort_order, id);
