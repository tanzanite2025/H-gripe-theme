ALTER TABLE site_logo_assets
    ALTER COLUMN mime_type SET DEFAULT 'image/svg+xml',
    ALTER COLUMN width SET DEFAULT 48,
    ALTER COLUMN height SET DEFAULT 48;

ALTER TABLE site_logo_assets
    ADD CONSTRAINT site_logo_assets_width_check CHECK (width = 48),
    ADD CONSTRAINT site_logo_assets_height_check CHECK (height = 48);
