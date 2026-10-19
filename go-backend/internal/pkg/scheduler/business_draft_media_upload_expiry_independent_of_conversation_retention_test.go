package scheduler

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBusinessDraftMediaUploadExpiryRunsWhenConversationRetentionIsDisabled(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.BusinessDraftMediaUploadReceipt{}, &outbox.Event{}))
	store, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://scheduler.example.test"})
	require.NoError(t, err)
	assets := service.NewMediaService(repository.NewMediaRepository(db), store, nil, "https://scheduler.example.test", 20<<30)
	assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	id := uuid.NewString()
	key, err := storage.AllocateBusinessDraftMediaObjectKey(id, ".png")
	require.NoError(t, err)
	encoded, _ := json.Marshal([]string{key})
	receipt := media.BusinessDraftMediaUploadReceipt{ID: id, UploaderID: 42, UploadOrigin: media.MediaUploadOriginAdministratorCatalog, OriginalObjectKey: key, ObjectKeysJSON: encoded, ContentSHA256: strings.Repeat("a", 64), Size: 12, MimeType: "image/png", Status: "uploading", CreatedAt: time.Now().UTC().Add(-48 * time.Hour), ExpiresAt: time.Now().UTC().Add(-time.Hour)}
	require.NoError(t, db.Create(&receipt).Error)
	scheduler := NewCustomerServiceRetentionScheduler(nil, config.WorkerConfig{CustomerServiceRetentionEnabled: false, CustomerServiceRetentionBatchLimit: 10})
	scheduler.ConfigureBusinessDraftMediaUploadExpiry(assets)
	result, err := scheduler.RunBusinessDraftMediaUploadExpiryOnce(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, result.Retired)
	var event outbox.Event
	require.NoError(t, db.First(&event).Error)
	require.NoError(t, service.NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil).Handle(t.Context(), event))
}
