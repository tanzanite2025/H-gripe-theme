ALTER TABLE spoke_build_presets
    ADD COLUMN IF NOT EXISTS actual_front_left_length_mm DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS actual_front_right_length_mm DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS actual_rear_left_length_mm DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS actual_rear_right_length_mm DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS actual_length_notes TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_spoke_build_presets_actual_front_left_length'
    ) THEN
        ALTER TABLE spoke_build_presets
            ADD CONSTRAINT ck_spoke_build_presets_actual_front_left_length
            CHECK (actual_front_left_length_mm IS NULL OR actual_front_left_length_mm > 0 AND actual_front_left_length_mm <= 500);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_spoke_build_presets_actual_front_right_length'
    ) THEN
        ALTER TABLE spoke_build_presets
            ADD CONSTRAINT ck_spoke_build_presets_actual_front_right_length
            CHECK (actual_front_right_length_mm IS NULL OR actual_front_right_length_mm > 0 AND actual_front_right_length_mm <= 500);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_spoke_build_presets_actual_rear_left_length'
    ) THEN
        ALTER TABLE spoke_build_presets
            ADD CONSTRAINT ck_spoke_build_presets_actual_rear_left_length
            CHECK (actual_rear_left_length_mm IS NULL OR actual_rear_left_length_mm > 0 AND actual_rear_left_length_mm <= 500);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_spoke_build_presets_actual_rear_right_length'
    ) THEN
        ALTER TABLE spoke_build_presets
            ADD CONSTRAINT ck_spoke_build_presets_actual_rear_right_length
            CHECK (actual_rear_right_length_mm IS NULL OR actual_rear_right_length_mm > 0 AND actual_rear_right_length_mm <= 500);
    END IF;
END $$;
