ALTER TABLE shipment_records
    ADD COLUMN IF NOT EXISTS tracking_shipment_id BIGINT,
    ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(120);
