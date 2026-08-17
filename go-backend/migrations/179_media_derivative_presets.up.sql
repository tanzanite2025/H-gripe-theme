CREATE TABLE IF NOT EXISTS media_derivative_presets (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(40) NOT NULL,
    label VARCHAR(120) NOT NULL DEFAULT '',
    max_width INTEGER NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    generation_version INTEGER NOT NULL DEFAULT 1,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_media_derivative_presets_code
        CHECK (TRIM(code) <> ''),
    CONSTRAINT ck_media_derivative_presets_max_width
        CHECK (max_width > 0),
    CONSTRAINT ck_media_derivative_presets_generation_version
        CHECK (generation_version > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_media_derivative_presets_code
    ON media_derivative_presets(code)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_media_derivative_presets_enabled_order
    ON media_derivative_presets(enabled, sort_order, id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_media_derivative_presets_deleted_at
    ON media_derivative_presets(deleted_at);

ALTER TABLE media_asset_derivatives
    ADD COLUMN IF NOT EXISTS preset_version INTEGER NOT NULL DEFAULT 1;

UPDATE media_asset_derivatives
SET preset_version = 1
WHERE preset_version IS NULL OR preset_version < 1;

ALTER TABLE media_asset_derivatives
    ADD CONSTRAINT ck_media_asset_derivatives_preset_version
        CHECK (preset_version > 0);

CREATE INDEX IF NOT EXISTS idx_media_asset_derivatives_preset_version
    ON media_asset_derivatives(preset, preset_version);

WITH defaults(code, label, max_width, sort_order, generation_version, is_system) AS (
    VALUES
        ('thumbnail', '缩略图', 320, 10, 1, TRUE),
        ('card', '卡片图', 640, 20, 1, TRUE),
        ('large', '大图', 1600, 30, 1, TRUE)
)
INSERT INTO media_derivative_presets (
    code,
    label,
    max_width,
    sort_order,
    enabled,
    generation_version,
    is_system,
    created_at,
    updated_at
)
SELECT
    defaults.code,
    defaults.label,
    defaults.max_width,
    defaults.sort_order,
    TRUE,
    defaults.generation_version,
    defaults.is_system,
    NOW(),
    NOW()
FROM defaults
WHERE NOT EXISTS (
    SELECT 1
    FROM media_derivative_presets AS existing
    WHERE existing.code = defaults.code
      AND existing.deleted_at IS NULL
);
