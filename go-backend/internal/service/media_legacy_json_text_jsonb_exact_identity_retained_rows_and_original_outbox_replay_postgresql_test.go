package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestPostgreSQLLegacyTextJSONAndJSONBReferencesRetainDeletedRowsAndRetryOriginalCleanupEventAfterDetachment(t *testing.T) {
	environment, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	db := environment.db
	for _, statement := range []string{
		"CREATE TABLE showcases (id BIGINT PRIMARY KEY, title TEXT, images JSON)",
		"CREATE TABLE reviews (id BIGINT PRIMARY KEY, product_id BIGINT, images TEXT, deleted_at TIMESTAMPTZ)",
		"CREATE TABLE warranty_claims (id BIGINT PRIMARY KEY, images JSONB, video_url TEXT, deleted_at TIMESTAMPTZ)",
		"CREATE TABLE suggestion_feedback (id BIGINT PRIMARY KEY, attachments JSON, deleted_at TIMESTAMPTZ)",
		"CREATE TABLE ticket_auto_replies (id BIGINT PRIMARY KEY, attachments JSONB)",
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	root := t.TempDir()
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://media.example.test"})
	require.NoError(t, err)
	service := NewMediaService(repository.NewMediaRepository(db), provider, nil, "https://shop.example.test", 20<<30)
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	key := "catalog/postgresql-legacy-json.png"
	asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, key)
	writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
	payload := strings.ReplaceAll(encodeLegacyMediaReferenceRetentionTestJSON(t, []map[string]string{{"url": "https://shop.example.test/uploads/" + key}}), "/", `\/`)
	for _, source := range legacyJSONMediaReferenceSourcesForRetentionTest() {
		if source.Table == "ticket_messages" {
			require.NoError(t, db.Table(source.Table).Create(map[string]any{"id": 1000, "ticket_id": environment.conversation.ID,
				"content": "Synthetic PostgreSQL legacy attachment", "attachments": payload}).Error)
		} else {
			insertLegacyMediaReferenceRetentionTestRow(t, db, source, 1000, payload, source.HasSoftDeletion, "")
		}
	}
	report, err := service.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 6, report.Total)
	retainedCount := 0
	for _, reference := range report.References {
		if reference.UsageState == "retained_deleted_parent" {
			retainedCount++
		}
	}
	require.Equal(t, 3, retainedCount)
	require.ErrorIs(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	foreignJSON := encodeLegacyMediaReferenceRetentionTestJSON(t, []string{"https://foreign.example.test/uploads/" + key})
	for _, source := range legacyJSONMediaReferenceSourcesForRetentionTest() {
		require.NoError(t, db.Table(source.Table).Where("id = ?", 1000).Update(source.JSONColumn, foreignJSON).Error)
	}
	report, err = service.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Zero(t, report.Total)
	failingProvider := &mediaCleanupStorageWithRetainedOriginalForDelayedConflictTest{StorageService: provider, deleteError: errors.New("synthetic PostgreSQL legacy cleanup outage")}
	service.storage = failingProvider
	require.NoError(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	var event outbox.Event
	require.NoError(t, db.Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(asset.ID)).First(&event).Error)
	require.NoError(t, db.Table("reviews").Where("id = ?", 1000).Update("images", payload).Error)
	failingProvider.deleteError = nil
	handler := NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil)
	require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
	_, err = os.Stat(filepath.Join(root, filepath.FromSlash(key)))
	require.NoError(t, err)
	require.NoError(t, db.Table("reviews").Where("id = ?", 1000).Update("deleted_at", nil).Error)
	require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
	require.NoError(t, db.Table("reviews").Where("id = ?", 1000).Update("images", "[]").Error)
	require.NoError(t, handler.Handle(t.Context(), event))
	require.NoError(t, handler.Handle(t.Context(), event))
	_, err = os.Stat(filepath.Join(root, filepath.FromSlash(key)))
	require.ErrorIs(t, err, os.ErrNotExist)
	var count int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("event_key = ?", event.EventKey).Count(&count).Error)
	require.EqualValues(t, 1, count)
}
