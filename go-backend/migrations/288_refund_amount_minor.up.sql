ALTER TABLE refunds
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3),
    ADD COLUMN IF NOT EXISTS amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS gift_card_refund_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS requested_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS discount_clawback_amount_minor BIGINT;

UPDATE refunds r
SET currency = UPPER(COALESCE(NULLIF(r.currency, ''), t.currency, o.currency, 'USD'))
FROM transactions t
LEFT JOIN orders o ON o.id = r.order_id
WHERE t.id = r.transaction_id AND (r.currency IS NULL OR r.currency = '');

UPDATE refunds r
SET currency = UPPER(COALESCE(NULLIF(r.currency, ''), o.currency, 'USD'))
FROM orders o
WHERE o.id = r.order_id AND (r.currency IS NULL OR r.currency = '');

UPDATE refunds
SET amount_minor = CASE WHEN UPPER(currency) IN ('JPY', 'KRW', 'CLP') THEN ROUND(amount)::BIGINT ELSE ROUND(amount * 100)::BIGINT END,
    gift_card_refund_amount_minor = CASE WHEN UPPER(currency) IN ('JPY', 'KRW', 'CLP') THEN ROUND(gift_card_refund_amount)::BIGINT ELSE ROUND(gift_card_refund_amount * 100)::BIGINT END,
    requested_amount_minor = CASE WHEN UPPER(currency) IN ('JPY', 'KRW', 'CLP') THEN ROUND(requested_amount)::BIGINT ELSE ROUND(requested_amount * 100)::BIGINT END,
    discount_clawback_amount_minor = CASE WHEN UPPER(currency) IN ('JPY', 'KRW', 'CLP') THEN ROUND(discount_clawback_amount)::BIGINT ELSE ROUND(discount_clawback_amount * 100)::BIGINT END
WHERE amount_minor IS NULL OR gift_card_refund_amount_minor IS NULL OR requested_amount_minor IS NULL OR discount_clawback_amount_minor IS NULL;

ALTER TABLE refunds
    ALTER COLUMN currency SET NOT NULL,
    ALTER COLUMN amount_minor SET NOT NULL,
    ALTER COLUMN gift_card_refund_amount_minor SET NOT NULL,
    ALTER COLUMN requested_amount_minor SET NOT NULL,
    ALTER COLUMN discount_clawback_amount_minor SET NOT NULL,
    ADD CONSTRAINT chk_refunds_amount_minor_non_negative CHECK (amount_minor >= 0),
    ADD CONSTRAINT chk_refunds_gift_card_amount_minor_non_negative CHECK (gift_card_refund_amount_minor >= 0),
    ADD CONSTRAINT chk_refunds_requested_amount_minor_non_negative CHECK (requested_amount_minor >= 0),
    ADD CONSTRAINT chk_refunds_discount_clawback_minor_non_negative CHECK (discount_clawback_amount_minor >= 0);
