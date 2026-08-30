package repository

import (
	"testing"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"

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

func TestStorefrontRouteCatalogRepositoryListFiltersBySearchProfileStatus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&urlmanagementdomain.StorefrontURLSearchProfile{},
	))

	now := time.Now().UTC()
	entries := []seodomain.StorefrontRouteCatalogEntry{
		{
			RouteKey:      "static:configured",
			Path:          "/en/configured",
			Locale:        "en",
			SourceType:    seodomain.RouteSourceStatic,
			Title:         "Configured",
			Summary:       "Configured summary",
			CanonicalPath: "/en/configured",
			IsIndexable:   true,
			EntryStatus:   seodomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
		{
			RouteKey:      "static:unconfigured",
			Path:          "/en/unconfigured",
			Locale:        "en",
			SourceType:    seodomain.RouteSourceStatic,
			Title:         "Unconfigured",
			Summary:       "Unconfigured summary",
			CanonicalPath: "/en/unconfigured",
			IsIndexable:   true,
			EntryStatus:   seodomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
	}
	require.NoError(t, db.Create(&entries).Error)
	require.NoError(t, db.Create(&urlmanagementdomain.StorefrontURLSearchProfile{
		RouteEntryID:   entries[0].ID,
		Enabled:        true,
		SearchWeight:   100,
		DisplayTitle:   "Configured",
		DisplaySummary: "Configured summary",
	}).Error)

	repo := NewStorefrontRouteCatalogRepository(db)

	configured, total, err := repo.List(StorefrontRouteCatalogListFilter{
		Page:                1,
		PageSize:            20,
		SearchProfileStatus: "configured",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, configured, 1)
	require.Equal(t, entries[0].ID, configured[0].ID)

	unconfigured, total, err := repo.List(StorefrontRouteCatalogListFilter{
		Page:                1,
		PageSize:            20,
		SearchProfileStatus: "unconfigured",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, unconfigured, 1)
	require.Equal(t, entries[1].ID, unconfigured[0].ID)
}
