ALTER TABLE shipping_yanwen_published_channels
    DROP COLUMN IF EXISTS require_receiver_tax_number,
    DROP COLUMN IF EXISTS require_ioss,
    DROP COLUMN IF EXISTS require_eori;
