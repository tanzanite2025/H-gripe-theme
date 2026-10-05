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

func TestStorefrontURLIssueRepositoryHidesHistoricalButShowsFreshRuntimeIssuesForStaleRoutes(t *testing.T) {
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
	latestCheckResultID := uint(42)
	previousCheckResultID := uint(41)
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
		{
			RouteKey:        "manifest:checked-legacy:ar",
			Path:            "/ar/checked-legacy",
			Locale:          "ar",
			SourceType:      seodomain.RouteSourceStatic,
			EntryStatus:     seodomain.RouteEntryStatusStale,
			IsCheckable:     true,
			LastCheckStatus: seodomain.RouteCheckStatusRedirectChain,
			LastSeenAt:      now,
		},
		{
			RouteKey:        "manifest:previously-checked-legacy:ar",
			Path:            "/ar/previously-checked-legacy",
			Locale:          "ar",
			SourceType:      seodomain.RouteSourceStatic,
			EntryStatus:     seodomain.RouteEntryStatusStale,
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
		{
			RouteEntryID:        entries[2].ID,
			IssueType:           urlmanagementdomain.URLIssueTypeRedirectChain,
			Severity:            urlmanagementdomain.URLIssueSeverityHigh,
			State:               urlmanagementdomain.URLIssueStateOpen,
			LatestCheckResultID: &latestCheckResultID,
			FirstDetectedAt:     now,
			LastDetectedAt:      now,
		},
		{
			RouteEntryID:        entries[3].ID,
			IssueType:           urlmanagementdomain.URLIssueTypeNotFound,
			Severity:            urlmanagementdomain.URLIssueSeverityHigh,
			State:               urlmanagementdomain.URLIssueStateVerified,
			LatestCheckResultID: &previousCheckResultID,
			FirstDetectedAt:     now,
			LastDetectedAt:      now,
		},
	}).Error)

	issues, total, err := NewStorefrontURLIssueRepository(db).List(StorefrontURLIssueListFilter{
		Page:     1,
		PageSize: 20,
		State:    "all",
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, issues, 2)
	listedPaths := map[string]bool{}
	for _, issue := range issues {
		require.NotNil(t, issue.RouteEntry)
		listedPaths[issue.RouteEntry.Path] = true
	}
	require.True(t, listedPaths["/ar/checked-legacy"])
	require.True(t, listedPaths["/ar/resources/blog"])

	stats, err := NewStorefrontURLIssueRepository(db).Stats()
	require.NoError(t, err)
	require.Equal(t, int64(2), stats.Active)
	require.Equal(t, int64(2), stats.Open)
	require.Equal(t, int64(2), stats.High)
	require.Zero(t, stats.Verified)
}

func TestAutomaticallyVerifyIssuesNoLongerDetectedForRouteEntryKeepsHistory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&urlmanagementdomain.StorefrontURLIssue{},
		&urlmanagementdomain.StorefrontURLIssueEvent{},
	))

	checkedAt := time.Now().UTC().Add(-time.Hour)
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey:        "manifest:current:en",
		Path:            "/resources/blog",
		Locale:          "en",
		SourceType:      seodomain.RouteSourceStatic,
		EntryStatus:     seodomain.RouteEntryStatusActive,
		IsCheckable:     true,
		LastSeenAt:      checkedAt,
		LastCheckStatus: seodomain.RouteCheckStatusOK,
	}
	require.NoError(t, db.Create(&entry).Error)
	issue := urlmanagementdomain.StorefrontURLIssue{
		RouteEntryID:    entry.ID,
		IssueType:       urlmanagementdomain.URLIssueTypeNotFound,
		Severity:        urlmanagementdomain.URLIssueSeverityHigh,
		State:           urlmanagementdomain.URLIssueStateOpen,
		FirstDetectedAt: checkedAt,
		LastDetectedAt:  checkedAt,
	}
	require.NoError(t, db.Create(&issue).Error)

	verifiedAt := time.Now().UTC()
	verifiedCount, err := NewStorefrontURLIssueRepository(db).
		AutomaticallyVerifyIssuesNoLongerDetectedForRouteEntry(entry.ID, nil, verifiedAt)
	require.NoError(t, err)
	require.Equal(t, int64(1), verifiedCount)

	var refreshed urlmanagementdomain.StorefrontURLIssue
	require.NoError(t, db.First(&refreshed, issue.ID).Error)
	require.Equal(t, urlmanagementdomain.URLIssueStateVerified, refreshed.State)
	require.Equal(t, urlmanagementdomain.URLIssueResolutionRuntimeFixed, refreshed.ResolutionType)
	require.NotNil(t, refreshed.VerifiedAt)

	var event urlmanagementdomain.StorefrontURLIssueEvent
	require.NoError(t, db.Where("issue_id = ? AND event_type = ?", issue.ID, urlmanagementdomain.URLIssueEventVerificationPassed).
		First(&event).Error)
	require.Contains(t, event.Note, "自动验证关闭")
}

func TestRetireAllActiveStaleRouteIssuesMarksOnlyStalePaths(t *testing.T) {
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
			RouteKey: "manifest:stale:en", Path: "/legacy", Locale: "en", SourceType: seodomain.RouteSourceStatic,
			EntryStatus: seodomain.RouteEntryStatusStale, IsCheckable: true, LastSeenAt: now,
		},
		{
			RouteKey: "manifest:active:en", Path: "/current", Locale: "en", SourceType: seodomain.RouteSourceStatic,
			EntryStatus: seodomain.RouteEntryStatusActive, IsCheckable: true, LastSeenAt: now,
		},
	}
	require.NoError(t, db.Create(&entries).Error)
	issues := []urlmanagementdomain.StorefrontURLIssue{
		{RouteEntryID: entries[0].ID, IssueType: urlmanagementdomain.URLIssueTypeStaleRoute, Severity: urlmanagementdomain.URLIssueSeverityMedium, State: urlmanagementdomain.URLIssueStateOpen, FirstDetectedAt: now, LastDetectedAt: now},
		{RouteEntryID: entries[1].ID, IssueType: urlmanagementdomain.URLIssueTypeStaleRoute, Severity: urlmanagementdomain.URLIssueSeverityMedium, State: urlmanagementdomain.URLIssueStateOpen, FirstDetectedAt: now, LastDetectedAt: now},
	}
	require.NoError(t, db.Create(&issues).Error)

	retiredAt := now.Add(time.Minute)
	retiredCount, err := NewStorefrontURLIssueRepository(db).
		RetireAllActiveStaleRouteIssues(7, retiredAt)
	require.NoError(t, err)
	require.Equal(t, int64(1), retiredCount)

	var staleIssue urlmanagementdomain.StorefrontURLIssue
	require.NoError(t, db.First(&staleIssue, issues[0].ID).Error)
	require.Equal(t, urlmanagementdomain.URLIssueStateVerified, staleIssue.State)
	require.Equal(t, urlmanagementdomain.URLIssueResolutionRetired, staleIssue.ResolutionType)
	_, err = NewStorefrontURLIssueRepository(db).RecordDetection(
		entries[0].ID,
		urlmanagementdomain.URLIssueTypeStaleRoute,
		urlmanagementdomain.URLIssueSeverityMedium,
		nil,
		now.Add(2*time.Minute),
	)
	require.NoError(t, err)
	require.NoError(t, db.First(&staleIssue, issues[0].ID).Error)
	require.Equal(t, urlmanagementdomain.URLIssueStateVerified, staleIssue.State)

	var activeIssue urlmanagementdomain.StorefrontURLIssue
	require.NoError(t, db.First(&activeIssue, issues[1].ID).Error)
	require.Equal(t, urlmanagementdomain.URLIssueStateOpen, activeIssue.State)
}
