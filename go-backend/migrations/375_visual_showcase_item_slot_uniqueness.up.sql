WITH ranked_active_items AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY showcase_key, locale, desktop_order
            ORDER BY id DESC
        ) AS duplicate_rank
    FROM visual_showcase_items
    WHERE deleted_at IS NULL
)
UPDATE visual_showcase_items
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE id IN (
    SELECT id
    FROM ranked_active_items
    WHERE duplicate_rank > 1
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_visual_showcase_items_active_slot
    ON visual_showcase_items (showcase_key, locale, desktop_order)
    WHERE deleted_at IS NULL;
