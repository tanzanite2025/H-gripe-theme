package service

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t *testing.T) (*gorm.DB, *HomeVisualTileService) {
	t.Helper()
	db := newSiteLogoTestDB(t)
	require.NoError(t, db.AutoMigrate(&homevisualtile.Tile{}, &media.MediaAsset{}))
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://media.example.test"})
	require.NoError(t, err)
	service := NewHomeVisualTileService(repository.NewHomeVisualTileRepository(db), provider, "https://shop.example.test")
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	return db, service
}

func uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t *testing.T, service *HomeVisualTileService, slot int) HomeVisualTileInput {
	t.Helper()
	upload, err := service.UploadAdminImage(t.Context(), HomeHeroVisualShowcaseTileSetKey, "en", multipartFileHeader(t, "showcase.png", "image/png", testOpaquePNG(t, 600, 600)))
	require.NoError(t, err)
	return HomeVisualTileInput{ImageURL: upload.ImageURL, ThumbnailURL: upload.ThumbnailURL, StorageKey: upload.StorageKey,
		Title: "Showcase", AltText: "Showcase image", DesktopOrder: slot, Width: upload.Width, Height: upload.Height, IsPublished: true}
}

func requireDedicatedVisualShowcaseObjectPhysicalExistence(t *testing.T, service *HomeVisualTileService, key string, expected bool) {
	t.Helper()
	object, err := service.storage.(storage.ObjectOpener).Open(t.Context(), key)
	if !expected {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)
	require.NotNil(t, object)
	require.NoError(t, object.ReadCloser.Close())
}

func TestVisualShowcaseFailedSaveNeverDeletesAFileReusedByAnotherSlot(t *testing.T) {
	db, service := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
	first := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
	second := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 2)
	_, err := service.ReplaceAdminItems(t.Context(), "home-hero", "en", []HomeVisualTileInput{first, second})
	require.NoError(t, err)
	service.ConfigureObjectCleanupOutbox(nil)
	replacement := second
	replacement.DesktopOrder = 1
	_, err = service.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", replacement)
	require.ErrorIs(t, err, ErrObjectStorageCleanupUnavailable)
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, first.StorageKey, true)
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, second.StorageKey, true)
	items, err := service.GetAdminItems("home-hero", "en")
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, first.StorageKey, items[0].StorageKey)
	var events int64
	require.NoError(t, db.Model(&outbox.Event{}).Count(&events).Error)
	require.Zero(t, events)
}

func TestVisualShowcaseNewOriginalAndThumbnailBindingsAreValidatedAtomically(t *testing.T) {
	_, service := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
	original := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
	_, err := service.ReplaceAdminItems(t.Context(), "home-hero", "en", []HomeVisualTileInput{original})
	require.NoError(t, err)
	for _, invalidKind := range []string{"foreign thumbnail", "wrong port thumbnail", "missing thumbnail", "missing original", "wrong explicit key", "private origin thumbnail"} {
		t.Run(invalidKind, func(t *testing.T) {
			candidate := original
			candidate.DesktopOrder = 2
			switch invalidKind {
			case "foreign thumbnail":
				candidate.ThumbnailURL = "https://foreign.example.test/uploads/" + original.StorageKey
			case "wrong port thumbnail":
				candidate.ThumbnailURL = "https://media.example.test:9443/uploads/" + original.StorageKey
			case "missing thumbnail":
				candidate.ThumbnailURL = "/uploads/visual-showcase/home-hero/en/missing.png"
			case "missing original":
				candidate.StorageKey = "visual-showcase/home-hero/en/missing.png"
				candidate.ImageURL = "/uploads/" + candidate.StorageKey
				candidate.ThumbnailURL = candidate.ImageURL
			case "wrong explicit key":
				candidate.StorageKey = "visual-showcase/home-hero/en/another.png"
			case "private origin thumbnail":
				candidate.ThumbnailURL = "/uploads/customer-service/private/conversations/1/secret.png"
			}
			_, err := service.ReplaceAdminItems(t.Context(), "home-hero", "en", []HomeVisualTileInput{original, candidate})
			require.ErrorIs(t, err, ErrHomeVisualTileImageInvalid)
			items, err := service.GetAdminItems("home-hero", "en")
			require.NoError(t, err)
			require.Len(t, items, 1)
			require.Equal(t, original.StorageKey, items[0].StorageKey)
			requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, original.StorageKey, true)
		})
	}
}

func TestVisualShowcaseCleanupRetainsDraftOriginalURLAndIndependentThumbnailAcrossScopes(t *testing.T) {
	db, service := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
	original := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
	thumbnail := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 2)
	original.ThumbnailURL = thumbnail.ImageURL
	_, err := service.ReplaceAdminItems(t.Context(), "home-hero", "en", []HomeVisualTileInput{original})
	require.NoError(t, err)
	// Historical URL-only drafts remain retention evidence; they are not public.
	draft := homevisualtile.Tile{TileSetKey: "editorial", Locale: "zh_cn", ImageURL: thumbnail.ImageURL,
		ThumbnailURL: original.ImageURL, Title: "Draft", AltText: "Draft image", DesktopOrder: 1, IsPublished: true}
	require.NoError(t, db.Create(&draft).Error)
	require.NoError(t, db.Model(&draft).Update("is_published", false).Error)
	_, err = service.ReplaceAdminItems(t.Context(), "home-hero", "en", nil)
	require.NoError(t, err)
	var event outbox.Event
	require.NoError(t, db.First(&event).Error)
	handler := NewObjectStorageCleanupOutboxHandler(nil, nil, service, nil)
	require.NoError(t, handler.Handle(t.Context(), event))
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, original.StorageKey, true)
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, thumbnail.StorageKey, true)
	access := NewPublicUploadAccessService(nil, nil)
	access.ConfigureHomeVisualTileService(service)
	allowed, err := access.CanServePublicUpload(t.Context(), thumbnail.StorageKey)
	require.NoError(t, err)
	require.False(t, allowed)
	_, err = service.ReplaceAdminItems(t.Context(), "editorial", "zh_cn", nil)
	require.NoError(t, err)
	var events []outbox.Event
	require.NoError(t, db.Order("id ASC").Find(&events).Error)
	for _, cleanupEvent := range events {
		require.NoError(t, handler.Handle(t.Context(), cleanupEvent))
	}
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, original.StorageKey, false)
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, thumbnail.StorageKey, false)
	_, err = service.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", original)
	require.ErrorIs(t, err, ErrHomeVisualTileImageInvalid)
	items, err := service.GetAdminItems("home-hero", "en")
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestVisualShowcaseHistoricalThumbnailCanStayAtItsExistingSlotButCannotBeCopied(t *testing.T) {
	db, service := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
	input := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
	_, err := service.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", input)
	require.NoError(t, err)
	input.ThumbnailURL = "https://legacy.example.test/historical.png"
	require.NoError(t, db.Model(&homevisualtile.Tile{}).Where("locale = ?", "en").Update("thumbnail_url", input.ThumbnailURL).Error)
	input.Title = "Updated description"
	_, err = service.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", input)
	require.NoError(t, err)
	_, err = service.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "zh_cn", input)
	require.ErrorIs(t, err, ErrHomeVisualTileImageInvalid)
	items, err := service.GetAdminItems("home-hero", "zh_cn")
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestVisualShowcaseLastSlotDetachmentCreatesNewCleanupAfterAnEarlierRetainedObjectEventCompleted(t *testing.T) {
	db, service := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
	first := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
	second := first
	second.DesktopOrder = 2
	_, err := service.ReplaceAdminItems(t.Context(), "home-hero", "en", []HomeVisualTileInput{first, second})
	require.NoError(t, err)
	replacement := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
	_, err = service.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", replacement)
	require.NoError(t, err)
	var firstEvent outbox.Event
	require.NoError(t, db.First(&firstEvent).Error)
	handler := NewObjectStorageCleanupOutboxHandler(nil, nil, service, nil)
	require.NoError(t, handler.Handle(t.Context(), firstEvent))
	require.NoError(t, db.Model(&firstEvent).Update("status", outbox.EventStatusProcessed).Error)
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, first.StorageKey, true)
	replacement.DesktopOrder = 2
	_, err = service.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", replacement)
	require.NoError(t, err)
	var events []outbox.Event
	require.NoError(t, db.Order("id ASC").Find(&events).Error)
	require.Len(t, events, 2)
	require.NotEqual(t, events[0].EventKey, events[1].EventKey)
	require.NoError(t, handler.Handle(t.Context(), events[1]))
	require.NoError(t, handler.Handle(t.Context(), events[1]), "delivery remains idempotent")
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, first.StorageKey, false)
	requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, replacement.StorageKey, true)
}

func TestVisualShowcasePublicAccessUsesOwnerPublicationAndCannotBeOverriddenByMediaRegistration(t *testing.T) {
	db, service := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
	input := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
	_, err := service.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", input)
	require.NoError(t, err)
	require.NoError(t, db.Model(&homevisualtile.Tile{}).Where("showcase_key = ?", "home-hero").Update("is_published", false).Error)
	unknownKey := "visual-showcase/home-hero/en/unbound.png"
	createRegisteredPublicImageForProductMediaBindingTest(t, db, unknownKey)
	access := NewPublicUploadAccessService(NewMediaService(repository.NewMediaRepository(db), service.storage, nil, "", 20<<30), nil)
	allowed, err := access.CanServePublicUpload(context.Background(), input.StorageKey)
	require.NoError(t, err)
	require.False(t, allowed)
	access.ConfigureHomeVisualTileService(service)
	allowed, err = access.CanServePublicUpload(t.Context(), input.StorageKey)
	require.NoError(t, err)
	require.True(t, allowed, "hero displays all configured images by its existing contract")
	allowed, err = access.CanServePublicUpload(t.Context(), unknownKey)
	require.NoError(t, err)
	require.False(t, allowed)
	// An expired non-hero placement is retained but cannot authorize public reads.
	require.NoError(t, db.Model(&homevisualtile.Tile{}).Where("showcase_key = ?", "home-hero").Updates(map[string]any{"showcase_key": "editorial", "is_published": true, "published_until": time.Now().Add(-time.Hour)}).Error)
	allowed, err = access.CanServePublicUpload(t.Context(), input.StorageKey)
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestVisualShowcaseRepositoryRejectsMissingValidatorAndDoesNotOverwriteCallerSuppliedForeignID(t *testing.T) {
	db, service := createDedicatedVisualShowcaseObjectLifecycleTestEnvironment(t)
	input := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
	_, err := service.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "en", input)
	require.NoError(t, err)
	var first homevisualtile.Tile
	require.NoError(t, db.First(&first).Error)
	clonedValue := first
	clonedValue.Locale = "zh_cn"
	clonedValue.Title = "Another locale"
	require.ErrorIs(t, repository.NewHomeVisualTileRepository(db).SaveSingleVisualShowcaseItem(&clonedValue, nil), repository.ErrVisualShowcaseTransactionalImageBindingValidationUnavailable)
	clonedValue.ID = first.ID
	require.NoError(t, service.repo.SaveSingleVisualShowcaseItem(&clonedValue, nil))
	require.NotEqual(t, first.ID, clonedValue.ID)
	var unchanged homevisualtile.Tile
	require.NoError(t, db.First(&unchanged, first.ID).Error)
	require.Equal(t, "en", unchanged.Locale)
	require.Equal(t, first.Title, unchanged.Title)
}
