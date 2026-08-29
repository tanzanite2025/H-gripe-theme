ALTER TABLE fitment_hub_specifications
    ADD COLUMN IF NOT EXISTS wr_mm DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS wl_mm DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS pcdr_mm DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS pcdl_mm DOUBLE PRECISION;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fitment_hub_specifications_spoke_geometry_complete_check'
    ) THEN
        ALTER TABLE fitment_hub_specifications
            ADD CONSTRAINT fitment_hub_specifications_spoke_geometry_complete_check
            CHECK (
                (wr_mm IS NULL AND wl_mm IS NULL AND pcdr_mm IS NULL AND pcdl_mm IS NULL)
                OR
                (
                    wr_mm IS NOT NULL
                    AND wl_mm IS NOT NULL
                    AND pcdr_mm IS NOT NULL
                    AND pcdl_mm IS NOT NULL
                    AND wr_mm > 0 AND wr_mm <= 100
                    AND wl_mm > 0 AND wl_mm <= 100
                    AND pcdr_mm >= 10 AND pcdr_mm <= 150
                    AND pcdl_mm >= 10 AND pcdl_mm <= 150
                )
            );
    END IF;
END $$;
