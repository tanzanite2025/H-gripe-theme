-- Payment, refund, dispute, and risk records are exact minor-unit snapshots.
-- Legacy major-unit columns are removed before launch so no workflow can
-- accidentally write a second monetary representation.
ALTER TABLE paypal_disputes
    ADD COLUMN IF NOT EXISTS amount_minor BIGINT;

UPDATE paypal_disputes
SET amount_minor = CASE
    WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
        THEN ROUND(COALESCE(amount, 0))::BIGINT
    ELSE ROUND(COALESCE(amount, 0) * 100)::BIGINT
END
WHERE amount_minor IS NULL;

ALTER TABLE paypal_disputes
    ALTER COLUMN amount_minor SET NOT NULL,
    ALTER COLUMN amount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_paypal_disputes_amount_minor_non_negative
        CHECK (amount_minor >= 0),
    DROP COLUMN IF EXISTS amount;

ALTER TABLE payment_refund_recommendations
    ADD COLUMN IF NOT EXISTS recommended_amount_minor BIGINT;

UPDATE payment_refund_recommendations
SET recommended_amount_minor = CASE
    WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
        THEN ROUND(COALESCE(recommended_amount, 0))::BIGINT
    ELSE ROUND(COALESCE(recommended_amount, 0) * 100)::BIGINT
END
WHERE recommended_amount_minor IS NULL;

ALTER TABLE payment_refund_recommendations
    ALTER COLUMN recommended_amount_minor SET NOT NULL,
    ALTER COLUMN recommended_amount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_payment_refund_recommendations_amount_minor_non_negative
        CHECK (recommended_amount_minor >= 0),
    DROP COLUMN IF EXISTS recommended_amount;

ALTER TABLE refunds
    DROP COLUMN IF EXISTS amount,
    DROP COLUMN IF EXISTS gift_card_refund_amount,
    DROP COLUMN IF EXISTS requested_amount,
    DROP COLUMN IF EXISTS discount_clawback_amount,
    DROP COLUMN IF EXISTS loyalty_cash_deduction_amount;

ALTER TABLE refund_line_items
    DROP COLUMN IF EXISTS unit_price,
    DROP COLUMN IF EXISTS line_subtotal_amount,
    DROP COLUMN IF EXISTS line_tax_amount,
    DROP COLUMN IF EXISTS line_discount_amount,
    DROP COLUMN IF EXISTS line_total_amount;

ALTER TABLE payment_refund_executions
    DROP COLUMN IF EXISTS amount;

ALTER TABLE stripe_disputes
    DROP COLUMN IF EXISTS amount;

ALTER TABLE payment_risk_events
    DROP COLUMN IF EXISTS amount;

ALTER TABLE payment_risk_checkout_decisions
    DROP COLUMN IF EXISTS amount;

ALTER TABLE payment_methods
    ADD COLUMN IF NOT EXISTS min_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS max_amount_minor BIGINT;

UPDATE payment_methods
-- Legacy thresholds have no currency. Because the project has not launched,
-- reset them rather than manufacturing a two-decimal currency assumption.
-- Operations must configure any non-zero thresholds before release.
SET min_amount_minor = 0,
    max_amount_minor = 0
WHERE min_amount_minor IS NULL OR max_amount_minor IS NULL;

ALTER TABLE payment_methods
    ALTER COLUMN min_amount_minor SET NOT NULL,
    ALTER COLUMN min_amount_minor SET DEFAULT 0,
    ALTER COLUMN max_amount_minor SET NOT NULL,
    ALTER COLUMN max_amount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_payment_methods_min_amount_minor_non_negative CHECK (min_amount_minor >= 0),
    ADD CONSTRAINT chk_payment_methods_max_amount_minor_non_negative CHECK (max_amount_minor >= 0),
    DROP COLUMN IF EXISTS fee_value,
    DROP COLUMN IF EXISTS min_amount,
    DROP COLUMN IF EXISTS max_amount;
