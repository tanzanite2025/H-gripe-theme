-- Customs declarations are transactional monetary values in the order line
-- currency. Persist only the exact smallest-unit integer representation.
ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS declared_value_minor BIGINT;

UPDATE order_items
SET declared_value_minor = CASE
    WHEN declared_value IS NULL THEN NULL
    WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
        THEN ROUND(declared_value)::BIGINT
    ELSE ROUND(declared_value * 100)::BIGINT
END
WHERE declared_value_minor IS NULL;

ALTER TABLE order_items
    ADD CONSTRAINT chk_order_items_declared_value_minor_non_negative
        CHECK (declared_value_minor IS NULL OR declared_value_minor >= 0),
    DROP COLUMN IF EXISTS declared_value;
