CREATE TABLE IF NOT EXISTS quick_buy_step_product_categories (
    id BIGSERIAL PRIMARY KEY,
    step_id BIGINT NOT NULL REFERENCES quick_buy_steps(id) ON DELETE CASCADE,
    product_category_id BIGINT NOT NULL REFERENCES product_categories(id) ON DELETE RESTRICT,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_step_product_categories_unique
    ON quick_buy_step_product_categories(step_id, product_category_id);

CREATE INDEX IF NOT EXISTS idx_quick_buy_step_product_categories_order
    ON quick_buy_step_product_categories(step_id, sort_order, id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quick_buy_step_product_categories_one_primary
    ON quick_buy_step_product_categories(step_id)
    WHERE is_primary = TRUE;

WITH wheelset_category AS (
    SELECT id
    FROM product_categories
    WHERE slug = 'wheelset'
),
quick_build_steps AS (
    SELECT
        step.id,
        step.sort_order
    FROM quick_buy_steps AS step
    JOIN quick_buy_flow_versions AS version
        ON version.id = step.flow_version_id
    JOIN quick_buy_flows AS flow
        ON flow.id = version.flow_id
    WHERE flow.slug = 'quick-build'
      AND step.step_key IN ('product-search', 'specifications', 'quantity')
)
INSERT INTO quick_buy_step_product_categories (
    step_id,
    product_category_id,
    is_primary,
    sort_order,
    created_at,
    updated_at
)
SELECT
    quick_build_steps.id,
    wheelset_category.id,
    TRUE,
    quick_build_steps.sort_order,
    NOW(),
    NOW()
FROM quick_build_steps
CROSS JOIN wheelset_category
ON CONFLICT (step_id, product_category_id) DO NOTHING;
