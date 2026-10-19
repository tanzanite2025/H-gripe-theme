package repository

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestPostgreSQLReadOnlyMediaUsageSnapshotKeepsCandidatesAndReferencesConsistentWithoutBlockingWriters(t *testing.T) {
	repo := createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t)
	asset := media.MediaAsset{Filename: "snapshot.png", URL: "https://media.example.test/uploads/snapshot.png", StorageKey: "snapshot.png"}
	require.NoError(t, repo.db.Create(&asset).Error)
	require.NoError(t, repo.db.AutoMigrate(&media.BusinessDraftMediaUploadReceipt{}))
	require.NoError(t, repo.db.Exec("CREATE TABLE posts(id BIGINT PRIMARY KEY,title TEXT,content TEXT,featured_img TEXT,deleted_at TIMESTAMPTZ)").Error)
	candidatesRead, continueScanning := make(chan struct{}), make(chan struct{})
	var isolation, readOnly string
	callbackName := "media_usage_summary_pause_after_candidate_snapshot"
	require.NoError(t, repo.db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table != "media_assets" {
			return
		}
		// The other-registrations scanner uses Table and Find as well.
		if _, ok := tx.Statement.Dest.(*[]media.MediaAsset); !ok {
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
		close(candidatesRead)
		select {
		case <-continueScanning:
		case <-time.After(10 * time.Second):
			tx.AddError(gorm.ErrInvalidTransaction)
		}
	}))
	t.Cleanup(func() { repo.db.Callback().Query().Remove(callbackName) })
	type snapshotResult struct {
		referenceCount int
		pendingDraft   bool
		err            error
	}
	results := make(chan snapshotResult, 1)
	go func() {
		count := 0
		pendingDraft := false
		err := repo.WalkReadOnlyMediaUsageSnapshot(t.Context(), MediaAssetFilter{}, 200, 20000, func(item media.MediaAsset, refs []media.AssetReference) {
			count += len(refs)
			pendingDraft = item.BusinessDraftUploadExpiresAt != nil
		})
		results <- snapshotResult{count, pendingDraft, err}
	}()
	select {
	case <-candidatesRead:
	case <-time.After(10 * time.Second):
		t.Fatal("candidate snapshot was not established")
	}
	writerError := repo.db.WithContext(t.Context()).Exec("INSERT INTO posts(id,title,content,featured_img) VALUES(1,'new concurrent article','',?)", asset.URL).Error
	receiptError := repo.db.WithContext(t.Context()).Create(&media.BusinessDraftMediaUploadReceipt{ID: "00000000-0000-0000-0000-000000000001", AssetID: &asset.ID,
		OriginalObjectKey: "snapshot.png", ObjectKeysJSON: datatypes.JSON(`[]`), Status: "uploaded", ExpiresAt: time.Now().Add(time.Hour)}).Error
	close(continueScanning)
	require.NoError(t, writerError)
	require.NoError(t, receiptError)
	result := <-results
	require.NoError(t, result.err)
	require.Equal(t, "repeatable read", isolation)
	require.Equal(t, "on", readOnly)
	require.Zero(t, result.referenceCount)
	require.False(t, result.pendingDraft)
	require.NoError(t, repo.db.Callback().Query().Remove(callbackName))
	var references []media.AssetReference
	require.NoError(t, repo.WalkReadOnlyMediaUsageSnapshot(t.Context(), MediaAssetFilter{}, 200, 20000, func(item media.MediaAsset, refs []media.AssetReference) {
		references = refs
		require.NotNil(t, item.BusinessDraftUploadExpiresAt)
	}))
	require.Len(t, references, 1)
	require.Equal(t, "post", references[0].ResourceType)
}

func TestPostgreSQLReadOnlyMediaUsageSnapshotReferenceBudgetStopsDelivery(t *testing.T) {
	repo := createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t)
	asset := media.MediaAsset{Filename: "budget.png", URL: "https://media.example.test/uploads/budget.png"}
	require.NoError(t, repo.db.Create(&asset).Error)
	require.NoError(t, repo.db.Exec("CREATE TABLE posts(id BIGINT PRIMARY KEY,title TEXT,content TEXT,featured_img TEXT,deleted_at TIMESTAMPTZ)").Error)
	require.NoError(t, repo.db.Exec("INSERT INTO posts(id,title,content,featured_img) VALUES(1,'one','',?),(2,'two','',?)", asset.URL, asset.URL).Error)
	consumed := false
	err := repo.WalkReadOnlyMediaUsageSnapshot(t.Context(), MediaAssetFilter{}, 200, 1, func(media.MediaAsset, []media.AssetReference) { consumed = true })
	require.ErrorIs(t, err, ErrMediaUsageSnapshotLimitExceeded)
	require.False(t, consumed)
}
