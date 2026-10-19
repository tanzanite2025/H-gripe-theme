package repository

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
)

func prepareMediaPhysicalCapacityReadOnlySnapshotTestTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, model := range []interface{}{&media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{}, &media.MediaUploadSession{}, &media.BusinessDraftMediaUploadReceipt{}} {
		if !db.Migrator().HasTable(model) {
			require.NoError(t, db.AutoMigrate(model))
		}
	}
}

func TestReadOnlyPhysicalCapacitySnapshotIncludesTombstonesAndExcludesConsumedOrIgnoredWorkflowReceipts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	connection, err := db.DB()
	require.NoError(t, err)
	connection.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = connection.Close() })
	prepareMediaPhysicalCapacityReadOnlySnapshotTestTables(t, db)
	asset := media.MediaAsset{Filename: "kept.png", URL: "https://capacity.example.test/kept.png", StorageKey: "kept.png", Size: 11}
	require.NoError(t, db.Create(&asset).Error)
	require.NoError(t, db.Delete(&asset).Error)
	derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "test", StorageKey: "derived.png", URL: "https://capacity.example.test/derived.png", Size: 4}
	require.NoError(t, db.Create(&derivative).Error)
	require.NoError(t, db.Delete(&derivative).Error)
	for index, status := range []string{outbox.EventStatusPending, outbox.EventStatusProcessed, outbox.EventStatusIgnored} {
		event := outbox.Event{EventKey: status, EventType: outbox.EventTypeObjectStorageCleanup, AggregateType: "media_asset", AggregateID: "test", Status: status, Payload: datatypes.JSON(`{}`)}
		require.NoError(t, db.Create(&event).Error)
		if index == 0 {
			require.NoError(t, db.Delete(&event).Error)
		}
	}
	for _, status := range []string{"uploading", "bound"} {
		require.NoError(t, db.Create(&media.MediaUploadSession{ID: uuid.NewString(), ObjectKey: status, Status: status, ExpiresAt: time.Now().Add(time.Hour)}).Error)
		require.NoError(t, db.Create(&media.BusinessDraftMediaUploadReceipt{ID: uuid.NewString(), OriginalObjectKey: status, Status: status, ObjectKeysJSON: datatypes.JSON(`[]`)}).Error)
	}
	repo := NewMediaRepository(db)
	snapshot, err := repo.ReadOnlyMediaPhysicalCapacityRegistrationSnapshot(t.Context(), 5)
	require.NoError(t, err)
	require.Len(t, snapshot.Originals, 1)
	require.True(t, snapshot.Originals[0].DeletedAt.Valid)
	require.Len(t, snapshot.Derivatives, 1)
	require.True(t, snapshot.Derivatives[0].DeletedAt.Valid)
	require.Len(t, snapshot.CleanupReceipts, 1)
	require.Equal(t, outbox.EventStatusProcessed, snapshot.CleanupReceipts[0].Status)
	require.Len(t, snapshot.PendingConversationUploads, 1)
	require.Equal(t, "uploading", snapshot.PendingConversationUploads[0].Status)
	require.Len(t, snapshot.BusinessDraftUploads, 1)
	_, err = repo.ReadOnlyMediaPhysicalCapacityRegistrationSnapshot(t.Context(), 4)
	require.ErrorIs(t, err, ErrMediaPhysicalCapacitySnapshotLimitExceeded)
	var count int64
	require.NoError(t, db.Unscoped().Model(&outbox.Event{}).Count(&count).Error)
	require.Equal(t, int64(3), count)
}

func TestPostgreSQLReadOnlyPhysicalCapacitySnapshotUsesRepeatableReadWithoutBlockingConcurrentOwnerRegistration(t *testing.T) {
	repo := createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t)
	prepareMediaPhysicalCapacityReadOnlySnapshotTestTables(t, repo.db)
	original := media.MediaAsset{Filename: "old.png", URL: "https://capacity.example.test/old.png", StorageKey: "old.png", Size: 10}
	require.NoError(t, repo.db.Create(&original).Error)
	originalsRead, continueReading := make(chan struct{}), make(chan struct{})
	var isolation, readOnly string
	callbackName := "capacity_snapshot_test_pause_after_originals"
	require.NoError(t, repo.db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table != "media_assets" {
			return
		}
		if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SHOW transaction_isolation").Scan(&isolation).Error; err != nil {
			tx.AddError(err)
			return
		}
		if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SHOW transaction_read_only").Scan(&readOnly).Error; err != nil {
			tx.AddError(err)
			return
		}
		close(originalsRead)
		select {
		case <-continueReading:
		case <-time.After(10 * time.Second):
			tx.AddError(gorm.ErrInvalidTransaction)
		}
	}))
	t.Cleanup(func() { repo.db.Callback().Query().Remove(callbackName) })
	type result struct {
		snapshot MediaPhysicalCapacityRegistrationSnapshot
		err      error
	}
	results := make(chan result, 1)
	go func() {
		snapshot, err := repo.ReadOnlyMediaPhysicalCapacityRegistrationSnapshot(t.Context(), 100)
		results <- result{snapshot, err}
	}()
	select {
	case <-originalsRead:
	case <-time.After(10 * time.Second):
		t.Fatal("snapshot did not read originals")
	}
	writer := repo.db.WithContext(t.Context()).Transaction(func(tx *gorm.DB) error {
		return tx.Create(&media.MediaAssetDerivative{MediaAssetID: original.ID, Preset: "new", StorageKey: "new.png", URL: "https://capacity.example.test/new.png", Size: 4}).Error
	})
	close(continueReading)
	require.NoError(t, writer)
	got := <-results
	require.NoError(t, got.err)
	require.Equal(t, "repeatable read", isolation)
	require.Equal(t, "on", readOnly)
	require.Len(t, got.snapshot.Originals, 1)
	require.Empty(t, got.snapshot.Derivatives)
}

func TestPostgreSQLPendingUploadExpectedCapacityMigrationPreservesUnknownHistoryAndRejectsNegativeBytes(t *testing.T) {
	repo := createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t)
	require.NoError(t, repo.db.Exec("CREATE TABLE media_upload_sessions(id TEXT PRIMARY KEY); INSERT INTO media_upload_sessions(id) VALUES('historical')").Error)
	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "428_add_pending_customer_service_upload_expected_capacity_bytes.up.sql"))
	require.NoError(t, err)
	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "428_add_pending_customer_service_upload_expected_capacity_bytes.down.sql"))
	require.NoError(t, err)
	require.NoError(t, repo.db.Exec(string(up)).Error)
	require.NoError(t, repo.db.Exec(string(up)).Error)
	var historical struct{ ExpectedUploadBytes *int64 }
	require.NoError(t, repo.db.Table("media_upload_sessions").Where("id = ?", "historical").Scan(&historical).Error)
	require.Nil(t, historical.ExpectedUploadBytes)
	require.Error(t, repo.db.Exec("INSERT INTO media_upload_sessions(id,expected_upload_bytes) VALUES('negative',-1)").Error)
	require.NoError(t, repo.db.Exec("INSERT INTO media_upload_sessions(id,expected_upload_bytes) VALUES('known',0)").Error)
	require.NoError(t, repo.db.Exec(string(down)).Error)
	var count int64
	require.NoError(t, repo.db.Table("media_upload_sessions").Count(&count).Error)
	require.Equal(t, int64(2), count)
}
