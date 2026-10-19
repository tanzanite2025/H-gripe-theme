package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t *testing.T, db *gorm.DB, provider storage.StorageService, key string) *media.MediaAsset {
	t.Helper()
	asset := &media.MediaAsset{Filename: filepath.Base(key), URL: provider.GetURL(key), StorageKey: key,
		MediaType: "image", MimeType: "image/png", Size: 5, Status: "active", Visibility: "public"}
	require.NoError(t, db.Create(asset).Error)
	return asset
}

func writeMediaCleanupOwnershipConflictTestObjects(t *testing.T, root string, keys ...string) {
	t.Helper()
	for _, key := range keys {
		name := filepath.Join(root, filepath.FromSlash(key))
		require.NoError(t, os.MkdirAll(filepath.Dir(name), 0o755))
		require.NoError(t, os.WriteFile(name, []byte("image"), 0o644))
	}
}

func TestMediaCleanupRetainsOriginalsAndDerivativesClaimedByOtherRegisteredOwners(t *testing.T) {
	for _, scenario := range []string{
		"original-key", "windows-key", "original-url-only", "storefront-url-only", "retained-original", "inactive-original",
		"original-claimed-by-derivative", "original-claimed-by-retained-derivative", "derivative-claimed-by-original", "derivative-claimed-by-derivative",
		"retired-derivative-claimed-by-original", "private-original-url-only", "conflict-after-first-page",
	} {
		t.Run(scenario, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
			root := t.TempDir()
			service := newMediaDeleteTestService(t, db, root)
			service.siteURL = "https://shop.example.test"
			originalKey, derivativeKey := "catalog/target.png", "media-derivatives/target-card.png"
			if scenario == "private-original-url-only" {
				originalKey = storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/target.png"
			}
			target := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, originalKey)
			originalRoot := root
			if scenario == "private-original-url-only" {
				originalRoot = root + "-private"
				require.NoError(t, db.Model(target).Update("visibility", "private").Error)
				derivativeKey = fmt.Sprintf("%s/%d/card.png", storage.CustomerServiceConversationAttachmentDerivativesPrefix, target.ID)
			}
			derivative := media.MediaAssetDerivative{MediaAssetID: target.ID, Preset: "card", PresetVersion: 1, URL: service.storage.GetURL(derivativeKey), StorageKey: derivativeKey}
			require.NoError(t, db.Create(&derivative).Error)
			writeMediaCleanupOwnershipConflictTestObjects(t, originalRoot, originalKey, derivativeKey)
			conflictingKey := originalKey
			if strings.HasPrefix(scenario, "derivative-claimed") || scenario == "retired-derivative-claimed-by-original" {
				conflictingKey = derivativeKey
			}
			if scenario == "retired-derivative-claimed-by-original" {
				require.NoError(t, db.Delete(&derivative).Error)
			}
			if scenario == "conflict-after-first-page" {
				for i := 0; i < 256; i++ {
					createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, fmt.Sprintf("unrelated/registered-image-%03d.png", i))
				}
			}
			var conflictingAsset *media.MediaAsset
			if strings.Contains(scenario, "claimed-by-derivative") || scenario == "original-claimed-by-retained-derivative" {
				conflictingAsset = createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, "catalog/another-owner.png")
				claim := media.MediaAssetDerivative{MediaAssetID: conflictingAsset.ID, Preset: "card", PresetVersion: 1,
					URL: service.storage.GetURL(conflictingKey) + "?other-owner=1", StorageKey: conflictingKey}
				require.NoError(t, db.Create(&claim).Error)
				if scenario == "original-claimed-by-retained-derivative" {
					require.NoError(t, db.Delete(&claim).Error)
				}
			} else {
				conflictingAsset = createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, "catalog/other-record.png")
				changes := map[string]any{"url": service.storage.GetURL(conflictingKey) + "?historical=2", "storage_key": conflictingKey}
				if scenario == "windows-key" {
					changes["storage_key"] = strings.ReplaceAll(conflictingKey, "/", "\\")
				}
				if scenario == "original-url-only" || scenario == "storefront-url-only" || scenario == "private-original-url-only" {
					changes["storage_key"] = ""
				}
				if scenario == "storefront-url-only" {
					changes["url"] = "https://shop.example.test/uploads/" + conflictingKey + "?version=old"
				}
				if scenario == "inactive-original" {
					changes["status"] = "inactive"
				}
				require.NoError(t, db.Model(conflictingAsset).Updates(changes).Error)
				if scenario == "retained-original" {
					require.NoError(t, db.Delete(conflictingAsset).Error)
				}
			}
			report, err := service.GetAssetReferences(target.ID)
			require.NoError(t, err)
			require.Len(t, report.References, 1)
			require.Equal(t, media.ReferenceCategoryRegistration, report.References[0].Category)
			require.Equal(t, conflictingAsset.ID, report.References[0].ParentResourceID)
			if scenario != "private-original-url-only" {
				review, err := service.ReviewMediaAssetLibraryOwnershipForManualClassification(t.Context(), target.ID)
				require.NoError(t, err)
				require.Equal(t, 1, review.ReferenceCategoryCounts[string(media.ReferenceCategoryRegistration)])
				require.Nil(t, review.SuggestedOwnership, "a duplicate registration is not evidence of a business purpose")
			}
			require.ErrorIs(t, service.DeleteAsset(t.Context(), target.ID, MediaAssetDeleteConfirmation(target.ID)), ErrMediaAssetInUse)
			for _, key := range []string{originalKey, derivativeKey} {
				_, err := os.Stat(filepath.Join(originalRoot, filepath.FromSlash(key)))
				require.NoError(t, err, "all objects must survive a conflict detected before retirement")
			}
			var eventCount int64
			require.NoError(t, db.Model(&outbox.Event{}).Count(&eventCount).Error)
			require.Zero(t, eventCount)
			_, err = service.repo.FindAssetByID(target.ID)
			require.NoError(t, err)
		})
	}
}

type mediaCleanupStorageWithRetainedOriginalForDelayedConflictTest struct {
	storage.StorageService
	deleteError error
}

func (s *mediaCleanupStorageWithRetainedOriginalForDelayedConflictTest) Delete(ctx context.Context, key string) error {
	if s.deleteError != nil {
		return s.deleteError
	}
	return s.StorageService.Delete(ctx, key)
}

func TestMediaCleanupDelayedOutboxRechecksRegistrationConflictsAndRetriesOriginalEventAfterRepair(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{})
	root := t.TempDir()
	service := newMediaDeleteTestService(t, db, root)
	key := "catalog/delayed-cleanup.png"
	asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, key)
	writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
	provider := &mediaCleanupStorageWithRetainedOriginalForDelayedConflictTest{StorageService: service.storage, deleteError: errors.New("synthetic object store outage")}
	service.storage = provider
	require.NoError(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&event).Error)
	duplicate := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, "catalog/temporary-registration.png")
	require.NoError(t, db.Model(duplicate).Updates(map[string]any{"storage_key": key, "url": service.storage.GetURL(key) + "?duplicate=1"}).Error)
	provider.deleteError = nil
	handler := NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil)
	require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(key)))
	require.NoError(t, err)
	var preserved media.MediaAsset
	require.NoError(t, db.First(&preserved, duplicate.ID).Error)
	var tombstone media.MediaAsset
	require.NoError(t, db.Unscoped().First(&tombstone, asset.ID).Error)
	require.True(t, tombstone.DeletedAt.Valid)
	// This is an explicit repair of synthetic test metadata, not automatic
	// winner selection or a production duplicate-registration repair command.
	require.NoError(t, db.Unscoped().Delete(duplicate).Error)
	require.NoError(t, handler.Handle(t.Context(), event))
	require.NoError(t, handler.Handle(t.Context(), event))
	_, err = os.Stat(filepath.Join(root, filepath.FromSlash(key)))
	require.True(t, errors.Is(err, os.ErrNotExist))
	var remainingEvents int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("event_key = ?", event.EventKey).Count(&remainingEvents).Error)
	require.EqualValues(t, 1, remainingEvents)
}

func TestMediaCleanupKeepsSameFilenameAndForeignHostRegistrationsIndependent(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	root := t.TempDir()
	service := newMediaDeleteTestService(t, db, root)
	asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, "catalog/one/image.png")
	neighbor := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, "catalog/two/image.png")
	foreign := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, "catalog/foreign/image.png")
	require.NoError(t, db.Model(foreign).Updates(map[string]any{"storage_key": "", "url": "https://foreign.example.test/uploads/" + asset.StorageKey}).Error)
	writeMediaCleanupOwnershipConflictTestObjects(t, root, asset.StorageKey, neighbor.StorageKey)
	require.NoError(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(asset.StorageKey)))
	require.True(t, errors.Is(err, os.ErrNotExist))
	_, err = os.Stat(filepath.Join(root, filepath.FromSlash(neighbor.StorageKey)))
	require.NoError(t, err)
	require.NoError(t, db.First(&media.MediaAsset{}, neighbor.ID).Error)
	require.NoError(t, db.First(&media.MediaAsset{}, foreign.ID).Error)
}

func TestMediaCleanupRetainsPrivateUnregisteredDerivativeNamespaceObjectsClaimedByAnotherRegistration(t *testing.T) {
	for _, claimType := range []string{"original", "retained-derivative"} {
		t.Run(claimType, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
			root := t.TempDir()
			service := newMediaDeleteTestService(t, db, root)
			asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, storage.CustomerServiceConversationAttachmentOriginalsPrefix+"/private.png")
			require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
			unregisteredKey := fmt.Sprintf("%s/%d/card/unregistered.png", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID)
			writeMediaCleanupOwnershipConflictTestObjects(t, root+"-private", asset.StorageKey, unregisteredKey)
			other := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, "catalog/other-owner.png")
			if claimType == "original" {
				require.NoError(t, db.Model(other).Updates(map[string]any{"storage_key": "", "url": service.storage.GetURL(unregisteredKey)}).Error)
			} else {
				derivative := media.MediaAssetDerivative{MediaAssetID: other.ID, Preset: "card", PresetVersion: 1, URL: service.storage.GetURL(unregisteredKey), StorageKey: unregisteredKey}
				require.NoError(t, db.Create(&derivative).Error)
				require.NoError(t, db.Delete(&derivative).Error)
			}
			require.ErrorIs(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
			// The namespace guard must also run for an already-retired asset.
			require.NoError(t, db.Delete(asset).Error)
			require.ErrorIs(t, service.cleanupDeletedAsset(t.Context(), asset.ID), ErrMediaAssetInUse)
			for _, key := range []string{asset.StorageKey, unregisteredKey} {
				_, err := os.Stat(filepath.Join(root+"-private", filepath.FromSlash(key)))
				require.NoError(t, err)
			}
		})
	}
}

func TestMediaCleanupIgnoresSameOwnerRegistrationsAndAdjacentPrivateDerivativeNamespaces(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	root := t.TempDir()
	service := newMediaDeleteTestService(t, db, root)
	asset := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, storage.CustomerServiceConversationAttachmentOriginalsPrefix+"/owned.png")
	require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
	derivativeKey := fmt.Sprintf("%s/%d/card/owned.png", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID)
	neighborKey := fmt.Sprintf("%s/%d0/card/owned.png", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID)
	for _, preset := range []string{"card", "preview"} {
		require.NoError(t, db.Create(&media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: preset, PresetVersion: 1,
			URL: service.storage.GetURL(derivativeKey) + "?preset=" + preset, StorageKey: derivativeKey}).Error)
	}
	neighbor := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, neighborKey)
	writeMediaCleanupOwnershipConflictTestObjects(t, root+"-private", asset.StorageKey, derivativeKey, neighborKey)
	require.NoError(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	for _, key := range []string{asset.StorageKey, derivativeKey} {
		_, err := os.Stat(filepath.Join(root+"-private", filepath.FromSlash(key)))
		require.ErrorIs(t, err, os.ErrNotExist)
	}
	_, err := os.Stat(filepath.Join(root+"-private", filepath.FromSlash(neighborKey)))
	require.NoError(t, err)
	require.NoError(t, db.First(&media.MediaAsset{}, neighbor.ID).Error)
}

type mediaCleanupStorageDeletionAttemptRecorder struct {
	storage.StorageService
	deleteCalls int
}

func (s *mediaCleanupStorageDeletionAttemptRecorder) Delete(context.Context, string) error {
	s.deleteCalls++
	return nil
}

func TestMediaCleanupRetainsProviderResolvedURLOnlyOriginalAndDerivativeRegistrationConflicts(t *testing.T) {
	for _, providerType := range []storage.StorageType{storage.StorageTypeLocal, storage.StorageTypeS3, storage.StorageTypeOSS} {
		for _, claimType := range []string{"original", "derivative"} {
			t.Run(string(providerType)+"/"+claimType, func(t *testing.T) {
				db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{})
				config := &storage.Config{Type: providerType, LocalPath: t.TempDir(), BaseURL: "https://media.example.test/assets",
					Bucket: "media-public", Region: "us-east-1", AccessKeyID: "synthetic-test-key", SecretAccessKey: "synthetic-test-secret"}
				if providerType == storage.StorageTypeOSS {
					config.Endpoint = "https://oss-cn-hangzhou.aliyuncs.com"
				}
				provider, err := storage.NewStorageService(config)
				require.NoError(t, err)
				recorder := &mediaCleanupStorageDeletionAttemptRecorder{StorageService: provider}
				service := NewMediaService(repository.NewMediaRepository(db), recorder, nil, "https://shop.example.test", 20<<30)
				service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
				key := "catalog/provider-owned.png"
				target := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, key)
				claimURL := provider.GetURL(key) + "?historical=1"
				switch providerType {
				case storage.StorageTypeS3:
					claimURL = "https://media-public.s3.us-east-1.amazonaws.com/" + key
				case storage.StorageTypeOSS:
					claimURL = "https://media-public.oss-cn-hangzhou.aliyuncs.com/" + key
				}
				other := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, "catalog/another-owner.png")
				if claimType == "original" {
					require.NoError(t, db.Model(other).Updates(map[string]any{"url": claimURL, "storage_key": ""}).Error)
				} else {
					require.NoError(t, db.Create(&media.MediaAssetDerivative{MediaAssetID: other.ID, Preset: "card", PresetVersion: 1, URL: claimURL}).Error)
				}
				require.ErrorIs(t, service.DeleteAsset(t.Context(), target.ID, MediaAssetDeleteConfirmation(target.ID)), ErrMediaAssetInUse)
				require.Zero(t, recorder.deleteCalls)
				_, err = service.repo.FindAssetByID(target.ID)
				require.NoError(t, err)
			})
		}
	}
}

func TestPostgreSQLMediaCleanupRetainsConflictingOriginalAndDerivativeRegistrationsOnRetirementAndReplay(t *testing.T) {
	environment, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	db := environment.db
	root := t.TempDir()
	providerStorage, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://media.example.test"})
	require.NoError(t, err)
	service := NewMediaService(repository.NewMediaRepository(db), providerStorage, nil, "", 20<<30)
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	key := "catalog/postgresql-shared-registration.png"
	target := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, key)
	writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
	duplicate := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, "catalog/postgresql-duplicate.png")
	require.NoError(t, db.Model(duplicate).Updates(map[string]any{"url": target.URL + "?retained-registration=1", "storage_key": ""}).Error)
	require.NoError(t, db.Delete(duplicate).Error)
	require.ErrorIs(t, service.DeleteAsset(t.Context(), target.ID, MediaAssetDeleteConfirmation(target.ID)), ErrMediaAssetInUse)
	// Explicitly repair only synthetic test records before exercising the
	// original durable event through a subsequent derivative conflict.
	require.NoError(t, db.Unscoped().Delete(duplicate).Error)
	provider := &mediaCleanupStorageWithRetainedOriginalForDelayedConflictTest{StorageService: service.storage, deleteError: errors.New("synthetic PostgreSQL test storage outage")}
	service.storage = provider
	require.NoError(t, service.DeleteAsset(t.Context(), target.ID, MediaAssetDeleteConfirmation(target.ID)))
	var event outbox.Event
	require.NoError(t, db.Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(target.ID)).First(&event).Error)
	other := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, service.storage, "catalog/postgresql-derivative-owner.png")
	claim := media.MediaAssetDerivative{MediaAssetID: other.ID, Preset: "card", PresetVersion: 1, URL: target.URL + "?derivative-claim=1", StorageKey: key}
	require.NoError(t, db.Create(&claim).Error)
	require.NoError(t, db.Delete(&claim).Error)
	provider.deleteError = nil
	handler := NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil)
	require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
	_, err = os.Stat(filepath.Join(root, filepath.FromSlash(key)))
	require.NoError(t, err)
	require.NoError(t, db.Unscoped().Delete(&claim).Error)
	require.NoError(t, handler.Handle(t.Context(), event))
	require.NoError(t, handler.Handle(t.Context(), event))
	_, err = os.Stat(filepath.Join(root, filepath.FromSlash(key)))
	require.ErrorIs(t, err, os.ErrNotExist)
	require.NoError(t, db.First(&media.MediaAsset{}, other.ID).Error)
}
