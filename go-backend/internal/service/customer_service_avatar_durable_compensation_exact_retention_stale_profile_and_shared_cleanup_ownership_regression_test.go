package service

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type customerServiceAvatarDurableCompensationRegressionEnvironment struct {
	db       *gorm.DB
	provider *dedicatedPublicUploadCompensationStorageWithControlledResultAndDeletion
	service  *CustomerServiceAvatarService
	handler  *CustomerServiceAvatarCleanupHandler
	profile  *user.AgentProfile
	key      string
	root     string
}

func configureCustomerServiceAvatarDurableCompensationRegressionEnvironment(t *testing.T, db *gorm.DB, underlying storage.StorageService, root string) customerServiceAvatarDurableCompensationRegressionEnvironment {
	t.Helper()
	key := CustomerServiceAvatarStoragePrefix + "/historical/face image.webp"
	writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
	provider := &dedicatedPublicUploadCompensationStorageWithControlledResultAndDeletion{StorageService: underlying, uploadReference: underlying.GetURL(key)}
	profile := createCustomerServiceAvatarProfile(t, db, "")
	service := NewCustomerServiceAvatarService(repository.NewUserRepository(db), provider, repository.NewOutboxRepository(db), "https://shop.example.test")
	handler := NewCustomerServiceAvatarCleanupHandler(service.userRepo, provider, service.siteURL)
	return customerServiceAvatarDurableCompensationRegressionEnvironment{db, provider, service, handler, profile, key, root}
}

func createCustomerServiceAvatarDurableCompensationRegressionEnvironment(t *testing.T) customerServiceAvatarDurableCompensationRegressionEnvironment {
	t.Helper()
	db := newCustomerServiceAvatarTestDB(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &post.Post{}, &product.Product{}, &product.ProductMedia{}))
	root := t.TempDir()
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://avatars.example.test/assets"})
	require.NoError(t, err)
	return configureCustomerServiceAvatarDurableCompensationRegressionEnvironment(t, db, provider, root)
}

func createCustomerServiceAvatarDurableCleanupEventForRegression(t *testing.T, env customerServiceAvatarDurableCompensationRegressionEnvironment) outbox.Event {
	t.Helper()
	event, err := env.service.enqueueCleanup(repository.NewOutboxRepository(env.db), env.profile.ID, env.provider.uploadReference)
	require.NoError(t, err)
	return *event
}

func TestLegacyCustomerServiceAvatarFailedUploadRetainsExactOwnerHistoryAndRegistrationClaimsUntilOriginalEventRetry(t *testing.T) {
	for _, claim := range []string{"none", "current profile", "live alias profile", "deleted profile", "profile beyond first page", "deleted post", "product thumbnail", "retained original registration", "retired derivative registration", "different directory same filename", "foreign origin same path"} {
		t.Run(claim, func(t *testing.T) {
			env := createCustomerServiceAvatarDurableCompensationRegressionEnvironment(t)
			alias := "https://shop.example.test/uploads/" + strings.ReplaceAll(env.key, " ", "%20") + "?v=old#avatar"
			detach := func() {}
			retained := claim != "none" && claim != "different directory same filename" && claim != "foreign origin same path"
			switch claim {
			case "current profile":
				require.NoError(t, env.db.Model(env.profile).Update("avatar", env.provider.uploadReference).Error)
				detach = func() { require.NoError(t, env.db.Model(env.profile).Update("avatar", "").Error) }
			case "live alias profile", "deleted profile", "profile beyond first page", "different directory same filename", "foreign origin same path":
				if claim == "profile beyond first page" {
					for i := 0; i < 260; i++ {
						require.NoError(t, env.db.Create(&user.AgentProfile{AgentID: uuid.NewString(), Name: "Unrelated"}).Error)
					}
				}
				if claim == "different directory same filename" {
					alias = env.provider.GetURL(CustomerServiceAvatarStoragePrefix + "/another/face image.webp")
				}
				if claim == "foreign origin same path" {
					alias = "https://foreign.example.test/uploads/" + env.key
				}
				profile := user.AgentProfile{AgentID: uuid.NewString(), Name: "Historical profile", Avatar: alias, Status: "inactive"}
				require.NoError(t, env.db.Create(&profile).Error)
				if claim == "deleted profile" {
					require.NoError(t, env.db.Delete(&profile).Error)
				}
				detach = func() { require.NoError(t, env.db.Unscoped().Delete(&profile).Error) }
			case "deleted post":
				detach = insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, env.db, env.key, "post_content")
			case "product thumbnail":
				placement := product.ProductMedia{ProductID: 999, URL: "/static/other.png", ThumbnailURL: alias}
				require.NoError(t, env.db.Create(&placement).Error)
				detach = func() { require.NoError(t, env.db.Unscoped().Delete(&placement).Error) }
			case "retained original registration":
				asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, env.db, env.provider, env.key)
				require.NoError(t, env.db.Delete(asset).Error)
				detach = func() { require.NoError(t, env.db.Unscoped().Delete(asset).Error) }
			case "retired derivative registration":
				asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, env.db, env.provider, "catalog/parent.png")
				derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "card", PresetVersion: 1, StorageKey: strings.ReplaceAll(env.key, "/", "\\"), URL: alias}
				require.NoError(t, env.db.Create(&derivative).Error)
				require.NoError(t, env.db.Delete(&derivative).Error)
				detach = func() { require.NoError(t, env.db.Unscoped().Delete(&derivative).Error) }
			}
			uploadFailure := errors.New("upload partially failed")
			env.provider.uploadError = uploadFailure
			err := errors.Join(uploadFailure, env.service.persistFailedAvatarUploadCompensationBeforeOwnerCheckedDeletion(t.Context(), env.profile.ID, env.provider.uploadReference))
			require.ErrorIs(t, err, uploadFailure)
			var events []outbox.Event
			require.NoError(t, env.db.Where("event_type = ?", outbox.EventTypeCustomerServiceAvatarCleanup).Find(&events).Error)
			require.Len(t, events, 1)
			if retained {
				require.Empty(t, env.provider.deletedKeys)
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
				worker := NewOutboxService(repository.NewOutboxRepository(env.db))
				worker.RegisterHandler(outbox.EventTypeCustomerServiceAvatarCleanup, env.handler.Handle)
				result, err := worker.ProcessPending(t.Context(), time.Now().Add(time.Minute), 10)
				require.NoError(t, err)
				require.Equal(t, 1, result.Failed)
				var failed outbox.Event
				require.NoError(t, env.db.First(&failed, events[0].ID).Error)
				require.Equal(t, events[0].EventKey, failed.EventKey)
				require.Equal(t, outbox.EventStatusFailed, failed.Status)
			}
			detach()
			require.NoError(t, env.handler.Handle(t.Context(), events[0]))
			requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
		})
	}
}

func TestLegacyCustomerServiceAvatarReturnedURLCompensationSurvivesCancellationAndPersistsBeforeDeletion(t *testing.T) {
	for _, failure := range []string{"cancellation", "delete failure", "persistence failure", "prior completed task"} {
		t.Run(failure, func(t *testing.T) {
			env := createCustomerServiceAvatarDurableCompensationRegressionEnvironment(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if failure == "cancellation" {
				cancel()
			}
			if failure == "delete failure" {
				env.provider.deleteError = errors.New("delete unavailable")
			}
			if failure == "prior completed task" {
				prior := createCustomerServiceAvatarDurableCleanupEventForRegression(t, env)
				require.NoError(t, env.db.Model(&prior).Update("status", outbox.EventStatusProcessed).Error)
			}
			callback := "observe_legacy_avatar_receipt_" + uuid.NewString()
			persisted := false
			require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
				if tx.Statement.Table == "outbox_events" && failure == "persistence failure" {
					tx.AddError(errors.New("outbox unavailable"))
				}
			}))
			require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callback+"_after", func(tx *gorm.DB) {
				if tx.Statement.Table == "outbox_events" && tx.Error == nil {
					persisted = true
				}
			}))
			t.Cleanup(func() {
				_ = env.db.Callback().Create().Remove(callback)
				_ = env.db.Callback().Create().Remove(callback + "_after")
			})
			env.provider.beforeDelete = func() { require.True(t, persisted) }
			err := env.service.persistFailedAvatarUploadCompensationBeforeOwnerCheckedDeletion(ctx, env.profile.ID, env.provider.uploadReference)
			var events []outbox.Event
			require.NoError(t, env.db.Order("id ASC").Find(&events).Error)
			if failure == "persistence failure" {
				require.ErrorIs(t, err, ErrObjectStorageCleanupUnavailable)
				require.Empty(t, events)
				require.Empty(t, env.provider.deletedKeys)
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
				return
			}
			require.NoError(t, err)
			if failure == "prior completed task" {
				require.Len(t, events, 2)
				require.NotEqual(t, events[0].EventKey, events[1].EventKey)
			} else {
				require.Len(t, events, 1)
			}
			env.provider.deleteError = nil
			require.NoError(t, env.handler.Handle(t.Context(), events[len(events)-1]))
			requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
		})
	}
}

func TestLegacyCustomerServiceAvatarCompensationAndCleanupRejectForeignOrUnresolvableObjectsWithoutDeletionAuthority(t *testing.T) {
	for _, identity := range []string{"foreign host", "other namespace", "namespace only", "missing URL"} {
		t.Run(identity, func(t *testing.T) {
			env := createCustomerServiceAvatarDurableCompensationRegressionEnvironment(t)
			switch identity {
			case "foreign host":
				env.provider.uploadReference = "https://foreign.example.test/uploads/" + env.key
			case "other namespace":
				env.provider.uploadReference = env.provider.GetURL("catalog/face image.webp")
			case "namespace only":
				env.provider.uploadReference = env.provider.GetURL(CustomerServiceAvatarStoragePrefix)
			case "missing URL":
				env.provider.uploadReference = ""
			}
			err := env.service.persistFailedAvatarUploadCompensationBeforeOwnerCheckedDeletion(t.Context(), env.profile.ID, env.provider.uploadReference)
			require.Error(t, err)
			payload, err := json.Marshal(outbox.CustomerServiceAvatarCleanupPayload{URL: env.provider.uploadReference})
			require.NoError(t, err)
			require.NoError(t, env.handler.Handle(t.Context(), outbox.Event{EventType: outbox.EventTypeCustomerServiceAvatarCleanup, Payload: payload}))
			require.Empty(t, env.provider.deletedKeys)
			var count int64
			require.NoError(t, env.db.Model(&outbox.Event{}).Count(&count).Error)
			require.Zero(t, count)
			requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
		})
	}
}

func TestCustomerServiceAvatarNewBindingRequiresConfiguredExistenceValidationAndGenericProfileEditsCannotRestoreStaleAvatar(t *testing.T) {
	for _, scenario := range []string{"missing bytes", "missing validator", "stale profile metadata save", "generic profile creation with avatar", "missing profile metadata save"} {
		t.Run(scenario, func(t *testing.T) {
			env := createCustomerServiceAvatarDurableCompensationRegressionEnvironment(t)
			switch scenario {
			case "missing bytes", "missing validator":
				repo := env.service.userRepo
				if scenario == "missing validator" {
					repo = repository.NewUserRepository(env.db)
				} else {
					require.NoError(t, env.provider.StorageService.Delete(t.Context(), env.key))
				}
				_, err := repo.WithContext(t.Context()).ReplaceCustomerServiceAgentAvatar(*env.profile.UserID, env.provider.uploadReference, nil)
				if scenario == "missing validator" {
					require.ErrorIs(t, err, repository.ErrCustomerServiceAvatarMustUseOwningLifecycle)
				} else {
					require.ErrorIs(t, err, ErrMediaAssetURLUnavailable)
				}
			case "stale profile metadata save":
				require.NoError(t, env.db.Model(env.profile).Update("avatar", env.provider.uploadReference).Error)
				var stale user.AgentProfile
				require.NoError(t, env.db.First(&stale, env.profile.ID).Error)
				require.NoError(t, env.service.Remove(t.Context(), *env.profile.UserID))
				stale.Name = "New support name"
				stale.WhatsApp = "+1 555 0100"
				stale.Status = "inactive"
				require.NoError(t, env.service.userRepo.UpdateCustomerServiceAgentProfile(&stale))
				var saved user.AgentProfile
				require.NoError(t, env.db.First(&saved, stale.ID).Error)
				require.Equal(t, stale.Name, saved.Name)
				require.Equal(t, stale.WhatsApp, saved.WhatsApp)
				require.Equal(t, stale.Status, saved.Status)
				require.Empty(t, saved.Avatar)
			case "generic profile creation with avatar":
				require.ErrorIs(t, env.service.userRepo.CreateCustomerServiceAgentProfile(&user.AgentProfile{AgentID: uuid.NewString(), Avatar: env.provider.uploadReference}), repository.ErrCustomerServiceAvatarMustUseOwningLifecycle)
			case "missing profile metadata save":
				require.NoError(t, env.db.Delete(env.profile).Error)
				require.ErrorIs(t, env.service.userRepo.UpdateCustomerServiceAgentProfile(env.profile), gorm.ErrRecordNotFound)
			}
		})
	}
}

func TestCustomerServiceAvatarSharedMediaAdmissionRetirementAndReplayRespectNamespaceOwnershipAfterLastProfileRemoval(t *testing.T) {
	for _, collision := range []string{"original key", "Windows key", "URL in key", "URL only", "derivative key"} {
		t.Run(collision, func(t *testing.T) {
			env := createCustomerServiceAvatarDurableCompensationRegressionEnvironment(t)
			mediaService := NewMediaService(repository.NewMediaRepository(env.db), env.provider, nil, "https://shop.example.test", 20<<30)
			mediaService.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
			asset := media.MediaAsset{Filename: filepath.Base(env.key), URL: env.provider.uploadReference, StorageKey: env.key, Status: "active", Visibility: "public", MediaType: "image"}
			require.ErrorIs(t, mediaService.repo.CreateAsset(&asset), repository.ErrMediaObjectRegistrationDedicatedOwnerConflict)
			asset.ID = 0
			if collision == "Windows key" {
				asset.StorageKey = strings.ReplaceAll(env.key, "/", "\\")
			}
			if collision == "URL in key" {
				asset.StorageKey = env.provider.uploadReference
			}
			if collision == "URL only" {
				asset.StorageKey = ""
			}
			if collision == "derivative key" {
				asset.StorageKey = "catalog/parent.png"
				asset.URL = env.provider.GetURL(asset.StorageKey)
				writeMediaCleanupOwnershipConflictTestObjects(t, env.root, asset.StorageKey)
			}
			require.NoError(t, env.db.Create(&asset).Error)
			if collision == "derivative key" {
				require.NoError(t, env.db.Create(&media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "card", PresetVersion: 1, StorageKey: env.key, URL: env.provider.uploadReference}).Error)
			}
			deleteError := mediaService.DeleteAsset(context.Background(), asset.ID, MediaAssetDeleteConfirmation(asset.ID))
			if collision == "URL in key" {
				require.Error(t, deleteError)
			} else {
				require.ErrorIs(t, deleteError, ErrMediaAssetInUse)
			}
			require.NoError(t, env.db.Delete(&asset).Error)
			cleanupError := mediaService.cleanupDeletedAsset(t.Context(), asset.ID)
			if collision == "URL in key" {
				require.Error(t, cleanupError)
			} else {
				require.ErrorIs(t, cleanupError, ErrMediaAssetInUse)
			}
			requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
			require.Empty(t, env.provider.deletedKeys)
		})
	}
}

func TestCustomerServiceAvatarPublicAccessRecognizesExactLiveAliasesAndKeepsRetainedHistoryPrivate(t *testing.T) {
	for _, source := range []string{"live alias", "deleted profile", "historical post", "foreign origin"} {
		t.Run(source, func(t *testing.T) {
			env := createCustomerServiceAvatarDurableCompensationRegressionEnvironment(t)
			alias := "https://shop.example.test/uploads/" + strings.ReplaceAll(env.key, " ", "%20") + "?version=old"
			if source == "historical post" {
				insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, env.db, env.key, "post_content")
			} else {
				if source == "foreign origin" {
					alias = "https://foreign.example.test/uploads/" + env.key
				}
				require.NoError(t, env.db.Model(env.profile).Update("avatar", alias).Error)
				if source == "deleted profile" {
					require.NoError(t, env.db.Delete(env.profile).Error)
				}
			}
			allowed, err := env.service.CanServePublicAvatar(t.Context(), env.key)
			require.NoError(t, err)
			require.Equal(t, source == "live alias", allowed)
			allowed, err = env.service.CanServePublicAvatar(t.Context(), CustomerServiceAvatarStoragePrefix+"/different/face image.webp")
			require.NoError(t, err)
			require.False(t, allowed)
			requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
		})
	}
}

func TestCustomerServiceAvatarOwnerChangesRollBackWhenCleanupPersistenceFails(t *testing.T) {
	for _, operation := range []string{"remove", "replace"} {
		t.Run(operation, func(t *testing.T) {
			env := createCustomerServiceAvatarDurableCompensationRegressionEnvironment(t)
			previousKey := CustomerServiceAvatarStoragePrefix + "/current/old.webp"
			writeMediaCleanupOwnershipConflictTestObjects(t, env.root, previousKey)
			previousURL := env.provider.GetURL(previousKey)
			require.NoError(t, env.db.Model(env.profile).Update("avatar", previousURL).Error)
			callback := "reject_avatar_owner_cleanup_event_" + uuid.NewString()
			require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
				if tx.Statement.Table == "outbox_events" {
					tx.AddError(errors.New("cleanup persistence unavailable"))
				}
			}))
			t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
			var err error
			if operation == "remove" {
				err = env.service.Remove(t.Context(), *env.profile.UserID)
			} else {
				_, err = env.service.Upload(t.Context(), *env.profile.UserID, customerServiceAvatarFile(t))
			}
			require.Error(t, err)
			var saved user.AgentProfile
			require.NoError(t, env.db.First(&saved, env.profile.ID).Error)
			require.Equal(t, previousURL, saved.Avatar)
			requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, previousKey)
			requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
			require.Empty(t, env.provider.deletedKeys)
			var events int64
			require.NoError(t, env.db.Model(&outbox.Event{}).Count(&events).Error)
			require.Zero(t, events)
		})
	}
}

func TestCustomerServiceAvatarCleanupReferenceScanFailureRetainsOriginalTaskAndBytesForRetry(t *testing.T) {
	env := createCustomerServiceAvatarDurableCompensationRegressionEnvironment(t)
	event := createCustomerServiceAvatarDurableCleanupEventForRegression(t, env)
	callback := "reject_avatar_reference_scan_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Query().Before("gorm:query").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "customer_service_agent_profiles" {
			tx.AddError(errors.New("avatar reference query unavailable"))
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Query().Remove(callback) })
	worker := NewOutboxService(repository.NewOutboxRepository(env.db))
	worker.RegisterHandler(outbox.EventTypeCustomerServiceAvatarCleanup, env.handler.Handle)
	result, err := worker.ProcessPending(t.Context(), time.Now().Add(time.Minute), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Failed)
	require.Empty(t, env.provider.deletedKeys)
	requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
	var retained outbox.Event
	require.NoError(t, env.db.First(&retained, event.ID).Error)
	require.Equal(t, outbox.EventStatusFailed, retained.Status)
	require.Equal(t, event.EventKey, retained.EventKey)
	require.NoError(t, env.db.Callback().Query().Remove(callback))
	require.NoError(t, env.handler.Handle(t.Context(), event))
	requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
}
