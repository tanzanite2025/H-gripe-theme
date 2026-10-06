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

func TestStorefrontRouteCatalogStatsIncludesLatestCheckTime(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&seodomain.StorefrontRouteCatalogEntry{}))

	now := time.Now().UTC().Truncate(time.Second)
	firstCheckAt := now.Add(-10 * time.Minute)
	latestCheckAt := now.Add(-time.Minute)
	entries := []seodomain.StorefrontRouteCatalogEntry{
		{
			RouteKey:        "static:older-check",
			Path:            "/older-check",
			Locale:          "en",
			SourceType:      seodomain.RouteSourceStatic,
			EntryStatus:     seodomain.RouteEntryStatusActive,
			LastCheckStatus: seodomain.RouteCheckStatusOK,
			LastCheckedAt:   &firstCheckAt,
			LastSeenAt:      now,
		},
		{
			RouteKey:        "static:latest-check",
			Path:            "/latest-check",
			Locale:          "en",
			SourceType:      seodomain.RouteSourceStatic,
			EntryStatus:     seodomain.RouteEntryStatusActive,
			LastCheckStatus: seodomain.RouteCheckStatusNotFound,
			LastCheckedAt:   &latestCheckAt,
			LastSeenAt:      now,
		},
	}
	require.NoError(t, db.Create(&entries).Error)

	stats, err := NewStorefrontRouteCatalogRepository(db).Stats()
	require.NoError(t, err)
	require.NotNil(t, stats.LastCheckedAt)
	require.WithinDuration(t, latestCheckAt, *stats.LastCheckedAt, time.Second)
}
