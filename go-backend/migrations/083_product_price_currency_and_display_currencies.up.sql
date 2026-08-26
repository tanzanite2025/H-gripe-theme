ALTER TABLE products
  ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

ALTER TABLE product_variants
  ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

ALTER TABLE cart_items
  ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

UPDATE product_variants
SET currency = COALESCE(
  (
    SELECT NULLIF(UPPER(TRIM(value)), '')
    FROM settings
    WHERE key = 'currency_primary_currency'
      AND locale = 'en'
    LIMIT 1
  ),
  (
    SELECT NULLIF(UPPER(TRIM(value)), '')
    FROM settings
    WHERE key = 'currency_default_checkout_currency'
      AND locale = 'en'
    LIMIT 1
  ),
  NULLIF(UPPER(TRIM(currency)), ''),
  'USD'
);

UPDATE products p
SET currency = COALESCE(
  (
    SELECT NULLIF(UPPER(TRIM(pv.currency)), '')
    FROM product_variants pv
    WHERE pv.product_id = p.id
      AND pv.deleted_at IS NULL
    ORDER BY pv.is_default DESC, pv.id ASC
    LIMIT 1
  ),
  (
    SELECT NULLIF(UPPER(TRIM(value)), '')
    FROM settings
    WHERE key = 'currency_primary_currency'
      AND locale = 'en'
    LIMIT 1
  ),
  (
    SELECT NULLIF(UPPER(TRIM(value)), '')
    FROM settings
    WHERE key = 'currency_default_checkout_currency'
      AND locale = 'en'
    LIMIT 1
  ),
  NULLIF(UPPER(TRIM(p.currency)), ''),
  'USD'
);

UPDATE cart_items ci
SET currency = COALESCE(
  (
    SELECT NULLIF(UPPER(TRIM(pv.currency)), '')
    FROM product_variants pv
    WHERE pv.id = ci.variant_id
    LIMIT 1
  ),
  (
    SELECT NULLIF(UPPER(TRIM(value)), '')
    FROM settings
    WHERE key = 'currency_primary_currency'
      AND locale = 'en'
    LIMIT 1
  ),
  (
    SELECT NULLIF(UPPER(TRIM(value)), '')
    FROM settings
    WHERE key = 'currency_default_checkout_currency'
      AND locale = 'en'
    LIMIT 1
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

DELETE FROM settings
WHERE key = 'currency_display_currencies'
  AND "group" = 'currency';

DELETE FROM settings
WHERE "group" = 'currency'
  AND key IN (
    'currency_accounting_currency',
    'currency_default_order_currency',
    'currency_accepted_currencies',
    'currency_default_checkout_currency',
    'currency_checkout_currencies'
  );
