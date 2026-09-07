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

func TestStorefrontURLIssueRepositoryHidesRuntimeIssuesForStaleRoutes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&urlmanagementdomain.StorefrontURLIssue{},
		&urlmanagementdomain.StorefrontURLIssueEvent{},
	))

	now := time.Now().UTC()
	entries := []seodomain.StorefrontRouteCatalogEntry{
		{
			RouteKey:        "manifest:legacy:ar",
			Path:            "/ar/blog",
			Locale:          "ar",
			SourceType:      seodomain.RouteSourceStatic,
			EntryStatus:     seodomain.RouteEntryStatusStale,
			IsCheckable:     true,
			LastCheckStatus: seodomain.RouteCheckStatusNotFound,
			LastSeenAt:      now,
		},
		{
			RouteKey:        "manifest:current:ar",
			Path:            "/ar/resources/blog",
			Locale:          "ar",
			SourceType:      seodomain.RouteSourceStatic,
			EntryStatus:     seodomain.RouteEntryStatusActive,
			IsCheckable:     true,
			LastCheckStatus: seodomain.RouteCheckStatusNotFound,
			LastSeenAt:      now,
		},
	}
	require.NoError(t, db.Create(&entries).Error)

	require.NoError(t, db.Create([]urlmanagementdomain.StorefrontURLIssue{
		{
			RouteEntryID:    entries[0].ID,
			IssueType:       urlmanagementdomain.URLIssueTypeNotFound,
			Severity:        urlmanagementdomain.URLIssueSeverityHigh,
			State:           urlmanagementdomain.URLIssueStateOpen,
			FirstDetectedAt: now,
			LastDetectedAt:  now,
		},
		{
			RouteEntryID:    entries[1].ID,
			IssueType:       urlmanagementdomain.URLIssueTypeNotFound,
			Severity:        urlmanagementdomain.URLIssueSeverityHigh,
			State:           urlmanagementdomain.URLIssueStateOpen,
			FirstDetectedAt: now,
			LastDetectedAt:  now,
		},
	}).Error)

	issues, total, err := NewStorefrontURLIssueRepository(db).List(StorefrontURLIssueListFilter{
		Page:     1,
		PageSize: 20,
		State:    "all",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, issues, 1)
	require.NotNil(t, issues[0].RouteEntry)
	require.Equal(t, "/ar/resources/blog", issues[0].RouteEntry.Path)

	stats, err := NewStorefrontURLIssueRepository(db).Stats()
	require.NoError(t, err)
	require.Equal(t, int64(1), stats.Active)
	require.Equal(t, int64(1), stats.Open)
	require.Equal(t, int64(1), stats.High)
}

func TestInvalidateRuntimeIssuesForCatalogSyncHidesOldObservations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&urlmanagementdomain.StorefrontURLIssue{},
		&urlmanagementdomain.StorefrontURLIssueEvent{},
	))

	checkedAt := time.Now().UTC().Add(-24 * time.Hour)
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey:        "manifest:current:en",
		Path:            "/resources/blog",
		Locale:          "en",
		SourceType:      seodomain.RouteSourceStatic,
		EntryStatus:     seodomain.RouteEntryStatusActive,
		IsCheckable:     true,
		LastCheckStatus: "",
		LastSeenAt:      time.Now().UTC(),
	}
	require.NoError(t, db.Create(&entry).Error)

	issue := urlmanagementdomain.StorefrontURLIssue{
		RouteEntryID:        entry.ID,
		IssueType:           urlmanagementdomain.URLIssueTypeNotFound,
		Severity:            urlmanagementdomain.URLIssueSeverityHigh,
		State:               urlmanagementdomain.URLIssueStateOpen,
		FirstDetectedAt:     checkedAt,
		LastDetectedAt:      checkedAt,
		LatestCheckResultID: nil,
	}
	require.NoError(t, db.Create(&issue).Error)

	syncedAt := time.Now().UTC()
	require.NoError(t, NewStorefrontURLIssueRepository(db).
		InvalidateRuntimeIssuesForCatalogSync(syncedAt))

	var refreshed urlmanagementdomain.StorefrontURLIssue
	require.NoError(t, db.First(&refreshed, issue.ID).Error)
	require.Equal(t, urlmanagementdomain.URLIssueStateVerified, refreshed.State)
	require.Equal(t, urlmanagementdomain.URLIssueResolutionNotApplicable, refreshed.ResolutionType)
	require.NotNil(t, refreshed.VerifiedAt)

	var eventCount int64
	require.NoError(t, db.Model(&urlmanagementdomain.StorefrontURLIssueEvent{}).
		Where("issue_id = ? AND event_type = ?", issue.ID, urlmanagementdomain.URLIssueEventSnapshotInvalidated).
		Count(&eventCount).Error)
	require.Equal(t, int64(1), eventCount)
}
