package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"testing"
	"time"

	mediadomain "commerce-platform/internal/domain/media"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMediaDerivativeRebuildSynchronizesStorefrontSnapshotsAndOwnedThumbnail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(
		&mediadomain.MediaAsset{},
		&mediadomain.MediaAssetDerivative{},
		&mediadomain.MediaDerivativePreset{},
		&mediadomain.MediaDerivativeRebuildJob{},
		&productdomain.ProductMedia{},
	))

	presetRepo := repository.NewMediaDerivativePresetRepository(db)
	require.NoError(t, SeedDefaultMediaDerivativePresets(presetRepo))
	mediaRepo := repository.NewMediaRepository(db)
	rebuildRepo := repository.NewMediaDerivativeRebuildJobRepository(db)

	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: t.TempDir(),
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)
	mediaService := NewMediaService(mediaRepo, storageService, nil, "https://shop.example.test", 20<<30)
	mediaService.ConfigureDerivativePresetRepository(presetRepo)
	mediaService.ConfigureDerivativeRebuildJobRepository(rebuildRepo)

	asset, err := mediaService.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "rebuild-wheel.png", "image/png", testOpaquePNG(t, 1200, 600)),
		MediaType:  "image",
		UploaderID: 42,
		Width:      1200,
		Height:     600,
	})
	require.NoError(t, err)

	variants, thumbnail, err := mediaService.ProductMediaImageVariants(asset.ID)
	require.NoError(t, err)
	require.NoError(t, db.Create(&productdomain.ProductMedia{
		ProductID:        100,
		MediaAssetID:     &asset.ID,
		MediaType:        "image",
		Role:             "primary",
		URL:              asset.URL,
		ThumbnailURL:     thumbnail,
		ImageVariantData: productdomain.ProductMediaImageVariantsJSON(variants),
		IsPrimary:        true,
		IsVisible:        true,
	}).Error)

	oldThumbnail := thumbnail
	thumbnailPreset, err := presetRepo.FindByCode("thumbnail")
	require.NoError(t, err)
	enabled := thumbnailPreset.Enabled
	_, err = mediaService.UpdateMediaDerivativePreset(thumbnailPreset.ID, MediaDerivativePresetInput{
		Label:     thumbnailPreset.Label,
		MaxWidth:  400,
		SortOrder: thumbnailPreset.SortOrder,
		Enabled:   &enabled,
	})
	require.NoError(t, err)

	firstRun, err := mediaService.ProcessNextMediaDerivativeRebuild(context.Background(), "test-rebuilder", 10)
	require.NoError(t, err)
	require.True(t, firstRun.Claimed)
	require.True(t, firstRun.Completed)
	require.Equal(t, 1, firstRun.ScannedAssets)
	require.Equal(t, 1, firstRun.GeneratedDerivatives)

	var mediaItem productdomain.ProductMedia
	require.NoError(t, db.Where("media_asset_id = ?", asset.ID).First(&mediaItem).Error)
	nextVariants := productdomain.ParseProductMediaImageVariants(mediaItem.ImageVariantData)
	require.NotEqual(t, oldThumbnail, nextVariants["thumbnail"].URL)
	require.Equal(t, nextVariants["thumbnail"].URL, mediaItem.ThumbnailURL)

	manualThumbnail := "https://cdn.example.test/manual-thumbnail.webp"
	require.NoError(t, db.Model(&productdomain.ProductMedia{}).
		Where("id = ?", mediaItem.ID).
		Update("thumbnail_url", manualThumbnail).Error)

	thumbnailPreset, err = presetRepo.FindByCode("thumbnail")
	require.NoError(t, err)
	enabled = thumbnailPreset.Enabled
	_, err = mediaService.UpdateMediaDerivativePreset(thumbnailPreset.ID, MediaDerivativePresetInput{
		Label:     thumbnailPreset.Label,
		MaxWidth:  420,
		SortOrder: thumbnailPreset.SortOrder,
		Enabled:   &enabled,
	})
	require.NoError(t, err)
	_, err = mediaService.ProcessNextMediaDerivativeRebuild(context.Background(), "test-rebuilder", 10)
	require.NoError(t, err)
	require.NoError(t, db.Where("id = ?", mediaItem.ID).First(&mediaItem).Error)
	require.Equal(t, manualThumbnail, mediaItem.ThumbnailURL)

	largePreset, err := presetRepo.FindByCode("large")
	require.NoError(t, err)
	_, err = mediaService.SetMediaDerivativePresetEnabled(largePreset.ID, false)
	require.NoError(t, err)
	_, err = mediaService.ProcessNextMediaDerivativeRebuild(context.Background(), "test-rebuilder", 10)
	require.NoError(t, err)
	require.NoError(t, db.Where("id = ?", mediaItem.ID).First(&mediaItem).Error)
	nextVariants = productdomain.ParseProductMediaImageVariants(mediaItem.ImageVariantData)
	_, hasLarge := nextVariants["large"]
	require.False(t, hasLarge)
}

func TestMediaDerivativeRebuildKeepsGeneratedThumbnailOwnershipAcrossFallbackPresets(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(
		&mediadomain.MediaAsset{},
		&mediadomain.MediaAssetDerivative{},
		&mediadomain.MediaDerivativePreset{},
		&mediadomain.MediaDerivativeRebuildJob{},
		&productdomain.ProductMedia{},
	))

	presetRepo := repository.NewMediaDerivativePresetRepository(db)
	require.NoError(t, SeedDefaultMediaDerivativePresets(presetRepo))
	mediaRepo := repository.NewMediaRepository(db)
	rebuildRepo := repository.NewMediaDerivativeRebuildJobRepository(db)
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: t.TempDir(),
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)
	mediaService := NewMediaService(mediaRepo, storageService, nil, "https://shop.example.test", 20<<30)
	mediaService.ConfigureDerivativePresetRepository(presetRepo)
	mediaService.ConfigureDerivativeRebuildJobRepository(rebuildRepo)

	asset, err := mediaService.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "fallback-wheel.png", "image/png", testOpaquePNG(t, 1200, 600)),
		MediaType:  "image",
		UploaderID: 42,
		Width:      1200,
		Height:     600,
	})
	require.NoError(t, err)
	variants, thumbnail, err := mediaService.ProductMediaImageVariants(asset.ID)
	require.NoError(t, err)
	require.NoError(t, db.Create(&productdomain.ProductMedia{
		ProductID:        100,
		MediaAssetID:     &asset.ID,
		MediaType:        "image",
		Role:             "primary",
		URL:              asset.URL,
		ThumbnailURL:     thumbnail,
		ImageVariantData: productdomain.ProductMediaImageVariantsJSON(variants),
		IsPrimary:        true,
		IsVisible:        true,
	}).Error)

	thumbnailPreset, err := presetRepo.FindByCode("thumbnail")
	require.NoError(t, err)
	_, err = mediaService.SetMediaDerivativePresetEnabled(thumbnailPreset.ID, false)
	require.NoError(t, err)
	_, err = mediaService.ProcessNextMediaDerivativeRebuild(context.Background(), "test-rebuilder", 10)
	require.NoError(t, err)

	var mediaItem productdomain.ProductMedia
	require.NoError(t, db.Where("media_asset_id = ?", asset.ID).First(&mediaItem).Error)
	variants = productdomain.ParseProductMediaImageVariants(mediaItem.ImageVariantData)
	cardThumbnail := variants["card"].URL
	require.NotEmpty(t, cardThumbnail)
	require.Equal(t, cardThumbnail, mediaItem.ThumbnailURL)

	cardPreset, err := presetRepo.FindByCode("card")
	require.NoError(t, err)
	enabled := cardPreset.Enabled
	_, err = mediaService.UpdateMediaDerivativePreset(cardPreset.ID, MediaDerivativePresetInput{
		Label:     cardPreset.Label,
		MaxWidth:  760,
		SortOrder: cardPreset.SortOrder,
		Enabled:   &enabled,
	})
	require.NoError(t, err)
	_, err = mediaService.ProcessNextMediaDerivativeRebuild(context.Background(), "test-rebuilder", 10)
	require.NoError(t, err)

	require.NoError(t, db.Where("id = ?", mediaItem.ID).First(&mediaItem).Error)
	variants = productdomain.ParseProductMediaImageVariants(mediaItem.ImageVariantData)
	require.NotEqual(t, cardThumbnail, variants["card"].URL)
	require.Equal(t, variants["card"].URL, mediaItem.ThumbnailURL)
}

func TestMediaDerivativePresetServiceCapsActiveConversions(t *testing.T) {
	mediaService, presetRepo, _ := newMediaDerivativePresetTestService(t)
	require.NoError(t, SeedDefaultMediaDerivativePresets(presetRepo))

	for index := 0; index < mediaDerivativeActivePresetLimit-3; index++ {
		_, err := mediaService.CreateMediaDerivativePreset(MediaDerivativePresetInput{
			Code:      fmt.Sprintf("custom-%d", index),
			Label:     fmt.Sprintf("Custom %d", index),
			MaxWidth:  100 + index,
			SortOrder: 100 + index,
		})
		require.NoError(t, err)
	}
	_, err := mediaService.CreateMediaDerivativePreset(MediaDerivativePresetInput{
		Code:      "one-too-many",
		Label:     "One too many",
		MaxWidth:  1200,
		SortOrder: 200,
	})
	require.ErrorIs(t, err, ErrMediaDerivativePresetLimitReached)
}

func TestDecodeDerivativeSourceRejectsOversizedStoredImage(t *testing.T) {
	payload := append(testOpaquePNG(t, 2, 2), bytes.Repeat([]byte{0}, mediaDerivativeMaxSourceBytes)...)
	_, err := decodeDerivativeSource(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(payload)), nil
	})
	require.ErrorIs(t, err, ErrMediaDerivativeGenerationFailed)
}

func TestDecodeDerivativeSourceAppliesJPEGExifOrientation(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 2, 1))
	source.Set(0, 0, color.RGBA{R: 255, A: 255})
	source.Set(1, 0, color.RGBA{G: 255, A: 255})
	var encoded bytes.Buffer
	require.NoError(t, jpeg.Encode(&encoded, source, &jpeg.Options{Quality: 95}))

	decoded, err := decodeDerivativeSource(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(testJPEGWithExifOrientation(t, encoded.Bytes(), 6))), nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, decoded.Bounds().Dx())
	require.Equal(t, 2, decoded.Bounds().Dy())
}

func TestMediaDerivativeGenerationSlotsRespectCanceledContext(t *testing.T) {
	releaseFirst, err := acquireMediaDerivativeGenerationSlot(context.Background())
	require.NoError(t, err)
	defer releaseFirst()
	releaseSecond, err := acquireMediaDerivativeGenerationSlot(context.Background())
	require.NoError(t, err)
	defer releaseSecond()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err = acquireMediaDerivativeGenerationSlot(ctx)
	require.ErrorIs(t, err, ErrMediaDerivativeGenerationFailed)
}

func testJPEGWithExifOrientation(t *testing.T, jpegData []byte, orientation byte) []byte {
	t.Helper()
	require.GreaterOrEqual(t, len(jpegData), 2)
	require.Equal(t, byte(0xff), jpegData[0])
	require.Equal(t, byte(0xd8), jpegData[1])

	tiff := []byte{
		'M', 'M', 0x00, 0x2a, 0x00, 0x00, 0x00, 0x08,
		0x00, 0x01,
		0x01, 0x12, 0x00, 0x03, 0x00, 0x00, 0x00, 0x01,
		0x00, orientation, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	exif := append([]byte("Exif\x00\x00"), tiff...)
	segmentLength := len(exif) + 2
	result := make([]byte, 0, len(jpegData)+len(exif)+4)
	result = append(result, jpegData[:2]...)
	result = append(result, 0xff, 0xe1, byte(segmentLength>>8), byte(segmentLength))
	result = append(result, exif...)
	result = append(result, jpegData[2:]...)
	return result
}
