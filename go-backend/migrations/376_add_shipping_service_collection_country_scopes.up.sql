ALTER TABLE shipping_fpx_channels
    ADD COLUMN IF NOT EXISTS countries TEXT NOT NULL DEFAULT '[]';

ALTER TABLE shipping_yanwen_published_channels
    ADD COLUMN IF NOT EXISTS countries TEXT NOT NULL DEFAULT '[]';
