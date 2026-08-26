ALTER TABLE product_procurement_records
    DROP COLUMN IF EXISTS customs_unit_cost;

ALTER TABLE product_profit_calculations
    DROP COLUMN IF EXISTS customs_unit_cost;

WITH base AS (
    SELECT
        id,
        effective_selling_price,
        ROUND(
            purchase_price
            + inbound_shipping_unit_cost
            + packaging_unit_cost
            + other_unit_cost,
            2
        ) AS landed_cost,
        COALESCE(warnings, '[]'::jsonb) AS warnings
    FROM product_profit_calculations
    WHERE effective_selling_price > 0
      AND calculation_status IN ('ready', 'warning')
),
recalculated AS (
    SELECT
        id,
        landed_cost,
        ROUND(effective_selling_price - landed_cost, 2) AS gross_profit,
        ROUND(
            (effective_selling_price - landed_cost)
            / NULLIF(effective_selling_price, 0)
            * 10000
        )::INTEGER AS gross_margin_bps,
        CASE
            WHEN effective_selling_price - landed_cost < 0
                AND NOT (warnings @> '["negative_gross_profit"]'::jsonb)
                THEN warnings || '["negative_gross_profit"]'::jsonb
            WHEN effective_selling_price - landed_cost >= 0
                THEN warnings - 'negative_gross_profit'
            ELSE warnings
        END AS warnings
    FROM base
)
UPDATE product_profit_calculations AS record
SET
    landed_cost = recalculated.landed_cost,
    gross_profit = recalculated.gross_profit,
    gross_margin_bps = recalculated.gross_margin_bps,
    calculation_status = CASE
        WHEN recalculated.warnings = '[]'::jsonb THEN 'ready'
        ELSE 'warning'
    END,
    warnings = recalculated.warnings,
    formula_version = 'gross-margin-v3-no-customs',
    calculated_at = NOW(),
    updated_at = NOW()
FROM recalculated
WHERE record.id = recalculated.id;
