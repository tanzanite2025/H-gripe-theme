CREATE TABLE IF NOT EXISTS order_evidence_packages (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    snapshot_id BIGINT NOT NULL REFERENCES order_evidence_snapshots(id) ON DELETE RESTRICT,
    package_version INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'incomplete',
    order_total_usd_snapshot NUMERIC(14,2) NOT NULL,
    is_high_value BOOLEAN NOT NULL DEFAULT FALSE,
    has_spoke_tension_qc BOOLEAN NOT NULL DEFAULT FALSE,
    schema_version INTEGER NOT NULL,
    created_by BIGINT NOT NULL DEFAULT 0,
    locked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT order_evidence_package_status_check
        CHECK (status IN ('incomplete', 'ready', 'locked', 'superseded')),
    CONSTRAINT order_evidence_package_version_check
        CHECK (package_version > 0),
    CONSTRAINT order_evidence_package_schema_version_check
        CHECK (schema_version = 1),
    CONSTRAINT order_evidence_package_total_usd_check
        CHECK (order_total_usd_snapshot >= 0),
    CONSTRAINT order_evidence_package_locked_at_check
        CHECK (
            (status = 'locked' AND locked_at IS NOT NULL)
            OR (status <> 'locked' AND locked_at IS NULL)
        ),
    CONSTRAINT uq_order_evidence_package_version
        UNIQUE (order_id, package_version)
);

CREATE TABLE IF NOT EXISTS order_evidence_items (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL REFERENCES order_evidence_packages(id) ON DELETE RESTRICT,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    order_item_id BIGINT REFERENCES order_items(id) ON DELETE RESTRICT,
    snapshot_id BIGINT REFERENCES order_evidence_snapshots(id) ON DELETE RESTRICT,
    item_type VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'missing',
    required_reason VARCHAR(32) NOT NULL,
    data_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    captured_at TIMESTAMPTZ,
    captured_by BIGINT NOT NULL DEFAULT 0,
    content_sha256 CHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT order_evidence_item_type_check
        CHECK (
            item_type IN (
                'configuration_confirmation',
                'product_identity',
                'outbound_weight_packaging',
                'signed_pod',
                'spoke_qc_tension'
            )
        ),
    CONSTRAINT order_evidence_item_status_check
        CHECK (status IN ('missing', 'draft', 'complete', 'waived')),
    CONSTRAINT order_evidence_item_reason_check
        CHECK (required_reason IN ('base', 'high_value', 'spoke_tension_qc')),
    CONSTRAINT order_evidence_item_order_item_scope_check
        CHECK (
            (
                item_type IN ('configuration_confirmation', 'product_identity', 'spoke_qc_tension')
                AND order_item_id IS NOT NULL
            )
            OR (
                item_type IN ('outbound_weight_packaging', 'signed_pod')
                AND order_item_id IS NULL
            )
        ),
    CONSTRAINT order_evidence_item_snapshot_scope_check
        CHECK (
            snapshot_id IS NULL
            OR (item_type = 'configuration_confirmation' AND status = 'complete')
        ),
    CONSTRAINT order_evidence_item_complete_data_check
        CHECK (
            status <> 'complete'
            OR (
                captured_at IS NOT NULL
                AND char_length(content_sha256) = 64
            )
        )
);

CREATE TABLE IF NOT EXISTS order_evidence_attachments (
    id BIGSERIAL PRIMARY KEY,
    evidence_item_id BIGINT NOT NULL REFERENCES order_evidence_items(id) ON DELETE RESTRICT,
    storage_key TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    mime_type VARCHAR(160) NOT NULL,
    size_bytes BIGINT NOT NULL,
    sha256 CHAR(64) NOT NULL,
    uploaded_by BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT order_evidence_attachment_size_check
        CHECK (size_bytes >= 0),
    CONSTRAINT order_evidence_attachment_sha256_check
        CHECK (sha256 ~ '^[0-9a-fA-F]{64}$'),
    CONSTRAINT uq_order_evidence_attachment_item_storage_key
        UNIQUE (evidence_item_id, storage_key)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_order_evidence_item_order_scope
    ON order_evidence_items(package_id, item_type)
    WHERE order_item_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_order_evidence_item_line_scope
    ON order_evidence_items(package_id, order_item_id, item_type)
    WHERE order_item_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_order_evidence_packages_order
    ON order_evidence_packages(order_id, package_version DESC);
CREATE INDEX IF NOT EXISTS idx_order_evidence_items_package
    ON order_evidence_items(package_id, id);
CREATE INDEX IF NOT EXISTS idx_order_evidence_items_order
    ON order_evidence_items(order_id, status);
CREATE INDEX IF NOT EXISTS idx_order_evidence_attachments_item
    ON order_evidence_attachments(evidence_item_id, id);
