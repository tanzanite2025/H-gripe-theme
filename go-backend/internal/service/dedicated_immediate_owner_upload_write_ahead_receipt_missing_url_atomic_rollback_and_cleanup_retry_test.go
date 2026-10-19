package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type dedicatedImmediateOwnerPreallocatedUploadControlledStorage struct {
	storage.StorageService
	writtenKeys             []string
	writeError              error
	returnReferenceOverride *string
	afterWrite              func()
}

func (s *dedicatedImmediateOwnerPreallocatedUploadControlledStorage) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func (s *dedicatedImmediateOwnerPreallocatedUploadControlledStorage) WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key, cacheControl string) (string, error) {
	s.writtenKeys = append(s.writtenKeys, key)
	reference, err := s.StorageService.(storage.PreallocatedDedicatedImmediateOwnerObjectWriter).WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(ctx, reader, key, cacheControl)
	if err != nil {
		return reference, err
	}
	if s.afterWrite != nil {
		s.afterWrite()
	}
	if s.returnReferenceOverride != nil {
		reference = *s.returnReferenceOverride
	}
	return reference, s.writeError
}

type dedicatedImmediateOwnerUploadReceiptRegressionEnvironment struct {
	db         *gorm.DB
	provider   *dedicatedImmediateOwnerPreallocatedUploadControlledStorage
	upload     func(context.Context) error
	currentKey func() string
	handle     func(context.Context, outbox.Event) error
}

func configureDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t *testing.T, db *gorm.DB, provider storage.StorageService, namespace string) dedicatedImmediateOwnerUploadReceiptRegressionEnvironment {
	t.Helper()
	controlled := &dedicatedImmediateOwnerPreallocatedUploadControlledStorage{StorageService: provider}
	outboxRepo := repository.NewOutboxRepository(db)
	settings := NewSettingService(repository.NewSettingRepository(db), nil, 0)
	env := dedicatedImmediateOwnerUploadReceiptRegressionEnvironment{db: db, provider: controlled}
	switch namespace {
	case SiteLogoStoragePrefix:
		service := NewSiteLogoService(repository.NewSiteLogoRepository(db), controlled, "https://shop.example.test")
		service.ConfigureSettingService(settings)
		service.ConfigureObjectCleanupOutbox(outboxRepo)
		env.upload = func(ctx context.Context) error {
			_, err := service.UploadCurrent(ctx, createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "logo.webp"), 11)
			return err
		}
		env.currentKey = func() string {
			current, err := service.repo.Current()
			require.NoError(t, err)
			if current == nil {
				return ""
			}
			return current.StorageKey
		}
		env.handle = NewObjectStorageCleanupOutboxHandler(nil, service, nil, nil).Handle
	case SiteFaviconStoragePrefix:
		service := NewSiteFaviconService(repository.NewSiteFaviconRepository(db), controlled, "https://shop.example.test")
		service.ConfigureSettingService(settings)
		service.ConfigureObjectCleanupOutbox(outboxRepo)
		env.upload = func(ctx context.Context) error {
			_, err := service.UploadCurrent(ctx, createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "favicon.webp"), 11)
			return err
		}
		env.currentKey = func() string {
			current, err := service.repo.Current()
			require.NoError(t, err)
			if current == nil {
				return ""
			}
			return current.StorageKey
		}
		handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
		handler.ConfigureSiteFaviconService(service)
		env.handle = handler.Handle
	default:
		profile := createCustomerServiceAvatarProfile(t, db, "")
		service := NewCustomerServiceAvatarService(repository.NewUserRepository(db), controlled, outboxRepo, "https://shop.example.test")
		env.upload = func(ctx context.Context) error {
			_, err := service.Upload(ctx, *profile.UserID, customerServiceAvatarFile(t))
			return err
		}
		env.currentKey = func() string {
			var saved user.AgentProfile
			require.NoError(t, db.First(&saved, profile.ID).Error)
			if saved.Avatar == "" {
				return ""
			}
			key, err := provider.ObjectKey(saved.Avatar)
			require.NoError(t, err)
			return key
		}
		env.handle = NewCustomerServiceAvatarCleanupHandler(service.userRepo, controlled, service.siteURL).Handle
	}
	return env
}

func createDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t *testing.T, namespace string) dedicatedImmediateOwnerUploadReceiptRegressionEnvironment {
	t.Helper()
	db := newCustomerServiceAvatarTestDB(t)
	require.NoError(t, db.AutoMigrate(&sitelogodomain.Asset{}, &sitefavicondomain.Asset{}, &setting.Setting{}))
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://media.example.test/uploads"})
	require.NoError(t, err)
	return configureDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, db, provider, namespace)
}

func TestDedicatedImmediateOwnerUploadRetainsExactWriteAheadReceiptForMissingURLCancellationAndOwnerRollback(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, CustomerServiceAvatarStoragePrefix} {
		for _, failure := range []string{"missing URL with error", "wrong same namespace URL", "foreign URL", "empty success URL", "cancellation", "owner rollback", "receipt exchange rollback"} {
			t.Run(namespace+"/"+failure, func(t *testing.T) {
				env := createDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, namespace)
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				var override string
				switch failure {
				case "missing URL with error":
					env.provider.returnReferenceOverride = &override
					env.provider.writeError = errors.New("provider wrote bytes but lost its response")
				case "wrong same namespace URL":
					override = env.provider.GetURL(namespace + "/existing-unrelated.webp")
					env.provider.returnReferenceOverride = &override
				case "foreign URL":
					override = "https://foreign.example.test/uploads/retained.webp"
					env.provider.returnReferenceOverride = &override
				case "empty success URL":
					env.provider.returnReferenceOverride = &override
				case "cancellation":
					env.provider.afterWrite = cancel
				case "owner rollback":
					callback := "reject_immediate_owner_binding_" + uuid.NewString()
					ownerTable := "site_logo_assets"
					if namespace == SiteFaviconStoragePrefix {
						ownerTable = "site_favicon_assets"
					}
					if namespace == CustomerServiceAvatarStoragePrefix {
						ownerTable = "customer_service_agent_profiles"
					}
					reject := func(tx *gorm.DB) {
						if tx.Statement.Table == ownerTable {
							tx.AddError(errors.New("owner binding unavailable"))
						}
					}
					require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callback, reject))
					require.NoError(t, env.db.Callback().Update().Before("gorm:update").Register(callback, reject))
					t.Cleanup(func() {
						_ = env.db.Callback().Create().Remove(callback)
						_ = env.db.Callback().Update().Remove(callback)
					})
				case "receipt exchange rollback":
					callback := "reject_immediate_receipt_exchange_" + uuid.NewString()
					require.NoError(t, env.db.Callback().Delete().Before("gorm:delete").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "outbox_events" {
							tx.AddError(errors.New("receipt exchange unavailable"))
						}
					}))
					t.Cleanup(func() { _ = env.db.Callback().Delete().Remove(callback) })
				}
				require.Error(t, env.upload(ctx))
				require.Empty(t, env.currentKey())
				require.Len(t, env.provider.writtenKeys, 1)
				key := env.provider.writtenKeys[0]
				var events []outbox.Event
				require.NoError(t, env.db.Find(&events).Error)
				require.Len(t, events, 1)
				require.Equal(t, outbox.EventStatusPending, events[0].Status)
				require.Contains(t, string(events[0].Payload), key)
				require.NotContains(t, string(events[0].Payload), "existing-unrelated")
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, key)
				// The same receipt survives provider failure and is replayed by the
				// existing worker once its write deadline has passed.
				worker := NewOutboxService(repository.NewOutboxRepository(env.db))
				worker.RegisterHandler(events[0].EventType, env.handle)
				result, err := worker.ProcessPending(t.Context(), time.Now().Add(6*time.Minute), 10)
				require.NoError(t, err)
				require.Equal(t, 1, result.Processed)
				requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, key)
			})
		}
	}
}

func TestDedicatedImmediateOwnerUploadNeverWritesWithoutCommittedReceiptAndTransfersSuccessfulReceiptAtomically(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, CustomerServiceAvatarStoragePrefix} {
		for _, failure := range []bool{true, false} {
			t.Run(namespace+"/"+map[bool]string{true: "receipt persistence failure", false: "successful owner exchange"}[failure], func(t *testing.T) {
				env := createDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, namespace)
				persisted := false
				callback := "observe_immediate_owner_receipt_commit_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "outbox_events" {
						if failure {
							tx.AddError(errors.New("receipt unavailable"))
						} else {
							persisted = true
						}
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
				env.provider.afterWrite = func() { require.True(t, persisted) }
				err := env.upload(t.Context())
				if failure {
					require.Error(t, err)
					require.Empty(t, env.provider.writtenKeys)
					require.Empty(t, env.currentKey())
				} else {
					require.NoError(t, err)
					require.Equal(t, env.provider.writtenKeys[0], env.currentKey())
				}
				var events []outbox.Event
				require.NoError(t, env.db.Unscoped().Find(&events).Error)
				require.Empty(t, events)
			})
		}
	}
}

func TestDedicatedImmediateOwnerUploadReceiptRetainsOriginalTaskAcrossHistoricalReferencesAndRetry(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, CustomerServiceAvatarStoragePrefix} {
		t.Run(namespace, func(t *testing.T) {
			env := createDedicatedImmediateOwnerUploadReceiptRegressionEnvironment(t, namespace)
			env.provider.writeError = errors.New("upload interrupted after write")
			require.Error(t, env.upload(t.Context()))
			key := env.provider.writtenKeys[0]
			require.NoError(t, env.db.AutoMigrate(&post.Post{}))
			detach := insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, env.db, key, "post_content")
			var original outbox.Event
			require.NoError(t, env.db.First(&original).Error)
			worker := NewOutboxService(repository.NewOutboxRepository(env.db))
			worker.RegisterHandler(original.EventType, env.handle)
			result, err := worker.ProcessPending(t.Context(), time.Now().Add(6*time.Minute), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Failed)
			var failed outbox.Event
			require.NoError(t, env.db.First(&failed, original.ID).Error)
			require.Equal(t, original.EventKey, failed.EventKey)
			require.JSONEq(t, string(original.Payload), string(failed.Payload))
			requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, key)
			detach()
			result, err = worker.ProcessPending(t.Context(), time.Now().Add(12*time.Minute), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Processed)
			requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, key)
			var payload map[string]interface{}
			require.NoError(t, json.Unmarshal(original.Payload, &payload))
		})
	}
}
