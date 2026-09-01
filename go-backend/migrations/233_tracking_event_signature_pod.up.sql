ALTER TABLE tracking_events
    ADD COLUMN IF NOT EXISTS recipient_signature_name VARCHAR(160),
    ADD COLUMN IF NOT EXISTS proof_of_delivery_url TEXT;
