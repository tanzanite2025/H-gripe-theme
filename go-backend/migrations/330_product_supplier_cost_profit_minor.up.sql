-- Supplier-cost and profitability snapshots are persisted only as exact
-- minor-unit integers.  Legacy NUMERIC major-unit columns are removed before
-- launch so there is one financial source of truth.
ALTER TABLE product_procurement_records
    ADD COLUMN IF NOT EXISTS purchase_price_minor BIGINT,
    ADD COLUMN IF NOT EXISTS inbound_shipping_unit_cost_minor BIGINT,
    ADD COLUMN IF NOT EXISTS packaging_unit_cost_minor BIGINT,
    ADD COLUMN IF NOT EXISTS other_unit_cost_minor BIGINT;

UPDATE product_procurement_records
SET purchase_price_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(purchase_price, 0))::BIGINT
        ELSE ROUND(COALESCE(purchase_price, 0) * 100)::BIGINT
    END,
    inbound_shipping_unit_cost_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(inbound_shipping_unit_cost, 0))::BIGINT
        ELSE ROUND(COALESCE(inbound_shipping_unit_cost, 0) * 100)::BIGINT
    END,
    packaging_unit_cost_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(packaging_unit_cost, 0))::BIGINT
        ELSE ROUND(COALESCE(packaging_unit_cost, 0) * 100)::BIGINT
    END,
    other_unit_cost_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(other_unit_cost, 0))::BIGINT
        ELSE ROUND(COALESCE(other_unit_cost, 0) * 100)::BIGINT
    END
WHERE purchase_price_minor IS NULL
   OR inbound_shipping_unit_cost_minor IS NULL
   OR packaging_unit_cost_minor IS NULL
   OR other_unit_cost_minor IS NULL;

ALTER TABLE product_procurement_records
    ALTER COLUMN purchase_price_minor SET NOT NULL,
    ALTER COLUMN purchase_price_minor SET DEFAULT 0,
    ALTER COLUMN inbound_shipping_unit_cost_minor SET NOT NULL,
    ALTER COLUMN inbound_shipping_unit_cost_minor SET DEFAULT 0,
    ALTER COLUMN packaging_unit_cost_minor SET NOT NULL,
    ALTER COLUMN packaging_unit_cost_minor SET DEFAULT 0,
    ALTER COLUMN other_unit_cost_minor SET NOT NULL,
    ALTER COLUMN other_unit_cost_minor SET DEFAULT 0,
    DROP COLUMN IF EXISTS purchase_price,
    DROP COLUMN IF EXISTS inbound_shipping_unit_cost,
    DROP COLUMN IF EXISTS packaging_unit_cost,
    DROP COLUMN IF EXISTS other_unit_cost;

ALTER TABLE product_profit_calculations
    ADD COLUMN IF NOT EXISTS list_price_minor BIGINT,
    ADD COLUMN IF NOT EXISTS sale_price_minor BIGINT,
    ADD COLUMN IF NOT EXISTS effective_selling_price_minor BIGINT,
    ADD COLUMN IF NOT EXISTS purchase_price_minor BIGINT,
    ADD COLUMN IF NOT EXISTS inbound_shipping_unit_cost_minor BIGINT,
    ADD COLUMN IF NOT EXISTS packaging_unit_cost_minor BIGINT,
    ADD COLUMN IF NOT EXISTS other_unit_cost_minor BIGINT,
    ADD COLUMN IF NOT EXISTS landed_cost_minor BIGINT,
    ADD COLUMN IF NOT EXISTS gross_profit_minor BIGINT;

UPDATE product_profit_calculations
SET list_price_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(list_price, 0))::BIGINT
        ELSE ROUND(COALESCE(list_price, 0) * 100)::BIGINT
    END,
    sale_price_minor = CASE
        WHEN sale_price IS NULL THEN NULL
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(sale_price)::BIGINT
        ELSE ROUND(sale_price * 100)::BIGINT
    END,
    effective_selling_price_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(effective_selling_price, 0))::BIGINT
        ELSE ROUND(COALESCE(effective_selling_price, 0) * 100)::BIGINT
    END,
    purchase_price_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(purchase_price, 0))::BIGINT
        ELSE ROUND(COALESCE(purchase_price, 0) * 100)::BIGINT
    END,
    inbound_shipping_unit_cost_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(inbound_shipping_unit_cost, 0))::BIGINT
        ELSE ROUND(COALESCE(inbound_shipping_unit_cost, 0) * 100)::BIGINT
    END,
    packaging_unit_cost_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(packaging_unit_cost, 0))::BIGINT
        ELSE ROUND(COALESCE(packaging_unit_cost, 0) * 100)::BIGINT
    END,
    other_unit_cost_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(other_unit_cost, 0))::BIGINT
        ELSE ROUND(COALESCE(other_unit_cost, 0) * 100)::BIGINT
    END,
    landed_cost_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(landed_cost, 0))::BIGINT
        ELSE ROUND(COALESCE(landed_cost, 0) * 100)::BIGINT
    END,
    gross_profit_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY','KRW','CLP')
            THEN ROUND(COALESCE(gross_profit, 0))::BIGINT
        ELSE ROUND(COALESCE(gross_profit, 0) * 100)::BIGINT
    END
WHERE list_price_minor IS NULL
   OR (sale_price IS NOT NULL AND sale_price_minor IS NULL)
   OR effective_selling_price_minor IS NULL
   OR purchase_price_minor IS NULL
   OR inbound_shipping_unit_cost_minor IS NULL
   OR packaging_unit_cost_minor IS NULL
   OR other_unit_cost_minor IS NULL
   OR landed_cost_minor IS NULL
   OR gross_profit_minor IS NULL;

ALTER TABLE product_profit_calculations
    ALTER COLUMN list_price_minor SET NOT NULL,
    ALTER COLUMN list_price_minor SET DEFAULT 0,
    ALTER COLUMN effective_selling_price_minor SET NOT NULL,
    ALTER COLUMN effective_selling_price_minor SET DEFAULT 0,
    ALTER COLUMN purchase_price_minor SET NOT NULL,
    ALTER COLUMN purchase_price_minor SET DEFAULT 0,
    ALTER COLUMN inbound_shipping_unit_cost_minor SET NOT NULL,
    ALTER COLUMN inbound_shipping_unit_cost_minor SET DEFAULT 0,
    ALTER COLUMN packaging_unit_cost_minor SET NOT NULL,
    ALTER COLUMN packaging_unit_cost_minor SET DEFAULT 0,
    ALTER COLUMN other_unit_cost_minor SET NOT NULL,
    ALTER COLUMN other_unit_cost_minor SET DEFAULT 0,
    ALTER COLUMN landed_cost_minor SET NOT NULL,
    ALTER COLUMN landed_cost_minor SET DEFAULT 0,
    ALTER COLUMN gross_profit_minor SET NOT NULL,
    ALTER COLUMN gross_profit_minor SET DEFAULT 0,
    DROP COLUMN IF EXISTS list_price,
    DROP COLUMN IF EXISTS sale_price,
    DROP COLUMN IF EXISTS effective_selling_price,
    DROP COLUMN IF EXISTS purchase_price,
    DROP COLUMN IF EXISTS inbound_shipping_unit_cost,
    DROP COLUMN IF EXISTS packaging_unit_cost,
    DROP COLUMN IF EXISTS other_unit_cost,
    DROP COLUMN IF EXISTS landed_cost,
    DROP COLUMN IF EXISTS gross_profit;
