ALTER TABLE products
  ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

ALTER TABLE product_variants
  ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

ALTER TABLE cart_items
  ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

UPDATE product_variants
SET currency = COALESCE(NULLIF(UPPER(TRIM(currency)), ''), 'USD');

UPDATE products p
SET currency = COALESCE(
  (
    SELECT pv.currency
    FROM product_variants pv
    WHERE pv.product_id = p.id
      AND pv.deleted_at IS NULL
    ORDER BY pv.is_default DESC, pv.id ASC
    LIMIT 1
  ),
  NULLIF(UPPER(TRIM(p.currency)), ''),
  'USD'
);

UPDATE cart_items ci
SET currency = COALESCE(
  (
    SELECT pv.currency
    FROM product_variants pv
    WHERE pv.id = ci.variant_id
  ),
  NULLIF(UPPER(TRIM(ci.currency)), ''),
  'USD'
);

CREATE INDEX IF NOT EXISTS idx_products_currency ON products(currency);
CREATE INDEX IF NOT EXISTS idx_product_variants_currency ON product_variants(currency);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_products_currency_iso'
  ) THEN
    ALTER TABLE products
      ADD CONSTRAINT chk_products_currency_iso CHECK (currency ~ '^[A-Z]{3}$');
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_product_variants_currency_iso'
  ) THEN
    ALTER TABLE product_variants
      ADD CONSTRAINT chk_product_variants_currency_iso CHECK (currency ~ '^[A-Z]{3}$');
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_cart_items_currency_iso'
  ) THEN
    ALTER TABLE cart_items
      ADD CONSTRAINT chk_cart_items_currency_iso CHECK (currency ~ '^[A-Z]{3}$');
  END IF;
END $$;

INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES (
  'currency_display_currencies',
  '',
  'string',
  'en',
  'currency',
  true,
  '后台明确添加的次展示币种，用于缓存汇率和前台价格标签',
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT (key, locale) DO UPDATE
SET type = EXCLUDED.type,
    "group" = EXCLUDED."group",
    is_public = EXCLUDED.is_public,
    description = EXCLUDED.description,
    updated_at = CURRENT_TIMESTAMP;

DELETE FROM settings
WHERE "group" = 'currency'
  AND key IN (
    'currency_accounting_currency',
    'currency_default_order_currency',
    'currency_accepted_currencies',
    'currency_default_checkout_currency',
    'currency_checkout_currencies'
  );
