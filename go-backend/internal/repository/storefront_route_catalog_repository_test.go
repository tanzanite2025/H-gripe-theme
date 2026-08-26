package repository

import (
	"testing"
	"time"

	seodomain "commerce-platform/internal/domain/seo"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestListSitemapEntriesExcludesProductRoutesUnderShop(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&seodomain.StorefrontRouteCatalogEntry{}))

	now := time.Now().UTC()
	entries := []seodomain.StorefrontRouteCatalogEntry{
		{
			RouteKey:      "product:1:en",
			Path:          "/products/real-wheel",
			Locale:        "en",
			SourceType:    seodomain.RouteSourceProduct,
			SourceKey:     "real-wheel",
			CanonicalPath: "/products/real-wheel",
			IsIndexable:   true,
			EntryStatus:   seodomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
		{
			RouteKey:      "product:2:zh_cn",
			Path:          "/zh_cn/shop/legacy-wheel",
			Locale:        "zh_cn",
			SourceType:    seodomain.RouteSourceProduct,
			SourceKey:     "legacy-wheel",
			CanonicalPath: "/zh_cn/shop/legacy-wheel",
			IsIndexable:   true,
			EntryStatus:   seodomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
		{
			RouteKey:      "product:3:en:wrong-shape",
			Path:          "/products/wrong-shape/extra",
			Locale:        "en",
			SourceType:    seodomain.RouteSourceProduct,
			SourceKey:     "wrong-shape",
			CanonicalPath: "/products/wrong-shape/extra",
			IsIndexable:   true,
			EntryStatus:   seodomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
		{
			RouteKey:      "manifest:shop:en",
			Path:          "/shop",
			Locale:        "en",
			SourceType:    seodomain.RouteSourceStatic,
			CanonicalPath: "/shop",
			IsIndexable:   true,
			EntryStatus:   seodomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
	}
	require.NoError(t, db.Create(&entries).Error)

	result, err := NewStorefrontRouteCatalogRepository(db).ListSitemapEntries(100)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "/products/real-wheel", result[0].Path)
	require.Equal(t, "/shop", result[1].Path)
}
