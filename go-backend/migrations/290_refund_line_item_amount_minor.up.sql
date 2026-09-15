ALTER TABLE refund_line_items
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3),
    ADD COLUMN IF NOT EXISTS unit_price_minor BIGINT,
    ADD COLUMN IF NOT EXISTS line_subtotal_minor BIGINT,
    ADD COLUMN IF NOT EXISTS line_tax_minor BIGINT,
    ADD COLUMN IF NOT EXISTS line_discount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS line_total_minor BIGINT;

UPDATE refund_line_items li
SET currency = UPPER(COALESCE(NULLIF(li.currency, ''), r.currency, o.currency, 'USD'))
FROM refunds r
LEFT JOIN orders o ON o.id = li.order_id
WHERE r.id = li.refund_id AND (li.currency IS NULL OR li.currency = '');

UPDATE refund_line_items li
SET unit_price_minor = CASE WHEN UPPER(li.currency) IN ('JPY', 'KRW', 'CLP') THEN ROUND(li.unit_price)::BIGINT ELSE ROUND(li.unit_price * 100)::BIGINT END,
    line_subtotal_minor = CASE WHEN UPPER(li.currency) IN ('JPY', 'KRW', 'CLP') THEN ROUND(li.line_subtotal_amount)::BIGINT ELSE ROUND(li.line_subtotal_amount * 100)::BIGINT END,
    line_tax_minor = CASE WHEN UPPER(li.currency) IN ('JPY', 'KRW', 'CLP') THEN ROUND(li.line_tax_amount)::BIGINT ELSE ROUND(li.line_tax_amount * 100)::BIGINT END,
    line_discount_minor = CASE WHEN UPPER(li.currency) IN ('JPY', 'KRW', 'CLP') THEN ROUND(li.line_discount_amount)::BIGINT ELSE ROUND(li.line_discount_amount * 100)::BIGINT END,
    line_total_minor = CASE WHEN UPPER(li.currency) IN ('JPY', 'KRW', 'CLP') THEN ROUND(li.line_total_amount)::BIGINT ELSE ROUND(li.line_total_amount * 100)::BIGINT END
WHERE unit_price_minor IS NULL OR line_subtotal_minor IS NULL OR line_tax_minor IS NULL OR line_discount_minor IS NULL OR line_total_minor IS NULL;

ALTER TABLE refund_line_items
    ALTER COLUMN currency SET NOT NULL,
    ALTER COLUMN unit_price_minor SET NOT NULL,
    ALTER COLUMN line_subtotal_minor SET NOT NULL,
    ALTER COLUMN line_tax_minor SET NOT NULL,
    ALTER COLUMN line_discount_minor SET NOT NULL,
    ALTER COLUMN line_total_minor SET NOT NULL,
    ADD CONSTRAINT chk_refund_line_items_unit_price_minor_non_negative CHECK (unit_price_minor >= 0),
    ADD CONSTRAINT chk_refund_line_items_subtotal_minor_non_negative CHECK (line_subtotal_minor >= 0),
    ADD CONSTRAINT chk_refund_line_items_tax_minor_non_negative CHECK (line_tax_minor >= 0),
    ADD CONSTRAINT chk_refund_line_items_discount_minor_non_negative CHECK (line_discount_minor >= 0),
    ADD CONSTRAINT chk_refund_line_items_total_minor_non_negative CHECK (line_total_minor >= 0);

CREATE INDEX IF NOT EXISTS idx_refund_line_items_currency ON refund_line_items(currency);
