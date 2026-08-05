-- Add explicit currency and finite inventory to loyalty gift-card options.

ALTER TABLE loyalty_program_redeem_options
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3),
    ADD COLUMN IF NOT EXISTS stock_quantity BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS redeemed_quantity BIGINT NOT NULL DEFAULT 0;

UPDATE loyalty_program_redeem_options option
SET currency = config.currency
FROM loyalty_program_configs config
WHERE option.config_id = config.id
  AND (option.currency IS NULL OR TRIM(option.currency) = '');

ALTER TABLE loyalty_program_redeem_options
    ALTER COLUMN currency SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'loyalty_program_redeem_options_stock_non_negative'
    ) THEN
        ALTER TABLE loyalty_program_redeem_options
            ADD CONSTRAINT loyalty_program_redeem_options_stock_non_negative
            CHECK (stock_quantity >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'loyalty_program_redeem_options_redeemed_non_negative'
    ) THEN
        ALTER TABLE loyalty_program_redeem_options
            ADD CONSTRAINT loyalty_program_redeem_options_redeemed_non_negative
            CHECK (redeemed_quantity >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'loyalty_program_redeem_options_redeemed_within_stock'
    ) THEN
        ALTER TABLE loyalty_program_redeem_options
            ADD CONSTRAINT loyalty_program_redeem_options_redeemed_within_stock
            CHECK (redeemed_quantity <= stock_quantity);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_loyalty_program_redeem_options_inventory
    ON loyalty_program_redeem_options(config_id, stock_quantity, redeemed_quantity);
