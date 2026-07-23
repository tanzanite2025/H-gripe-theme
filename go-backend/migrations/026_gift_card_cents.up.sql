ALTER TABLE gift_cards
  ADD COLUMN IF NOT EXISTS initial_value_cents BIGINT,
  ADD COLUMN IF NOT EXISTS balance_cents BIGINT;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'gift_cards' AND column_name = 'initial_value'
  ) THEN
    UPDATE gift_cards
    SET initial_value_cents = ROUND(initial_value * 100)::BIGINT
    WHERE initial_value_cents IS NULL;
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'gift_cards' AND column_name = 'balance'
  ) THEN
    UPDATE gift_cards
    SET balance_cents = ROUND(balance * 100)::BIGINT
    WHERE balance_cents IS NULL;
  END IF;
END $$;

UPDATE gift_cards
SET
  initial_value_cents = COALESCE(initial_value_cents, 0),
  balance_cents = COALESCE(balance_cents, 0);

ALTER TABLE gift_cards
  ALTER COLUMN initial_value_cents SET NOT NULL,
  ALTER COLUMN balance_cents SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'initial_value_cents_non_negative'
  ) THEN
    ALTER TABLE gift_cards
      ADD CONSTRAINT initial_value_cents_non_negative CHECK (initial_value_cents >= 0);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'balance_cents_non_negative'
  ) THEN
    ALTER TABLE gift_cards
      ADD CONSTRAINT balance_cents_non_negative CHECK (balance_cents >= 0);
  END IF;
END $$;

ALTER TABLE gift_cards
  DROP COLUMN IF EXISTS initial_value,
  DROP COLUMN IF EXISTS balance;

ALTER TABLE gift_card_transactions
  ADD COLUMN IF NOT EXISTS amount_cents BIGINT,
  ADD COLUMN IF NOT EXISTS balance_cents BIGINT;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'gift_card_transactions' AND column_name = 'amount'
  ) THEN
    UPDATE gift_card_transactions
    SET amount_cents = ROUND(amount * 100)::BIGINT
    WHERE amount_cents IS NULL;
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'gift_card_transactions' AND column_name = 'balance'
  ) THEN
    UPDATE gift_card_transactions
    SET balance_cents = ROUND(balance * 100)::BIGINT
    WHERE balance_cents IS NULL;
  END IF;
END $$;

UPDATE gift_card_transactions
SET
  amount_cents = COALESCE(amount_cents, 0),
  balance_cents = COALESCE(balance_cents, 0);

ALTER TABLE gift_card_transactions
  ALTER COLUMN amount_cents SET NOT NULL,
  ALTER COLUMN balance_cents SET NOT NULL;

ALTER TABLE gift_card_transactions
  DROP COLUMN IF EXISTS amount,
  DROP COLUMN IF EXISTS balance;
