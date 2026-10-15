package service

import (
	"encoding/json"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMediaUploadCustomerSubmissionsKeepReceiptWithoutAssigningSiteCopyright(t *testing.T) {
	mediaService, _ := newMediaAssetUploadTestService(t)
	require.NoError(t, mediaService.repo.WithTransaction(func(_ *repository.MediaRepository, tx *gorm.DB) error {
		if err := tx.AutoMigrate(&setting.Setting{}); err != nil {
			return err
		}
		settings := NewSettingService(repository.NewSettingRepository(tx), nil, 0)
		require.NoError(t, settings.BatchSet([]setting.Setting{
			{Key: "site_name", Value: "Site name", Type: "string", Locale: "en", Group: "site"},
			{Key: "copyright_holder", Value: "Site owner", Type: "string", Locale: "en", Group: "site"},
			{Key: "copyright_notice", Value: "Site copyright notice", Type: "string", Locale: "en", Group: "site"},
		}))
		mediaService.settings = settings
		for _, actorType := range []media.MediaUploadActorType{media.MediaUploadActorCustomer, media.MediaUploadActorVisitor, media.MediaUploadActorAdministrator} {
			input := MediaUploadInput{File: multipartFileHeader(t, "attachment.png", "image/png", mediaUploadTestPNG(t)),
				UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: actorType}}
			claimJSON, err := mediaService.buildCopyrightClaim(input, strings.Repeat("a", 64))
			require.NoError(t, err)
			var claim CopyrightEvidenceClaim
			require.NoError(t, json.Unmarshal([]byte(claimJSON), &claim))
			require.Equal(t, actorType, claim.UploaderActorType)
			require.Equal(t, strings.Repeat("a", 64), claim.OriginalSHA256)
			require.False(t, claim.ServerReceivedAt.IsZero())
			if actorType == media.MediaUploadActorAdministrator {
				require.Equal(t, "Site owner", claim.RightsHolder)
				require.Equal(t, "Site copyright notice", claim.CopyrightNotice)
			} else {
				require.Empty(t, claim.RightsHolder)
				require.Empty(t, claim.CopyrightNotice)
				require.Empty(t, claim.CopyrightPolicyURL)
			}
		}
		return nil
	}))
}
