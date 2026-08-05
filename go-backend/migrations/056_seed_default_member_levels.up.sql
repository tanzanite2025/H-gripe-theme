CREATE TABLE IF NOT EXISTS member_levels (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    min_points INTEGER NOT NULL,
    max_points INTEGER NOT NULL,
    discount_rate NUMERIC NOT NULL DEFAULT 0,
    benefits TEXT,
    icon TEXT,
    color TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

ALTER TABLE member_levels
    ADD COLUMN IF NOT EXISTS name TEXT,
    ADD COLUMN IF NOT EXISTS min_points INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS max_points INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS discount_rate NUMERIC NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS benefits TEXT,
    ADD COLUMN IF NOT EXISTS icon TEXT,
    ADD COLUMN IF NOT EXISTS color TEXT,
    ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_member_levels_deleted_at
    ON member_levels(deleted_at);

CREATE INDEX IF NOT EXISTS idx_member_levels_sort_order
    ON member_levels(sort_order, min_points, id);

INSERT INTO member_levels (
    name,
    min_points,
    max_points,
    discount_rate,
    benefits,
    icon,
    color,
    sort_order,
    created_at,
    updated_at
)
SELECT *
FROM (
    VALUES
        ('Ordinary', 0, 499, 0, '[]', 'circle', '#f8fafc', 0, NOW(), NOW()),
        ('Bronze', 500, 1999, 0, '[]', 'medal', '#b87333', 10, NOW(), NOW()),
        ('Silver', 2000, 4999, 0, '[]', 'medal', '#c0c0c0', 20, NOW(), NOW()),
        ('Gold', 5000, 9999, 0, '[]', 'medal', '#d4af37', 30, NOW(), NOW()),
        ('Platinum', 10000, 19999, 0, '[]', 'gem', '#e5e4e2', 40, NOW(), NOW()),
        ('Diamond', 20000, 999999999, 0, '[]', 'gem', '#b9f2ff', 50, NOW(), NOW())
) AS defaults (
    name,
    min_points,
    max_points,
    discount_rate,
    benefits,
    icon,
    color,
    sort_order,
    created_at,
    updated_at
)
WHERE NOT EXISTS (
    SELECT 1
    FROM member_levels
    WHERE deleted_at IS NULL
);
