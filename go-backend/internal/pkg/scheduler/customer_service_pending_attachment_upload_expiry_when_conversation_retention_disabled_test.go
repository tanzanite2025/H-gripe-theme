package scheduler

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestCustomerServicePendingUploadExpiryRunsThroughExistingSchedulerWhenConversationRetentionIsDisabled(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&ticket.Ticket{}, &ticket.TicketMessage{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.MediaUploadSession{},
		&ticket.TicketMessageMediaAttachment{}, &audit.AuditLog{}, &outbox.Event{}))
	privateRoot := t.TempDir()
	store, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: privateRoot, BaseURL: "https://public.example.test"})
	require.NoError(t, err)
	mediaService := service.NewMediaService(repository.NewMediaRepository(db), store, nil, "https://public.example.test", 20<<30)
	mediaService.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	conversation := ticket.Ticket{UserID: 42, Category: "customer_service", Status: "open", Subject: "Expiry independent from retention"}
	require.NoError(t, db.Create(&conversation).Error)
	key := storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/scheduler-abandoned.png"
	_, err = store.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("abandoned")), key)
	require.NoError(t, err)
	session := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext, BusinessContextID: conversation.ID,
		UploadActorType: media.MediaUploadActorAdministrator, UploadActorAccountID: 42, ObjectKey: key, Status: "failed", CreatedAt: time.Now().Add(-48 * time.Hour), ExpiresAt: time.Now().Add(-24 * time.Hour)}
	require.NoError(t, db.Create(&session).Error)
	retention := service.NewCustomerServiceRetentionService(repository.NewTicketRepository(db), nil, nil, nil, nil, service.NewAuditService(repository.NewAuditRepository(db)))
	retention.ConfigureMediaService(mediaService)
	retention.ConfigureCleanupOutbox(repository.NewOutboxRepository(db))
	scheduler := NewCustomerServiceRetentionScheduler(retention, config.WorkerConfig{CustomerServiceRetentionEnabled: false, CustomerServiceRetentionBatchLimit: 10})
	conversationResult, err := scheduler.RunOnce()
	require.NoError(t, err)
	require.Zero(t, conversationResult.Purged)
	result, err := scheduler.RunCustomerServiceAttachmentUploadExpiryOnce(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, result.Retired)
	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&event).Error)
	require.NoError(t, service.NewObjectStorageCleanupOutboxHandler(mediaService, nil, nil, nil).Handle(t.Context(), event))
	_, err = os.Stat(filepath.Join(privateRoot, filepath.FromSlash(key)))
	require.True(t, os.IsNotExist(err))
	require.NoError(t, db.First(&conversation, conversation.ID).Error)
}
