-- Warranty next-stage facts:
-- 1) order_item_id links a warranty claim to the exact purchased line item.
-- 2) warranty_service_records stores service history independently from processing notes.
--
-- Do not store order-item bindings or service history as free text in warranty_claims.resolution.

ALTER TABLE warranty_claims
  ADD COLUMN IF NOT EXISTS order_item_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_warranty_claims_order_item_id
  ON warranty_claims (order_item_id);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_warranty_claims_order_item'
  ) THEN
    ALTER TABLE warranty_claims
      ADD CONSTRAINT fk_warranty_claims_order_item
      FOREIGN KEY (order_item_id)
      REFERENCES order_items(id)
      ON UPDATE CASCADE
      ON DELETE SET NULL;
  END IF;
END
$$;

CREATE TABLE IF NOT EXISTS warranty_service_records (
  id BIGSERIAL PRIMARY KEY,
  claim_id BIGINT NOT NULL REFERENCES warranty_claims(id) ON UPDATE CASCADE ON DELETE CASCADE,
  registration_id BIGINT REFERENCES product_registrations(id) ON UPDATE CASCADE ON DELETE SET NULL,
  service_type VARCHAR(80) NOT NULL DEFAULT 'inspection',
  status VARCHAR(50) NOT NULL DEFAULT 'open',
  summary TEXT NOT NULL DEFAULT '',
  cost_amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
  currency VARCHAR(8) NOT NULL,
  performed_by BIGINT REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL,
  created_by BIGINT REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL,
  performed_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_warranty_service_records_claim_id
  ON warranty_service_records (claim_id);

CREATE INDEX IF NOT EXISTS idx_warranty_service_records_registration_id
  ON warranty_service_records (registration_id);

CREATE INDEX IF NOT EXISTS idx_warranty_service_records_status
  ON warranty_service_records (status);

CREATE INDEX IF NOT EXISTS idx_warranty_service_records_deleted_at
  ON warranty_service_records (deleted_at);
