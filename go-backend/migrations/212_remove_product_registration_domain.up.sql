ALTER TABLE warranty_service_records
    DROP COLUMN IF EXISTS registration_id;

ALTER TABLE warranty_claims
    DROP COLUMN IF EXISTS registration_id;

DROP TABLE IF EXISTS product_registrations CASCADE;
