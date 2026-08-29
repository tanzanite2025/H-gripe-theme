ALTER TABLE fitment_hub_specifications
    DROP CONSTRAINT IF EXISTS fitment_hub_specifications_spoke_geometry_complete_check;

ALTER TABLE fitment_hub_specifications
    DROP COLUMN IF EXISTS wr_mm,
    DROP COLUMN IF EXISTS wl_mm,
    DROP COLUMN IF EXISTS pcdr_mm,
    DROP COLUMN IF EXISTS pcdl_mm;
