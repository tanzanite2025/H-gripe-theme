-- Warranty claims can come from two long-term sources:
-- 1) registration-bound claims, where registration_id points at product_registrations.id
-- 2) order-verified claims, where registration linkage is intentionally absent until
--    the order line / serial-number association is implemented.
--
-- Do not store registration_id=0 as a fake fallback. Null means "not linked yet".

UPDATE warranty_claims
SET registration_id = NULL
WHERE registration_id = 0;

ALTER TABLE warranty_claims
  ALTER COLUMN registration_id DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_warranty_claims_registration_id
  ON warranty_claims (registration_id);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_warranty_claims_registration'
  ) THEN
    ALTER TABLE warranty_claims
      ADD CONSTRAINT fk_warranty_claims_registration
      FOREIGN KEY (registration_id)
      REFERENCES product_registrations(id)
      ON UPDATE CASCADE
      ON DELETE SET NULL;
  END IF;
END
$$;
