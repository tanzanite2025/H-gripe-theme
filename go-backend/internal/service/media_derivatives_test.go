package service

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	mediadomain "commerce-platform/internal/domain/media"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMediaUploadGeneratesPersistentImageDerivatives(t *testing.T) {
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
	require.NoError(t, db.AutoMigrate(&mediadomain.MediaAsset{}, &mediadomain.MediaAssetDerivative{}))

	uploadRoot := t.TempDir()
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: uploadRoot,
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)
	service := NewMediaService(repository.NewMediaRepository(db), storageService, nil, "https://shop.example.test", 20<<30)

	asset, err := service.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "wheel.png", "image/png", testOpaquePNG(t, 900, 600)),
		MediaType:  "image",
		UploaderID: 42,
		Width:      900,
		Height:     600,
	})
	require.NoError(t, err)
	require.Len(t, asset.Derivatives, 3)
	require.Equal(t, []MediaDerivativePresetDefinition{
		{Name: "thumbnail", Label: "缩略图", MaxWidth: 320, GenerationVersion: 1, SortOrder: 10, IsSystem: true},
		{Name: "card", Label: "卡片图", MaxWidth: 640, GenerationVersion: 1, SortOrder: 20, IsSystem: true},
		{Name: "large", Label: "大图", MaxWidth: 1600, GenerationVersion: 1, SortOrder: 30, IsSystem: true},
	}, MediaDerivativePresetDefinitions())

	byPreset := map[string]mediadomain.MediaAssetDerivative{}
	for _, derivative := range asset.Derivatives {
		byPreset[derivative.Preset] = derivative
		_, err := os.Stat(filepath.Join(uploadRoot, filepath.FromSlash(derivative.StorageKey)))
		require.NoError(t, err)
	}
	require.Equal(t, 320, byPreset["thumbnail"].Width)
	require.Equal(t, 213, byPreset["thumbnail"].Height)
	require.Equal(t, 1, byPreset["thumbnail"].PresetVersion)
	require.Equal(t, 640, byPreset["card"].Width)
	require.Equal(t, 427, byPreset["card"].Height)
	require.Equal(t, 1, byPreset["card"].PresetVersion)
	require.Equal(t, 900, byPreset["large"].Width)
	require.Equal(t, 600, byPreset["large"].Height)
	require.Equal(t, 1, byPreset["large"].PresetVersion)

	variantMap, thumbnailURL, err := service.ProductMediaImageVariants(asset.ID)
	require.NoError(t, err)
	require.Equal(t, byPreset["thumbnail"].URL, thumbnailURL)
	require.Equal(t, byPreset["card"].URL, variantMap["card"].URL)
}

func TestMediaDerivativeBackfillUpgradesLegacyAssetAndProductMedia(t *testing.T) {
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
		&productdomain.ProductMedia{},
	))

	uploadRoot := t.TempDir()
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: uploadRoot,
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)
	service := NewMediaService(repository.NewMediaRepository(db), storageService, nil, "https://shop.example.test", 20<<30)

	payload := testOpaquePNG(t, 900, 600)
	originalURL, err := storageService.UploadFromReader(context.Background(), bytes.NewReader(payload), "legacy-wheel.png")
	require.NoError(t, err)
	legacyAsset := mediadomain.MediaAsset{
		Filename:         "legacy-wheel.png",
		OriginalFilename: "legacy-wheel.png",
		URL:              originalURL,
		StorageKey:       storageObjectKey(storageService, originalURL),
		MimeType:         "image/png",
		MediaType:        "image",
		Size:             int64(len(payload)),
		Width:            900,
		Height:           600,
		UploaderID:       42,
		Status:           "active",
		Visibility:       "public",
	}
	require.NoError(t, db.Create(&legacyAsset).Error)
	require.NoError(t, db.Create(&productdomain.ProductMedia{
		ProductID:    100,
		MediaAssetID: &legacyAsset.ID,
		MediaType:    "image",
		Role:         "primary",
		URL:          originalURL,
		IsPrimary:    true,
		IsVisible:    true,
	}).Error)

	result, err := service.BackfillMissingImageDerivatives(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, result.ScannedAssets)
	require.Equal(t, 1, result.GeneratedAssets)
	require.Equal(t, 3, result.GeneratedDerivatives)
	require.EqualValues(t, 1, result.UpdatedProductMediaRows)

	derivatives, err := repository.NewMediaRepository(db).FindAssetDerivatives(legacyAsset.ID)
	require.NoError(t, err)
	require.Len(t, derivatives, 3)

	var mediaItem productdomain.ProductMedia
	require.NoError(t, db.Where("media_asset_id = ?", legacyAsset.ID).First(&mediaItem).Error)
	variants := productdomain.ParseProductMediaImageVariants(mediaItem.ImageVariantData)
	require.Equal(t, 3, len(variants))
	require.Equal(t, variants["thumbnail"].URL, mediaItem.ThumbnailURL)
	require.NotEqual(t, originalURL, variants["card"].URL)

	secondResult, err := service.BackfillMissingImageDerivatives(context.Background())
	require.NoError(t, err)
	require.Zero(t, secondResult.ScannedAssets)
	require.Zero(t, secondResult.GeneratedDerivatives)
}

func TestMediaDerivativeBackfillRegeneratesVersionedPreset(t *testing.T) {
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
		&productdomain.ProductMedia{},
	))

	presetRepo := repository.NewMediaDerivativePresetRepository(db)
	require.NoError(t, SeedDefaultMediaDerivativePresets(presetRepo))

	uploadRoot := t.TempDir()
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: uploadRoot,
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)
	service := NewMediaService(repository.NewMediaRepository(db), storageService, nil, "https://shop.example.test", 20<<30)
	service.ConfigureDerivativePresetRepository(presetRepo)

	asset, err := service.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "versioned-wheel.png", "image/png", testOpaquePNG(t, 1200, 600)),
		MediaType:  "image",
		UploaderID: 42,
		Width:      1200,
		Height:     600,
	})
	require.NoError(t, err)

	byPreset := map[string]mediadomain.MediaAssetDerivative{}
	for _, derivative := range asset.Derivatives {
		byPreset[derivative.Preset] = derivative
	}
	require.Equal(t, 640, byPreset["card"].Width)
	require.Equal(t, 1, byPreset["card"].PresetVersion)
	oldCardURL := byPreset["card"].URL

	require.NoError(t, db.Create(&productdomain.ProductMedia{
		ProductID:    100,
		MediaAssetID: &asset.ID,
		MediaType:    "image",
		Role:         "primary",
		URL:          asset.URL,
		IsPrimary:    true,
		IsVisible:    true,
	}).Error)
	require.NoError(t, db.Model(&mediadomain.MediaDerivativePreset{}).
		Where("code = ?", "card").
		Updates(map[string]interface{}{
			"max_width":          800,
			"generation_version": 2,
		}).Error)

	engine := NewMediaImageDimensionEngine(service)
	before, err := engine.List(MediaImageDimensionListInput{
		Page:     1,
		PageSize: 20,
		State:    MediaImageDimensionStateMissingVariants,
	})
	require.NoError(t, err)
	require.Len(t, before.Items, 1)
	require.Equal(t, MediaImageDimensionStateMissingVariants, before.Items[0].State)
	require.ElementsMatch(t, []string{"card"}, before.Items[0].MissingPresets)
	require.Equal(t, 2, presetDefinitionByName(before.Presets, "card").GenerationVersion)
	require.Equal(t, 800, presetDefinitionByName(before.Presets, "card").MaxWidth)

	result, err := service.BackfillMissingImageDerivatives(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, result.ScannedAssets)
	require.Equal(t, 1, result.GeneratedAssets)
	require.Equal(t, 1, result.GeneratedDerivatives)
	require.EqualValues(t, 1, result.UpdatedProductMediaRows)

	derivatives, err := repository.NewMediaRepository(db).FindAssetDerivatives(asset.ID)
	require.NoError(t, err)
	require.Len(t, derivatives, 3)
	activeByPreset := map[string]mediadomain.MediaAssetDerivative{}
	for _, derivative := range derivatives {
		activeByPreset[derivative.Preset] = derivative
	}
	require.Equal(t, 800, activeByPreset["card"].Width)
	require.Equal(t, 400, activeByPreset["card"].Height)
	require.Equal(t, 2, activeByPreset["card"].PresetVersion)
	require.NotEqual(t, oldCardURL, activeByPreset["card"].URL)

	var allCardRows []mediadomain.MediaAssetDerivative
	require.NoError(t, db.Unscoped().
		Where("media_asset_id = ? AND preset = ?", asset.ID, "card").
		Order("preset_version ASC").
		Find(&allCardRows).Error)
	require.Len(t, allCardRows, 2)
	require.True(t, allCardRows[0].DeletedAt.Valid)
	require.False(t, allCardRows[1].DeletedAt.Valid)

	var mediaItem productdomain.ProductMedia
	require.NoError(t, db.Where("media_asset_id = ?", asset.ID).First(&mediaItem).Error)
	variants := productdomain.ParseProductMediaImageVariants(mediaItem.ImageVariantData)
	require.Equal(t, activeByPreset["card"].URL, variants["card"].URL)

	after, err := engine.List(MediaImageDimensionListInput{
		Page:     1,
		PageSize: 20,
		State:    MediaImageDimensionStateReady,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, after.Summary.Ready)
	require.EqualValues(t, 0, after.Summary.Attention)
}

func TestMediaImageDimensionReconcileRepairsLegacyAsset(t *testing.T) {
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
	require.NoError(t, db.AutoMigrate(&mediadomain.MediaAsset{}, &mediadomain.MediaAssetDerivative{}))

	uploadRoot := t.TempDir()
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: uploadRoot,
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)
	service := NewMediaService(repository.NewMediaRepository(db), storageService, nil, "https://shop.example.test", 20<<30)
	engine := NewMediaImageDimensionEngine(service)

	payload := testOpaquePNG(t, 900, 600)
	originalURL, err := storageService.UploadFromReader(context.Background(), bytes.NewReader(payload), "legacy-dimensions.png")
	require.NoError(t, err)
	legacyAsset := mediadomain.MediaAsset{
		Filename:         "legacy-dimensions.png",
		OriginalFilename: "legacy-dimensions.png",
		URL:              originalURL,
		StorageKey:       storageObjectKey(storageService, originalURL),
		MimeType:         "image/png",
		MediaType:        "image",
		Size:             int64(len(payload)),
		UploaderID:       42,
		Status:           "active",
		Visibility:       "public",
	}
	require.NoError(t, db.Create(&legacyAsset).Error)

	before, err := engine.List(MediaImageDimensionListInput{
		Page:     1,
		PageSize: 20,
		State:    MediaImageDimensionStateAttention,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, before.Summary.Attention)
	require.Len(t, before.Items, 1)
	require.Equal(t, "missing_dimensions_and_variants", before.Items[0].State)
	require.ElementsMatch(t, []string{"thumbnail", "card", "large"}, before.Items[0].MissingPresets)
	require.Equal(t, []MediaDerivativePresetDefinition{
		{Name: "thumbnail", Label: "缩略图", MaxWidth: 320, GenerationVersion: 1, SortOrder: 10, IsSystem: true},
		{Name: "card", Label: "卡片图", MaxWidth: 640, GenerationVersion: 1, SortOrder: 20, IsSystem: true},
		{Name: "large", Label: "大图", MaxWidth: 1600, GenerationVersion: 1, SortOrder: 30, IsSystem: true},
	}, before.Presets)

	reconciled, err := engine.Reconcile(context.Background(), legacyAsset.ID)
	require.NoError(t, err)
	require.Equal(t, MediaImageDimensionStateReady, reconciled.State)
	require.Equal(t, 900, reconciled.Asset.Width)
	require.Equal(t, 600, reconciled.Asset.Height)
	require.Len(t, reconciled.Asset.Derivatives, 3)

	after, err := engine.List(MediaImageDimensionListInput{
		Page:     1,
		PageSize: 20,
		State:    MediaImageDimensionStateReady,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, after.Summary.Ready)
	require.EqualValues(t, 0, after.Summary.Attention)
	require.Len(t, after.Items, 1)
}

func presetDefinitionByName(presets []MediaDerivativePresetDefinition, name string) MediaDerivativePresetDefinition {
	for _, preset := range presets {
		if preset.Name == name {
			return preset
		}
	}
	return MediaDerivativePresetDefinition{}
}

func testOpaquePNG(t *testing.T, width int, height int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: 120,
				A: 255,
			})
		}
	}

	var buffer bytes.Buffer
	require.NoError(t, png.Encode(&buffer, img))
	return buffer.Bytes()
}
