package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestMediaObjectRegistrationRequiresConfiguredProviderVerificationForEveryRepositoryWriteEntrypoint(t *testing.T) {
	for _, operation := range []string{"original", "create-derivatives", "replace-derivatives"} {
		t.Run(operation, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
			repo := repository.NewMediaRepository(db)
			owner := &media.MediaAsset{Filename: "owner.png", URL: "https://media.example.test/uploads/owner.png", StorageKey: "owner.png", Status: "active"}
			require.NoError(t, db.Create(owner).Error)
			derivative := media.MediaAssetDerivative{MediaAssetID: owner.ID, Preset: "card", URL: "https://media.example.test/uploads/card.png", StorageKey: "card.png"}
			var err error
			switch operation {
			case "original":
				err = repo.CreateAsset(&media.MediaAsset{Filename: "new.png", URL: "https://media.example.test/uploads/new.png", StorageKey: "new.png"})
			case "create-derivatives":
				err = repo.CreateAssetDerivatives([]media.MediaAssetDerivative{derivative})
			case "replace-derivatives":
				_, err = repo.ReplaceAssetDerivatives([]media.MediaAssetDerivative{derivative})
			}
			require.ErrorIs(t, err, repository.ErrMediaObjectRegistrationVerificationUnavailable)
		})
	}
}

func TestMediaObjectRegistrationRejectsMissingBytesAndPreservesPreviousDerivativeMetadataAtomically(t *testing.T) {
	for _, operation := range []string{"original", "quota-original", "associated-derivative", "create-derivatives", "replace-derivatives"} {
		t.Run(operation, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
			root := t.TempDir()
			assets := newMediaDeleteTestService(t, db, root)
			owner := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, assets.storage, "catalog/existing-owner.png")
			previous := media.MediaAssetDerivative{MediaAssetID: owner.ID, Preset: "card", PresetVersion: 1, URL: assets.storage.GetURL("media-derivatives/previous.png"), StorageKey: "media-derivatives/previous.png"}
			require.NoError(t, db.Create(&previous).Error)
			writeMediaCleanupOwnershipConflictTestObjects(t, root, owner.StorageKey, previous.StorageKey, "media-derivatives/present.png")
			candidate := &media.MediaAsset{Filename: "missing.png", StorageKey: "catalog/missing.png", URL: assets.storage.GetURL("catalog/missing.png"),
				MediaType: "image", Status: "active", Visibility: "public", UploaderID: 701, UploaderActorType: media.MediaUploadActorAdministrator, Size: 1}
			derivatives := []media.MediaAssetDerivative{
				{MediaAssetID: owner.ID, Preset: "present", PresetVersion: 1, URL: assets.storage.GetURL("media-derivatives/present.png"), StorageKey: "media-derivatives/present.png"},
				{MediaAssetID: owner.ID, Preset: "card", PresetVersion: 2, URL: assets.storage.GetURL("media-derivatives/missing.png"), StorageKey: "media-derivatives/missing.png"},
			}
			var err error
			switch operation {
			case "original":
				err = assets.repo.CreateAsset(candidate)
			case "quota-original":
				err = assets.repo.CreateAssetWithinUploadActorStorageQuota(t.Context(), candidate, 20<<30)
			case "associated-derivative":
				candidate.StorageKey, candidate.URL = "catalog/new-owner.png", assets.storage.GetURL("catalog/new-owner.png")
				writeMediaCleanupOwnershipConflictTestObjects(t, root, candidate.StorageKey)
				candidate.Derivatives = []media.MediaAssetDerivative{{Preset: "card", PresetVersion: 1, URL: derivatives[1].URL, StorageKey: derivatives[1].StorageKey}}
				err = assets.repo.CreateAsset(candidate)
			case "create-derivatives":
				err = assets.repo.CreateAssetDerivatives(derivatives)
			case "replace-derivatives":
				_, err = assets.repo.ReplaceAssetDerivatives(derivatives)
			}
			require.ErrorIs(t, err, ErrMediaAssetURLUnavailable)
			var originals, variants int64
			require.NoError(t, db.Unscoped().Model(&media.MediaAsset{}).Count(&originals).Error)
			require.EqualValues(t, 1, originals)
			require.NoError(t, db.Unscoped().Model(&media.MediaAssetDerivative{}).Count(&variants).Error)
			require.EqualValues(t, 1, variants, "no partial derivative registration or retirement is allowed")
			var retained media.MediaAssetDerivative
			require.NoError(t, db.First(&retained, previous.ID).Error)
			require.Equal(t, previous.StorageKey, retained.StorageKey)
		})
	}
}

func TestMediaObjectRegistrationCannotAcquireDedicatedLogoFaviconOrVisualShowcaseObjectsThroughOriginalOrDerivativeWrites(t *testing.T) {
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		for _, operation := range []string{"original-url-only", "associated-derivative", "create-derivative", "replace-derivative"} {
			t.Run(namespace+"/"+operation, func(t *testing.T) {
				db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
				root := t.TempDir()
				assets := newMediaDeleteTestService(t, db, root)
				owner := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, assets.storage, "catalog/owner.png")
				key := namespace + "/dedicated.png"
				writeMediaCleanupOwnershipConflictTestObjects(t, root, owner.StorageKey, key, "catalog/new-owner.png")
				candidate := &media.MediaAsset{Filename: "dedicated.png", URL: assets.storage.GetURL(key)}
				derivative := media.MediaAssetDerivative{MediaAssetID: owner.ID, Preset: "card", PresetVersion: 1, URL: assets.storage.GetURL(key), StorageKey: key}
				var err error
				switch operation {
				case "original-url-only":
					err = assets.repo.CreateAsset(candidate)
				case "associated-derivative":
					candidate.StorageKey, candidate.URL = "catalog/new-owner.png", assets.storage.GetURL("catalog/new-owner.png")
					derivative.MediaAssetID = 0
					candidate.Derivatives = []media.MediaAssetDerivative{derivative}
					err = assets.repo.CreateAsset(candidate)
				case "create-derivative":
					err = assets.repo.CreateAssetDerivatives([]media.MediaAssetDerivative{derivative})
				case "replace-derivative":
					_, err = assets.repo.ReplaceAssetDerivatives([]media.MediaAssetDerivative{derivative})
				}
				require.ErrorIs(t, err, repository.ErrMediaObjectRegistrationDedicatedOwnerConflict)
				var originals, derivatives int64
				require.NoError(t, db.Model(&media.MediaAsset{}).Count(&originals).Error)
				require.NoError(t, db.Model(&media.MediaAssetDerivative{}).Count(&derivatives).Error)
				require.EqualValues(t, 1, originals)
				require.Zero(t, derivatives)
			})
		}
	}
}

func TestMediaObjectRegistrationUsesExactProviderAliasesWindowsKeysAndPrivateObjectLocations(t *testing.T) {
	for _, identity := range []string{"windows-key", "storefront-url-only", "private-url-only", "same-name-neighbor"} {
		t.Run(identity, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{})
			root := t.TempDir()
			assets := newMediaDeleteTestService(t, db, root)
			assets.siteURL = "https://shop.example.test"
			key := "catalog/image.png"
			if identity == "private-url-only" {
				key, root = storage.CustomerServiceConversationAttachmentOriginalsPrefix+"/image.png", root+"-private"
			}
			writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
			candidate := &media.MediaAsset{Filename: "image.png", URL: assets.storage.GetURL(key), StorageKey: key, Status: "active", Visibility: "public"}
			switch identity {
			case "windows-key":
				candidate.StorageKey = strings.ReplaceAll(key, "/", "\\")
			case "storefront-url-only":
				candidate.StorageKey, candidate.URL = "", "https://shop.example.test/uploads/"+key+"?version=1"
			case "private-url-only":
				candidate.StorageKey, candidate.Visibility = "", "private"
			case "same-name-neighbor":
				candidate.StorageKey, candidate.URL = "another/image.png", assets.storage.GetURL("another/image.png")
			}
			err := assets.repo.CreateAsset(candidate)
			if identity == "same-name-neighbor" {
				require.ErrorIs(t, err, ErrMediaAssetURLUnavailable, "a filename is not proof that this exact object exists")
			} else {
				require.NoError(t, err)
				require.NotZero(t, candidate.ID)
			}
		})
	}
}

type mediaObjectRegistrationVerificationStorageWithPrivateProviderFailure struct {
	storage.StorageService
}

func (s *mediaObjectRegistrationVerificationStorageWithPrivateProviderFailure) VerifyStoredObjectExists(context.Context, string) error {
	return errors.New("synthetic provider failure with private-url-and-credential-marker")
}

func TestMediaObjectRegistrationFailsClosedWhenProviderVerificationFailsWithoutLeakingProviderDetails(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{})
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	assets.storage = &mediaObjectRegistrationVerificationStorageWithPrivateProviderFailure{StorageService: assets.storage}
	err := assets.repo.CreateAsset(&media.MediaAsset{Filename: "private.png", StorageKey: "catalog/private.png", URL: assets.storage.GetURL("catalog/private.png")})
	require.ErrorIs(t, err, ErrMediaAssetURLUnavailable)
	require.NotContains(t, err.Error(), "private-url-and-credential-marker")
	var count int64
	require.NoError(t, db.Model(&media.MediaAsset{}).Count(&count).Error)
	require.Zero(t, count)
}
