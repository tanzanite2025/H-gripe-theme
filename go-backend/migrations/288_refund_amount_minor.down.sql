ALTER TABLE refunds
    DROP CONSTRAINT IF EXISTS chk_refunds_amount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_refunds_gift_card_amount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_refunds_requested_amount_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_refunds_discount_clawback_minor_non_negative;
ALTER TABLE refunds
    DROP COLUMN IF EXISTS discount_clawback_amount_minor,
    DROP COLUMN IF EXISTS requested_amount_minor,
    DROP COLUMN IF EXISTS gift_card_refund_amount_minor,
    DROP COLUMN IF EXISTS amount_minor,
    DROP COLUMN IF EXISTS currency;
