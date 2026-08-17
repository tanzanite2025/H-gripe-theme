DROP INDEX IF EXISTS idx_media_asset_derivatives_preset_version;

ALTER TABLE media_asset_derivatives
    DROP CONSTRAINT IF EXISTS ck_media_asset_derivatives_preset_version;

ALTER TABLE media_asset_derivatives
    DROP COLUMN IF EXISTS preset_version;

DROP INDEX IF EXISTS idx_media_derivative_presets_deleted_at;
DROP INDEX IF EXISTS idx_media_derivative_presets_enabled_order;
DROP INDEX IF EXISTS idx_media_derivative_presets_code;

DROP TABLE IF EXISTS media_derivative_presets;
