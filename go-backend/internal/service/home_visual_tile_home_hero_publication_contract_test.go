package service

import (
	"testing"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestHomeHeroVisualShowcasePublishesConfiguredItemsWithoutStatusToggle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&homevisualtile.Tile{}))
	require.NoError(t, db.Create(&homevisualtile.Tile{
		TileSetKey:   HomeHeroVisualShowcaseTileSetKey,
		Locale:       "en",
		ImageURL:     "/uploads/visual-showcase/home-hero/en/first.webp",
		StorageKey:   "visual-showcase/home-hero/en/first.webp",
		Title:        "First",
		AltText:      "First image",
		DesktopOrder: 1,
		Width:        HomeHeroVisualShowcaseImageDimension,
		Height:       HomeHeroVisualShowcaseImageDimension,
		IsPublished:  false,
	}).Error)

	service := NewHomeVisualTileService(repository.NewHomeVisualTileRepository(db), nil)
	result, err := service.GetPublishedResult(HomeHeroVisualShowcaseTileSetKey, "en")
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.Equal(t, "First", result.Items[0].Title)
}
