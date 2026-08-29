-- Make repeated block requests idempotent at the database boundary.
--
-- Older application instances used a read-then-create flow without a
-- uniqueness guarantee. Keep the newest identical active rule and disable
-- older duplicates before creating the constraint.
WITH ranked_active_rules AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY source, source_reference, cidr
            ORDER BY created_at DESC, id DESC
        ) AS duplicate_rank
    FROM global_ip_block_rules
    WHERE deleted_at IS NULL
      AND enabled = TRUE
)
UPDATE global_ip_block_rules
SET
    enabled = FALSE,
    disabled_at = COALESCE(disabled_at, NOW()),
    updated_at = NOW()
WHERE id IN (
    SELECT id
    FROM ranked_active_rules
    WHERE duplicate_rank > 1
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_global_ip_block_rules_active_identity
    ON global_ip_block_rules (source, source_reference, cidr)
    WHERE deleted_at IS NULL AND enabled = TRUE;
