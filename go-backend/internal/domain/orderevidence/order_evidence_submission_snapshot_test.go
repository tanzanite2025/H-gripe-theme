package orderevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOrderEvidenceSubmissionSnapshotIsImmutableAfterCreate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OrderEvidenceSubmissionSnapshot{}))

	data := []byte(`{"provider":"stripe","request":{"submit":true}}`)
	hash := sha256.Sum256(data)
	snapshot := &OrderEvidenceSubmissionSnapshot{
		Provider:       "stripe",
		DisputeID:      11,
		OrderID:        22,
		Version:        1,
		Status:         SubmissionSnapshotStatusLocked,
		SchemaVersion:  OrderEvidenceSubmissionSnapshotSchemaVersion,
		LockedAt:       time.Now().UTC(),
		SnapshotData:   data,
		SnapshotSHA256: hex.EncodeToString(hash[:]),
	}
	require.NoError(t, db.Create(snapshot).Error)

	require.ErrorIs(t, db.Model(&OrderEvidenceSubmissionSnapshot{}).
		Where("id = ?", snapshot.ID).
		Update("snapshot_data", `{"tampered":true}`).Error, ErrOrderEvidenceSubmissionSnapshotImmutable)
	require.ErrorIs(t, db.Delete(snapshot).Error, ErrOrderEvidenceSubmissionSnapshotImmutable)
}
