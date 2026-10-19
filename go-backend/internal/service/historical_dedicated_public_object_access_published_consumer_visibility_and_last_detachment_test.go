package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/setting"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func insertHistoricalDedicatedObjectConsumerForPublicVisibilityTest(t *testing.T, db *gorm.DB, key, scenario string) func() {
	t.Helper()
	alias := "https://shop.example.test/uploads/" + strings.ReplaceAll(key, " ", "%20") + "?version=old"
	switch {
	case strings.HasPrefix(scenario, "product_"):
		item := product.Product{Name: "Historical product", Slug: "historical-product", Status: "active"}
		if scenario == "product_description" {
			item.Description = `<img src="` + alias + `">`
		}
		if scenario == "product_inactive" {
			item.Status = "inactive"
		}
		require.NoError(t, db.Create(&item).Error)
		if scenario == "product_deleted" {
			require.NoError(t, db.Delete(&item).Error)
		}
		if scenario == "product_description" {
			return func() { require.NoError(t, db.Unscoped().Model(&item).Update("description", "").Error) }
		}
		placement := product.ProductMedia{ProductID: item.ID, URL: alias, IsVisible: true}
		if scenario == "product_thumbnail" {
			placement.URL, placement.ThumbnailURL = "/static/unrelated.png", alias
		}
		if scenario == "product_poster" {
			placement.URL, placement.PosterURL = "/static/unrelated.webm", alias
		}
		if scenario == "product_orphan" {
			placement.ProductID += 1000
		}
		require.NoError(t, db.Create(&placement).Error)
		if scenario == "product_hidden" {
			require.NoError(t, db.Model(&placement).Update("is_visible", false).Error)
		}
		if scenario == "product_deleted_placement" {
			require.NoError(t, db.Delete(&placement).Error)
		}
		return func() { require.NoError(t, db.Unscoped().Delete(&placement).Error) }
	case strings.HasPrefix(scenario, "post_"):
		item := post.Post{Title: "Historical article", Slug: "historical-article", Status: "published", Content: `<img src="` + alias + `">`}
		if scenario == "post_draft" {
			item.Status = "draft"
		}
		require.NoError(t, db.Create(&item).Error)
		if scenario == "post_deleted" {
			require.NoError(t, db.Delete(&item).Error)
		}
		return func() { require.NoError(t, db.Unscoped().Delete(&item).Error) }
	case strings.HasPrefix(scenario, "faq_"):
		item := faq.FAQ{Question: "Historical question", Answer: `<a href="` + alias + `">Download</a>`, Status: "published"}
		if scenario == "faq_draft" {
			item.Status = "draft"
		}
		require.NoError(t, db.Create(&item).Error)
		return func() { require.NoError(t, db.Unscoped().Delete(&item).Error) }
	case strings.HasPrefix(scenario, "gallery_"):
		item := gallery.Gallery{Name: "Historical gallery", Slug: "historical-gallery", Status: "published"}
		if scenario == "gallery_draft_cover" {
			item.Status = "draft"
		}
		if strings.HasSuffix(scenario, "cover") {
			item.CoverImage = alias
		}
		require.NoError(t, db.Create(&item).Error)
		if strings.HasSuffix(scenario, "cover") {
			return func() { require.NoError(t, db.Unscoped().Delete(&item).Error) }
		}
		placement := gallery.GalleryImage{GalleryID: item.ID, URL: "/static/unrelated.png", Thumbnail: alias}
		if scenario == "gallery_orphan_image" {
			placement.GalleryID += 1000
		}
		require.NoError(t, db.Create(&placement).Error)
		if scenario == "gallery_deleted_image" {
			require.NoError(t, db.Delete(&placement).Error)
		}
		return func() { require.NoError(t, db.Unscoped().Delete(&placement).Error) }
	case strings.HasPrefix(scenario, "setting_") || scenario == "identity_mirror_only":
		item := setting.Setting{Key: setting.WebsiteProfileKeyAvatarURL, Value: alias, Locale: "en", IsPublic: true}
		if scenario == "identity_mirror_only" {
			item.Key = setting.SiteLogoSettingKey
		}
		require.NoError(t, db.Create(&item).Error)
		if scenario == "setting_hidden" {
			require.NoError(t, db.Model(&item).Update("is_public", false).Error)
		}
		if scenario == "setting_deleted" {
			require.NoError(t, db.Delete(&item).Error)
		}
		return func() { require.NoError(t, db.Unscoped().Delete(&item).Error) }
	case strings.HasPrefix(scenario, "swatch_"):
		require.NoError(t, db.AutoMigrate(&product.ProductSpecificationTemplate{}, &product.SpecDefinition{}, &product.ProductSpecOptionItem{}, &product.ProductVariantOptionValue{}))
		template := product.ProductSpecificationTemplate{Name: "Options", Slug: "historical-options"}
		require.NoError(t, db.Create(&template).Error)
		definition := product.SpecDefinition{ProductSpecificationTemplateID: template.ID, Name: "Finish", Slug: "finish", Role: "variant", IsVisible: true}
		if scenario == "swatch_attribute_definition" {
			definition.Role = "attribute"
		}
		require.NoError(t, db.Create(&definition).Error)
		if scenario == "swatch_hidden_definition" {
			require.NoError(t, db.Model(&definition).Update("is_visible", false).Error)
		}
		item := product.Product{Name: "Swatch owner", Slug: "swatch-owner", ProductSpecificationTemplateID: &template.ID, Status: "active"}
		if scenario == "swatch_wrong_template" {
			item.ProductSpecificationTemplateID = nil
		}
		require.NoError(t, db.Create(&item).Error)
		placement := product.ProductVariantOptionValue{ProductID: item.ID, SpecDefinitionID: definition.ID, ValueKey: "black", Label: "Black", SwatchURL: alias, IsEnabled: true}
		require.NoError(t, db.Create(&placement).Error)
		if scenario == "swatch_disabled" {
			require.NoError(t, db.Model(&placement).Update("is_enabled", false).Error)
		}
		return func() { require.NoError(t, db.Delete(&placement).Error) }
	case scenario == "registration_only":
		asset := media.MediaAsset{URL: alias, StorageKey: key, MediaType: "image", Status: "active", Visibility: "public"}
		require.NoError(t, db.Create(&asset).Error)
		return func() { require.NoError(t, db.Unscoped().Delete(&asset).Error) }
	case scenario == "private_attachment":
		require.NoError(t, db.Exec("CREATE TABLE ticket_messages (id INTEGER PRIMARY KEY, ticket_id INTEGER, attachments TEXT)").Error)
		encoded, err := json.Marshal([]map[string]string{{"url": alias}})
		require.NoError(t, err)
		require.NoError(t, db.Table("ticket_messages").Create(map[string]any{"id": 1, "ticket_id": 42, "attachments": string(encoded)}).Error)
		return func() {
			require.NoError(t, db.Table("ticket_messages").Where("id = ?", 1).Update("attachments", "[]").Error)
		}
	default:
		t.Fatalf("unsupported public visibility scenario %s", scenario)
		return nil
	}
}

func TestHistoricalDedicatedObjectPublicAccessRequiresVisiblePublishedConsumerAndEndsAtLastDetachment(t *testing.T) {
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		for _, scenario := range []struct {
			name   string
			public bool
		}{
			{"product_visible", true}, {"product_thumbnail", true}, {"product_poster", true}, {"product_description", true},
			{"product_hidden", false}, {"product_inactive", false}, {"product_deleted", false}, {"product_deleted_placement", false}, {"product_orphan", false},
			{"post_published", true}, {"post_draft", false}, {"post_deleted", false}, {"faq_published", true}, {"faq_draft", false},
			{"gallery_published_cover", true}, {"gallery_draft_cover", false}, {"gallery_thumbnail", true}, {"gallery_deleted_image", false}, {"gallery_orphan_image", false},
			{"setting_public", true}, {"setting_hidden", false}, {"setting_deleted", false}, {"identity_mirror_only", false},
			{"swatch_visible", true}, {"swatch_disabled", false}, {"swatch_hidden_definition", false}, {"swatch_attribute_definition", false}, {"swatch_wrong_template", false},
			{"registration_only", false}, {"private_attachment", false},
		} {
			t.Run(namespace+"/"+scenario.name, func(t *testing.T) {
				db, provider, handler, event, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, namespace)
				detach := insertHistoricalDedicatedObjectConsumerForPublicVisibilityTest(t, db, key, scenario.name)
				access := NewPublicUploadAccessService(nil, nil)
				access.ConfigureSiteLogoService(handler.siteLogo)
				access.ConfigureSiteFaviconService(handler.siteFavicon)
				access.ConfigureHomeVisualTileService(handler.homeVisual)
				allowed, err := access.CanServePublicUpload(t.Context(), key)
				require.NoError(t, err)
				require.Equal(t, scenario.public, allowed)
				require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse, "nonpublic consumers still retain physical objects")
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, provider, key)
				detach()
				allowed, err = access.CanServePublicUpload(t.Context(), key)
				require.NoError(t, err)
				require.False(t, allowed, "access ends before the eventual physical cleanup")
				require.NoError(t, handler.Handle(t.Context(), event))
			})
		}
	}
}

func TestHistoricalDedicatedObjectPublicAccessPropagatesDatabaseErrorsAndCancellation(t *testing.T) {
	db, _, handler, _, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, "site-logo")
	canceledContext, cancel := context.WithCancel(t.Context())
	cancel()
	allowed, err := handler.siteLogo.CanServePublicLogo(canceledContext, key)
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, allowed)
	require.NoError(t, db.Migrator().DropTable("posts"))
	require.NoError(t, db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY)").Error)
	allowed, err = handler.siteLogo.CanServePublicLogo(t.Context(), key)
	require.Error(t, err)
	require.False(t, allowed)
}

func TestHistoricalDedicatedObjectPublicAccessDoesNotDependOnUnrelatedPrivateAttachmentJSON(t *testing.T) {
	db, _, handler, _, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, "site-logo")
	require.NoError(t, db.Exec("CREATE TABLE ticket_messages (id INTEGER PRIMARY KEY, ticket_id INTEGER, attachments TEXT)").Error)
	require.NoError(t, db.Table("ticket_messages").Create(map[string]any{"id": 1, "ticket_id": 42, "attachments": "invalid historical JSON"}).Error)
	allowed, err := handler.siteLogo.CanServePublicLogo(t.Context(), key)
	require.NoError(t, err)
	require.False(t, allowed)
	insertHistoricalDedicatedObjectConsumerForPublicVisibilityTest(t, db, key, "post_published")
	allowed, err = handler.siteLogo.CanServePublicLogo(t.Context(), key)
	require.NoError(t, err)
	require.True(t, allowed)
}
