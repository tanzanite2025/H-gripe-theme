ALTER TABLE after_sales_refund_reviews
    ADD COLUMN IF NOT EXISTS proposed_amount_minor BIGINT;

UPDATE after_sales_refund_reviews
SET proposed_amount_minor = CASE
    WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
        THEN ROUND(COALESCE(proposed_amount, 0))::BIGINT
    ELSE ROUND(COALESCE(proposed_amount, 0) * 100)::BIGINT
END
WHERE proposed_amount_minor IS NULL;

ALTER TABLE after_sales_refund_reviews
    ALTER COLUMN proposed_amount_minor SET NOT NULL,
    ALTER COLUMN proposed_amount_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_after_sales_refund_reviews_proposed_amount_minor_non_negative
        CHECK (proposed_amount_minor >= 0),
    DROP COLUMN IF EXISTS proposed_amount;
