CREATE TABLE IF NOT EXISTS fitment_hub_specifications (
    id BIGSERIAL PRIMARY KEY,
    spec_code VARCHAR(80) NOT NULL,
    display_name VARCHAR(160) NOT NULL,
    position VARCHAR(16) NOT NULL,
    axle_type VARCHAR(32) NOT NULL,
    axle_spacing_mm INTEGER NOT NULL,
    notes TEXT,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fitment_hub_specifications_position_check CHECK (position IN ('front', 'rear')),
    CONSTRAINT fitment_hub_specifications_axle_type_check CHECK (
        axle_type IN ('quick_release', 'thru_axle', 'bolt_on', 'other')
    ),
    CONSTRAINT fitment_hub_specifications_spacing_check CHECK (axle_spacing_mm > 0)
);

CREATE INDEX IF NOT EXISTS idx_fitment_hub_specifications_position
    ON fitment_hub_specifications(position);
CREATE INDEX IF NOT EXISTS idx_fitment_hub_specifications_axle_type
    ON fitment_hub_specifications(axle_type);
CREATE INDEX IF NOT EXISTS idx_fitment_hub_specifications_enabled
    ON fitment_hub_specifications(is_enabled);
CREATE INDEX IF NOT EXISTS idx_fitment_hub_specifications_deleted
    ON fitment_hub_specifications(deleted_at);

CREATE UNIQUE INDEX IF NOT EXISTS uk_fitment_hub_specifications_code
    ON fitment_hub_specifications(LOWER(BTRIM(spec_code)))
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS fitment_frame_hub_specifications (
    frame_entry_id BIGINT NOT NULL
        REFERENCES fitment_frame_entries(id) ON DELETE CASCADE,
    hub_specification_id BIGINT NOT NULL
        REFERENCES fitment_hub_specifications(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (frame_entry_id, hub_specification_id)
);

CREATE INDEX IF NOT EXISTS idx_fitment_frame_hub_specifications_hub
    ON fitment_frame_hub_specifications(hub_specification_id);
