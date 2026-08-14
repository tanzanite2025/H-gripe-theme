CREATE TABLE IF NOT EXISTS showcases (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    kind VARCHAR(20) NOT NULL DEFAULT 'user',
    title VARCHAR(255),
    region VARCHAR(100),
    location VARCHAR(100),
    nickname VARCHAR(100),
    bike_model VARCHAR(100),
    notes TEXT,
    product_refs JSON,
    images JSON,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    rejected_reason TEXT,
    approved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_showcases_user_id
    ON showcases(user_id);

CREATE INDEX IF NOT EXISTS idx_showcases_kind
    ON showcases(kind);

CREATE INDEX IF NOT EXISTS idx_showcases_status
    ON showcases(status);

CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,
    showcase_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    author VARCHAR(100),
    content TEXT NOT NULL,
    location VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_comments_showcase_id
    ON comments(showcase_id);

CREATE INDEX IF NOT EXISTS idx_comments_user_id
    ON comments(user_id);

CREATE INDEX IF NOT EXISTS idx_comments_status
    ON comments(status);

ALTER TABLE showcases
    ADD COLUMN IF NOT EXISTS order_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_showcases_order_id
    ON showcases(order_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_showcases_order'
    ) THEN
        ALTER TABLE showcases
            ADD CONSTRAINT fk_showcases_order
            FOREIGN KEY (order_id) REFERENCES orders(id)
            ON DELETE SET NULL;
    END IF;
END $$;
