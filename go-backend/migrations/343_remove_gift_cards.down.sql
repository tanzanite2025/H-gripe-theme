-- The retired gift-card data, redemption settings, and loyalty ledger rows are
-- intentionally not recreated. Restore only empty schema objects so rollback
-- can continue through the historical migrations that precede this retirement.

CREATE TABLE IF NOT EXISTS gift_cards (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(100) NOT NULL,
    initial_value_minor BIGINT NOT NULL,
    balance_minor BIGINT NOT NULL,
    currency VARCHAR(8) NOT NULL,
    status VARCHAR(50),
    recipient_email VARCHAR(255),
    recipient_name VARCHAR(255),
    sender_name VARCHAR(255),
    message TEXT,
    cover_image TEXT,
    expires_at TIMESTAMPTZ,
    owner_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    origin VARCHAR(32) NOT NULL DEFAULT 'admin',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT initial_value_minor_non_negative CHECK (initial_value_minor >= 0),
    CONSTRAINT balance_minor_non_negative CHECK (balance_minor >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_gift_cards_code ON gift_cards(code);
CREATE INDEX IF NOT EXISTS idx_gift_cards_status ON gift_cards(status);
CREATE INDEX IF NOT EXISTS idx_gift_cards_deleted_at ON gift_cards(deleted_at);
CREATE INDEX IF NOT EXISTS idx_gift_cards_owner_user_id ON gift_cards(owner_user_id);

ALTER TABLE loyalty_program_configs
    ADD COLUMN IF NOT EXISTS min_redeem_points INTEGER NOT NULL DEFAULT 1000,
    ADD COLUMN IF NOT EXISTS max_value_per_day_minor BIGINT NOT NULL DEFAULT 50000,
    ADD COLUMN IF NOT EXISTS card_expiry_days INTEGER NOT NULL DEFAULT 365;

CREATE TABLE IF NOT EXISTS loyalty_program_redeem_options (
    id BIGSERIAL PRIMARY KEY,
    config_id BIGINT NOT NULL REFERENCES loyalty_program_configs(id) ON DELETE CASCADE,
    value_minor BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    stock_quantity BIGINT NOT NULL DEFAULT 0,
    redeemed_quantity BIGINT NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT loyalty_program_redeem_options_positive_value CHECK (value_minor > 0),
    CONSTRAINT loyalty_program_redeem_options_unique_value UNIQUE (config_id, value_minor),
    CONSTRAINT loyalty_program_redeem_options_stock_non_negative CHECK (stock_quantity >= 0),
    CONSTRAINT loyalty_program_redeem_options_redeemed_non_negative CHECK (redeemed_quantity >= 0),
    CONSTRAINT loyalty_program_redeem_options_redeemed_within_stock CHECK (redeemed_quantity <= stock_quantity)
);

CREATE INDEX IF NOT EXISTS idx_loyalty_program_redeem_options_inventory
    ON loyalty_program_redeem_options(config_id, stock_quantity, redeemed_quantity);

CREATE TABLE IF NOT EXISTS gift_card_redemptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    gift_card_id BIGINT NOT NULL UNIQUE REFERENCES gift_cards(id) ON DELETE RESTRICT,
    loyalty_transaction_id BIGINT UNIQUE REFERENCES loyalty_transactions(id) ON DELETE RESTRICT,
    program_config_id BIGINT NOT NULL REFERENCES loyalty_program_configs(id) ON DELETE RESTRICT,
    idempotency_key VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    gift_card_value_minor BIGINT NOT NULL,
    points_spent INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT gift_card_redemptions_positive_value CHECK (gift_card_value_minor > 0),
    CONSTRAINT gift_card_redemptions_positive_points CHECK (points_spent > 0),
    CONSTRAINT gift_card_redemptions_user_idempotency UNIQUE (user_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_gift_card_redemptions_user_id
    ON gift_card_redemptions(user_id);
CREATE INDEX IF NOT EXISTS idx_gift_card_redemptions_program_config_id
    ON gift_card_redemptions(program_config_id);

CREATE TABLE IF NOT EXISTS gift_card_transactions (
    id BIGSERIAL PRIMARY KEY,
    gift_card_id BIGINT NOT NULL REFERENCES gift_cards(id) ON DELETE RESTRICT,
    order_id BIGINT,
    redemption_id BIGINT REFERENCES gift_card_redemptions(id) ON DELETE SET NULL,
    refund_id BIGINT REFERENCES refunds(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL,
    amount_minor BIGINT NOT NULL,
    balance_minor BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    note TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_gift_card_transactions_gift_card_id
    ON gift_card_transactions(gift_card_id);
CREATE INDEX IF NOT EXISTS idx_gift_card_transactions_order_id
    ON gift_card_transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_gift_card_transactions_redemption_id
    ON gift_card_transactions(redemption_id);
CREATE INDEX IF NOT EXISTS idx_gift_card_transactions_refund_id
    ON gift_card_transactions(refund_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_gift_card_transactions_refund_card
    ON gift_card_transactions(refund_id, gift_card_id)
    WHERE type = 'refund' AND refund_id IS NOT NULL;

ALTER TABLE refunds
    ADD COLUMN IF NOT EXISTS gift_card_refund_amount_minor BIGINT NOT NULL DEFAULT 0;
