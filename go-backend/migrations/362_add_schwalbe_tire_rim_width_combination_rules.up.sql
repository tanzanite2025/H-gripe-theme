-- Official Schwalbe/ETRTO guidance for possible tire-width and inner-rim-width
-- combinations. These rules are independent from sales-product templates and
-- do not claim that every tire model is compatible with every listed rim.
CREATE TABLE IF NOT EXISTS schwalbe_tire_rim_width_combination_rules (
    id BIGSERIAL PRIMARY KEY,
    tire_width_min_mm INTEGER NOT NULL,
    tire_width_max_mm INTEGER NOT NULL,
    inner_rim_width_min_mm INTEGER NOT NULL,
    inner_rim_width_max_mm INTEGER NOT NULL,
    source_basis TEXT NOT NULL,
    source_version VARCHAR(64) NOT NULL,
    source_url TEXT NOT NULL,
    source_checked_at DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_schwalbe_tire_rim_width_rule_positive_ranges
        CHECK (
            tire_width_min_mm > 0
            AND tire_width_max_mm > 0
            AND inner_rim_width_min_mm > 0
            AND inner_rim_width_max_mm > 0
        ),
    CONSTRAINT chk_schwalbe_tire_rim_width_rule_range_order
        CHECK (
            tire_width_min_mm <= tire_width_max_mm
            AND inner_rim_width_min_mm <= inner_rim_width_max_mm
        ),
    CONSTRAINT uq_schwalbe_tire_rim_width_combination_rule
        UNIQUE (
            tire_width_min_mm,
            tire_width_max_mm,
            inner_rim_width_min_mm,
            inner_rim_width_max_mm
        )
);

CREATE INDEX IF NOT EXISTS idx_schwalbe_tire_rim_width_rule_tire_width
    ON schwalbe_tire_rim_width_combination_rules (
        tire_width_min_mm,
        tire_width_max_mm
    );

CREATE INDEX IF NOT EXISTS idx_schwalbe_tire_rim_width_rule_inner_rim_width
    ON schwalbe_tire_rim_width_combination_rules (
        inner_rim_width_min_mm,
        inner_rim_width_max_mm
    );

COMMENT ON TABLE schwalbe_tire_rim_width_combination_rules IS
    'Schwalbe possible combinations of tire width and inner rim width, based on ETRTO Standard 2024 and Schwalbe guidance (matrix 05/2024). These ranges are not a model-specific compatibility certification and do not replace frame-clearance checks. Hookless or straight-side rims require TLE/TLR tires and the rim manufacturer''s tire-width, pressure, and tire-type requirements.';

COMMENT ON COLUMN schwalbe_tire_rim_width_combination_rules.source_basis IS
    'The standard or manufacturer guidance used to define the possible-combination range.';

COMMENT ON COLUMN schwalbe_tire_rim_width_combination_rules.source_version IS
    'Version printed on the source matrix, for example 05/2024.';

INSERT INTO schwalbe_tire_rim_width_combination_rules (
    tire_width_min_mm,
    tire_width_max_mm,
    inner_rim_width_min_mm,
    inner_rim_width_max_mm,
    source_basis,
    source_version,
    source_url,
    source_checked_at
) VALUES
    (20, 21, 15, 17, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (22, 24, 15, 20, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (25, 27, 15, 22, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (28, 28, 16, 23, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (29, 34, 16, 25, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (35, 46, 17, 27, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (47, 57, 17, 30, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (58, 65, 21, 35, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (66, 71, 25, 43, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (72, 83, 31, 53, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (84, 95, 41, 64, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (96, 113, 59, 89, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28'),
    (114, 132, 72, 100, 'ETRTO Standard 2024; Schwalbe possible combinations guidance', '05/2024', 'https://s3-fsn.scale.sc/schwalbe.public/prod-sde/media/f5/9b/3e/1716878352/Reifen_Felgenkombination_ETRTO_EN_(2).pdf?ts=1716878352', DATE '2026-09-28')
ON CONFLICT (
    tire_width_min_mm,
    tire_width_max_mm,
    inner_rim_width_min_mm,
    inner_rim_width_max_mm
) DO UPDATE SET
    source_basis = EXCLUDED.source_basis,
    source_version = EXCLUDED.source_version,
    source_url = EXCLUDED.source_url,
    source_checked_at = EXCLUDED.source_checked_at,
    updated_at = NOW();
