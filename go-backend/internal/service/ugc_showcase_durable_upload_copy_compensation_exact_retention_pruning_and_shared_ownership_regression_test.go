package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type ugcShowcasePartialUploadCopyAndCleanupRegressionStorage struct {
	storage.StorageService
	uploadReferences        []string
	uploadError             error
	uploadCalls             int
	preallocatedPendingKeys []string
	uploadErrorAt           int
	copyError               error
	copyErrorAt             int
	copiedKeys              []string
	deletedKeys             []string
	deleteError             error
	afterUpload             func()
	beforeCopy              func()
	beforeDelete            func()
}

func (s *ugcShowcasePartialUploadCopyAndCleanupRegressionStorage) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func (s *ugcShowcasePartialUploadCopyAndCleanupRegressionStorage) WritePrivateObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	s.uploadCalls++
	s.preallocatedPendingKeys = append(s.preallocatedPendingKeys, key)
	reference, err := s.StorageService.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(ctx, reader, key)
	if err != nil {
		return reference, err
	}
	if s.afterUpload != nil {
		s.afterUpload()
	}
	if len(s.uploadReferences) >= s.uploadCalls {
		reference = s.uploadReferences[s.uploadCalls-1]
	}
	if s.uploadError != nil && (s.uploadErrorAt == 0 || s.uploadCalls == s.uploadErrorAt) {
		return reference, s.uploadError
	}
	return reference, nil
}

func (s *ugcShowcasePartialUploadCopyAndCleanupRegressionStorage) UploadWithPrefixPrivate(context.Context, *multipart.FileHeader, string) (string, error) {
	s.uploadCalls++
	if s.afterUpload != nil {
		s.afterUpload()
	}
	reference := ""
	if s.uploadCalls <= len(s.uploadReferences) {
		reference = s.uploadReferences[s.uploadCalls-1]
	}
	if s.uploadError != nil && (s.uploadErrorAt == 0 || s.uploadErrorAt == s.uploadCalls) {
		return reference, s.uploadError
	}
	return reference, nil
}

func (s *ugcShowcasePartialUploadCopyAndCleanupRegressionStorage) CopyObject(ctx context.Context, source, destination string) error {
	if s.beforeCopy != nil {
		s.beforeCopy()
	}
	s.copiedKeys = append(s.copiedKeys, destination)
	if err := s.StorageService.CopyObject(ctx, source, destination); err != nil {
		return err
	}
	if s.copyError != nil && (s.copyErrorAt == 0 || s.copyErrorAt == len(s.copiedKeys)) {
		return s.copyError
	}
	return nil
}

func (s *ugcShowcasePartialUploadCopyAndCleanupRegressionStorage) Delete(ctx context.Context, reference string) error {
	if s.beforeDelete != nil {
		s.beforeDelete()
	}
	key, err := s.ObjectKey(reference)
	if err != nil {
		return err
	}
	s.deletedKeys = append(s.deletedKeys, key)
	if s.deleteError != nil {
		return s.deleteError
	}
	return s.StorageService.Delete(ctx, key)
}

func (s *ugcShowcasePartialUploadCopyAndCleanupRegressionStorage) VerifyStoredObjectExists(ctx context.Context, key string) error {
	return s.StorageService.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(ctx, key)
}

type ugcShowcaseDurableCompensationRegressionEnvironment struct {
	db          *gorm.DB
	service     *UGCShowcaseService
	provider    *ugcShowcasePartialUploadCopyAndCleanupRegressionStorage
	publicRoot  string
	privateRoot string
	key         string
}

func configureUGCShowcaseDurableCompensationRegressionEnvironment(t *testing.T, db *gorm.DB, provider storage.StorageService, publicRoot, privateRoot string) ugcShowcaseDurableCompensationRegressionEnvironment {
	t.Helper()
	controlled := &ugcShowcasePartialUploadCopyAndCleanupRegressionStorage{StorageService: provider}
	service := NewUGCShowcaseService(repository.NewUGCShowcaseRepository(db), controlled, "https://shop.example.test")
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	return ugcShowcaseDurableCompensationRegressionEnvironment{db, service, controlled, publicRoot, privateRoot, "showcase/pending/historical/ride image.webp"}
}

func createUGCShowcaseDurableCompensationRegressionEnvironment(t *testing.T) ugcShowcaseDurableCompensationRegressionEnvironment {
	t.Helper()
	db, _ := newTestShowcaseService(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &post.Post{}, &product.Product{}, &product.ProductMedia{}))
	publicRoot, privateRoot := filepath.Join(t.TempDir(), "public"), filepath.Join(t.TempDir(), "private")
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: publicRoot, PrivateLocalPath: privateRoot, BaseURL: "https://media.example.test/assets", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	return configureUGCShowcaseDurableCompensationRegressionEnvironment(t, db, provider, publicRoot, privateRoot)
}

func createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t *testing.T, env ugcShowcaseDurableCompensationRegressionEnvironment, status string, references ...string) *ugcshowcase.UGCShowcase {
	t.Helper()
	encoded, err := json.Marshal(references)
	require.NoError(t, err)
	item := &ugcshowcase.UGCShowcase{UserID: 7, Status: status, Images: encoded, CreatedAt: time.Now().Add(-10 * 24 * time.Hour)}
	require.NoError(t, env.db.Create(item).Error)
	return item
}

func writeUGCShowcaseRegressionObject(t *testing.T, env ugcShowcaseDurableCompensationRegressionEnvironment, key string) {
	t.Helper()
	root := env.publicRoot
	if storage.IsPrivateObjectKey(key) {
		root = env.privateRoot
	}
	writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
}

func createUGCShowcaseRegressionCleanupEvent(t *testing.T, env ugcShowcaseDurableCompensationRegressionEnvironment, ownerID uint, key string, failedUpload bool) outbox.Event {
	t.Helper()
	resourceType, resourceID := objectCleanupResourceUGCShowcase, fmt.Sprint(ownerID)
	if failedUpload {
		resourceType, resourceID = objectCleanupResourceFailedUGCShowcaseUpload, "failed-upload-"+uuid.NewString()
	}
	event, err := newObjectStorageCleanupEvent(resourceType, resourceID, outbox.AggregateTypeUGCShowcase, resourceID, []string{key})
	require.NoError(t, err)
	require.NoError(t, repository.NewOutboxRepository(env.db).CreateEvent(event))
	return *event
}

func requireUGCShowcaseRegressionObjectExists(t *testing.T, env ugcShowcaseDurableCompensationRegressionEnvironment, key string, exists bool) {
	t.Helper()
	err := env.provider.VerifyStoredObjectExists(t.Context(), key)
	if exists {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}
}

func TestUGCShowcaseCleanupRetainsExactHistoricalClaimsAndRetriesOriginalEventAfterLastDetachment(t *testing.T) {
	for _, claim := range []string{"none", "other pending", "other rejected", "escaped beyond first page", "private origin", "retained post", "retained original", "retired derivative", "foreign origin", "same filename elsewhere"} {
		t.Run(claim, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			writeUGCShowcaseRegressionObject(t, env, env.key)
			owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusRejected, env.key)
			event := createUGCShowcaseRegressionCleanupEvent(t, env, owner.ID, env.key, false)
			detach := func() {}
			retained := claim != "none" && claim != "foreign origin" && claim != "same filename elsewhere"
			switch claim {
			case "other pending", "other rejected", "escaped beyond first page", "private origin", "foreign origin", "same filename elsewhere":
				if claim == "escaped beyond first page" {
					for i := 0; i < 260; i++ {
						createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusRejected, env.key+fmt.Sprint(i))
					}
				}
				reference := "https://shop.example.test/uploads/" + strings.ReplaceAll(env.key, " ", "%20") + "?v=old#image"
				if claim == "private origin" {
					reference = "https://private.example.test/" + strings.ReplaceAll(env.key, " ", "%20")
				}
				if claim == "foreign origin" {
					reference = "https://foreign.example.test/uploads/" + strings.ReplaceAll(env.key, " ", "%20")
				}
				if claim == "same filename elsewhere" {
					reference = "showcase/pending/another/ride image.webp"
				}
				status := ugcshowcase.StatusPending
				if claim == "other rejected" {
					status = ugcshowcase.StatusRejected
				}
				other := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, status, reference)
				if claim == "escaped beyond first page" {
					require.NoError(t, env.db.Model(other).Update("images", strings.ReplaceAll(string(other.Images), "/", `\/`)).Error)
				}
				detach = func() { require.NoError(t, env.db.Delete(other).Error) }
			case "retained post":
				detach = insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, env.db, env.key, "post_content")
			case "retained original":
				asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, env.db, env.provider, env.key)
				require.NoError(t, env.db.Delete(asset).Error)
				detach = func() { require.NoError(t, env.db.Unscoped().Delete(asset).Error) }
			case "retired derivative":
				asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, env.db, env.provider, "catalog/parent.png")
				derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "old", PresetVersion: 1, StorageKey: strings.ReplaceAll(env.key, "/", `\`), URL: "/uploads/" + env.key}
				require.NoError(t, env.db.Create(&derivative).Error)
				require.NoError(t, env.db.Delete(&derivative).Error)
				detach = func() { require.NoError(t, env.db.Unscoped().Delete(&derivative).Error) }
			}
			handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, env.service)
			err := handler.Handle(t.Context(), event)
			if retained {
				require.ErrorIs(t, err, ErrMediaAssetInUse)
				require.Empty(t, env.provider.deletedKeys)
				requireUGCShowcaseRegressionObjectExists(t, env, env.key, true)
				detach()
				require.NoError(t, handler.Handle(t.Context(), event))
			} else {
				require.NoError(t, err)
			}
			requireUGCShowcaseRegressionObjectExists(t, env, env.key, false)
			var saved ugcshowcase.UGCShowcase
			require.NoError(t, env.db.First(&saved, owner.ID).Error)
			require.JSONEq(t, "[]", string(saved.Images))
		})
	}
}

func configureEligibleOrderForUGCShowcaseUploadCompensationRegression(t *testing.T, env ugcShowcaseDurableCompensationRegressionEnvironment) {
	t.Helper()
	completedAt := time.Now()
	require.NoError(t, env.db.Create(&order.Order{ID: 91, UserID: 7, OrderNumber: "UGC-" + uuid.NewString(), Status: "completed", CompletedAt: &completedAt, TotalAmountMinor: 10000, Currency: "USD"}).Error)
	env.service.ConfigureUploadEligibility(NewUGCShowcaseUploadEligibilityService(repository.NewOrderRepository(env.db)))
}

func (s *UGCShowcaseService) uploadLegacyPendingShowcaseImageWithReturnedIdentityForRegression(ctx context.Context, file *multipart.FileHeader) (string, error) {
	privateUploader, ok := s.storage.(storage.PrivateObjectUploader)
	if !ok {
		return "", ErrShowcaseStorageUnavailable
	}
	reference, uploadErr := privateUploader.UploadWithPrefixPrivate(ctx, file, showcasePendingUploadPrefix)
	if strings.TrimSpace(reference) == "" && uploadErr != nil {
		return "", uploadErr
	}
	key, identityErr := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, showcasePendingUploadPrefix, reference)
	if identityErr != nil {
		return "", errors.Join(uploadErr, fmt.Errorf("%w: pending upload returned no trusted object identity", ErrObjectStorageCleanupUnavailable))
	}
	return key, uploadErr
}

func attemptLegacyUGCShowcaseReturnedUploadCompensationForRegression(ctx context.Context, svc *UGCShowcaseService, files []*multipart.FileHeader) error {
	var keys []string
	for _, file := range files {
		key, err := svc.uploadLegacyPendingShowcaseImageWithReturnedIdentityForRegression(ctx, file)
		if key != "" {
			keys = append(keys, key)
		}
		if err != nil {
			return errors.Join(err, svc.persistFailedUGCShowcaseImageUploadCompensationBeforeOwnerCheckedCleanup(ctx, keys))
		}
	}
	encoded, _ := json.Marshal(keys)
	if err := svc.createPendingSubmission(ctx, &ugcshowcase.UGCShowcase{UserID: 7, Status: ugcshowcase.StatusPending, Images: encoded}); err != nil {
		return errors.Join(err, svc.persistFailedUGCShowcaseImageUploadCompensationBeforeOwnerCheckedCleanup(ctx, keys))
	}
	return nil
}

func TestLegacyUGCShowcaseReturnedUploadCompensationPersistsBeforeDeletionAndSurvivesCancellationOrPersistenceFailure(t *testing.T) {
	for _, failure := range []string{"partial upload", "canceled partial upload", "canceled successful upload", "owner insert failure", "retained owner", "outbox failure", "delete failure", "foreign result", "wrong namespace", "missing result"} {
		t.Run(failure, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			configureEligibleOrderForUGCShowcaseUploadCompensationRegression(t, env)
			writeUGCShowcaseRegressionObject(t, env, env.key)
			env.provider.uploadReferences = []string{env.provider.GetURL(env.key)}
			env.provider.uploadError = errors.New("provider partial upload failed")
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			switch failure {
			case "canceled partial upload":
				env.provider.afterUpload = cancel
			case "canceled successful upload":
				env.provider.uploadError = nil
				env.provider.afterUpload = cancel
			case "owner insert failure":
				env.provider.uploadError = nil
				require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register("reject_ugc_owner", func(tx *gorm.DB) {
					if tx.Statement.Table == "showcases" {
						tx.AddError(errors.New("owner insert failed"))
					}
				}))
			case "retained owner":
				createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key)
			case "outbox failure":
				require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register("reject_ugc_compensation", func(tx *gorm.DB) {
					if tx.Statement.Table == "outbox_events" {
						tx.AddError(errors.New("outbox insert failed"))
					}
				}))
			case "delete failure":
				env.provider.deleteError = errors.New("provider deletion failed")
			case "foreign result":
				env.provider.uploadReferences = []string{"https://foreign.example.test/uploads/" + env.key}
			case "wrong namespace":
				env.provider.uploadReferences = []string{env.provider.GetURL("site-logo/ride.webp")}
			case "missing result":
				env.provider.uploadReferences = nil
			}
			err := attemptLegacyUGCShowcaseReturnedUploadCompensationForRegression(ctx, env.service, []*multipart.FileHeader{{Filename: "ride.webp", Size: 100}})
			require.Error(t, err)
			var events []outbox.Event
			require.NoError(t, env.db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).Find(&events).Error)
			noEvent := failure == "outbox failure" || failure == "foreign result" || failure == "wrong namespace" || failure == "missing result"
			if noEvent {
				require.Empty(t, events)
				require.Empty(t, env.provider.deletedKeys)
			} else {
				require.Len(t, events, 1)
			}
			retained := noEvent || failure == "retained owner" || failure == "delete failure"
			requireUGCShowcaseRegressionObjectExists(t, env, env.key, retained)
			if failure == "outbox failure" {
				require.ErrorIs(t, err, ErrObjectStorageCleanupUnavailable)
			}
			if failure == "delete failure" {
				env.provider.deleteError = nil
				require.NoError(t, NewObjectStorageCleanupOutboxHandler(nil, nil, nil, env.service).Handle(t.Context(), events[0]))
				requireUGCShowcaseRegressionObjectExists(t, env, env.key, false)
			}
		})
	}
}

func TestUGCShowcasePublicationPartiallyWrittenDestinationAndOwnerTransactionFailureUseDurableCompensation(t *testing.T) {
	for _, failure := range []string{"first copy", "later copy", "owner update", "cleanup event persistence", "prune persistence"} {
		t.Run(failure, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			writeUGCShowcaseRegressionObject(t, env, env.key)
			otherKey := "showcase/pending/second.webp"
			writeUGCShowcaseRegressionObject(t, env, otherKey)
			owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key, otherKey)
			switch failure {
			case "first copy", "later copy":
				env.provider.copyError = errors.New("copy wrote bytes then failed")
				env.provider.copyErrorAt = 1
				if failure == "later copy" {
					env.provider.copyErrorAt = 2
				}
			case "owner update":
				require.NoError(t, env.db.Callback().Update().Before("gorm:update").Register("reject_publication_owner", func(tx *gorm.DB) {
					if tx.Statement.Table == "showcases" {
						tx.AddError(errors.New("owner update failed"))
					}
				}))
			case "cleanup event persistence":
				require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register("reject_source_cleanup_event", func(tx *gorm.DB) {
					if tx.Statement.Table == "outbox_events" {
						var event *outbox.Event
						if candidate, ok := tx.Statement.Dest.(*outbox.Event); ok {
							event = candidate
						}
						if event != nil && strings.Contains(string(event.Payload), `"resource_type":"ugc_showcase"`) {
							tx.AddError(errors.New("source task insert failed"))
						}
					}
				}))
			case "prune persistence":
				require.NoError(t, env.service.Reject(t.Context(), owner.ID, "reject"))
				require.NoError(t, env.db.Callback().Update().Before("gorm:update").Register("reject_prune", func(tx *gorm.DB) {
					if tx.Statement.Table == "showcases" {
						tx.AddError(errors.New("prune update failed"))
					}
				}))
				event := createUGCShowcaseRegressionCleanupEvent(t, env, owner.ID, env.key, false)
				require.Error(t, NewObjectStorageCleanupOutboxHandler(nil, nil, nil, env.service).Handle(t.Context(), event))
				require.NoError(t, env.db.Callback().Update().Remove("reject_prune"))
				require.NoError(t, NewObjectStorageCleanupOutboxHandler(nil, nil, nil, env.service).Handle(t.Context(), event))
				return
			}
			require.Error(t, env.service.Approve(t.Context(), owner.ID))
			require.NotEmpty(t, env.provider.copiedKeys)
			var precommitted outbox.Event
			require.NoError(t, env.db.Where("event_key LIKE ?", repository.UGCShowcasePrecommittedObjectWriteReceiptEventKeyPrefix+"%").First(&precommitted).Error)
			require.NoError(t, NewObjectStorageCleanupOutboxHandler(nil, nil, nil, env.service).Handle(t.Context(), precommitted))
			for _, key := range env.provider.copiedKeys {
				requireUGCShowcaseRegressionObjectExists(t, env, key, false)
			}
			requireUGCShowcaseRegressionObjectExists(t, env, env.key, true)
			var saved ugcshowcase.UGCShowcase
			require.NoError(t, env.db.First(&saved, owner.ID).Error)
			require.Equal(t, ugcshowcase.StatusPending, saved.Status)
			var events []outbox.Event
			require.NoError(t, env.db.Find(&events).Error)
			require.Len(t, events, 1)
			require.Contains(t, string(events[0].Payload), objectCleanupResourceFailedUGCShowcaseUpload)
		})
	}
}

func TestUGCShowcaseImageBindingAndPublicAccessUseOwningNamespaceExactIdentityAndAllPages(t *testing.T) {
	for _, boundary := range []string{"missing bytes", "unconfigured repository", "foreign namespace", "stale rejected restore", "public beyond first page", "public rejected before approved", "escaped rejected expiry"} {
		t.Run(boundary, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			writeUGCShowcaseRegressionObject(t, env, env.key)
			switch boundary {
			case "missing bytes", "unconfigured repository", "foreign namespace":
				key := env.key
				if boundary == "missing bytes" {
					require.NoError(t, env.provider.StorageService.Delete(t.Context(), key))
				}
				if boundary == "foreign namespace" {
					key = "site-logo/ride.webp"
					writeUGCShowcaseRegressionObject(t, env, key)
				}
				encoded, _ := json.Marshal([]string{key})
				item := &ugcshowcase.UGCShowcase{UserID: 7, Status: ugcshowcase.StatusPending, Images: encoded}
				repo := env.service.repo
				if boundary == "unconfigured repository" {
					repo = repository.NewUGCShowcaseRepository(env.db)
				}
				require.Error(t, repo.WithContext(t.Context()).Create(item))
			case "stale rejected restore":
				owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key)
				require.NoError(t, env.service.Reject(t.Context(), owner.ID, "reject"))
				require.ErrorIs(t, env.service.repo.UpdateStatus(owner.ID, ugcshowcase.StatusPending, ""), ErrShowcaseInvalidTransition)
			case "public beyond first page", "public rejected before approved":
				key := "showcase/approved/historical/ride image.webp"
				if boundary == "public beyond first page" {
					for i := 0; i < 260; i++ {
						createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusRejected, key+fmt.Sprint(i))
					}
				}
				createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusRejected, key)
				owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusApproved, "https://shop.example.test/uploads/"+strings.ReplaceAll(key, " ", "%20"))
				require.NoError(t, env.db.Model(owner).Update("images", strings.ReplaceAll(string(owner.Images), "/", `\/`)).Error)
				access, err := env.service.PublicImageAccess(t.Context(), key)
				require.NoError(t, err)
				require.True(t, access.Allowed)
				require.NoError(t, env.db.Delete(owner).Error)
				access, err = env.service.PublicImageAccess(t.Context(), key)
				require.NoError(t, err)
				require.False(t, access.Allowed)
			case "escaped rejected expiry":
				owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusRejected, env.key)
				require.NoError(t, env.db.Model(owner).Update("images", strings.ReplaceAll(string(owner.Images), "/", `\/`)).Error)
				result, err := env.service.CleanupExpiredPendingImages(t.Context(), time.Now(), 24*time.Hour, 1)
				require.NoError(t, err)
				require.Equal(t, 1, result.DeletedPendingImages)
			}
		})
	}
}

func TestUGCShowcaseSharedMediaCannotAdmitRetireOrReplayCleanupOfPendingAndApprovedObjects(t *testing.T) {
	for _, namespace := range []string{showcasePendingUploadPrefix, showcaseApprovedUploadPrefix} {
		for _, operation := range []string{"register", "retire", "old original cleanup", "failed compensation"} {
			t.Run(namespace+"/"+operation, func(t *testing.T) {
				env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
				key := namespace + "/orphan.webp"
				writeUGCShowcaseRegressionObject(t, env, key)
				shared := NewMediaService(repository.NewMediaRepository(env.db), env.provider, nil, "https://shop.example.test", 20<<30)
				shared.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
				switch operation {
				case "register":
					require.Error(t, shared.repo.CreateAsset(&media.MediaAsset{URL: env.provider.GetURL(key), StorageKey: key, Filename: "orphan.webp", Status: "active"}))
				case "retire":
					asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, env.db, env.provider, key)
					require.ErrorIs(t, shared.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
				case "old original cleanup":
					asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, env.db, env.provider, key)
					require.NoError(t, env.db.Delete(asset).Error)
					event, err := newObjectStorageCleanupEvent(objectCleanupResourceMediaAsset, fmt.Sprint(asset.ID), outbox.AggregateTypeMediaAsset, fmt.Sprint(asset.ID), []string{key})
					require.NoError(t, err)
					require.Error(t, NewObjectStorageCleanupOutboxHandler(shared, nil, nil, nil).Handle(t.Context(), *event))
				case "failed compensation":
					event, err := newObjectStorageCleanupEvent(objectCleanupResourceUnregisteredMediaUpload, uuid.NewString(), outbox.AggregateTypeMediaAsset, "failed", []string{key})
					require.NoError(t, err)
					require.Error(t, NewObjectStorageCleanupOutboxHandler(shared, nil, nil, nil).Handle(t.Context(), *event))
				}
				requireUGCShowcaseRegressionObjectExists(t, env, key, true)
				require.Empty(t, env.provider.deletedKeys)
			})
		}
	}
}

func TestUGCShowcaseOriginalOutboxEventRemainsRetryableForReferencesProviderAndQueryFailures(t *testing.T) {
	for _, namespace := range []string{showcasePendingUploadPrefix, showcaseApprovedUploadPrefix} {
		for _, failure := range []string{"reference", "provider", "query"} {
			t.Run(namespace+"/"+failure, func(t *testing.T) {
				env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
				key := namespace + "/retry.webp"
				writeUGCShowcaseRegressionObject(t, env, key)
				event := createUGCShowcaseRegressionCleanupEvent(t, env, 0, key, true)
				var owner *ugcshowcase.UGCShowcase
				if failure == "reference" {
					owner = createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusApproved, key)
				}
				if failure == "provider" {
					env.provider.deleteError = errors.New("provider unavailable")
				}
				if failure == "query" {
					require.NoError(t, env.db.Callback().Query().Before("gorm:query").Register("fail_ugc_retention_query", func(tx *gorm.DB) {
						if tx.Statement.Table == "showcases" {
							tx.AddError(errors.New("retention query unavailable"))
						}
					}))
				}
				worker := NewOutboxService(repository.NewOutboxRepository(env.db))
				worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, NewObjectStorageCleanupOutboxHandler(nil, nil, nil, env.service).Handle)
				result, err := worker.ProcessPending(t.Context(), time.Now().Add(time.Minute), 10)
				require.NoError(t, err)
				require.Equal(t, 1, result.Failed)
				var failed outbox.Event
				require.NoError(t, env.db.First(&failed, event.ID).Error)
				require.Equal(t, event.EventKey, failed.EventKey)
				require.Equal(t, outbox.EventStatusFailed, failed.Status)
				requireUGCShowcaseRegressionObjectExists(t, env, key, true)
				if owner != nil {
					require.NoError(t, env.db.Delete(owner).Error)
				}
				env.provider.deleteError = nil
				if failure == "query" {
					require.NoError(t, env.db.Callback().Query().Remove("fail_ugc_retention_query"))
				}
				result, err = worker.ProcessPending(t.Context(), time.Now().Add(time.Hour), 10)
				require.NoError(t, err)
				require.Equal(t, 1, result.Processed)
				var processed outbox.Event
				require.NoError(t, env.db.First(&processed, event.ID).Error)
				require.Equal(t, event.EventKey, processed.EventKey)
				require.Equal(t, outbox.EventStatusProcessed, processed.Status)
				requireUGCShowcaseRegressionObjectExists(t, env, key, false)
			})
		}
	}
}

func TestLegacyUGCShowcaseReturnedMultipleUploadFailureAndCompletedSameKeyTaskCannotLoseCompensation(t *testing.T) {
	for _, scenario := range []string{"second partial upload", "same key after completed event"} {
		t.Run(scenario, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			configureEligibleOrderForUGCShowcaseUploadCompensationRegression(t, env)
			writeUGCShowcaseRegressionObject(t, env, env.key)
			files := []*multipart.FileHeader{{Filename: "one.webp", Size: 100}}
			env.provider.uploadReferences = []string{env.provider.GetURL(env.key)}
			env.provider.uploadError = errors.New("partial upload")
			if scenario == "second partial upload" {
				key := "showcase/pending/second.webp"
				writeUGCShowcaseRegressionObject(t, env, key)
				env.provider.uploadReferences = append(env.provider.uploadReferences, env.provider.GetURL(key))
				env.provider.uploadErrorAt = 2
				files = append(files, &multipart.FileHeader{Filename: "two.webp", Size: 100})
			} else {
				prior := createUGCShowcaseRegressionCleanupEvent(t, env, 0, env.key, true)
				require.NoError(t, env.db.Model(&prior).Update("status", outbox.EventStatusProcessed).Error)
			}
			persisted := false
			require.NoError(t, env.db.Callback().Create().After("gorm:create").Register("observe_ugc_compensation", func(tx *gorm.DB) {
				if tx.Statement.Table == "outbox_events" && tx.Error == nil {
					persisted = true
				}
			}))
			env.provider.beforeDelete = func() { require.True(t, persisted) }
			err := attemptLegacyUGCShowcaseReturnedUploadCompensationForRegression(t.Context(), env.service, files)
			require.Error(t, err)
			for _, reference := range env.provider.uploadReferences {
				key, err := env.provider.ObjectKey(reference)
				require.NoError(t, err)
				requireUGCShowcaseRegressionObjectExists(t, env, key, false)
			}
			var events []outbox.Event
			require.NoError(t, env.db.Order("id ASC").Find(&events).Error)
			if scenario == "same key after completed event" {
				require.Len(t, events, 2)
				require.NotEqual(t, events[0].EventKey, events[1].EventKey)
			} else {
				require.Len(t, events, 1)
			}
		})
	}
}

func TestUGCShowcaseFailedPublicationRetainsCopiedBytesWhenOutboxUnavailableOrHistoricalClaimExists(t *testing.T) {
	for _, scenario := range []string{"outbox unavailable", "historical copied destination"} {
		t.Run(scenario, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			writeUGCShowcaseRegressionObject(t, env, env.key)
			owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key)
			env.provider.copyError = errors.New("copy wrote bytes before failure")
			var historical *ugcshowcase.UGCShowcase
			require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register("control_ugc_copy_compensation", func(tx *gorm.DB) {
				if tx.Statement.Table != "outbox_events" {
					return
				}
				if scenario == "outbox unavailable" {
					tx.AddError(errors.New("outbox unavailable"))
					return
				}
				var payload outbox.ObjectStorageCleanupPayload
				if candidate, ok := tx.Statement.Dest.(*outbox.Event); ok {
					if err := json.Unmarshal(candidate.Payload, &payload); err != nil {
						tx.AddError(err)
						return
					}
					if payload.ResourceType != objectCleanupResourceFailedUGCShowcaseUpload {
						return
					}
					encoded, _ := json.Marshal(payload.ObjectKeys)
					historical = &ugcshowcase.UGCShowcase{UserID: 9, Status: ugcshowcase.StatusApproved, Images: encoded}
					tx.AddError(tx.Session(&gorm.Session{NewDB: true}).Create(historical).Error)
				}
			}))
			err := env.service.Approve(t.Context(), owner.ID)
			require.Error(t, err)
			if scenario == "outbox unavailable" {
				require.Empty(t, env.provider.copiedKeys)
				require.Empty(t, env.provider.deletedKeys)
				return
			}
			require.Len(t, env.provider.copiedKeys, 1)
			key := env.provider.copiedKeys[0]
			requireUGCShowcaseRegressionObjectExists(t, env, key, true)
			require.Empty(t, env.provider.deletedKeys)
			if scenario == "outbox unavailable" {
				require.ErrorIs(t, err, ErrObjectStorageCleanupUnavailable)
			} else {
				require.NotNil(t, historical)
				require.NoError(t, env.db.Delete(historical).Error)
				var event outbox.Event
				require.NoError(t, env.db.First(&event).Error)
				require.NoError(t, NewObjectStorageCleanupOutboxHandler(nil, nil, nil, env.service).Handle(t.Context(), event))
				requireUGCShowcaseRegressionObjectExists(t, env, key, false)
			}
		})
	}
}

func TestUGCShowcaseRejectAndExpiryCannotCommitWithoutTheirCleanupEvent(t *testing.T) {
	for _, operation := range []string{"reject", "expire"} {
		t.Run(operation, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			writeUGCShowcaseRegressionObject(t, env, env.key)
			owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, env.key)
			require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register("fail_ugc_retirement_event", func(tx *gorm.DB) {
				if tx.Statement.Table == "outbox_events" {
					tx.AddError(errors.New("outbox unavailable"))
				}
			}))
			if operation == "reject" {
				require.Error(t, env.service.Reject(t.Context(), owner.ID, "reject"))
			} else {
				_, err := env.service.CleanupExpiredPendingImages(t.Context(), time.Now(), 24*time.Hour, 10)
				require.Error(t, err)
			}
			var saved ugcshowcase.UGCShowcase
			require.NoError(t, env.db.First(&saved, owner.ID).Error)
			require.Equal(t, ugcshowcase.StatusPending, saved.Status)
			require.JSONEq(t, string(owner.Images), string(saved.Images))
			require.Empty(t, env.provider.deletedKeys)
		})
	}
}

func TestUGCShowcasePublicationRejectsOtherPrivateLifecyclesForeignSourcesAndDirectoryRoots(t *testing.T) {
	for _, reference := range []string{"order-evidence/1/2/receipt.webp", "showcase/pending", "showcase/approved", "https://foreign.example.test/uploads/showcase/pending/ride.webp"} {
		t.Run(reference, func(t *testing.T) {
			env := createUGCShowcaseDurableCompensationRegressionEnvironment(t)
			owner := createHistoricalUGCShowcaseImageOwnerForRetentionRegression(t, env, ugcshowcase.StatusPending, reference)
			require.ErrorIs(t, env.service.Approve(t.Context(), owner.ID), ErrShowcaseImagesInvalid)
			require.Empty(t, env.provider.copiedKeys)
			require.Empty(t, env.provider.deletedKeys)
		})
	}
}
