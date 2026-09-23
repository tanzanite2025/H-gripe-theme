ALTER TABLE shipping_templates
    ADD COLUMN IF NOT EXISTS free_threshold_minor BIGINT,
    ADD COLUMN IF NOT EXISTS default_fee_minor BIGINT;

UPDATE shipping_templates
SET free_threshold_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
        THEN ROUND(COALESCE(free_threshold, 0))::BIGINT
        ELSE ROUND(COALESCE(free_threshold, 0) * 100)::BIGINT END,
    default_fee_minor = CASE WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
        THEN ROUND(COALESCE(default_fee, 0))::BIGINT
        ELSE ROUND(COALESCE(default_fee, 0) * 100)::BIGINT END
WHERE free_threshold_minor IS NULL OR default_fee_minor IS NULL;

ALTER TABLE shipping_templates
    ALTER COLUMN free_threshold_minor SET NOT NULL,
    ALTER COLUMN free_threshold_minor SET DEFAULT 0,
    ALTER COLUMN default_fee_minor SET NOT NULL,
    ALTER COLUMN default_fee_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_shipping_templates_money_minor_non_negative
        CHECK (free_threshold_minor >= 0 AND default_fee_minor >= 0);

ALTER TABLE shipping_rules
    ADD COLUMN IF NOT EXISTS min_value_minor BIGINT,
    ADD COLUMN IF NOT EXISTS max_value_minor BIGINT,
    ADD COLUMN IF NOT EXISTS fee_minor BIGINT,
    ADD COLUMN IF NOT EXISTS additional_minor BIGINT;

UPDATE shipping_rules r
SET min_value_minor = CASE WHEN LOWER(COALESCE(t.type, '')) IN ('price','amount')
        THEN CASE WHEN UPPER(COALESCE(r.currency, t.currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(r.min_value, 0))::BIGINT
            ELSE ROUND(COALESCE(r.min_value, 0) * 100)::BIGINT END
        ELSE 0 END,
    max_value_minor = CASE WHEN LOWER(COALESCE(t.type, '')) IN ('price','amount')
        THEN CASE WHEN UPPER(COALESCE(r.currency, t.currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(r.max_value, 0))::BIGINT
            ELSE ROUND(COALESCE(r.max_value, 0) * 100)::BIGINT END
        ELSE 0 END,
    fee_minor = CASE WHEN UPPER(COALESCE(r.currency, t.currency, 'USD')) IN ('JPY','KRW','CLP')
        THEN ROUND(COALESCE(r.fee, 0))::BIGINT
        ELSE ROUND(COALESCE(r.fee, 0) * 100)::BIGINT END,
    additional_minor = CASE WHEN UPPER(COALESCE(r.currency, t.currency, 'USD')) IN ('JPY','KRW','CLP')
        THEN ROUND(COALESCE(r.additional, 0))::BIGINT
        ELSE ROUND(COALESCE(r.additional, 0) * 100)::BIGINT END
FROM shipping_templates t
WHERE t.id = r.template_id
  AND (r.min_value_minor IS NULL OR r.max_value_minor IS NULL OR r.fee_minor IS NULL OR r.additional_minor IS NULL);

ALTER TABLE shipping_rules
    ALTER COLUMN min_value_minor SET NOT NULL,
    ALTER COLUMN min_value_minor SET DEFAULT 0,
    ALTER COLUMN max_value_minor SET NOT NULL,
    ALTER COLUMN max_value_minor SET DEFAULT 0,
    ALTER COLUMN fee_minor SET NOT NULL,
    ALTER COLUMN fee_minor SET DEFAULT 0,
    ALTER COLUMN additional_minor SET NOT NULL,
    ALTER COLUMN additional_minor SET DEFAULT 0,
    ADD CONSTRAINT chk_shipping_rules_money_minor_non_negative
        CHECK (min_value_minor >= 0 AND max_value_minor >= 0 AND fee_minor >= 0 AND additional_minor >= 0);

