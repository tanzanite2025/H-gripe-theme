ALTER TABLE spoke_build_presets
    DROP CONSTRAINT IF EXISTS ck_spoke_build_presets_actual_front_left_length,
    DROP CONSTRAINT IF EXISTS ck_spoke_build_presets_actual_front_right_length,
    DROP CONSTRAINT IF EXISTS ck_spoke_build_presets_actual_rear_left_length,
    DROP CONSTRAINT IF EXISTS ck_spoke_build_presets_actual_rear_right_length,
    DROP COLUMN IF EXISTS actual_front_left_length_mm,
    DROP COLUMN IF EXISTS actual_front_right_length_mm,
    DROP COLUMN IF EXISTS actual_rear_left_length_mm,
    DROP COLUMN IF EXISTS actual_rear_right_length_mm,
    DROP COLUMN IF EXISTS actual_length_notes;
