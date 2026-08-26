ALTER TABLE shipment_records
    ADD COLUMN IF NOT EXISTS product_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS details_bound BOOLEAN NOT NULL DEFAULT FALSE;

-- Rows created by the previous auto-sync implementation may already contain
-- operator-entered evidence. Preserve those attachments as bound while rows
-- that only mirror shipment facts remain optional/unbound.
UPDATE shipment_records
SET details_bound = TRUE
WHERE details_bound = FALSE
  AND (
      COALESCE(NULLIF(TRIM(shipping_note), ''), '') <> ''
      OR shipping_images <> '[]'::jsonb
      OR product_codes <> '[]'::jsonb
      OR warranty_months <> 12
  );
