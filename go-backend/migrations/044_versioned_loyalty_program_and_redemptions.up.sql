-- Versioned loyalty configuration and lossless gift-card redemption history.

CREATE TABLE IF NOT EXISTS loyalty_program_configs (
    id BIGSERIAL PRIMARY KEY,
    version INTEGER NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate_points INTEGER NOT NULL,
    min_redeem_points INTEGER NOT NULL,
    max_value_per_day_cents BIGINT NOT NULL,
    card_expiry_days INTEGER NOT NULL,
    referral_referrer_points INTEGER NOT NULL,
    referral_referee_points INTEGER NOT NULL,
    checkin_base_points INTEGER NOT NULL,
    checkin_streak_interval_days INTEGER NOT NULL,
    checkin_streak_bonus_points INTEGER NOT NULL,
    checkin_max_points INTEGER NOT NULL,
    created_by BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_loyalty_program_configs_active
    ON loyalty_program_configs(status)
    WHERE status = 'active';

CREATE TABLE IF NOT EXISTS loyalty_program_redeem_options (
    id BIGSERIAL PRIMARY KEY,
    config_id BIGINT NOT NULL REFERENCES loyalty_program_configs(id) ON DELETE CASCADE,
    value_cents BIGINT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT loyalty_program_redeem_options_positive_value CHECK (value_cents > 0),
    CONSTRAINT loyalty_program_redeem_options_unique_value UNIQUE (config_id, value_cents)
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'loyalty_program_redeem_options_positive_value'
    ) THEN
        ALTER TABLE loyalty_program_redeem_options
            ADD CONSTRAINT loyalty_program_redeem_options_positive_value CHECK (value_cents > 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'loyalty_program_redeem_options_unique_value'
    ) THEN
        ALTER TABLE loyalty_program_redeem_options
            ADD CONSTRAINT loyalty_program_redeem_options_unique_value UNIQUE (config_id, value_cents);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'loyalty_program_redeem_options'::regclass
          AND contype = 'f'
          AND pg_get_constraintdef(oid) LIKE 'FOREIGN KEY (config_id)%REFERENCES loyalty_program_configs%'
    ) THEN
        ALTER TABLE loyalty_program_redeem_options
            ADD CONSTRAINT loyalty_program_redeem_options_config_fk
            FOREIGN KEY (config_id) REFERENCES loyalty_program_configs(id)
            ON DELETE CASCADE;
    END IF;
END $$;

ALTER TABLE gift_cards
    ADD COLUMN IF NOT EXISTS owner_user_id BIGINT,
    ADD COLUMN IF NOT EXISTS origin VARCHAR(32) NOT NULL DEFAULT 'admin';

CREATE INDEX IF NOT EXISTS idx_gift_cards_owner_user_id
    ON gift_cards(owner_user_id);

ALTER TABLE gift_card_transactions
    ADD COLUMN IF NOT EXISTS redemption_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_gift_card_transactions_redemption_id
    ON gift_card_transactions(redemption_id);

ALTER TABLE loyalty_transactions
    ADD COLUMN IF NOT EXISTS program_config_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_loyalty_transactions_program_config_id
    ON loyalty_transactions(program_config_id);

ALTER TABLE referrals
    ADD COLUMN IF NOT EXISTS referrer_points INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS referred_points INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS completed_order_id BIGINT;

ALTER TABLE check_ins
    ADD COLUMN IF NOT EXISTS is_canonical BOOLEAN NOT NULL DEFAULT TRUE;

WITH ranked_check_ins AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY user_id, check_in_date
            ORDER BY created_at ASC, id ASC
        ) AS row_number
    FROM check_ins
)
UPDATE check_ins
SET is_canonical = FALSE
FROM ranked_check_ins ranked
WHERE check_ins.id = ranked.id
  AND ranked.row_number > 1;

DROP INDEX IF EXISTS idx_check_ins_user_date;

CREATE UNIQUE INDEX IF NOT EXISTS idx_check_ins_user_date_canonical
    ON check_ins(user_id, check_in_date)
    WHERE is_canonical = TRUE;

DO $$
DECLARE
    config_id BIGINT;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM loyalty_program_configs) THEN
        INSERT INTO loyalty_program_configs (
            version,
            status,
            enabled,
            currency,
            exchange_rate_points,
            min_redeem_points,
            max_value_per_day_cents,
            card_expiry_days,
            referral_referrer_points,
            referral_referee_points,
            checkin_base_points,
            checkin_streak_interval_days,
            checkin_streak_bonus_points,
            checkin_max_points
        )
        VALUES (
            1,
            'active',
            COALESCE((
                SELECT value IN ('1', 'true')
                FROM settings
                WHERE key = 'tz_redeem_enabled' AND locale = 'en'
                LIMIT 1
            ), TRUE),
            'USD',
            COALESCE((
                SELECT value::INTEGER
                FROM settings
                WHERE key = 'tz_redeem_exchange_rate' AND locale = 'en'
                LIMIT 1
            ), 100),
            COALESCE((
                SELECT value::INTEGER
                FROM settings
                WHERE key = 'tz_redeem_min_points' AND locale = 'en'
                LIMIT 1
            ), 1000),
            ROUND(COALESCE((
                SELECT value::NUMERIC
                FROM settings
                WHERE key = 'tz_redeem_max_value_per_day' AND locale = 'en'
                LIMIT 1
            ), 500) * 100)::BIGINT,
            COALESCE((
                SELECT value::INTEGER
                FROM settings
                WHERE key = 'tz_redeem_card_expiry_days' AND locale = 'en'
                LIMIT 1
            ), 365),
            COALESCE((
                SELECT value::INTEGER
                FROM settings
                WHERE key = 'tz_loyalty_referral_referrer_points' AND locale = 'en'
                LIMIT 1
            ), 100),
            COALESCE((
                SELECT value::INTEGER
                FROM settings
                WHERE key = 'tz_loyalty_referral_referee_points' AND locale = 'en'
                LIMIT 1
            ), 50),
            COALESCE((
                SELECT value::INTEGER
                FROM settings
                WHERE key = 'tz_loyalty_checkin_base_points' AND locale = 'en'
                LIMIT 1
            ), 10),
            COALESCE((
                SELECT value::INTEGER
                FROM settings
                WHERE key = 'tz_loyalty_checkin_streak_interval_days' AND locale = 'en'
                LIMIT 1
            ), 7),
            COALESCE((
                SELECT value::INTEGER
                FROM settings
                WHERE key = 'tz_loyalty_checkin_streak_bonus_points' AND locale = 'en'
                LIMIT 1
            ), 5),
            COALESCE((
                SELECT value::INTEGER
                FROM settings
                WHERE key = 'tz_loyalty_checkin_max_points' AND locale = 'en'
                LIMIT 1
            ), 50)
        )
        RETURNING id INTO config_id;

        INSERT INTO loyalty_program_redeem_options (config_id, value_cents, sort_order)
        SELECT
            config_id,
            ROUND(TRIM(value)::NUMERIC * 100)::BIGINT,
            ordinal::INTEGER - 1
        FROM regexp_split_to_table(
            COALESCE((
                SELECT value
                FROM settings
                WHERE key = 'tz_redeem_preset_values' AND locale = 'en'
                LIMIT 1
            ), '10,50,100,200,500'),
            ','
        ) WITH ORDINALITY AS values(value, ordinal)
        WHERE TRIM(value) <> ''
          AND TRIM(value)::NUMERIC > 0
        ON CONFLICT (config_id, value_cents) DO NOTHING;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'gift_cards_owner_user_fk'
    ) THEN
        ALTER TABLE gift_cards
            ADD CONSTRAINT gift_cards_owner_user_fk
            FOREIGN KEY (owner_user_id) REFERENCES users(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS gift_card_redemptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    gift_card_id BIGINT NOT NULL UNIQUE REFERENCES gift_cards(id) ON DELETE RESTRICT,
    loyalty_transaction_id BIGINT UNIQUE REFERENCES loyalty_transactions(id) ON DELETE RESTRICT,
    program_config_id BIGINT NOT NULL REFERENCES loyalty_program_configs(id) ON DELETE RESTRICT,
    idempotency_key VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    gift_card_value_cents BIGINT NOT NULL,
    points_spent INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT gift_card_redemptions_positive_value CHECK (gift_card_value_cents > 0),
    CONSTRAINT gift_card_redemptions_positive_points CHECK (points_spent > 0),
    CONSTRAINT gift_card_redemptions_user_idempotency UNIQUE (user_id, idempotency_key)
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'gift_card_redemptions_positive_value'
    ) THEN
        ALTER TABLE gift_card_redemptions
            ADD CONSTRAINT gift_card_redemptions_positive_value CHECK (gift_card_value_cents > 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'gift_card_redemptions_positive_points'
    ) THEN
        ALTER TABLE gift_card_redemptions
            ADD CONSTRAINT gift_card_redemptions_positive_points CHECK (points_spent > 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'gift_card_redemptions_user_idempotency'
    ) THEN
        ALTER TABLE gift_card_redemptions
            ADD CONSTRAINT gift_card_redemptions_user_idempotency UNIQUE (user_id, idempotency_key);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'gift_card_redemptions'::regclass
          AND contype = 'f'
          AND pg_get_constraintdef(oid) LIKE 'FOREIGN KEY (user_id)%REFERENCES users%'
    ) THEN
        ALTER TABLE gift_card_redemptions
            ADD CONSTRAINT gift_card_redemptions_user_fk
            FOREIGN KEY (user_id) REFERENCES users(id)
            ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'gift_card_redemptions'::regclass
          AND contype = 'f'
          AND pg_get_constraintdef(oid) LIKE 'FOREIGN KEY (gift_card_id)%REFERENCES gift_cards%'
    ) THEN
        ALTER TABLE gift_card_redemptions
            ADD CONSTRAINT gift_card_redemptions_gift_card_fk
            FOREIGN KEY (gift_card_id) REFERENCES gift_cards(id)
            ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'gift_card_redemptions'::regclass
          AND contype = 'f'
          AND pg_get_constraintdef(oid) LIKE 'FOREIGN KEY (loyalty_transaction_id)%REFERENCES loyalty_transactions%'
    ) THEN
        ALTER TABLE gift_card_redemptions
            ADD CONSTRAINT gift_card_redemptions_loyalty_transaction_fk
            FOREIGN KEY (loyalty_transaction_id) REFERENCES loyalty_transactions(id)
            ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'gift_card_redemptions'::regclass
          AND contype = 'f'
          AND pg_get_constraintdef(oid) LIKE 'FOREIGN KEY (program_config_id)%REFERENCES loyalty_program_configs%'
    ) THEN
        ALTER TABLE gift_card_redemptions
            ADD CONSTRAINT gift_card_redemptions_program_config_fk
            FOREIGN KEY (program_config_id) REFERENCES loyalty_program_configs(id)
            ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'gift_card_transactions_redemption_fk'
    ) THEN
        ALTER TABLE gift_card_transactions
            ADD CONSTRAINT gift_card_transactions_redemption_fk
            FOREIGN KEY (redemption_id) REFERENCES gift_card_redemptions(id)
            ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'loyalty_transactions_program_config_fk'
    ) THEN
        ALTER TABLE loyalty_transactions
            ADD CONSTRAINT loyalty_transactions_program_config_fk
            FOREIGN KEY (program_config_id) REFERENCES loyalty_program_configs(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_gift_card_redemptions_user_id
    ON gift_card_redemptions(user_id);

CREATE INDEX IF NOT EXISTS idx_gift_card_redemptions_program_config_id
    ON gift_card_redemptions(program_config_id);

-- Preserve ownership and traceability for redemptions created before this migration.
UPDATE gift_cards gc
SET
    owner_user_id = lt.user_id,
    origin = 'loyalty_redemption'
FROM loyalty_transactions lt
WHERE lt.type = 'spend'
  AND lt.source = 'giftcard'
  AND lt.source_id = gc.id
  AND gc.owner_user_id IS NULL;

INSERT INTO gift_card_redemptions (
    user_id,
    gift_card_id,
    loyalty_transaction_id,
    program_config_id,
    idempotency_key,
    currency,
    gift_card_value_cents,
    points_spent,
    status
)
SELECT
    lt.user_id,
    gc.id,
    lt.id,
    (SELECT id FROM loyalty_program_configs WHERE status = 'active' ORDER BY version DESC LIMIT 1),
    'legacy-loyalty-transaction-' || lt.id,
    gc.currency,
    gc.initial_value_cents,
    ABS(lt.points),
    'completed'
FROM loyalty_transactions lt
JOIN gift_cards gc ON gc.id = lt.source_id
WHERE lt.type = 'spend'
  AND lt.source = 'giftcard'
  AND NOT EXISTS (
      SELECT 1
      FROM gift_card_redemptions existing
      WHERE existing.loyalty_transaction_id = lt.id
  )
ON CONFLICT DO NOTHING;

INSERT INTO gift_card_transactions (
    gift_card_id,
    redemption_id,
    type,
    amount_cents,
    balance_cents,
    note
)
SELECT
    redemption.gift_card_id,
    redemption.id,
    'issue',
    card.initial_value_cents,
    card.balance_cents,
    'Legacy loyalty redemption'
FROM gift_card_redemptions redemption
JOIN gift_cards card ON card.id = redemption.gift_card_id
WHERE NOT EXISTS (
    SELECT 1
    FROM gift_card_transactions existing
    WHERE existing.gift_card_id = redemption.gift_card_id
      AND existing.type = 'issue'
);
