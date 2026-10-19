package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPostgreSQLMediaAssociationDiagnosisSnapshotKeepsAbsentParentConsistentWithoutBlockingItsRecreation(t *testing.T) {
	repo := createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t)
	require.NoError(t, repo.db.Exec("CREATE TABLE galleries(id BIGINT PRIMARY KEY)").Error)
	var isolation, readOnly string
	err := repo.ReadMediaAssociationDiagnosisSnapshot(t.Context(), func(snapshot *MediaRepository, tx *gorm.DB) error {
		require.NoError(t, tx.Raw("SHOW transaction_isolation").Scan(&isolation).Error)
		require.NoError(t, tx.Raw("SHOW transaction_read_only").Scan(&readOnly).Error)
		var before, after int64
		require.NoError(t, tx.Table("galleries").Where("id = 33").Count(&before).Error)
		require.Zero(t, before)
		require.NoError(t, repo.db.WithContext(t.Context()).Exec("INSERT INTO galleries(id) VALUES(33)").Error)
		require.NoError(t, tx.Table("galleries").Where("id = 33").Count(&after).Error)
		require.Zero(t, after)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, "repeatable read", isolation)
	require.Equal(t, "on", readOnly)
	var committed int64
	require.NoError(t, repo.db.Table("galleries").Where("id = 33").Count(&committed).Error)
	require.EqualValues(t, 1, committed)
}
