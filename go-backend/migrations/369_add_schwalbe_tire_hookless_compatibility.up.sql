-- Keep Hookless compatibility as an independently sourced product fact.
-- Do not infer it from TLE/TLR, casing, category, or mountain-bike usage.
CREATE TABLE IF NOT EXISTS schwalbe_tire_hookless_compatibility (
    scope_key VARCHAR(240) PRIMARY KEY,
    article_no VARCHAR(32),
    model_name VARCHAR(180),
    status VARCHAR(32) NOT NULL,
    source_basis TEXT NOT NULL,
    source_version VARCHAR(64) NOT NULL,
    source_url TEXT NOT NULL,
    source_checked_at DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_schwalbe_hookless_compatibility_scope
        CHECK (article_no IS NOT NULL OR model_name IS NOT NULL),
    CONSTRAINT chk_schwalbe_hookless_compatibility_status
        CHECK (status IN ('supported', 'not_supported', 'unknown'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_schwalbe_hookless_compatibility_article
    ON schwalbe_tire_hookless_compatibility (article_no)
    WHERE article_no IS NOT NULL;

-- A model fallback may coexist with article-level overrides for that model.
CREATE UNIQUE INDEX IF NOT EXISTS uq_schwalbe_hookless_compatibility_model
    ON schwalbe_tire_hookless_compatibility (model_name)
    WHERE article_no IS NULL AND model_name IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_schwalbe_hookless_compatibility_lookup
    ON schwalbe_tire_hookless_compatibility (article_no, model_name);

COMMENT ON TABLE schwalbe_tire_hookless_compatibility IS
    'Explicit Schwalbe Hookless compatibility facts. Missing records remain unknown; no status is inferred from seal, casing, category, or model family heuristics.';

COMMENT ON COLUMN schwalbe_tire_hookless_compatibility.scope_key IS
    'Stable maintenance key, normally article:<Article No.> or model:<exact model name>.';

COMMENT ON COLUMN schwalbe_tire_hookless_compatibility.source_basis IS
    'The official source and filter meaning used to establish the status.';

INSERT INTO schwalbe_tire_hookless_compatibility (
    scope_key,
    model_name,
    status,
    source_basis,
    source_version,
    source_url,
    source_checked_at
) VALUES
    ('model:Al Mighty', 'Al Mighty', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:Big Betty', 'Big Betty', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:Dirty Dan', 'Dirty Dan', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:Nobby Nic', 'Nobby Nic', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:Racing Ralph', 'Racing Ralph', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:Racing Ray', 'Racing Ray', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:Rick', 'Rick', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:Thunder Burt', 'Thunder Burt', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY Magic Mary', 'GRAVITY Magic Mary', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY PRO Albert Radial', 'GRAVITY PRO Albert Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY PRO Eddy Current Radial', 'GRAVITY PRO Eddy Current Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY PRO Magic Mary', 'GRAVITY PRO Magic Mary', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY PRO Magic Mary Radial', 'GRAVITY PRO Magic Mary Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY PRO Romy', 'GRAVITY PRO Romy', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY PRO Romy Radial', 'GRAVITY PRO Romy Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY PRO Shredda F Radial', 'GRAVITY PRO Shredda F Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY PRO Shredda R Radial', 'GRAVITY PRO Shredda R Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY PRO Tacky Chan', 'GRAVITY PRO Tacky Chan', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY PRO Tacky Chan Radial', 'GRAVITY PRO Tacky Chan Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:GRAVITY Tacky Chan', 'GRAVITY Tacky Chan', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL Magic Mary', 'TRAIL Magic Mary', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL PRO Albert Radial', 'TRAIL PRO Albert Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL PRO Magic Mary', 'TRAIL PRO Magic Mary', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL PRO Magic Mary Radial', 'TRAIL PRO Magic Mary Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL PRO Nobby Nic', 'TRAIL PRO Nobby Nic', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL PRO Romy', 'TRAIL PRO Romy', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL PRO Romy Radial', 'TRAIL PRO Romy Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL PRO Tacky Chan', 'TRAIL PRO Tacky Chan', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL PRO Tacky Chan Radial', 'TRAIL PRO Tacky Chan Radial', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL Romy', 'TRAIL Romy', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:TRAIL Tacky Chan', 'TRAIL Tacky Chan', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:RACE PRO Romy', 'RACE PRO Romy', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:RACE PRO Tacky Chan', 'RACE PRO Tacky Chan', 'supported', 'Schwalbe MTB filter: Hookless compatible = Yes', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=5e2c9aee6ff74f20a34e0fbafc97a5a0', DATE '2026-09-28'),
    ('model:Billy Bonkers', 'Billy Bonkers', 'not_supported', 'Schwalbe MTB filter: Hookless compatible = No', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=521a732c26eb4eb588616d909a10112b', DATE '2026-09-28'),
    ('model:Johnny Watts LR', 'Johnny Watts LR', 'not_supported', 'Schwalbe MTB filter: Hookless compatible = No', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=521a732c26eb4eb588616d909a10112b', DATE '2026-09-28'),
    ('model:Little Joe', 'Little Joe', 'not_supported', 'Schwalbe MTB filter: Hookless compatible = No', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=521a732c26eb4eb588616d909a10112b', DATE '2026-09-28'),
    ('model:Marathon Plus MTB', 'Marathon Plus MTB', 'not_supported', 'Schwalbe MTB filter: Hookless compatible = No', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=521a732c26eb4eb588616d909a10112b', DATE '2026-09-28'),
    ('model:Rapid Rob', 'Rapid Rob', 'not_supported', 'Schwalbe MTB filter: Hookless compatible = No', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=521a732c26eb4eb588616d909a10112b', DATE '2026-09-28'),
    ('model:Smart Sam', 'Smart Sam', 'not_supported', 'Schwalbe MTB filter: Hookless compatible = No', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=521a732c26eb4eb588616d909a10112b', DATE '2026-09-28'),
    ('model:SX-R', 'SX-R', 'not_supported', 'Schwalbe MTB filter: Hookless compatible = No', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=521a732c26eb4eb588616d909a10112b', DATE '2026-09-28'),
    ('model:Tough Tom', 'Tough Tom', 'not_supported', 'Schwalbe MTB filter: Hookless compatible = No', '2026-09-28', 'https://www.schwalbe.com/en/bike-tires/mtb/?properties=521a732c26eb4eb588616d909a10112b', DATE '2026-09-28')
ON CONFLICT (scope_key) DO UPDATE SET
    article_no = EXCLUDED.article_no,
    model_name = EXCLUDED.model_name,
    status = EXCLUDED.status,
    source_basis = EXCLUDED.source_basis,
    source_version = EXCLUDED.source_version,
    source_url = EXCLUDED.source_url,
    source_checked_at = EXCLUDED.source_checked_at,
    updated_at = NOW();
