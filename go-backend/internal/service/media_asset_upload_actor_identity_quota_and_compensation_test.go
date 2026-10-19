package service

import (
	"encoding/json"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMediaUploadActorIdentityRejectsMissingMalformedAndConflictingIdentities(t *testing.T) {
	cases := []struct {
		name     string
		legacyID uint
		identity media.MediaUploadActorIdentity
	}{
		{name: "missing"},
		{name: "administrator without account", identity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator}},
		{name: "unknown actor", identity: media.MediaUploadActorIdentity{ActorType: "untrusted", AccountID: 9}},
		{name: "visitor without digest", identity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorVisitor}},
		{name: "visitor malformed digest", identity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorVisitor, VisitorSessionHash: strings.Repeat("z", 64)}},
		{name: "visitor noncanonical digest", identity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorVisitor, VisitorSessionHash: strings.Repeat("A", 64)}},
		{name: "visitor with account", identity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorVisitor, AccountID: 9, VisitorSessionHash: strings.Repeat("a", 64)}},
		{name: "account with visitor digest", identity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorCustomer, AccountID: 9, VisitorSessionHash: strings.Repeat("a", 64)}},
		{name: "conflicting legacy id", legacyID: 8, identity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorCustomer, AccountID: 9}},
		{name: "partial identity cannot fall back", legacyID: 9, identity: media.MediaUploadActorIdentity{VisitorSessionHash: strings.Repeat("a", 64)}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mediaService, recordingStorage := newMediaAssetUploadTestService(t)
			asset, err := mediaService.UploadAsset(t.Context(), MediaUploadInput{
				File: multipartFileHeader(t, "attachment.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image",
				UploaderID: testCase.legacyID, UploadActorIdentity: testCase.identity,
			})
			require.ErrorIs(t, err, ErrMediaUploadIdentityRequired)
			require.Nil(t, asset)
			require.Zero(t, recordingStorage.uploadCalls)
		})
	}
}

func TestMediaUploadActorIdentityPersistsTypedActorsAndKeepsLegacyCallerCompatibility(t *testing.T) {
	for _, actorType := range []media.MediaUploadActorType{media.MediaUploadActorLegacyAccount, media.MediaUploadActorAdministrator, media.MediaUploadActorCustomer, media.MediaUploadActorVisitor} {
		t.Run(string(actorType), func(t *testing.T) {
			mediaService, _ := newMediaAssetUploadTestService(t)
			input := MediaUploadInput{File: multipartFileHeader(t, "attachment.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image"}
			switch actorType {
			case media.MediaUploadActorLegacyAccount:
				input.UploaderID = 42
			case media.MediaUploadActorVisitor:
				input.UploadActorIdentity = media.MediaUploadActorIdentity{ActorType: actorType, VisitorSessionHash: strings.Repeat("a", 64)}
			default:
				input.UploadActorIdentity = media.MediaUploadActorIdentity{ActorType: actorType, AccountID: 42}
			}
			asset, err := mediaService.UploadAsset(t.Context(), input)
			require.NoError(t, err)
			stored, err := mediaService.repo.FindAssetByID(asset.ID)
			require.NoError(t, err)
			require.Equal(t, actorType, stored.UploaderActorType)
			require.Equal(t, input.UploadActorIdentity.VisitorSessionHash, stored.UploaderVisitorSessionHash)
			if actorType == media.MediaUploadActorVisitor {
				encoded, err := json.Marshal(stored)
				require.NoError(t, err)
				require.NotContains(t, string(encoded), stored.UploaderVisitorSessionHash)
				require.Zero(t, stored.UploaderID)
			} else {
				require.EqualValues(t, 42, stored.UploaderID)
			}
		})
	}
}

func TestMediaUploadActorQuotaSeparatesAccountNamespacesAndConservativelyCountsLegacyRows(t *testing.T) {
	mediaService, recordingStorage := newMediaAssetUploadTestService(t)
	payload := mediaUploadTestPNG(t)
	mediaService.accountStorageQuotaBytes = 4 + int64(len(payload))
	seedMediaUploadActorQuotaAsset(t, mediaService, "legacy", media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorLegacyAccount, AccountID: 42}, 4, true)
	seedMediaUploadActorQuotaAsset(t, mediaService, "administrator", media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: 42}, 6, false)
	_, err := mediaService.UploadAsset(t.Context(), MediaUploadInput{
		File: multipartFileHeader(t, "administrator.png", "image/png", payload), MediaType: "image",
		UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: 42},
	})
	require.ErrorIs(t, err, ErrMediaAccountStorageQuotaExceeded)
	require.Zero(t, recordingStorage.uploadCalls)
	_, err = mediaService.UploadAsset(t.Context(), MediaUploadInput{
		File: multipartFileHeader(t, "customer.png", "image/png", payload), MediaType: "image",
		UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorCustomer, AccountID: 42},
	})
	require.NoError(t, err)
	_, err = mediaService.UploadAsset(t.Context(), MediaUploadInput{
		File: multipartFileHeader(t, "legacy.png", "image/png", payload), MediaType: "image", UploaderID: 42,
	})
	require.ErrorIs(t, err, ErrMediaAccountStorageQuotaExceeded)
}

func TestMediaUploadVisitorQuotaSeparatesSessionsCountsRetainedFilesAndAllowsExactLimit(t *testing.T) {
	mediaService, recordingStorage := newMediaAssetUploadTestService(t)
	payload := mediaUploadTestPNG(t)
	mediaService.visitorSessionStorageQuotaBytes = int64(len(payload))
	retainedVisitor := media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorVisitor, VisitorSessionHash: strings.Repeat("a", 64)}
	seedMediaUploadActorQuotaAsset(t, mediaService, "retained-visitor", retainedVisitor, int64(len(payload)), true)
	_, err := mediaService.UploadAsset(t.Context(), MediaUploadInput{
		File: multipartFileHeader(t, "retained.png", "image/png", payload), MediaType: "image", UploadActorIdentity: retainedVisitor,
	})
	require.ErrorIs(t, err, ErrMediaVisitorSessionStorageQuotaExceeded)
	require.Zero(t, recordingStorage.uploadCalls)
	otherVisitor := media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorVisitor, VisitorSessionHash: strings.Repeat("b", 64)}
	_, err = mediaService.UploadAsset(t.Context(), MediaUploadInput{
		File: multipartFileHeader(t, "other.png", "image/png", payload), MediaType: "image", UploadActorIdentity: otherVisitor,
	})
	require.NoError(t, err)
	_, err = mediaService.UploadAsset(t.Context(), MediaUploadInput{
		File: multipartFileHeader(t, "overflow.png", "image/png", payload), MediaType: "image", UploadActorIdentity: otherVisitor,
	})
	require.ErrorIs(t, err, ErrMediaVisitorSessionStorageQuotaExceeded)
	require.Equal(t, 1, recordingStorage.uploadCalls)
}

func seedMediaUploadActorQuotaAsset(t *testing.T, mediaService *MediaService, filename string, identity media.MediaUploadActorIdentity, size int64, softDeleted bool) {
	t.Helper()
	asset := &media.MediaAsset{
		Filename: filename, URL: "https://media.example.test/uploads/seed-" + filename, Size: size,
		UploaderActorType: identity.ActorType, UploaderID: identity.AccountID, UploaderVisitorSessionHash: identity.VisitorSessionHash,
	}
	// Synthetic historical quota rows need no physical upload. New production
	// registrations now verify bytes while holding the object ownership lock.
	require.NoError(t, mediaService.repo.WithTransaction(func(_ *repository.MediaRepository, tx *gorm.DB) error { return tx.Create(asset).Error }))
	if softDeleted {
		require.NoError(t, mediaService.repo.WithTransaction(func(_ *repository.MediaRepository, tx *gorm.DB) error { return tx.Delete(asset).Error }))
	}
}
