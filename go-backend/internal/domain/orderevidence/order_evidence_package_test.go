package orderevidence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderEvidencePackageValidateRequiresSnapshotAndLockConsistency(t *testing.T) {
	pkg := OrderEvidencePackage{
		OrderID:        10,
		SnapshotID:     20,
		PackageVersion: 1,
		Status:         PackageStatusIncomplete,
		SchemaVersion:  OrderEvidencePackageSchemaVersion,
	}
	require.NoError(t, pkg.Validate())

	pkg.SnapshotID = 0
	require.ErrorContains(t, pkg.Validate(), "snapshot_id is required")

	pkg.SnapshotID = 20
	pkg.Status = PackageStatusLocked
	require.ErrorContains(t, pkg.Validate(), "requires locked_at")

	lockedAt := time.Now().UTC()
	pkg.LockedAt = &lockedAt
	require.NoError(t, pkg.Validate())

	pkg.Status = PackageStatusReady
	require.ErrorContains(t, pkg.Validate(), "cannot have locked_at")
	assert.Equal(t, PackageStatusReady, pkg.Status)
}
