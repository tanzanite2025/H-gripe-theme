package service

import (
	"strings"
	"testing"

	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestMediaRetentionScansFormalAndLegacyArticleCoverColumnsWithoutChangingSchema(t *testing.T) {
	for _, columns := range [][]string{{"featured_image"}, {"featured_img"}, {"featured_image", "featured_img"}} {
		t.Run(strings.Join(columns, "_and_"), func(t *testing.T) {
			env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, nil)
			declarations := ""
			for _, column := range columns {
				declarations += ", " + column + " TEXT"
			}
			require.NoError(t, env.db.Exec("CREATE TABLE posts(id INTEGER PRIMARY KEY,title TEXT,content TEXT,deleted_at DATETIME"+declarations+")").Error)
			for index, column := range columns {
				asset := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/"+column+".png")
				require.NoError(t, env.db.Exec("INSERT INTO posts(id,title,content,deleted_at,"+column+") VALUES(?,'Retained article','',CURRENT_TIMESTAMP,?)", index+1, asset.URL).Error)
				report, err := env.assets.GetAssetReferences(asset.ID)
				require.NoError(t, err)
				require.Len(t, report.References, 1)
				require.Equal(t, column, report.References[0].Field)
				require.Equal(t, "retained_deleted_parent", report.References[0].UsageState)
				require.ErrorIs(t, env.assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
			}
			configuration := &storage.Config{Type: storage.StorageTypeLocal, LocalPath: "unused-column-audit", BaseURL: "https://media.example.test"}
			audit := NewHistoricalMediaReadOnlyInventoryService(env.db, configuration, configuration, "https://shop.example.test")
			inventory, err := audit.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
			require.NoError(t, err)
			coverSources := 0
			for _, source := range inventory.Sources {
				if source.Table == "posts" && source.Field != "content" {
					require.Contains(t, columns, source.Field)
					require.True(t, source.Complete)
					require.False(t, source.Unavailable)
					require.Equal(t, 1, source.RowsScanned)
					coverSources++
				}
			}
			require.Equal(t, len(columns), coverSources)
			for _, column := range []string{"featured_image", "featured_img"} {
				require.Equal(t, strings.Contains(strings.Join(columns, ","), column), env.db.Migrator().HasColumn("posts", column))
			}
		})
	}
}

func TestMediaRetentionScansWarrantyImagesWithAndWithoutOptionalVideoColumn(t *testing.T) {
	for _, hasVideo := range []bool{false, true} {
		name := "formal_images_only"
		if hasVideo {
			name = "images_and_application_video"
		}
		t.Run(name, func(t *testing.T) {
			env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, nil)
			extra := ""
			if hasVideo {
				extra = ", video_url TEXT"
			}
			require.NoError(t, env.db.Exec("CREATE TABLE warranty_claims(id INTEGER PRIMARY KEY,images TEXT,deleted_at DATETIME"+extra+")").Error)
			image := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/warranty.png")
			require.NoError(t, env.db.Exec("INSERT INTO warranty_claims(id,images) VALUES(1,?)", `["`+image.URL+`"]`).Error)
			require.ErrorIs(t, env.assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)), ErrMediaAssetInUse)
			const dedicatedKey = "site-logo/historical/warranty.png"
			require.NoError(t, env.db.Exec("INSERT INTO warranty_claims(id,images) VALUES(2,?)", `["/uploads/`+dedicatedKey+`"]`).Error)
			mediaRepository := repository.NewMediaRepository(env.db)
			mediaRepository.ConfigurePublicBusinessMediaReferenceIdentityResolver(env.assets)
			state, err := mediaRepository.FindDedicatedPublicObjectBusinessReferenceState(dedicatedKey)
			require.NoError(t, err)
			require.Len(t, state.RetainedReferences, 1)
			require.Equal(t, "images", state.RetainedReferences[0].Field)
			require.False(t, state.HasPublishedConsumer)
			if hasVideo {
				video := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/warranty-video.mp4")
				require.NoError(t, env.db.Exec("UPDATE warranty_claims SET video_url=? WHERE id=1", video.URL).Error)
				report, err := env.assets.GetAssetReferences(video.ID)
				require.NoError(t, err)
				require.Len(t, report.References, 1)
				require.Equal(t, "video_url", report.References[0].Field)
				require.ErrorIs(t, env.assets.DeleteAsset(t.Context(), video.ID, MediaAssetDeleteConfirmation(video.ID)), ErrMediaAssetInUse)
				require.NoError(t, env.db.Exec("UPDATE warranty_claims SET images='[]', video_url=? WHERE id=2", "/uploads/"+dedicatedKey).Error)
				state, err = mediaRepository.FindDedicatedPublicObjectBusinessReferenceState(dedicatedKey)
				require.NoError(t, err)
				require.Len(t, state.RetainedReferences, 1)
				require.Equal(t, "video_url", state.RetainedReferences[0].Field)
			}
			require.Equal(t, hasVideo, env.db.Migrator().HasColumn("warranty_claims", "video_url"))
		})
	}
}
