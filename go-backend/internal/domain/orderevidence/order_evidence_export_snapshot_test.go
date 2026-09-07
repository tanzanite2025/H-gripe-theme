package orderevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestOrderEvidenceExportSnapshotRejectsHashTamperingAndMutation(t *testing.T) {
	data := datatypes.JSON([]byte(`{"export_type":"order_evidence_manifest","schema_version":1}`))
	hash := sha256.Sum256(data)
	snapshot := OrderEvidenceExportSnapshot{
		OrderID:                100,
		EvidencePackageID:      200,
		EvidencePackageVersion: 1,
		Version:                1,
		Status:                 ExportSnapshotStatusLocked,
		SchemaVersion:          OrderEvidenceExportSnapshotSchemaVersion,
		LockedAt:               time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC),
		SnapshotData:           data,
		SnapshotSHA256:         hex.EncodeToString(hash[:]),
	}

	require.NoError(t, snapshot.Validate())

	snapshot.SnapshotData = datatypes.JSON([]byte(`{"export_type":"tampered"}`))
	require.ErrorContains(t, snapshot.Validate(), "sha256 does not match")
	assert.ErrorIs(t, snapshot.BeforeUpdate(nil), ErrOrderEvidenceExportSnapshotImmutable)
	assert.ErrorIs(t, snapshot.BeforeDelete(nil), ErrOrderEvidenceExportSnapshotImmutable)
}
