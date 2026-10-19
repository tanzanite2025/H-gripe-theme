package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type businessDraftMediaReceiptTestEnvironment struct {
	db      *gorm.DB
	service *MediaService
	store   storage.StorageService
}

func createBusinessDraftMediaReceiptTestEnvironment(t *testing.T, db *gorm.DB) businessDraftMediaReceiptTestEnvironment {
	t.Helper()
	if db == nil {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		require.NoError(t, err)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.SetMaxOpenConns(1)
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
	for _, model := range []any{&media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{}, &product.ProductCategory{}} {
		if !db.Migrator().HasTable(model) {
			require.NoError(t, db.AutoMigrate(model))
		}
	}
	if db.Name() == "postgres" {
		migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "426_add_business_draft_media_preallocated_upload_receipts.up.sql"))
		require.NoError(t, err)
		require.NoError(t, db.Exec(string(migration)).Error)
	} else {
		require.NoError(t, db.AutoMigrate(&media.BusinessDraftMediaUploadReceipt{}))
	}
	store, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://draft.example.test/uploads"})
	require.NoError(t, err)
	service := NewMediaService(repository.NewMediaRepository(db), store, nil, "https://draft.example.test", 20<<30)
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	return businessDraftMediaReceiptTestEnvironment{db: db, service: service, store: store}
}

func uploadBusinessDraftMediaReceiptTestAsset(t *testing.T, env businessDraftMediaReceiptTestEnvironment) *media.MediaAsset {
	t.Helper()
	asset, err := env.service.UploadBusinessDraftMediaAsset(t.Context(), businessDraftMediaReceiptTestInput(t))
	require.NoError(t, err)
	require.Len(t, asset.Derivatives, 3)
	require.NotNil(t, asset.BusinessDraftUploadExpiresAt)
	return asset
}

func businessDraftMediaReceiptTestInput(t *testing.T) MediaUploadInput {
	t.Helper()
	return MediaUploadInput{File: multipartFileHeader(t, "draft.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploadOrigin: media.MediaUploadOriginAdministratorCatalog,
		UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: 42}}
}

func expireBusinessDraftMediaReceiptTestClock(t *testing.T, env businessDraftMediaReceiptTestEnvironment) {
	t.Helper()
	require.NoError(t, env.db.Model(&media.BusinessDraftMediaUploadReceipt{}).Where("status IN ?", []string{"ready", "uploading"}).Updates(map[string]any{"expires_at": time.Now().UTC().Add(-time.Hour), "next_expiry_attempt_at": time.Now().UTC().Add(-time.Hour)}).Error)
}

func assertBusinessDraftMediaReceiptTestObjectsExist(t *testing.T, env businessDraftMediaReceiptTestEnvironment, asset *media.MediaAsset, exists bool) {
	t.Helper()
	keys := []string{asset.StorageKey}
	for _, derivative := range asset.Derivatives {
		keys = append(keys, derivative.StorageKey)
	}
	for _, key := range keys {
		err := env.store.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(t.Context(), key)
		if exists {
			require.NoError(t, err, key)
		} else {
			require.Error(t, err, key)
		}
	}
}

func replayBusinessDraftMediaReceiptTestCleanupEvents(t *testing.T, env businessDraftMediaReceiptTestEnvironment) {
	t.Helper()
	var events []outbox.Event
	require.NoError(t, env.db.Find(&events).Error)
	require.NotEmpty(t, events)
	handler := NewObjectStorageCleanupOutboxHandler(env.service, nil, nil, nil)
	for _, event := range events {
		require.NoError(t, handler.Handle(t.Context(), event))
		require.NoError(t, handler.Handle(t.Context(), event))
	}
}

func TestBusinessDraftMediaReceiptExpiresUnboundOriginalAndEveryDerivativeThroughDurableOutbox(t *testing.T) {
	env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
	asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
	var receipt media.BusinessDraftMediaUploadReceipt
	require.NoError(t, env.db.First(&receipt).Error)
	require.Equal(t, "ready", receipt.Status)
	require.Equal(t, asset.ID, *receipt.AssetID)
	keys, err := repository.BusinessDraftMediaUploadReceiptObjectKeys(&receipt)
	require.NoError(t, err)
	require.Len(t, keys, 4)
	assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, true)
	expireBusinessDraftMediaReceiptTestClock(t, env)
	result, err := env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Retired, "result=%+v", result)
	require.Zero(t, result.Errors)
	require.ErrorIs(t, env.db.First(&media.BusinessDraftMediaUploadReceipt{}).Error, gorm.ErrRecordNotFound)
	assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, true)
	replayBusinessDraftMediaReceiptTestCleanupEvents(t, env)
	assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, false)
	require.ErrorIs(t, env.db.Unscoped().First(&media.MediaAsset{}, asset.ID).Error, gorm.ErrRecordNotFound)
}

func TestBusinessDraftMediaReceiptBusinessBindingCommitsOrRollsBackWithOwningCategory(t *testing.T) {
	for _, failBusinessWrite := range []bool{false, true} {
		t.Run(map[bool]string{false: "committed", true: "rolled back"}[failBusinessWrite], func(t *testing.T) {
			env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
			asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
			repo := repository.NewProductCategoryRepository(env.db)
			repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(env.service)
			if failBusinessWrite {
				require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register("reject_category_business_write", func(tx *gorm.DB) {
					if tx.Statement.Table == "product_categories" {
						tx.AddError(errors.New("category write rejected"))
					}
				}))
			}
			category := product.ProductCategory{Name: "Draft bound category", Slug: uuid.NewString(), ImageMediaAssetID: &asset.ID}
			err := repo.Create(&category)
			if failBusinessWrite {
				require.Error(t, err)
				require.NoError(t, env.db.Callback().Create().Remove("reject_category_business_write"))
			} else {
				require.NoError(t, err)
			}
			var receipt media.BusinessDraftMediaUploadReceipt
			require.NoError(t, env.db.First(&receipt).Error)
			if failBusinessWrite {
				require.Equal(t, "ready", receipt.Status)
			} else {
				require.Equal(t, "bound", receipt.Status)
			}
			returned, readErr := env.service.GetAsset(asset.ID)
			require.NoError(t, readErr)
			if failBusinessWrite {
				require.NotNil(t, returned.BusinessDraftUploadExpiresAt)
			} else {
				require.Nil(t, returned.BusinessDraftUploadExpiresAt)
			}
			expireBusinessDraftMediaReceiptTestClock(t, env)
			result, err := env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
			require.NoError(t, err)
			if failBusinessWrite {
				require.Equal(t, 1, result.Retired)
			} else {
				require.Zero(t, result.Scanned)
				require.ErrorIs(t, env.service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
				require.NoError(t, env.db.Delete(&category).Error)
				require.NoError(t, env.service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
				assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, false)
			}
		})
	}
}

type businessDraftMediaPartialWriteTestStorage struct {
	storage.StorageService
	beforeWrite             func(string)
	failOnWrite             int
	writes                  int
	ignoreWriteCancellation bool
	returnForeignReference  bool
	corruptWrittenPayload   bool
}

func (s *businessDraftMediaPartialWriteTestStorage) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}
func (s *businessDraftMediaPartialWriteTestStorage) WriteBusinessDraftMediaObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	s.writes++
	if s.beforeWrite != nil {
		s.beforeWrite(key)
	}
	providerContext := ctx
	if s.ignoreWriteCancellation {
		providerContext = context.WithoutCancel(ctx)
	}
	if s.corruptWrittenPayload {
		payload, err := io.ReadAll(reader)
		if err != nil {
			return "", err
		}
		if len(payload) > 0 {
			payload[0] ^= 0xff
		}
		reader = bytes.NewReader(payload)
	}
	reference, err := s.StorageService.(storage.PreallocatedBusinessDraftMediaObjectWriter).WriteBusinessDraftMediaObjectAtPreallocatedKey(providerContext, reader, key)
	if err != nil {
		return reference, err
	}
	if s.writes == s.failOnWrite {
		return "", errors.New("provider wrote bytes but lost response")
	}
	if s.returnForeignReference {
		return "https://foreign.example.test/uploads/unrelated.png", nil
	}
	return reference, nil
}

func TestBusinessDraftMediaReceiptRejectsForeignProviderResponseAndCorruptOriginBytesButRetainsOwnedRecoveryKeys(t *testing.T) {
	for _, scenario := range []string{"foreign response", "corrupt bytes"} {
		t.Run(scenario, func(t *testing.T) {
			env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
			provider := &businessDraftMediaPartialWriteTestStorage{StorageService: env.store, returnForeignReference: scenario == "foreign response", corruptWrittenPayload: scenario == "corrupt bytes"}
			env.service.storage = provider
			asset, err := env.service.UploadBusinessDraftMediaAsset(t.Context(), businessDraftMediaReceiptTestInput(t))
			require.ErrorIs(t, err, ErrMediaAssetURLUnavailable)
			require.Nil(t, asset)
			var receipt media.BusinessDraftMediaUploadReceipt
			require.NoError(t, env.db.First(&receipt).Error)
			keys, err := repository.BusinessDraftMediaUploadReceiptObjectKeys(&receipt)
			require.NoError(t, err)
			require.Len(t, keys, 4)
			result, err := env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Retired)
			replayBusinessDraftMediaReceiptTestCleanupEvents(t, env)
			for _, key := range keys {
				require.Error(t, env.store.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(t.Context(), key))
			}
		})
	}
}

func TestBusinessDraftMediaReceiptMustCommitBeforeAnyProviderWrite(t *testing.T) {
	env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
	provider := &businessDraftMediaPartialWriteTestStorage{StorageService: env.store}
	env.service.storage = provider
	require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register("reject_preallocated_draft_receipt", func(tx *gorm.DB) {
		if tx.Statement.Table == "business_draft_media_upload_receipts" {
			tx.AddError(errors.New("receipt database unavailable"))
		}
	}))
	_, err := env.service.UploadBusinessDraftMediaAsset(t.Context(), businessDraftMediaReceiptTestInput(t))
	require.ErrorContains(t, err, "receipt database unavailable")
	require.Zero(t, provider.writes)
}

func TestBusinessDraftMediaReceiptKeepsEveryPreallocatedKeyAfterPartialWriteOrRegistrationFailure(t *testing.T) {
	for _, scenario := range []string{"first write", "second write", "registration"} {
		t.Run(scenario, func(t *testing.T) {
			env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
			provider := &businessDraftMediaPartialWriteTestStorage{StorageService: env.store}
			if scenario == "first write" {
				provider.failOnWrite = 1
			}
			if scenario == "second write" {
				provider.failOnWrite = 2
			}
			if scenario == "registration" {
				require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register("reject_draft_asset", func(tx *gorm.DB) {
					if tx.Statement.Table == "media_assets" {
						tx.AddError(errors.New("registration rejected"))
					}
				}))
			}
			env.service.storage = provider
			asset, err := env.service.UploadBusinessDraftMediaAsset(t.Context(), businessDraftMediaReceiptTestInput(t))
			require.Error(t, err)
			require.Nil(t, asset)
			require.NoError(t, env.db.Callback().Create().Remove("reject_draft_asset"))
			var receipt media.BusinessDraftMediaUploadReceipt
			require.NoError(t, env.db.First(&receipt).Error)
			require.Equal(t, "uploading", receipt.Status)
			require.Nil(t, receipt.AssetID)
			keys, err := repository.BusinessDraftMediaUploadReceiptObjectKeys(&receipt)
			require.NoError(t, err)
			require.Len(t, keys, 4)
			result, err := env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
			require.NoError(t, err)
			require.Equal(t, 1, result.Retired)
			require.Zero(t, result.Errors)
			replayBusinessDraftMediaReceiptTestCleanupEvents(t, env)
			for _, key := range keys {
				require.Error(t, env.store.(storage.ObjectExistenceVerifier).VerifyStoredObjectExists(t.Context(), key))
			}
		})
	}
}

func TestBusinessDraftMediaReceiptRetainsHistoricalReferencesAndOriginalExpiryCandidateUntilLastDetachment(t *testing.T) {
	env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
	asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
	// Historical rows bypass the new write validator, as restored old rows can.
	category := product.ProductCategory{Name: "Historical consumer", Slug: uuid.NewString(), ImageMediaAssetID: &asset.ID, ImageURL: asset.URL}
	require.NoError(t, env.db.Create(&category).Error)
	expireBusinessDraftMediaReceiptTestClock(t, env)
	result, err := env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Blocked)
	require.Zero(t, result.Retired)
	assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, true)
	require.NoError(t, env.db.Model(&category).Updates(map[string]any{"image_media_asset_id": nil, "image_url": ""}).Error)
	expireBusinessDraftMediaReceiptTestClock(t, env)
	result, err = env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Retired)
	replayBusinessDraftMediaReceiptTestCleanupEvents(t, env)
	assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, false)
}

func TestBusinessDraftMediaReceiptPermanentLibraryUploadIsNeverEnrolledOrExpired(t *testing.T) {
	env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
	asset, err := env.service.UploadAsset(t.Context(), businessDraftMediaReceiptTestInput(t))
	require.NoError(t, err)
	result, err := env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Zero(t, result.Scanned)
	var count int64
	require.NoError(t, env.db.Model(&media.BusinessDraftMediaUploadReceipt{}).Count(&count).Error)
	require.Zero(t, count)
	assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, true)
}

func TestBusinessDraftMediaReceiptBlockedHistoryDoesNotStarveTheNextBoundedExpiryBatch(t *testing.T) {
	env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
	blocked := uploadBusinessDraftMediaReceiptTestAsset(t, env)
	category := product.ProductCategory{Name: "Retained history", Slug: uuid.NewString(), ImageMediaAssetID: &blocked.ID, ImageURL: blocked.URL}
	require.NoError(t, env.db.Create(&category).Error)
	next := uploadBusinessDraftMediaReceiptTestAsset(t, env)
	expireBusinessDraftMediaReceiptTestClock(t, env)
	require.NoError(t, env.db.Model(&media.BusinessDraftMediaUploadReceipt{}).Where("asset_id = ?", blocked.ID).Update("expires_at", time.Now().UTC().Add(-2*time.Hour)).Error)
	result, err := env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 1)
	require.NoError(t, err)
	require.Equal(t, 1, result.Blocked)
	result, err = env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 1)
	require.NoError(t, err)
	require.Equal(t, 1, result.Retired)
	require.ErrorIs(t, env.db.First(&media.MediaAsset{}, next.ID).Error, gorm.ErrRecordNotFound)
	assertBusinessDraftMediaReceiptTestObjectsExist(t, env, blocked, true)
}

func TestBusinessDraftMediaReceiptRejectsGenericRegistrationCleanupAndUnboundDerivativeRebuild(t *testing.T) {
	env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
	asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
	input := businessDraftMediaReceiptTestInput(t)
	input.UploadOrigin = media.MediaUploadOriginAdministratorGeneral
	input.StoragePrefix = storage.BusinessDraftMediaObjectPrefix
	_, uploadErr := env.service.UploadAsset(t.Context(), input)
	require.ErrorIs(t, uploadErr, ErrMediaAssetForbidden)
	clone := *asset
	clone.ID = 0
	clone.Derivatives = nil
	require.ErrorIs(t, env.service.repo.CreateAsset(&clone), repository.ErrBusinessDraftMediaUploadReceiptUnavailable)
	require.Error(t, env.service.cleanupUnregisteredOrRetiredMediaObjectFromDurableEvent(t.Context(), objectCleanupResourceUnregisteredMediaUpload, "untrusted", []string{asset.StorageKey}))
	_, err := env.service.generateAssetDerivativesFromSource(t.Context(), asset, func() (io.ReadCloser, error) {
		object, err := env.store.(storage.ObjectOpener).Open(t.Context(), asset.StorageKey)
		if err != nil {
			return nil, err
		}
		return object.ReadCloser, nil
	}, MediaDerivativePresetDefinitions())
	require.ErrorIs(t, err, repository.ErrBusinessDraftMediaUploadReceiptUnavailable)
	assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, true)
}

func TestBusinessDraftMediaReceiptExpiryAndOutboxCreationRollBackTogether(t *testing.T) {
	env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
	asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
	expireBusinessDraftMediaReceiptTestClock(t, env)
	require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register("reject_draft_cleanup_outbox", func(tx *gorm.DB) {
		if tx.Statement.Table == "outbox_events" {
			tx.AddError(errors.New("outbox unavailable"))
		}
	}))
	result, err := env.service.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.Errors)
	var receipt media.BusinessDraftMediaUploadReceipt
	require.NoError(t, env.db.First(&receipt).Error)
	require.NoError(t, env.db.First(&media.MediaAsset{}, asset.ID).Error)
	assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, true)
}

func TestBusinessDraftMediaReceiptRejectsForgedActorAndUnsupportedOriginBeforeCreatingReceipt(t *testing.T) {
	for _, scenario := range []string{"visitor", "general origin", "dedicated namespace"} {
		t.Run(scenario, func(t *testing.T) {
			env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
			input := businessDraftMediaReceiptTestInput(t)
			switch scenario {
			case "visitor":
				input.UploadActorIdentity = media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorCustomer, AccountID: 42}
			case "general origin":
				input.UploadOrigin = media.MediaUploadOriginAdministratorGeneral
			case "dedicated namespace":
				input.StoragePrefix = "order-evidence/1/2"
			}
			_, err := env.service.UploadBusinessDraftMediaAsset(t.Context(), input)
			require.Error(t, err)
			var count int64
			require.NoError(t, env.db.Model(&media.BusinessDraftMediaUploadReceipt{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestBusinessDraftMediaReceiptCleanupRejectsForeignNamespacesAndMismatchedReceipt(t *testing.T) {
	env := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
	asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
	id, _ := storage.BusinessDraftMediaReceiptIDFromObjectKey(asset.StorageKey)
	for _, key := range []string{asset.StorageKey, "order-evidence/1/2/proof.png", "catalog/product-images/public.png"} {
		require.Error(t, env.service.cleanupExpiredBusinessDraftMediaObjectsFromDurableReceipt(t.Context(), uuid.NewString(), []string{key}))
	}
	require.ErrorIs(t, env.service.cleanupExpiredBusinessDraftMediaObjectsFromDurableReceipt(t.Context(), id, []string{asset.StorageKey}), repository.ErrBusinessDraftMediaUploadReceiptUnavailable)
}
