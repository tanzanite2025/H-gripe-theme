-- Gift-card and loyalty redemption amounts are currency minor units, not
-- cents. Rename the physical columns so zero-decimal currencies are handled
-- correctly and the schema matches the Money domain vocabulary.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'gift_cards' AND column_name = 'initial_value_cents') THEN
        ALTER TABLE gift_cards RENAME COLUMN initial_value_cents TO initial_value_minor;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'gift_cards' AND column_name = 'balance_cents') THEN
        ALTER TABLE gift_cards RENAME COLUMN balance_cents TO balance_minor;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'gift_card_transactions' AND column_name = 'amount_cents') THEN
        ALTER TABLE gift_card_transactions RENAME COLUMN amount_cents TO amount_minor;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'gift_card_transactions' AND column_name = 'balance_cents') THEN
        ALTER TABLE gift_card_transactions RENAME COLUMN balance_cents TO balance_minor;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'gift_card_redemptions' AND column_name = 'gift_card_value_cents') THEN
        ALTER TABLE gift_card_redemptions RENAME COLUMN gift_card_value_cents TO gift_card_value_minor;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'loyalty_program_configs' AND column_name = 'max_value_per_day_cents') THEN
        ALTER TABLE loyalty_program_configs RENAME COLUMN max_value_per_day_cents TO max_value_per_day_minor;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'loyalty_program_redeem_options' AND column_name = 'value_cents') THEN
        ALTER TABLE loyalty_program_redeem_options RENAME COLUMN value_cents TO value_minor;
    END IF;
END $$;

DO $$
BEGIN
    -- Constraint names are scoped to their table, not globally. Qualify the
    -- relation in every lookup so a same-named constraint on another table
    -- cannot make this migration target the wrong object.
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'gift_cards'::regclass
          AND conname = 'initial_value_cents_non_negative'
    ) AND NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'gift_cards'::regclass
          AND conname = 'initial_value_minor_non_negative'
    ) THEN
        ALTER TABLE gift_cards RENAME CONSTRAINT initial_value_cents_non_negative TO initial_value_minor_non_negative;
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'gift_cards'::regclass
          AND conname = 'balance_cents_non_negative'
    ) AND NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'gift_cards'::regclass
          AND conname = 'balance_minor_non_negative'
    ) THEN
        ALTER TABLE gift_cards RENAME CONSTRAINT balance_cents_non_negative TO balance_minor_non_negative;
    END IF;
END $$;

ALTER TABLE gift_card_transactions
    DROP CONSTRAINT IF EXISTS balance_cents_non_negative;

ALTER TABLE loyalty_program_redeem_options
    DROP CONSTRAINT IF EXISTS loyalty_program_redeem_options_positive_value;

ALTER TABLE gift_card_redemptions
    DROP CONSTRAINT IF EXISTS gift_card_redemptions_positive_value;

ALTER TABLE loyalty_program_redeem_options
    ADD CONSTRAINT loyalty_program_redeem_options_positive_value
        CHECK (value_minor > 0);

ALTER TABLE gift_card_redemptions
    ADD CONSTRAINT gift_card_redemptions_positive_value
        CHECK (gift_card_value_minor > 0);
