-- Stage three: persist template candidate values independently from the
-- legacy SpecDefinition.options JSON field.

CREATE TABLE IF NOT EXISTS product_spec_option_items (
    id BIGSERIAL PRIMARY KEY,
    spec_definition_id BIGINT NOT NULL
        REFERENCES product_spec_definitions(id) ON DELETE CASCADE,
    value_key VARCHAR(160) NOT NULL,
    default_label VARCHAR(160) NOT NULL,
    color_hex VARCHAR(20),
    swatch_media_asset_id BIGINT,
    swatch_url VARCHAR(800),
    is_enabled_by_default BOOLEAN NOT NULL DEFAULT TRUE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    default_price_delta_minor BIGINT,
    default_price_currency CHAR(3),
    sort_order INTEGER NOT NULL DEFAULT 0,
    revision INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_product_spec_option_item_key UNIQUE (spec_definition_id, value_key),
    CONSTRAINT ck_product_spec_option_item_price_pair CHECK (
        (default_price_delta_minor IS NULL AND default_price_currency IS NULL)
        OR (default_price_delta_minor IS NOT NULL AND default_price_currency IS NOT NULL)
    ),
    CONSTRAINT ck_product_spec_option_item_price_non_negative CHECK (
        default_price_delta_minor IS NULL OR default_price_delta_minor >= 0
    )
);

-- Materialize the pre-stage JSON candidates once. The JSON column remains only
-- as a migration source; runtime reads use product_spec_option_items.
INSERT INTO product_spec_option_items (
    spec_definition_id,
    value_key,
    default_label,
    is_enabled_by_default,
    sort_order,
    revision
)
SELECT definition.id,
       candidate.value_key,
       candidate.value_key,
       TRUE,
       candidate.sort_order,
       1
  FROM product_spec_definitions AS definition
 CROSS JOIN LATERAL (
       SELECT value.value_key,
              value.ordinality::INTEGER * 10 AS sort_order
         FROM jsonb_array_elements_text(
                CASE
                    WHEN btrim(COALESCE(definition.options, '')) ~ '^\\s*\\[.*\\]\\s*$'
                        THEN definition.options::jsonb
                    ELSE '[]'::jsonb
                END
              ) WITH ORDINALITY AS value(value_key, ordinality)
       ) AS candidate
 WHERE NOT EXISTS (
       SELECT 1
         FROM product_spec_option_items AS existing
        WHERE existing.spec_definition_id = definition.id
          AND existing.value_key = candidate.value_key
       )
ON CONFLICT (spec_definition_id, value_key) DO NOTHING;

CREATE UNIQUE INDEX IF NOT EXISTS uq_product_spec_option_item_default
    ON product_spec_option_items(spec_definition_id)
    WHERE is_default = TRUE;

CREATE INDEX IF NOT EXISTS idx_product_spec_option_items_definition_sort
    ON product_spec_option_items(spec_definition_id, sort_order, id);
