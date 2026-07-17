CREATE TABLE IF NOT EXISTS browsing_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    view_count INTEGER NOT NULL DEFAULT 1,
    last_viewed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_browsing_history_user_id ON browsing_history(user_id);
CREATE INDEX IF NOT EXISTS idx_browsing_history_product_id ON browsing_history(product_id);
CREATE INDEX IF NOT EXISTS idx_browsing_history_last_viewed_at ON browsing_history(last_viewed_at);
CREATE UNIQUE INDEX IF NOT EXISTS uk_browsing_history_user_product
    ON browsing_history(user_id, product_id);
