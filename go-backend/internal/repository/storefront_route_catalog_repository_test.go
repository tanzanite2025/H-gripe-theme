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

func TestStorefrontRouteCatalogRepositoryListCheckableOnlyExcludesStaleAndUncheckable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&seodomain.StorefrontRouteCatalogEntry{}))

	now := time.Now().UTC()
	entries := []seodomain.StorefrontRouteCatalogEntry{
		{
			RouteKey:    "static:checkable",
			Path:        "/checkable",
			Locale:      "en",
			SourceType:  seodomain.RouteSourceStatic,
			IsCheckable: true,
			EntryStatus: seodomain.RouteEntryStatusActive,
			LastSeenAt:  now,
		},
		{
			RouteKey:    "static:uncheckable",
			Path:        "/uncheckable",
			Locale:      "en",
			SourceType:  seodomain.RouteSourceStatic,
			IsCheckable: false,
			EntryStatus: seodomain.RouteEntryStatusActive,
			LastSeenAt:  now,
		},
		{
			RouteKey:    "static:stale",
			Path:        "/stale",
			Locale:      "en",
			SourceType:  seodomain.RouteSourceStatic,
			IsCheckable: true,
			EntryStatus: seodomain.RouteEntryStatusStale,
			LastSeenAt:  now,
		},
	}
	require.NoError(t, db.Create(&entries).Error)
	require.NoError(t, db.Model(&seodomain.StorefrontRouteCatalogEntry{}).
		Where("route_key = ?", "static:uncheckable").
		Update("is_checkable", false).Error)

	result, total, err := NewStorefrontRouteCatalogRepository(db).List(StorefrontRouteCatalogListFilter{
		Page:          1,
		PageSize:      20,
		CheckableOnly: true,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, result, 1)
	require.Equal(t, "/checkable", result[0].Path)
}

func TestStorefrontRouteCatalogRepositoryStatsFiltersByLocale(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&seodomain.StorefrontRouteCatalogEntry{}))

	now := time.Now().UTC()
	entries := []seodomain.StorefrontRouteCatalogEntry{
		{
			RouteKey:        "static:en:healthy",
			Path:            "/en/healthy",
			Locale:          "en",
			SourceType:      seodomain.RouteSourceStatic,
			IsCheckable:     true,
			EntryStatus:     seodomain.RouteEntryStatusActive,
			LastCheckStatus: seodomain.RouteCheckStatusOK,
			LastSeenAt:      now,
		},
		{
			RouteKey:        "static:en:missing",
			Path:            "/en/missing",
			Locale:          "en",
			SourceType:      seodomain.RouteSourceStatic,
			IsCheckable:     true,
			EntryStatus:     seodomain.RouteEntryStatusActive,
			LastCheckStatus: seodomain.RouteCheckStatusNotFound,
			LastSeenAt:      now,
		},
		{
			RouteKey:    "static:zh:unchecked",
			Path:        "/zh_cn/unchecked",
			Locale:      "zh_cn",
			SourceType:  seodomain.RouteSourceStatic,
			IsCheckable: true,
			EntryStatus: seodomain.RouteEntryStatusActive,
			LastSeenAt:  now,
		},
		{
			RouteKey:    "static:zh:stale",
			Path:        "/zh_cn/stale",
			Locale:      "zh_cn",
			SourceType:  seodomain.RouteSourceStatic,
			IsCheckable: true,
			EntryStatus: seodomain.RouteEntryStatusStale,
			LastSeenAt:  now,
		},
	}
	require.NoError(t, db.Create(&entries).Error)

	repo := NewStorefrontRouteCatalogRepository(db)

	allStats, err := repo.Stats()
	require.NoError(t, err)
	require.Equal(t, int64(4), allStats.Total)
	require.Equal(t, int64(3), allStats.Checkable)

	enStats, err := repo.StatsForLocale("en")
	require.NoError(t, err)
	require.Equal(t, int64(2), enStats.Total)
	require.Equal(t, int64(1), enStats.OK)
	require.Equal(t, int64(1), enStats.NotFound)
	require.Equal(t, int64(2), enStats.Checked)
	require.Equal(t, int64(2), enStats.Checkable)
	require.Equal(t, int64(1), enStats.NeedsAttention)

	zhStats, err := repo.StatsForLocale("zh_cn")
	require.NoError(t, err)
	require.Equal(t, int64(2), zhStats.Total)
	require.Equal(t, int64(1), zhStats.Unchecked)
	require.Equal(t, int64(1), zhStats.Checkable)
	require.Equal(t, int64(1), zhStats.Stale)
}

func TestStatsForLocaleAndScopeRestrictsCanonicalMetrics(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&seodomain.StorefrontRouteCatalogEntry{}))
	now := time.Now().UTC()
	require.NoError(t, db.Create(&[]seodomain.StorefrontRouteCatalogEntry{
		{RouteKey: "canonical:en:mismatch", Path: "/en/mismatch", Locale: "en", SourceType: seodomain.RouteSourceStatic, EntryStatus: seodomain.RouteEntryStatusActive, LastCheckStatus: seodomain.RouteCheckStatusCanonicalMisfit, LastSeenAt: now},
		{RouteKey: "canonical:en:duplicate", Path: "/en/duplicate", Locale: "en", SourceType: seodomain.RouteSourceStatic, EntryStatus: seodomain.RouteEntryStatusDuplicate, LastSeenAt: now},
		{RouteKey: "canonical:en:healthy", Path: "/en/healthy", Locale: "en", SourceType: seodomain.RouteSourceStatic, EntryStatus: seodomain.RouteEntryStatusActive, LastCheckStatus: seodomain.RouteCheckStatusOK, LastSeenAt: now},
		{RouteKey: "canonical:zh:mismatch", Path: "/zh_cn/mismatch", Locale: "zh_cn", SourceType: seodomain.RouteSourceStatic, EntryStatus: seodomain.RouteEntryStatusActive, LastCheckStatus: seodomain.RouteCheckStatusCanonicalMisfit, LastSeenAt: now},
	}).Error)

	repo := NewStorefrontRouteCatalogRepository(db)
	stats, err := repo.StatsForLocaleAndScope("en", "canonical")
	require.NoError(t, err)
	require.Equal(t, int64(2), stats.Total)
	require.Equal(t, int64(1), stats.CanonicalMismatch)
	require.Equal(t, int64(1), stats.Duplicate)
	require.Equal(t, int64(2), stats.NeedsAttention)

	allStats, err := repo.StatsForLocale("en")
	require.NoError(t, err)
	require.Equal(t, int64(3), allStats.Total)
}

func TestUpsertSnapshotClearsCurrentCheckProjectionAndPreservesHistory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&seodomain.StorefrontRouteCheckResult{},
	))

	checkedAt := time.Now().UTC().Add(-24 * time.Hour)
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey:        "manifest:resources-blog:en",
		Path:            "/blog",
		Locale:          "en",
		SourceType:      seodomain.RouteSourceStatic,
		CanonicalPath:   "/blog",
		EntryStatus:     seodomain.RouteEntryStatusActive,
		IsCheckable:     true,
		LastSeenAt:      checkedAt,
		LastCheckStatus: seodomain.RouteCheckStatusNotFound,
		LastHTTPStatus:  404,
		LastFinalURL:    "http://localhost:9200/blog",
		LastResponseMS:  1,
		LastCheckedAt:   &checkedAt,
	}
	require.NoError(t, db.Create(&entry).Error)
	require.NoError(t, db.Create(&seodomain.StorefrontRouteCheckResult{
		RouteEntryID: entry.ID,
		CheckedAt:    checkedAt,
		HTTPStatus:   404,
		Status:       seodomain.RouteCheckStatusNotFound,
	}).Error)

	seenAt := time.Now().UTC()
	err = NewStorefrontRouteCatalogRepository(db).UpsertSnapshot([]seodomain.StorefrontRouteCatalogEntry{{
		RouteKey:        entry.RouteKey,
		Path:            "/resources/blog",
		Locale:          "en",
		SourceType:      seodomain.RouteSourceStatic,
		CanonicalPath:   "/resources/blog",
		ManifestVersion: "2026-08-23",
		EntryStatus:     seodomain.RouteEntryStatusActive,
		IsCheckable:     true,
		LastSeenAt:      seenAt,
	}}, seenAt)
	require.NoError(t, err)

	var refreshed seodomain.StorefrontRouteCatalogEntry
	require.NoError(t, db.First(&refreshed, entry.ID).Error)
	require.Equal(t, "/resources/blog", refreshed.Path)
	require.Empty(t, refreshed.LastCheckStatus)
	require.Zero(t, refreshed.LastHTTPStatus)
	require.Empty(t, refreshed.LastFinalURL)
	require.Zero(t, refreshed.LastResponseMS)
	require.Nil(t, refreshed.LastCheckedAt)

	var historyCount int64
	require.NoError(t, db.Model(&seodomain.StorefrontRouteCheckResult{}).
		Where("route_entry_id = ?", entry.ID).
		Count(&historyCount).Error)
	require.Equal(t, int64(1), historyCount)
}

func TestUpsertSnapshotClearsCheckProjectionForStaleEntries(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&seodomain.StorefrontRouteCatalogEntry{}))

	checkedAt := time.Now().UTC().Add(-24 * time.Hour)
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey:        "manifest:legacy:en",
		Path:            "/blog",
		Locale:          "en",
		SourceType:      seodomain.RouteSourceStatic,
		CanonicalPath:   "/blog",
		EntryStatus:     seodomain.RouteEntryStatusActive,
		IsCheckable:     true,
		LastSeenAt:      checkedAt,
		LastCheckStatus: seodomain.RouteCheckStatusNotFound,
		LastHTTPStatus:  404,
		LastCheckedAt:   &checkedAt,
	}
	require.NoError(t, db.Create(&entry).Error)

	seenAt := time.Now().UTC()
	require.NoError(t, NewStorefrontRouteCatalogRepository(db).UpsertSnapshot(nil, seenAt))

	var refreshed seodomain.StorefrontRouteCatalogEntry
	require.NoError(t, db.First(&refreshed, entry.ID).Error)
	require.Equal(t, seodomain.RouteEntryStatusStale, refreshed.EntryStatus)
	require.Empty(t, refreshed.LastCheckStatus)
	require.Zero(t, refreshed.LastHTTPStatus)
	require.Nil(t, refreshed.LastCheckedAt)
}
