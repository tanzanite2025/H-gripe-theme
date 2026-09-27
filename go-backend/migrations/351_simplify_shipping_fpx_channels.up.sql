ALTER TABLE shipping_fpx_channels
    DROP COLUMN IF EXISTS max_length_cm,
    DROP COLUMN IF EXISTS max_width_cm,
    DROP COLUMN IF EXISTS max_height_cm,
    DROP COLUMN IF EXISTS volumetric_divisor,
    DROP COLUMN IF EXISTS max_weight_grams,
    DROP COLUMN IF EXISTS notes,
    DROP COLUMN IF EXISTS sort_order;

DROP INDEX IF EXISTS idx_shipping_fpx_channels_sort_order;
