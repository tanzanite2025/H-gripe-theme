CREATE TABLE IF NOT EXISTS schwalbe_tire_specifications (
    article_no VARCHAR(32) PRIMARY KEY,
    ean VARCHAR(32),
    model_name VARCHAR(180) NOT NULL,
    etrto VARCHAR(32) NOT NULL,
    inch_designation VARCHAR(32),
    weight_g NUMERIC(7,2),
    version_label VARCHAR(120),
    compound VARCHAR(120),
    color VARCHAR(120),
    bead VARCHAR(64),
    e_bike_rating VARCHAR(64),
    epi INTEGER,
    load_kg NUMERIC(7,2),
    seal VARCHAR(120),
    tread VARCHAR(120),
    min_pressure_bar NUMERIC(5,2),
    max_pressure_bar NUMERIC(5,2),
    min_pressure_psi NUMERIC(6,2),
    max_pressure_psi NUMERIC(6,2),
    source_url TEXT NOT NULL,
    source_checked_at DATE NOT NULL,
    verification_status VARCHAR(16) NOT NULL DEFAULT 'pending',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_schwalbe_tire_verification_status
        CHECK (verification_status IN ('pending', 'verified', 'rejected')),
    CONSTRAINT chk_schwalbe_tire_pressure_bar_order
        CHECK (
            min_pressure_bar IS NULL
            OR max_pressure_bar IS NULL
            OR min_pressure_bar <= max_pressure_bar
        ),
    CONSTRAINT chk_schwalbe_tire_pressure_psi_order
        CHECK (
            min_pressure_psi IS NULL
            OR max_pressure_psi IS NULL
            OR min_pressure_psi <= max_pressure_psi
        ),
    CONSTRAINT chk_schwalbe_tire_non_negative_measurements
        CHECK (
            (weight_g IS NULL OR weight_g >= 0)
            AND (load_kg IS NULL OR load_kg >= 0)
            AND (epi IS NULL OR epi >= 0)
        )
);

CREATE INDEX IF NOT EXISTS idx_schwalbe_tire_specs_etrto
    ON schwalbe_tire_specifications (etrto);

CREATE INDEX IF NOT EXISTS idx_schwalbe_tire_specs_inch
    ON schwalbe_tire_specifications (inch_designation);

CREATE INDEX IF NOT EXISTS idx_schwalbe_tire_specs_verified
    ON schwalbe_tire_specifications (verification_status, is_active);
