ALTER TABLE shipping_yanwen_published_channels
    DROP COLUMN IF EXISTS countries;

ALTER TABLE shipping_fpx_channels
    DROP COLUMN IF EXISTS countries;
