-- Shipping display prices are a storefront read model. Keep source money on
-- the transactional template/rule rows and move mutable converted JSON into
-- an independently refreshed table.
CREATE TABLE IF NOT EXISTS shipping_display_price_snapshots (
    id BIGSERIAL PRIMARY KEY,
    scope_key VARCHAR(64) NOT NULL UNIQUE,
    template_id BIGINT NOT NULL REFERENCES shipping_templates(id) ON DELETE CASCADE,
    rule_id BIGINT REFERENCES shipping_rules(id) ON DELETE CASCADE,
    source_currency VARCHAR(3) NOT NULL,
    source_default_fee_minor BIGINT,
    source_free_threshold_minor BIGINT,
    source_min_value_minor BIGINT,
    source_max_value_minor BIGINT,
    source_fee_minor BIGINT,
    source_additional_minor BIGINT,
    display_prices JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_shipping_display_price_snapshot_scope CHECK (
        (rule_id IS NULL AND scope_key = 'template:' || template_id::text)
        OR (rule_id IS NOT NULL AND scope_key = 'rule:' || rule_id::text)
    ),
    CONSTRAINT chk_shipping_display_price_snapshot_template_amounts CHECK (
        (rule_id IS NOT NULL)
        OR (source_default_fee_minor IS NOT NULL AND source_free_threshold_minor IS NOT NULL)
    ),
    CONSTRAINT chk_shipping_display_price_snapshot_rule_amounts CHECK (
        (rule_id IS NULL)
        OR (
            source_min_value_minor IS NOT NULL AND
            source_max_value_minor IS NOT NULL AND
            source_fee_minor IS NOT NULL AND
            source_additional_minor IS NOT NULL
        )
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_display_price_snapshots_template_scope
    ON shipping_display_price_snapshots(template_id)
    WHERE rule_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_display_price_snapshots_rule_scope
    ON shipping_display_price_snapshots(rule_id)
    WHERE rule_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_shipping_display_price_snapshots_template_id
    ON shipping_display_price_snapshots(template_id);

-- Preserve the existing storefront projection before removing its core-table
-- columns. Empty JSON remains an explicit empty read-model snapshot.
INSERT INTO shipping_display_price_snapshots (
    scope_key,
    template_id,
    source_currency,
    source_default_fee_minor,
    source_free_threshold_minor,
    display_prices
)
SELECT
    'template:' || id::text,
    id,
    UPPER(COALESCE(NULLIF(TRIM(currency), ''), 'USD')),
    COALESCE(default_fee_minor, 0),
    COALESCE(free_threshold_minor, 0),
    COALESCE(display_price_snapshots, '{}'::jsonb)
FROM shipping_templates
ON CONFLICT (scope_key) DO UPDATE SET
    source_currency = EXCLUDED.source_currency,
    source_default_fee_minor = EXCLUDED.source_default_fee_minor,
    source_free_threshold_minor = EXCLUDED.source_free_threshold_minor,
    display_prices = EXCLUDED.display_prices,
    updated_at = NOW();

INSERT INTO shipping_display_price_snapshots (
    scope_key,
    template_id,
    rule_id,
    source_currency,
    source_min_value_minor,
    source_max_value_minor,
    source_fee_minor,
    source_additional_minor,
    display_prices
)
SELECT
    'rule:' || r.id::text,
    r.template_id,
    r.id,
    UPPER(COALESCE(NULLIF(TRIM(r.currency), ''), NULLIF(TRIM(t.currency), ''), 'USD')),
    COALESCE(r.min_value_minor, 0),
    COALESCE(r.max_value_minor, 0),
    COALESCE(r.fee_minor, 0),
    COALESCE(r.additional_minor, 0),
    COALESCE(r.display_price_snapshots, '{}'::jsonb)
FROM shipping_rules r
INNER JOIN shipping_templates t ON t.id = r.template_id
ON CONFLICT (scope_key) DO UPDATE SET
    source_currency = EXCLUDED.source_currency,
    source_min_value_minor = EXCLUDED.source_min_value_minor,
    source_max_value_minor = EXCLUDED.source_max_value_minor,
    source_fee_minor = EXCLUDED.source_fee_minor,
    source_additional_minor = EXCLUDED.source_additional_minor,
    display_prices = EXCLUDED.display_prices,
    updated_at = NOW();

ALTER TABLE shipping_templates
    DROP COLUMN IF EXISTS display_price_snapshots;

ALTER TABLE shipping_rules
    DROP COLUMN IF EXISTS display_price_snapshots;
