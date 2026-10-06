package repository

import (
	"testing"

	"commerce-platform/internal/domain/homevisualtile"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestHomeVisualTileRepositorySaveSingleVisualShowcaseItemKeepsOtherSlots(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&homevisualtile.Tile{}))

	first := homevisualtile.Tile{
		TileSetKey:   "home-hero",
		Locale:       "en",
		ImageURL:     "/uploads/visual-showcase/home-hero/en/first.webp",
		StorageKey:   "visual-showcase/home-hero/en/first.webp",
		Title:        "First",
		AltText:      "First image",
		DesktopOrder: 1,
		Width:        600,
		Height:       600,
		IsPublished:  true,
	}
	second := first
	second.ID = 0
	second.ImageURL = "/uploads/visual-showcase/home-hero/en/second.webp"
	second.StorageKey = "visual-showcase/home-hero/en/second.webp"
	second.Title = "Second"
	second.AltText = "Second image"
	second.DesktopOrder = 2
	require.NoError(t, db.Create(&first).Error)
	require.NoError(t, db.Create(&second).Error)

	repository := NewHomeVisualTileRepository(db)
	replacement := first
	replacement.ImageURL = "/uploads/visual-showcase/home-hero/en/replacement.webp"
	replacement.StorageKey = "visual-showcase/home-hero/en/replacement.webp"
	replacement.Title = "Replacement"
	replacement.AltText = "Replacement image"
	var previous *homevisualtile.Tile
	require.NoError(t, repository.SaveSingleVisualShowcaseItem(&replacement, func(_ *gorm.DB, old *homevisualtile.Tile) error {
		previous = old
		return nil
	}))

	var savedItems []homevisualtile.Tile
	require.NoError(t, db.Where("showcase_key = ? AND locale = ?", "home-hero", "en").Order("desktop_order").Find(&savedItems).Error)
	require.Len(t, savedItems, 2)
	require.Equal(t, "Replacement", savedItems[0].Title)
	require.Equal(t, "Second", savedItems[1].Title)
	require.NotNil(t, previous)
	require.Equal(t, "First", previous.Title)
}
