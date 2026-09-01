ALTER TABLE tracking_events
    DROP COLUMN IF EXISTS proof_of_delivery_url,
    DROP COLUMN IF EXISTS recipient_signature_name;
