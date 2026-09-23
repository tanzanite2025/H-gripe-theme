ALTER TABLE shipment_records
    DROP COLUMN IF EXISTS tracking_shipment_id,
    DROP COLUMN IF EXISTS tracking_number;
