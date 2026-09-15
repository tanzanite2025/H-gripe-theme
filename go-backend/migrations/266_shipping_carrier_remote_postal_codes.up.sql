ALTER TABLE shipping_carrier_services
    ADD COLUMN IF NOT EXISTS remote_postal_codes TEXT NOT NULL DEFAULT '[]';
