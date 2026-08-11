CREATE TABLE IF NOT EXISTS spoke_rim_brands (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(80) NOT NULL,
    name VARCHAR(160) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_spoke_rim_brands_code
    ON spoke_rim_brands(code);
CREATE INDEX IF NOT EXISTS idx_spoke_rim_brands_sort_order
    ON spoke_rim_brands(sort_order);
CREATE INDEX IF NOT EXISTS idx_spoke_rim_brands_deleted_at
    ON spoke_rim_brands(deleted_at);

CREATE TABLE IF NOT EXISTS spoke_rim_models (
    id BIGSERIAL PRIMARY KEY,
    brand_id BIGINT NOT NULL REFERENCES spoke_rim_brands(id) ON DELETE CASCADE,
    code VARCHAR(120) NOT NULL,
    name VARCHAR(180) NOT NULL,
    erd_mm DOUBLE PRECISION,
    weight_g DOUBLE PRECISION,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_spoke_rim_models_erd_range CHECK (erd_mm IS NULL OR erd_mm BETWEEN 250 AND 800),
    CONSTRAINT ck_spoke_rim_models_weight_range CHECK (weight_g IS NULL OR weight_g BETWEEN 0 AND 5000)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_spoke_rim_models_code
    ON spoke_rim_models(code);
CREATE INDEX IF NOT EXISTS idx_spoke_rim_models_brand_id
    ON spoke_rim_models(brand_id);
CREATE INDEX IF NOT EXISTS idx_spoke_rim_models_sort_order
    ON spoke_rim_models(sort_order);
CREATE INDEX IF NOT EXISTS idx_spoke_rim_models_deleted_at
    ON spoke_rim_models(deleted_at);

CREATE TABLE IF NOT EXISTS spoke_hub_brands (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(80) NOT NULL,
    name VARCHAR(160) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_spoke_hub_brands_code
    ON spoke_hub_brands(code);
CREATE INDEX IF NOT EXISTS idx_spoke_hub_brands_sort_order
    ON spoke_hub_brands(sort_order);
CREATE INDEX IF NOT EXISTS idx_spoke_hub_brands_deleted_at
    ON spoke_hub_brands(deleted_at);

CREATE TABLE IF NOT EXISTS spoke_hub_models (
    id BIGSERIAL PRIMARY KEY,
    brand_id BIGINT NOT NULL REFERENCES spoke_hub_brands(id) ON DELETE CASCADE,
    code VARCHAR(120) NOT NULL,
    name VARCHAR(180) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    front_left_flange_mm DOUBLE PRECISION,
    front_right_flange_mm DOUBLE PRECISION,
    front_left_flange_pcd_mm DOUBLE PRECISION,
    front_right_flange_pcd_mm DOUBLE PRECISION,
    front_spoke_hole_diameter_mm DOUBLE PRECISION,
    rear_left_flange_mm DOUBLE PRECISION,
    rear_right_flange_mm DOUBLE PRECISION,
    rear_left_flange_pcd_mm DOUBLE PRECISION,
    rear_right_flange_pcd_mm DOUBLE PRECISION,
    rear_spoke_hole_diameter_mm DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_spoke_hub_models_front_left_flange_range CHECK (front_left_flange_mm IS NULL OR front_left_flange_mm BETWEEN 0 AND 100),
    CONSTRAINT ck_spoke_hub_models_front_right_flange_range CHECK (front_right_flange_mm IS NULL OR front_right_flange_mm BETWEEN 0 AND 100),
    CONSTRAINT ck_spoke_hub_models_front_left_pcd_range CHECK (front_left_flange_pcd_mm IS NULL OR front_left_flange_pcd_mm BETWEEN 10 AND 150),
    CONSTRAINT ck_spoke_hub_models_front_right_pcd_range CHECK (front_right_flange_pcd_mm IS NULL OR front_right_flange_pcd_mm BETWEEN 10 AND 150),
    CONSTRAINT ck_spoke_hub_models_front_hole_range CHECK (front_spoke_hole_diameter_mm IS NULL OR front_spoke_hole_diameter_mm BETWEEN 0 AND 10),
    CONSTRAINT ck_spoke_hub_models_rear_left_flange_range CHECK (rear_left_flange_mm IS NULL OR rear_left_flange_mm BETWEEN 0 AND 100),
    CONSTRAINT ck_spoke_hub_models_rear_right_flange_range CHECK (rear_right_flange_mm IS NULL OR rear_right_flange_mm BETWEEN 0 AND 100),
    CONSTRAINT ck_spoke_hub_models_rear_left_pcd_range CHECK (rear_left_flange_pcd_mm IS NULL OR rear_left_flange_pcd_mm BETWEEN 10 AND 150),
    CONSTRAINT ck_spoke_hub_models_rear_right_pcd_range CHECK (rear_right_flange_pcd_mm IS NULL OR rear_right_flange_pcd_mm BETWEEN 10 AND 150),
    CONSTRAINT ck_spoke_hub_models_rear_hole_range CHECK (rear_spoke_hole_diameter_mm IS NULL OR rear_spoke_hole_diameter_mm BETWEEN 0 AND 10)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_spoke_hub_models_code
    ON spoke_hub_models(code);
CREATE INDEX IF NOT EXISTS idx_spoke_hub_models_brand_id
    ON spoke_hub_models(brand_id);
CREATE INDEX IF NOT EXISTS idx_spoke_hub_models_sort_order
    ON spoke_hub_models(sort_order);
CREATE INDEX IF NOT EXISTS idx_spoke_hub_models_deleted_at
    ON spoke_hub_models(deleted_at);

CREATE TABLE IF NOT EXISTS spoke_build_presets (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(140) NOT NULL,
    name VARCHAR(220) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    keywords_json TEXT NOT NULL DEFAULT '[]',
    rim_brand_id BIGINT NOT NULL REFERENCES spoke_rim_brands(id) ON DELETE CASCADE,
    rim_model_id BIGINT NOT NULL REFERENCES spoke_rim_models(id) ON DELETE CASCADE,
    hub_brand_id BIGINT NOT NULL REFERENCES spoke_hub_brands(id) ON DELETE CASCADE,
    hub_model_id BIGINT NOT NULL REFERENCES spoke_hub_models(id) ON DELETE CASCADE,
    wheel_position VARCHAR(16) NOT NULL DEFAULT 'auto',
    spoke_count INTEGER NOT NULL,
    crossing INTEGER NOT NULL DEFAULT 0,
    nipple_type VARCHAR(20) NOT NULL DEFAULT 'standard',
    nipple_length_mm DOUBLE PRECISION,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_spoke_build_presets_wheel_position CHECK (wheel_position IN ('auto', 'front', 'rear')),
    CONSTRAINT ck_spoke_build_presets_spoke_count CHECK (spoke_count IN (16, 18, 20, 24, 28, 32, 36)),
    CONSTRAINT ck_spoke_build_presets_crossing CHECK (crossing IN (0, 1, 2, 3, 4)),
    CONSTRAINT ck_spoke_build_presets_nipple_type CHECK (nipple_type IN ('standard', 'hidden')),
    CONSTRAINT ck_spoke_build_presets_nipple_length CHECK (nipple_length_mm IS NULL OR nipple_length_mm BETWEEN 0 AND 40)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_spoke_build_presets_code
    ON spoke_build_presets(code);
CREATE INDEX IF NOT EXISTS idx_spoke_build_presets_rim_brand_id
    ON spoke_build_presets(rim_brand_id);
CREATE INDEX IF NOT EXISTS idx_spoke_build_presets_rim_model_id
    ON spoke_build_presets(rim_model_id);
CREATE INDEX IF NOT EXISTS idx_spoke_build_presets_hub_brand_id
    ON spoke_build_presets(hub_brand_id);
CREATE INDEX IF NOT EXISTS idx_spoke_build_presets_hub_model_id
    ON spoke_build_presets(hub_model_id);
CREATE INDEX IF NOT EXISTS idx_spoke_build_presets_sort_order
    ON spoke_build_presets(sort_order);
CREATE INDEX IF NOT EXISTS idx_spoke_build_presets_deleted_at
    ON spoke_build_presets(deleted_at);

INSERT INTO spoke_rim_brands (code, name, sort_order)
VALUES
    ('dt_swiss', 'DT Swiss', 0),
    ('mavic', 'Mavic', 1),
    ('kinlin', 'Kinlin', 2)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

INSERT INTO spoke_rim_models (brand_id, code, name, erd_mm, sort_order)
VALUES
    ((SELECT id FROM spoke_rim_brands WHERE code = 'dt_swiss'), 'rr411_db', 'RR 411 db', 598, 0),
    ((SELECT id FROM spoke_rim_brands WHERE code = 'dt_swiss'), 'rr511_db', 'RR 511 db', 581, 1),
    ((SELECT id FROM spoke_rim_brands WHERE code = 'dt_swiss'), 'rr421_db', 'RR 421 db', 594, 2),
    ((SELECT id FROM spoke_rim_brands WHERE code = 'dt_swiss'), 'r460_db', 'R 460 db', 596, 3),
    ((SELECT id FROM spoke_rim_brands WHERE code = 'dt_swiss'), 'gr531_db', 'GR 531 db', 597, 4),
    ((SELECT id FROM spoke_rim_brands WHERE code = 'dt_swiss'), 'g540_db', 'G 540 db', 592, 5),
    ((SELECT id FROM spoke_rim_brands WHERE code = 'mavic'), 'open_pro_ust_disc', 'Open Pro UST Disc', 589, 0),
    ((SELECT id FROM spoke_rim_brands WHERE code = 'mavic'), 'open_pro_ust', 'Open Pro UST', 589, 1),
    ((SELECT id FROM spoke_rim_brands WHERE code = 'mavic'), 'a_1028', 'A 1028', 614, 2),
    ((SELECT id FROM spoke_rim_brands WHERE code = 'kinlin'), 'xr26t', 'XR-26T', 592, 0),
    ((SELECT id FROM spoke_rim_brands WHERE code = 'kinlin'), 'xr31t', 'XR-31T', 580, 1)
ON CONFLICT (code) DO UPDATE SET
    brand_id = EXCLUDED.brand_id,
    name = EXCLUDED.name,
    erd_mm = EXCLUDED.erd_mm,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

INSERT INTO spoke_hub_brands (code, name, sort_order)
VALUES
    ('dt_swiss', 'DT Swiss', 0),
    ('shimano', 'Shimano', 1),
    ('novatec', 'Novatec', 2)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

INSERT INTO spoke_hub_models (
    brand_id,
    code,
    name,
    front_left_flange_mm,
    front_right_flange_mm,
    front_left_flange_pcd_mm,
    front_right_flange_pcd_mm,
    rear_left_flange_mm,
    rear_right_flange_mm,
    rear_left_flange_pcd_mm,
    rear_right_flange_pcd_mm,
    sort_order
)
VALUES
    ((SELECT id FROM spoke_hub_brands WHERE code = 'dt_swiss'), '180_road_db_cl', '180 Road db CL', 22.5, 35.6, 44, 42, 33, 20.2, 46, 46, 0),
    ((SELECT id FROM spoke_hub_brands WHERE code = 'dt_swiss'), '240_road_db_cl', '240 EXP Road db CL', 22.5, 35.6, 44, 42, 33, 20.2, 46, 46, 1),
    ((SELECT id FROM spoke_hub_brands WHERE code = 'dt_swiss'), '350_road_db_cl', '350 Road db CL', 22.5, 35.6, 44, 42, 33, 20.2, 46, 46, 2),
    ((SELECT id FROM spoke_hub_brands WHERE code = 'dt_swiss'), '350_classic_db_is', '350 Classic db IS (6-bolt)', 22.5, 35.6, 58, 45, 35.5, 21.2, 58, 52, 3),
    ((SELECT id FROM spoke_hub_brands WHERE code = 'shimano'), 'hb_r7070', '105 HB-R7070', 22, 35.6, 44, 44, 36.5, 21.6, 45, 45, 0),
    ((SELECT id FROM spoke_hub_brands WHERE code = 'novatec'), 'd791sb_d792sb', 'D791SB / D792SB', 27, 32, 58, 45, 35, 21, 58, 49, 0)
ON CONFLICT (code) DO UPDATE SET
    brand_id = EXCLUDED.brand_id,
    name = EXCLUDED.name,
    front_left_flange_mm = EXCLUDED.front_left_flange_mm,
    front_right_flange_mm = EXCLUDED.front_right_flange_mm,
    front_left_flange_pcd_mm = EXCLUDED.front_left_flange_pcd_mm,
    front_right_flange_pcd_mm = EXCLUDED.front_right_flange_pcd_mm,
    rear_left_flange_mm = EXCLUDED.rear_left_flange_mm,
    rear_right_flange_mm = EXCLUDED.rear_right_flange_mm,
    rear_left_flange_pcd_mm = EXCLUDED.rear_left_flange_pcd_mm,
    rear_right_flange_pcd_mm = EXCLUDED.rear_right_flange_pcd_mm,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

INSERT INTO spoke_build_presets (
    code,
    name,
    description,
    keywords_json,
    rim_brand_id,
    rim_model_id,
    hub_brand_id,
    hub_model_id,
    wheel_position,
    spoke_count,
    crossing,
    nipple_type,
    nipple_length_mm,
    sort_order
)
VALUES
    (
        'tz_ar45_dt350_fr',
        'AR 45 Disc + DT Swiss 350',
        'Popular all-rounder build. Reliable ratchet hub with aero rim.',
        '["350","240","dt swiss","45mm","ar45","disc","road"]',
        (SELECT id FROM spoke_rim_brands WHERE code = 'dt_swiss'),
        (SELECT id FROM spoke_rim_models WHERE code = 'rr411_db'),
        (SELECT id FROM spoke_hub_brands WHERE code = 'dt_swiss'),
        (SELECT id FROM spoke_hub_models WHERE code = '350_road_db_cl'),
        'auto',
        24,
        2,
        'standard',
        14,
        0
    ),
    (
        'tz_ar50_dt240_fr',
        'AR 50 Disc + DT Swiss 240 EXP',
        'Lightweight racing build. Top-tier hub performance.',
        '["240","dt swiss","50mm","ar50","exp","racing"]',
        (SELECT id FROM spoke_rim_brands WHERE code = 'dt_swiss'),
        (SELECT id FROM spoke_rim_models WHERE code = 'rr511_db'),
        (SELECT id FROM spoke_hub_brands WHERE code = 'dt_swiss'),
        (SELECT id FROM spoke_hub_models WHERE code = '240_road_db_cl'),
        'auto',
        24,
        2,
        'hidden',
        12,
        1
    ),
    (
        'mavic_open_dt350',
        'Mavic Open Pro UST + DT Swiss 350',
        'Classic training wheelset. Bombproof reliability.',
        '["mavic","open pro","350","training"]',
        (SELECT id FROM spoke_rim_brands WHERE code = 'mavic'),
        (SELECT id FROM spoke_rim_models WHERE code = 'open_pro_ust_disc'),
        (SELECT id FROM spoke_hub_brands WHERE code = 'dt_swiss'),
        (SELECT id FROM spoke_hub_models WHERE code = '350_road_db_cl'),
        'auto',
        28,
        3,
        'standard',
        12,
        2
    )
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    keywords_json = EXCLUDED.keywords_json,
    rim_brand_id = EXCLUDED.rim_brand_id,
    rim_model_id = EXCLUDED.rim_model_id,
    hub_brand_id = EXCLUDED.hub_brand_id,
    hub_model_id = EXCLUDED.hub_model_id,
    wheel_position = EXCLUDED.wheel_position,
    spoke_count = EXCLUDED.spoke_count,
    crossing = EXCLUDED.crossing,
    nipple_type = EXCLUDED.nipple_type,
    nipple_length_mm = EXCLUDED.nipple_length_mm,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();
