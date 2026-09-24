-- The removed columns cannot be restored without a source-of-truth major
-- amount. Recreate nullable columns for manual rollback only; canonical minor
-- values remain authoritative.
ALTER TABLE shipping_templates
    ADD COLUMN IF NOT EXISTS free_threshold DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS default_fee DOUBLE PRECISION;

ALTER TABLE shipping_rules
    ADD COLUMN IF NOT EXISTS fee DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS additional DOUBLE PRECISION;
