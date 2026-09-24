-- Restore the historical column names only when explicitly rolling back.
ALTER TABLE gift_card_redemptions
    DROP CONSTRAINT IF EXISTS gift_card_redemptions_positive_value;
ALTER TABLE loyalty_program_redeem_options
    DROP CONSTRAINT IF EXISTS loyalty_program_redeem_options_positive_value;

-- Restore the historical gift-card constraint names before restoring the
-- historical column names. Constraint names are table-local, so qualify the
-- lookup to avoid collisions with similarly named constraints elsewhere.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'gift_cards'::regclass
          AND conname = 'initial_value_minor_non_negative'
    ) AND NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'gift_cards'::regclass
          AND conname = 'initial_value_cents_non_negative'
    ) THEN
        ALTER TABLE gift_cards RENAME CONSTRAINT initial_value_minor_non_negative TO initial_value_cents_non_negative;
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'gift_cards'::regclass
          AND conname = 'balance_minor_non_negative'
    ) AND NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'gift_cards'::regclass
          AND conname = 'balance_cents_non_negative'
    ) THEN
        ALTER TABLE gift_cards RENAME CONSTRAINT balance_minor_non_negative TO balance_cents_non_negative;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'gift_cards' AND column_name = 'initial_value_minor') THEN
        ALTER TABLE gift_cards RENAME COLUMN initial_value_minor TO initial_value_cents;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'gift_cards' AND column_name = 'balance_minor') THEN
        ALTER TABLE gift_cards RENAME COLUMN balance_minor TO balance_cents;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'gift_card_transactions' AND column_name = 'amount_minor') THEN
        ALTER TABLE gift_card_transactions RENAME COLUMN amount_minor TO amount_cents;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'gift_card_transactions' AND column_name = 'balance_minor') THEN
        ALTER TABLE gift_card_transactions RENAME COLUMN balance_minor TO balance_cents;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'gift_card_redemptions' AND column_name = 'gift_card_value_minor') THEN
        ALTER TABLE gift_card_redemptions RENAME COLUMN gift_card_value_minor TO gift_card_value_cents;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'loyalty_program_configs' AND column_name = 'max_value_per_day_minor') THEN
        ALTER TABLE loyalty_program_configs RENAME COLUMN max_value_per_day_minor TO max_value_per_day_cents;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'loyalty_program_redeem_options' AND column_name = 'value_minor') THEN
        ALTER TABLE loyalty_program_redeem_options RENAME COLUMN value_minor TO value_cents;
    END IF;
END $$;

-- The up migration recreated these checks against the minor-unit columns;
-- recreate the original expressions after restoring their cents names.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'loyalty_program_redeem_options'::regclass
          AND conname = 'loyalty_program_redeem_options_positive_value'
    ) THEN
        ALTER TABLE loyalty_program_redeem_options
            ADD CONSTRAINT loyalty_program_redeem_options_positive_value
            CHECK (value_cents > 0);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'gift_card_redemptions'::regclass
          AND conname = 'gift_card_redemptions_positive_value'
    ) THEN
        ALTER TABLE gift_card_redemptions
            ADD CONSTRAINT gift_card_redemptions_positive_value
            CHECK (gift_card_value_cents > 0);
    END IF;
END $$;
