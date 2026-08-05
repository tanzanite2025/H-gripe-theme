DROP INDEX IF EXISTS idx_loyalty_transactions_order_earn_once;

ALTER TABLE loyalty_program_configs
    DROP CONSTRAINT IF EXISTS loyalty_program_configs_purchase_earn_non_negative;

ALTER TABLE loyalty_program_configs
    DROP COLUMN IF EXISTS purchase_earn_points_per_currency_unit;
