package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t *testing.T, namespace string) (*gorm.DB, storage.StorageService, *ObjectStorageCleanupOutboxHandler, outbox.Event, string) {
	t.Helper()
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{},
		&sitelogodomain.Asset{}, &sitefavicondomain.Asset{}, &homevisualtile.Tile{},
		&product.Product{}, &product.ProductMedia{}, &post.Post{}, &faq.FAQ{}, &gallery.Gallery{}, &gallery.GalleryImage{}, &setting.Setting{})
	root := t.TempDir()
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://dedicated.example.test/assets"})
	require.NoError(t, err)
	key := namespace + "/historical/retained image.png"
	objectPath := filepath.Join(root, filepath.FromSlash(key))
	require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o755))
	require.NoError(t, os.WriteFile(objectPath, []byte("dedicated object"), 0o644))
	logos := NewSiteLogoService(repository.NewSiteLogoRepository(db), provider, "https://shop.example.test")
	favicons := NewSiteFaviconService(repository.NewSiteFaviconRepository(db), provider, "https://shop.example.test")
	visuals := NewHomeVisualTileService(repository.NewHomeVisualTileRepository(db), provider, "https://shop.example.test")
	handler := NewObjectStorageCleanupOutboxHandler(nil, logos, visuals, nil)
	handler.ConfigureSiteFaviconService(favicons)
	resource := map[string]string{"site-logo": objectCleanupResourceSiteLogo, "site-favicon": objectCleanupResourceSiteFavicon, "visual-showcase": objectCleanupResourceHomeVisualTile}[namespace]
	event, err := newObjectStorageCleanupEvent(resource, "1", resource, "1", []string{key})
	require.NoError(t, err)
	require.NoError(t, repository.NewOutboxRepository(db).CreateEvent(event))
	return db, provider, handler, *event, key
}

func insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t *testing.T, db *gorm.DB, key, consumer string) func() {
	t.Helper()
	alias := "https://shop.example.test/uploads/" + strings.ReplaceAll(key, " ", "%20") + "?revision=old&size=full"
	if strings.HasPrefix(consumer, "product_") {
		item := product.Product{SKU: "retained-product", Slug: "retained-product", Name: "Retained product", PriceMinor: 100, Status: "draft"}
		if consumer == "product_description" {
			item.Description = `<img src="` + strings.ReplaceAll(alias, "&", "&amp;") + `">`
		}
		if consumer == "product_short_desc" {
			item.ShortDesc = `<video poster="` + alias + `"></video>`
		}
		require.NoError(t, db.Create(&item).Error)
		require.NoError(t, db.Delete(&item).Error)
		if consumer == "product_description" || consumer == "product_short_desc" {
			return func() {
				require.NoError(t, db.Unscoped().Model(&item).Updates(map[string]any{"description": "", "short_description": ""}).Error)
			}
		}
		placement := product.ProductMedia{ProductID: item.ID, MediaType: "image", URL: "/static/unrelated.png"}
		switch consumer {
		case "product_url", "product_deleted_placement":
			placement.URL = alias
		case "product_thumbnail":
			placement.ThumbnailURL = alias
		case "product_poster":
			placement.PosterURL = alias
		case "product_orphan":
			placement.URL, placement.ProductID = alias, item.ID+1000
		}
		require.NoError(t, db.Create(&placement).Error)
		if consumer == "product_deleted_placement" {
			require.NoError(t, db.Delete(&placement).Error)
		}
		return func() { require.NoError(t, db.Unscoped().Delete(&placement).Error) }
	}
	switch consumer {
	case "post_content", "post_featured":
		item := post.Post{Title: "Retained post", Slug: "retained-post", Status: "draft"}
		if consumer == "post_content" {
			item.Content = `<a href="` + alias + `">Download historical image</a>`
		} else {
			item.FeaturedImg = alias
		}
		require.NoError(t, db.Create(&item).Error)
		require.NoError(t, db.Delete(&item).Error)
		return func() { require.NoError(t, db.Unscoped().Delete(&item).Error) }
	case "faq_answer_link":
		item := faq.FAQ{Question: "Retained question", Answer: `<a href="` + alias + `">Historical file</a>`}
		require.NoError(t, db.Create(&item).Error)
		require.NoError(t, db.Delete(&item).Error)
		return func() { require.NoError(t, db.Unscoped().Delete(&item).Error) }
	case "gallery_thumbnail":
		item := gallery.GalleryImage{GalleryID: 999, URL: "/static/other.png", Thumbnail: alias}
		require.NoError(t, db.Create(&item).Error)
		require.NoError(t, db.Delete(&item).Error)
		return func() { require.NoError(t, db.Unscoped().Delete(&item).Error) }
	case "hidden_deleted_setting":
		item := setting.Setting{Key: setting.WebsiteProfileKeyAvatarURL, Locale: "en", Value: alias}
		require.NoError(t, db.Create(&item).Error)
		require.NoError(t, db.Model(&item).Update("is_public", false).Error)
		require.NoError(t, db.Delete(&item).Error)
		return func() { require.NoError(t, db.Unscoped().Delete(&item).Error) }
	case "legacy_attachments":
		require.NoError(t, db.Exec("CREATE TABLE ticket_messages (id INTEGER PRIMARY KEY, ticket_id INTEGER, attachments TEXT)").Error)
		value, err := json.Marshal([]map[string]string{{"url": alias}})
		require.NoError(t, err)
		require.NoError(t, db.Table("ticket_messages").Create(map[string]any{"id": 1, "ticket_id": 42, "attachments": string(value)}).Error)
		return func() {
			require.NoError(t, db.Table("ticket_messages").Where("id = ?", 1).Update("attachments", "[]").Error)
		}
	default:
		t.Fatalf("unsupported historical consumer %s", consumer)
		return nil
	}
}

func requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t *testing.T, provider storage.StorageService, key string) {
	t.Helper()
	object, err := provider.(storage.ObjectOpener).Open(t.Context(), key)
	require.NoError(t, err)
	require.NoError(t, object.ReadCloser.Close())
}

func TestDedicatedPublicObjectCleanupRetainsHistoricalConsumersAndRetriesOriginalOutboxEventAfterLastDetachment(t *testing.T) {
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		for _, consumer := range []string{"product_url", "product_thumbnail", "product_poster", "product_deleted_placement", "product_orphan", "product_description", "product_short_desc", "post_content", "post_featured", "faq_answer_link", "gallery_thumbnail", "hidden_deleted_setting", "legacy_attachments"} {
			t.Run(namespace+"/"+consumer, func(t *testing.T) {
				db, provider, handler, event, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, namespace)
				detach := insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, db, key, consumer)
				require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, provider, key)
				worker := NewOutboxService(repository.NewOutboxRepository(db))
				worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, handler.Handle)
				now := time.Now().UTC().Add(time.Second)
				result, err := worker.ProcessPending(t.Context(), now, 10)
				require.NoError(t, err)
				require.Equal(t, 1, result.Failed)
				var retained outbox.Event
				require.NoError(t, db.First(&retained, event.ID).Error)
				require.Equal(t, outbox.EventStatusFailed, retained.Status)
				require.Equal(t, event.EventKey, retained.EventKey)
				detach()
				result, err = worker.ProcessPending(t.Context(), retained.AvailableAt.Add(time.Second), 10)
				require.NoError(t, err)
				require.Equal(t, 1, result.Processed)
				require.NoError(t, db.First(&retained, event.ID).Error)
				require.Equal(t, outbox.EventStatusProcessed, retained.Status)
				require.Equal(t, 2, retained.Attempts)
				var count int64
				require.NoError(t, db.Model(&outbox.Event{}).Count(&count).Error)
				require.EqualValues(t, 1, count)
				_, err = provider.(storage.ObjectOpener).Open(t.Context(), key)
				require.Error(t, err)
				require.NoError(t, handler.Handle(t.Context(), event), "cleanup retries remain idempotent")
			})
		}
	}
}

func TestDedicatedPublicObjectCleanupWaitsUntilBothProductAndPostHaveDetached(t *testing.T) {
	db, provider, handler, event, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, "visual-showcase")
	detachProduct := insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, db, key, "product_url")
	detachPost := insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, db, key, "post_content")
	detachProduct()
	require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
	requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, provider, key)
	detachPost()
	require.NoError(t, handler.Handle(t.Context(), event))
}

func TestSharedMediaDeletionAndPreviouslyQueuedCleanupCannotDeleteDedicatedOriginalOrDerivativeObjects(t *testing.T) {
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		for _, collision := range []string{"original_key", "original_url_only", "original_legacy_relative_url", "derivative_key", "derivative_url_only", "derivative_legacy_relative_url", "derivative_stored_uploads_path", "derivative_stored_absolute_url", "derivative_url_with_different_stored_key", "derivative_retained_deleted"} {
			t.Run(namespace+"/"+collision, func(t *testing.T) {
				db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
				root := t.TempDir()
				service := newMediaDeleteTestService(t, db, root)
				key := namespace + "/retired-but-dedicated.png"
				keys := []string{key, "catalog/shared-original.png", "derivatives/ordinary.webp", "derivatives/conflicting.webp"}
				for _, objectKey := range keys {
					objectPath := filepath.Join(root, filepath.FromSlash(objectKey))
					require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o755))
					require.NoError(t, os.WriteFile(objectPath, []byte(objectKey), 0o644))
				}
				asset := media.MediaAsset{Filename: "original.png", URL: service.storage.GetURL("catalog/shared-original.png"), StorageKey: "catalog/shared-original.png", MediaType: "image", Status: "active", Visibility: "public"}
				if strings.HasPrefix(collision, "original") {
					asset.StorageKey, asset.URL = key, service.storage.GetURL(key)
					if collision == "original_url_only" {
						asset.StorageKey = ""
					}
					if collision == "original_legacy_relative_url" {
						asset.StorageKey, asset.URL = "", key
					}
				}
				require.NoError(t, db.Create(&asset).Error)
				require.NoError(t, db.Create(&media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "ordinary", URL: service.storage.GetURL(keys[2]), StorageKey: keys[2]}).Error)
				if strings.HasPrefix(collision, "derivative") {
					derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "conflict", URL: service.storage.GetURL(key), StorageKey: key}
					if collision == "derivative_url_only" {
						derivative.StorageKey = ""
					}
					if collision == "derivative_legacy_relative_url" {
						derivative.StorageKey, derivative.URL = "", key
					}
					if collision == "derivative_url_with_different_stored_key" {
						derivative.StorageKey = keys[3]
					}
					if collision == "derivative_stored_uploads_path" {
						derivative.StorageKey, derivative.URL = "/uploads/"+key, "/static/unrelated.png"
					}
					if collision == "derivative_stored_absolute_url" {
						derivative.StorageKey, derivative.URL = service.storage.GetURL(key), "/static/unrelated.png"
					}
					require.NoError(t, db.Create(&derivative).Error)
					if collision == "derivative_retained_deleted" {
						require.NoError(t, db.Delete(&derivative).Error)
					}
				}
				require.ErrorIs(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
				_, err := service.repo.FindAssetByID(asset.ID)
				require.NoError(t, err, "ownership failure must not retire shared metadata")
				var events int64
				require.NoError(t, db.Model(&outbox.Event{}).Count(&events).Error)
				require.Zero(t, events)
				// An event committed by an earlier version must obey the same guard.
				require.NoError(t, db.Delete(&asset).Error)
				event, err := newObjectStorageCleanupEvent(objectCleanupResourceMediaAsset, fmt.Sprint(asset.ID), outbox.AggregateTypeMediaAsset, fmt.Sprint(asset.ID), nil)
				require.NoError(t, err)
				require.ErrorIs(t, NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil).Handle(t.Context(), *event), ErrMediaAssetInUse)
				for _, objectKey := range keys {
					_, err := os.Stat(filepath.Join(root, filepath.FromSlash(objectKey)))
					require.NoError(t, err, "all ownership checks must precede the first physical deletion")
				}
				_, err = service.repo.FindAssetByIDUnscoped(asset.ID)
				require.NoError(t, err)
			})
		}
	}
}

func TestDedicatedPublicObjectCleanupMatchesOnlyExactTrustedAliasesAndFindsConsumersAfterFirstPage(t *testing.T) {
	db, provider, handler, event, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, "site-logo")
	for index := 0; index < 257; index++ {
		item := product.Product{SKU: fmt.Sprint(index), Slug: fmt.Sprint(index), Name: "Unrelated history", PriceMinor: 100,
			Description: `<p>/uploads/` + key + `</p><img src="https://foreign.example.test/uploads/` + key + `"><img src="https://shop.example.test:8443/uploads/` + key + `">`}
		require.NoError(t, db.Create(&item).Error)
		placement := product.ProductMedia{ProductID: item.ID, URL: "/uploads/" + key + ".backup", ThumbnailURL: "/uploads/another/retained image.png"}
		require.NoError(t, db.Create(&placement).Error)
	}
	retained := product.ProductMedia{ProductID: 9999, URL: provider.GetURL(key) + "?revision=1"}
	require.NoError(t, db.Create(&retained).Error)
	require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse, "use the dedicated provider and scan beyond the first page")
	requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, provider, key)
	require.NoError(t, db.Unscoped().Delete(&retained).Error)
	require.NoError(t, handler.Handle(t.Context(), event), "foreign URLs, plaintext, prefixes and same basenames are not consumers")
}

func TestDedicatedPublicObjectCleanupRetainsConflictingSharedRegistrationsIncludingDeletedMetadata(t *testing.T) {
	for _, registration := range []string{"original_key", "original_url_only", "deleted_original", "deleted_derivative"} {
		t.Run(registration, func(t *testing.T) {
			db, provider, handler, event, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, "site-favicon")
			asset := media.MediaAsset{Filename: "retained.png", StorageKey: key, URL: "https://other-store.example.test/independent.png", MediaType: "image"}
			if registration == "original_url_only" {
				asset.StorageKey, asset.URL = "", provider.GetURL(key)
			}
			if registration == "deleted_derivative" {
				asset.StorageKey = "catalog/unrelated.png"
			}
			require.NoError(t, db.Create(&asset).Error)
			if registration == "deleted_original" {
				require.NoError(t, db.Delete(&asset).Error)
			}
			if registration == "deleted_derivative" {
				derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "legacy", StorageKey: key}
				require.NoError(t, db.Create(&derivative).Error)
				require.NoError(t, db.Delete(&derivative).Error)
			}
			require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
			requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, provider, key)
			require.NoError(t, db.Unscoped().Where("media_asset_id = ?", asset.ID).Delete(&media.MediaAssetDerivative{}).Error)
			require.NoError(t, db.Unscoped().Delete(&asset).Error)
			require.NoError(t, handler.Handle(t.Context(), event))
		})
	}
}

func TestSharedMediaDeletionPreservesDedicatedFilesWithSameBasename(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{})
	root := t.TempDir()
	service := newMediaDeleteTestService(t, db, root)
	for _, key := range []string{"catalog/same.png", "site-logo/same.png", "site-favicon/same.png", "visual-showcase/same.png"} {
		objectPath := filepath.Join(root, filepath.FromSlash(key))
		require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o755))
		require.NoError(t, os.WriteFile(objectPath, []byte(key), 0o644))
	}
	asset := media.MediaAsset{Filename: "same.png", URL: service.storage.GetURL("catalog/same.png"), StorageKey: "catalog/same.png", MediaType: "image"}
	require.NoError(t, db.Create(&asset).Error)
	require.NoError(t, service.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	_, err := os.Stat(filepath.Join(root, "catalog", "same.png"))
	require.True(t, os.IsNotExist(err))
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, service.storage, namespace+"/same.png")
	}
}
