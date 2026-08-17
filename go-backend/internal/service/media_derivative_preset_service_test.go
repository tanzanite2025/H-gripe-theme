package service

import (
	"context"
	"testing"

	mediadomain "commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMediaDerivativePresetServiceLifecycle(t *testing.T) {
	service, presetRepo, db := newMediaDerivativePresetTestService(t)
	enabled := true

	preset, err := service.CreateMediaDerivativePreset(MediaDerivativePresetInput{
		Code:      "Homepage-Hero",
		Label:     "首页主图",
		MaxWidth:  1200,
		SortOrder: 40,
		Enabled:   &enabled,
	})
	require.NoError(t, err)
	require.Equal(t, "homepage-hero", preset.Code)
	require.Equal(t, 1, preset.GenerationVersion)
	require.True(t, preset.Enabled)

	_, err = service.CreateMediaDerivativePreset(MediaDerivativePresetInput{
		Code:      "homepage-hero",
		Label:     "重复定义",
		MaxWidth:  1200,
		SortOrder: 50,
	})
	require.ErrorIs(t, err, ErrMediaDerivativePresetConflict)

	updated, err := service.UpdateMediaDerivativePreset(preset.ID, MediaDerivativePresetInput{
		Label:     "首页横幅",
		MaxWidth:  1440,
		SortOrder: 60,
		Enabled:   &enabled,
	})
	require.NoError(t, err)
	require.Equal(t, 2, updated.GenerationVersion)
	require.Equal(t, 1440, updated.MaxWidth)

	updated, err = service.UpdateMediaDerivativePreset(preset.ID, MediaDerivativePresetInput{
		Label:     "首页横幅图",
		MaxWidth:  1440,
		SortOrder: 70,
		Enabled:   &enabled,
	})
	require.NoError(t, err)
	require.Equal(t, 2, updated.GenerationVersion)

	_, err = service.UpdateMediaDerivativePreset(preset.ID, MediaDerivativePresetInput{
		Code:      "renamed-banner",
		Label:     "首页横幅图",
		MaxWidth:  1440,
		SortOrder: 70,
		Enabled:   &enabled,
	})
	require.ErrorIs(t, err, ErrMediaDerivativePresetCodeImmutable)

	asset := mediadomain.MediaAsset{
		Filename:         "homepage.png",
		OriginalFilename: "homepage.png",
		URL:              "https://media.example.test/homepage.png",
		StorageKey:       "uploads/homepage.png",
		MimeType:         "image/png",
		MediaType:        "image",
		Width:            1600,
		Height:           900,
		Status:           "active",
		Visibility:       "public",
	}
	require.NoError(t, db.Create(&asset).Error)
	require.NoError(t, db.Create(&mediadomain.MediaAssetDerivative{
		MediaAssetID:  asset.ID,
		Preset:        preset.Code,
		PresetVersion: updated.GenerationVersion,
		URL:           "https://media.example.test/derivatives/homepage-hero.png",
		StorageKey:    "media-derivatives/1/homepage-hero/v2/homepage.png",
		MimeType:      "image/png",
		Width:         1440,
		Height:        810,
	}).Error)

	presets, err := service.ListMediaDerivativePresets()
	require.NoError(t, err)
	require.Len(t, presets, 1)
	require.EqualValues(t, 1, presets[0].GeneratedDerivatives)

	err = service.DeleteMediaDerivativePreset(preset.ID)
	require.ErrorIs(t, err, ErrMediaDerivativePresetInUse)

	disabled, err := service.SetMediaDerivativePresetEnabled(preset.ID, false)
	require.NoError(t, err)
	require.False(t, disabled.Enabled)

	systemPreset := mediadomain.MediaDerivativePreset{
		Code:              "thumbnail",
		Label:             "缩略图",
		MaxWidth:          320,
		SortOrder:         10,
		Enabled:           true,
		GenerationVersion: 1,
		IsSystem:          true,
	}
	require.NoError(t, presetRepo.Create(&systemPreset))
	require.ErrorIs(t, service.DeleteMediaDerivativePreset(systemPreset.ID), ErrMediaDerivativePresetProtected)

	removable, err := service.CreateMediaDerivativePreset(MediaDerivativePresetInput{
		Code:      "unused-preview",
		Label:     "未使用预览图",
		MaxWidth:  480,
		SortOrder: 90,
	})
	require.NoError(t, err)
	require.NoError(t, service.DeleteMediaDerivativePreset(removable.ID))
	_, err = service.GetMediaDerivativePreset(removable.ID)
	require.ErrorIs(t, err, ErrMediaDerivativePresetNotFound)
}

func TestMediaDerivativePresetServiceDrivesCustomUploadGeneration(t *testing.T) {
	mediaService, presetRepo, _ := newMediaDerivativePresetTestService(t)
	require.NoError(t, SeedDefaultMediaDerivativePresets(presetRepo))

	preset, err := mediaService.CreateMediaDerivativePreset(MediaDerivativePresetInput{
		Code:      "article-banner",
		Label:     "文章横幅",
		MaxWidth:  700,
		SortOrder: 40,
	})
	require.NoError(t, err)

	uploadRoot := t.TempDir()
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: uploadRoot,
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)
	mediaService.storage = storageService

	asset, err := mediaService.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "article.png", "image/png", testOpaquePNG(t, 900, 600)),
		MediaType:  "image",
		UploaderID: 42,
		Width:      900,
		Height:     600,
	})
	require.NoError(t, err)
	require.Len(t, asset.Derivatives, 4)

	var custom *mediadomain.MediaAssetDerivative
	for index := range asset.Derivatives {
		if asset.Derivatives[index].Preset == preset.Code {
			custom = &asset.Derivatives[index]
			break
		}
	}
	require.NotNil(t, custom)
	require.Equal(t, 700, custom.Width)
	require.Equal(t, 1, custom.PresetVersion)
}

func newMediaDerivativePresetTestService(t *testing.T) (*MediaService, *repository.MediaDerivativePresetRepository, *gorm.DB) {
	t.Helper()

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
	))

	presetRepo := repository.NewMediaDerivativePresetRepository(db)
	mediaService := NewMediaService(repository.NewMediaRepository(db), nil, nil, "https://shop.example.test", 20<<30)
	mediaService.ConfigureDerivativePresetRepository(presetRepo)
	return mediaService, presetRepo, db
}
