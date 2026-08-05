-- Make purchase earning an explicit loyalty-program rule.

ALTER TABLE loyalty_program_configs
    ADD COLUMN IF NOT EXISTS purchase_earn_points_per_currency_unit INTEGER NOT NULL DEFAULT 1;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'loyalty_program_configs_purchase_earn_non_negative'
    ) THEN
        ALTER TABLE loyalty_program_configs
            ADD CONSTRAINT loyalty_program_configs_purchase_earn_non_negative
            CHECK (purchase_earn_points_per_currency_unit >= 0);
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_loyalty_transactions_order_earn_once
    ON loyalty_transactions(user_id, source, source_id, type)
    WHERE type = 'earn'
      AND source = 'order'
      AND source_id > 0;
