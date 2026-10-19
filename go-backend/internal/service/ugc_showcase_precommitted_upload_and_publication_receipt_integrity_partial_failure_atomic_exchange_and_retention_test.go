package service

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func loadUGCShowcasePrecommittedWriteReceiptAndPayload(t *testing.T, db *gorm.DB) (outbox.Event, outbox.ObjectStorageCleanupPayload) {
	t.Helper()
	var event outbox.Event
	require.NoError(t, db.Where("event_key LIKE ?", repository.UGCShowcasePrecommittedObjectWriteReceiptEventKeyPrefix+"%").First(&event).Error)
	var payload outbox.ObjectStorageCleanupPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	return event, payload
}

func uploadUGCShowcaseTwoImagesThroughPrecommittedEntryPoint(ctx context.Context, t *testing.T, env ugcShowcaseDurableCompensationRegressionEnvironment) (*ugcshowcase.UGCShowcase, error) {
	t.Helper()
	return env.service.UploadPhotos(ctx, 7, 91, []*multipart.FileHeader{
		createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "first.webp"),
		createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "second.webp"),
	}, map[string]string{"nickname": "Rider"})
}

func processUGCShowcaseWriteReceiptsWithExistingWorker(t *testing.T, env ugcShowcaseDurableCompensationRegressionEnvironment, now time.Time) OutboxDispatchResult {
	t.Helper()
	worker := NewOutboxService(repository.NewOutboxRepository(env.db))
	worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, NewObjectStorageCleanupOutboxHandler(nil, nil, nil, env.service).Handle)
	result, err := worker.ProcessPending(t.Context(), now, 100)
	require.NoError(t, err)
	return result
}

func TestUGCShowcasePrecommittedUploadAndPublicationExchangeReceiptsForRealOwnersAndPreservePendingPrivacy(t *testing.T) {
	env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
	configureEligibleOrderForUGCShowcaseUploadCompensationRegression(t, env)
	item, err := uploadUGCShowcaseTwoImagesThroughPrecommittedEntryPoint(t.Context(), t, env)
	require.NoError(t, err)
	var pending []string
	require.NoError(t, json.Unmarshal(item.Images, &pending))
	require.Len(t, pending, 2)
	var count int64
	require.NoError(t, env.db.Model(&outbox.Event{}).Count(&count).Error)
	require.Zero(t, count)
	for _, key := range pending {
		require.True(t, strings.HasPrefix(key, showcasePendingUploadPrefix+"/"))
		require.NotContains(t, key, "://")
		_, err := os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(key)))
		require.NoError(t, err)
		_, err = os.Stat(filepath.Join(env.publicRoot, filepath.FromSlash(key)))
		require.True(t, os.IsNotExist(err))
		access, err := env.service.PublicImageAccess(t.Context(), key)
		require.NoError(t, err)
		require.False(t, access.Allowed)
	}
	require.NoError(t, env.service.Approve(t.Context(), item.ID))
	var saved ugcshowcase.UGCShowcase
	require.NoError(t, env.db.First(&saved, item.ID).Error)
	require.Equal(t, ugcshowcase.StatusApproved, saved.Status)
	var approved []string
	require.NoError(t, json.Unmarshal(saved.Images, &approved))
	require.Len(t, approved, 2)
	var events []outbox.Event
	require.NoError(t, env.db.Find(&events).Error)
	require.Len(t, events, 1)
	require.NotContains(t, events[0].EventKey, repository.UGCShowcasePrecommittedObjectWriteReceiptEventKeyPrefix)
	result := processUGCShowcaseWriteReceiptsWithExistingWorker(t, env, time.Now().Add(time.Minute))
	require.Equal(t, 1, result.Processed)
	for _, key := range pending {
		requireUGCShowcaseRegressionObjectExists(t, env, key, false)
	}
	for _, key := range approved {
		requireUGCShowcaseRegressionObjectExists(t, env, key, true)
		access, err := env.service.PublicImageAccess(t.Context(), key)
		require.NoError(t, err)
		require.True(t, access.Allowed)
	}
}

func TestUGCShowcasePrecommittedPendingUploadFailuresKeepAllExactKeysAndNeverBindOrDeleteReturnedIdentities(t *testing.T) {
	for _, failure := range []string{"first partial write", "later partial write", "missing response", "foreign response", "different pending key", "cancellation", "owner insert", "receipt exchange", "wrong digest"} {
		t.Run(failure, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			configureEligibleOrderForUGCShowcaseUploadCompensationRegression(t, env)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			unrelated := "showcase/pending/historical/do-not-delete.webp"
			writeUGCShowcaseRegressionObject(t, env, unrelated)
			switch failure {
			case "first partial write", "later partial write":
				env.provider.uploadError = errors.New("write succeeded but response lost")
				env.provider.uploadErrorAt = 1
				if failure == "later partial write" {
					env.provider.uploadErrorAt = 2
				}
			case "missing response":
				env.provider.uploadReferences = []string{""}
			case "foreign response":
				env.provider.uploadReferences = []string{"https://foreign.example.test/uploads/" + unrelated}
			case "different pending key":
				env.provider.uploadReferences = []string{env.provider.GetURL(unrelated)}
			case "cancellation":
				env.provider.afterUpload = cancel
			case "wrong digest":
				env.provider.afterUpload = func() {
					key := env.provider.preallocatedPendingKeys[len(env.provider.preallocatedPendingKeys)-1]
					require.NoError(t, os.WriteFile(filepath.Join(env.privateRoot, filepath.FromSlash(key)), []byte("corrupt"), 0o600))
				}
			case "owner insert":
				callback := "reject_ugc_precommitted_owner_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "showcases" {
						tx.AddError(errors.New("owner insert failure"))
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
			case "receipt exchange":
				callback := "reject_ugc_precommitted_receipt_exchange_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Delete().Before("gorm:delete").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "outbox_events" {
						tx.AddError(errors.New("receipt exchange failure"))
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Delete().Remove(callback) })
			}
			item, err := uploadUGCShowcaseTwoImagesThroughPrecommittedEntryPoint(ctx, t, env)
			require.Error(t, err)
			require.Nil(t, item)
			event, payload := loadUGCShowcasePrecommittedWriteReceiptAndPayload(t, env.db)
			require.Len(t, payload.ObjectKeys, 2)
			require.Equal(t, outbox.EventStatusPending, event.Status)
			require.Empty(t, env.provider.deletedKeys)
			var owners int64
			require.NoError(t, env.db.Model(&ugcshowcase.UGCShowcase{}).Count(&owners).Error)
			require.Zero(t, owners)
			for _, key := range env.provider.preallocatedPendingKeys {
				require.Contains(t, payload.ObjectKeys, key)
			}
			result := processUGCShowcaseWriteReceiptsWithExistingWorker(t, env, time.Now().Add(6*time.Minute))
			require.Equal(t, 1, result.Processed)
			for _, key := range payload.ObjectKeys {
				requireUGCShowcaseRegressionObjectExists(t, env, key, false)
			}
			requireUGCShowcaseRegressionObjectExists(t, env, unrelated, true)
		})
	}
}

func TestUGCShowcasePrecommittedReceiptFailureUnsupportedWriterAndExternalTransactionPreventObjectIO(t *testing.T) {
	for _, operation := range []string{"upload", "publish"} {
		for _, failure := range []string{"receipt insert", "external owner transaction", "external outbox transaction", "missing outbox"} {
			t.Run(operation+"/"+failure, func(t *testing.T) {
				env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
				configureEligibleOrderForUGCShowcaseUploadCompensationRegression(t, env)
				writeUGCShowcaseRegressionObject(t, env, env.key)
				owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key)
				if failure == "receipt insert" {
					callback := "reject_ugc_precommitted_insert_" + uuid.NewString()
					require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "outbox_events" {
							tx.AddError(errors.New("receipt unavailable"))
						}
					}))
					t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
				}
				if strings.HasPrefix(failure, "external") {
					tx := env.db.Begin()
					require.NoError(t, tx.Error)
					defer tx.Rollback()
					if failure == "external owner transaction" {
						env.service.repo = env.service.repo.WithTx(tx)
					} else {
						env.service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(tx))
					}
				}
				if failure == "missing outbox" {
					env.service.ConfigureObjectCleanupOutbox(nil)
				}
				if operation == "upload" {
					_, err := uploadUGCShowcaseTwoImagesThroughPrecommittedEntryPoint(t.Context(), t, env)
					require.Error(t, err)
				} else {
					require.Error(t, env.service.Approve(t.Context(), owner.ID))
				}
				require.Empty(t, env.provider.preallocatedPendingKeys)
				require.Empty(t, env.provider.copiedKeys)
			})
		}
	}
	t.Run("unsupported private writer", func(t *testing.T) {
		env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
		configureEligibleOrderForUGCShowcaseUploadCompensationRegression(t, env)
		env.service.storage = struct{ storage.StorageService }{env.provider}
		_, err := uploadUGCShowcaseTwoImagesThroughPrecommittedEntryPoint(t.Context(), t, env)
		require.ErrorIs(t, err, ErrShowcaseStorageUnavailable)
		require.Empty(t, env.provider.preallocatedPendingKeys)
	})
}

func TestUGCShowcasePrecommittedPublicationFailureRetainsAllDestinationsAndPendingOwnerForRetry(t *testing.T) {
	for _, failure := range []string{"first copy", "later copy", "cancellation", "wrong copied bytes", "owner status", "source cleanup event", "receipt exchange"} {
		t.Run(failure, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			keys := []string{env.key, "showcase/pending/historical/second.webp"}
			for _, key := range keys {
				writeUGCShowcaseRegressionObject(t, env, key)
			}
			owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, keys...)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			removeInjection := func() {}
			switch failure {
			case "first copy", "later copy":
				env.provider.copyError = errors.New("copy wrote then failed")
				env.provider.copyErrorAt = 1
				if failure == "later copy" {
					env.provider.copyErrorAt = 2
				}
			case "cancellation":
				env.provider.beforeCopy = cancel
			case "wrong copied bytes":
				env.provider.beforeCopy = func() {
					require.NoError(t, os.WriteFile(filepath.Join(env.privateRoot, filepath.FromSlash(env.key)), []byte("changed after source digest"), 0o600))
				}
			case "owner status":
				callback := "reject_ugc_precommitted_publication_owner_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "showcases" {
						tx.AddError(errors.New("owner update failure"))
					}
				}))
				removeInjection = func() { _ = env.db.Callback().Update().Remove(callback) }
				t.Cleanup(removeInjection)
			case "source cleanup event":
				callback := "reject_ugc_source_cleanup_after_copy_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
					if event, ok := tx.Statement.Dest.(*outbox.Event); ok && strings.Contains(string(event.Payload), `"resource_type":"ugc_showcase"`) {
						tx.AddError(errors.New("source cleanup insert failure"))
					}
				}))
				removeInjection = func() { _ = env.db.Callback().Create().Remove(callback) }
				t.Cleanup(removeInjection)
			case "receipt exchange":
				callback := "reject_ugc_publication_receipt_exchange_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Delete().Before("gorm:delete").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "outbox_events" {
						tx.AddError(errors.New("exchange failed"))
					}
				}))
				removeInjection = func() { _ = env.db.Callback().Delete().Remove(callback) }
				t.Cleanup(removeInjection)
			}
			require.Error(t, env.service.Approve(ctx, owner.ID))
			removeInjection()
			event, payload := loadUGCShowcasePrecommittedWriteReceiptAndPayload(t, env.db)
			require.Len(t, payload.ObjectKeys, 2)
			for _, key := range env.provider.copiedKeys {
				require.Contains(t, payload.ObjectKeys, key)
			}
			var saved ugcshowcase.UGCShowcase
			require.NoError(t, env.db.First(&saved, owner.ID).Error)
			require.Equal(t, ugcshowcase.StatusPending, saved.Status)
			require.JSONEq(t, string(owner.Images), string(saved.Images))
			require.Empty(t, env.provider.deletedKeys)
			result := processUGCShowcaseWriteReceiptsWithExistingWorker(t, env, time.Now().Add(6*time.Minute))
			require.Equal(t, 1, result.Processed)
			for _, key := range payload.ObjectKeys {
				requireUGCShowcaseRegressionObjectExists(t, env, key, false)
			}
			for _, key := range keys {
				requireUGCShowcaseRegressionObjectExists(t, env, key, true)
			}
			var completed outbox.Event
			require.NoError(t, env.db.First(&completed, event.ID).Error)
			require.Equal(t, event.EventKey, completed.EventKey)
			env.provider.copyError, env.provider.beforeCopy = nil, nil
			require.NoError(t, env.service.Approve(t.Context(), owner.ID))
		})
	}
}

func TestUGCShowcasePrecommittedFailedWriteRetainsHistoricalReferencesAndReplaysOriginalTask(t *testing.T) {
	for _, operation := range []string{"upload", "publish"} {
		t.Run(operation, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			if operation == "upload" {
				configureEligibleOrderForUGCShowcaseUploadCompensationRegression(t, env)
				env.provider.uploadError = errors.New("write response lost")
				_, err := uploadUGCShowcaseTwoImagesThroughPrecommittedEntryPoint(t.Context(), t, env)
				require.Error(t, err)
			} else {
				writeUGCShowcaseRegressionObject(t, env, env.key)
				owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key)
				env.provider.copyError = errors.New("copy response lost")
				require.Error(t, env.service.Approve(t.Context(), owner.ID))
			}
			original, payload := loadUGCShowcasePrecommittedWriteReceiptAndPayload(t, env.db)
			key := payload.ObjectKeys[0]
			if operation == "upload" {
				key = env.provider.preallocatedPendingKeys[0]
			}
			detach := insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, env.db, key, "post_content")
			result := processUGCShowcaseWriteReceiptsWithExistingWorker(t, env, time.Now().Add(6*time.Minute))
			require.Equal(t, 1, result.Failed)
			retained, _ := loadUGCShowcasePrecommittedWriteReceiptAndPayload(t, env.db)
			require.Equal(t, original.ID, retained.ID)
			require.Equal(t, original.Payload, retained.Payload)
			requireUGCShowcaseRegressionObjectExists(t, env, key, true)
			detach()
			result = processUGCShowcaseWriteReceiptsWithExistingWorker(t, env, retained.AvailableAt.Add(time.Second))
			require.Equal(t, 1, result.Processed)
			requireUGCShowcaseRegressionObjectExists(t, env, key, false)
		})
	}
}

func TestUGCShowcasePrecommittedPublicationDeduplicatesCopiesAndNeverCleansReusedApprovedObjects(t *testing.T) {
	env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
	approved := "showcase/approved/historical/approved.webp"
	for _, key := range []string{env.key, approved} {
		writeUGCShowcaseRegressionObject(t, env, key)
	}
	owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key, env.key, approved)
	require.NoError(t, env.service.Approve(t.Context(), owner.ID))
	require.Len(t, env.provider.copiedKeys, 1)
	var saved ugcshowcase.UGCShowcase
	require.NoError(t, env.db.First(&saved, owner.ID).Error)
	var images []string
	require.NoError(t, json.Unmarshal(saved.Images, &images))
	require.Len(t, images, 3)
	require.Equal(t, images[0], images[1])
	require.Equal(t, approved, images[2])
	result := processUGCShowcaseWriteReceiptsWithExistingWorker(t, env, time.Now().Add(time.Minute))
	require.Equal(t, 1, result.Processed)
	requireUGCShowcaseRegressionObjectExists(t, env, approved, true)
	for _, key := range images {
		require.True(t, strings.HasPrefix(key, showcaseApprovedUploadPrefix+"/"))
	}
}

func TestUGCShowcasePrecommittedPublicationKeepsDistinctSourcesWithTheSameLegacySuffixSeparate(t *testing.T) {
	env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
	pending, legacy := "showcase/pending/collision/ride.webp", "collision/ride.webp"
	for _, key := range []string{pending, legacy} {
		writeUGCShowcaseRegressionObject(t, env, key)
	}
	require.NoError(t, os.WriteFile(filepath.Join(env.privateRoot, filepath.FromSlash(pending)), []byte("private source"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(env.publicRoot, filepath.FromSlash(legacy)), []byte("legacy source"), 0o600))
	owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, pending, legacy)
	require.NoError(t, env.service.Approve(t.Context(), owner.ID))
	current, err := env.service.Get(owner.ID)
	require.NoError(t, err)
	var published []string
	require.NoError(t, json.Unmarshal(current.Images, &published))
	require.Len(t, published, 2)
	require.NotEqual(t, published[0], published[1])
	for index, expected := range []string{"private source", "legacy source"} {
		data, err := os.ReadFile(filepath.Join(env.publicRoot, filepath.FromSlash(published[index])))
		require.NoError(t, err)
		require.Equal(t, expected, string(data))
	}
	result := processUGCShowcaseWriteReceiptsWithExistingWorker(t, env, time.Now().Add(time.Minute))
	require.Equal(t, 1, result.Processed)
	requireUGCShowcaseRegressionObjectExists(t, env, legacy, true)
}
