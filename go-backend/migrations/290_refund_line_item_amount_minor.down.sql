DROP INDEX IF EXISTS idx_refund_line_items_currency;
ALTER TABLE refund_line_items
    DROP CONSTRAINT IF EXISTS chk_refund_line_items_unit_price_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_refund_line_items_subtotal_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_refund_line_items_tax_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_refund_line_items_discount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_refund_line_items_total_minor_non_negative,
    DROP COLUMN IF EXISTS unit_price_minor,
    DROP COLUMN IF EXISTS line_subtotal_minor,
    DROP COLUMN IF EXISTS line_tax_minor,
    DROP COLUMN IF EXISTS line_discount_minor,
    DROP COLUMN IF EXISTS line_total_minor,
    DROP COLUMN IF EXISTS currency;
