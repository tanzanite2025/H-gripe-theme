CREATE TABLE IF NOT EXISTS order_evidence_export_snapshots (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    evidence_package_id BIGINT NOT NULL REFERENCES order_evidence_packages(id) ON DELETE RESTRICT,
    evidence_package_version INTEGER NOT NULL,
    version INTEGER NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'locked',
    schema_version INTEGER NOT NULL,
    locked_at TIMESTAMPTZ NOT NULL,
    snapshot_data JSONB NOT NULL,
    snapshot_sha256 CHAR(64) NOT NULL,
    created_by BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT order_evidence_export_snapshot_package_version_check
        CHECK (evidence_package_version > 0),
    CONSTRAINT order_evidence_export_snapshot_version_check
        CHECK (version > 0),
    CONSTRAINT order_evidence_export_snapshot_status_check
        CHECK (status = 'locked'),
    CONSTRAINT order_evidence_export_snapshot_schema_version_check
        CHECK (schema_version = 1),
    CONSTRAINT uq_order_evidence_export_snapshot_version
        UNIQUE (evidence_package_id, version)
);

CREATE INDEX IF NOT EXISTS idx_order_evidence_export_snapshots_order
    ON order_evidence_export_snapshots(order_id, evidence_package_version DESC, version DESC);
