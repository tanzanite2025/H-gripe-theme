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

func openCurrentStorefrontRouteSnapshotTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{}, &seodomain.StorefrontRouteCheckResult{},
		&urlmanagementdomain.StorefrontURLIssue{}, &urlmanagementdomain.StorefrontURLIssueEvent{},
		&urlmanagementdomain.StorefrontURLSearchProfile{},
	))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	return db
}

func TestCurrentStorefrontSnapshotDeletesRemovedURLsAndRelatedRecordsWhilePreservingCurrentSearchSettings(t *testing.T) {
	db := openCurrentStorefrontRouteSnapshotTestDatabase(t)
	previousSync := time.Now().UTC().Add(-time.Hour)
	entries := []seodomain.StorefrontRouteCatalogEntry{
		{RouteKey: "manifest:current:en", Path: "/current", Locale: "en", SourceType: "static", EntryStatus: "active", LastSeenAt: previousSync},
		{RouteKey: "manifest:removed:en", Path: "/removed", Locale: "en", SourceType: "static", EntryStatus: "active", LastSeenAt: previousSync},
	}
	require.NoError(t, db.Create(&entries).Error)
	for _, entry := range entries {
		require.NoError(t, db.Create(&urlmanagementdomain.StorefrontURLSearchProfile{RouteEntryID: entry.ID, DisplayTitle: "Customized search title"}).Error)
	}
	oldCheck := seodomain.StorefrontRouteCheckResult{RouteEntryID: entries[1].ID, CheckedAt: previousSync, Status: "not_found", HTTPStatus: 404}
	require.NoError(t, db.Create(&oldCheck).Error)
	oldIssue := urlmanagementdomain.StorefrontURLIssue{
		RouteEntryID: entries[1].ID, IssueType: "not_found", Severity: "high", State: "resolved",
		LatestCheckResultID: &oldCheck.ID, FirstDetectedAt: previousSync, LastDetectedAt: previousSync,
	}
	require.NoError(t, db.Create(&oldIssue).Error)
	require.NoError(t, db.Create(&urlmanagementdomain.StorefrontURLIssueEvent{IssueID: oldIssue.ID, EventType: "detected", Metadata: "{}", CreatedAt: previousSync}).Error)
	require.NoError(t, db.Exec(`CREATE TABLE site_quality_targets (id INTEGER PRIMARY KEY, route_entry_id INTEGER, enabled BOOLEAN, next_scheduled_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO site_quality_targets (route_entry_id, enabled, next_scheduled_at) VALUES (?, TRUE, ?)`, entries[1].ID, previousSync).Error)

	require.NoError(t, NewStorefrontRouteCatalogRepository(db).UpsertSnapshot(entries[:1], time.Now().UTC()))
	var count int64
	for _, model := range []interface{}{&seodomain.StorefrontRouteCheckResult{}, &urlmanagementdomain.StorefrontURLIssue{}, &urlmanagementdomain.StorefrontURLIssueEvent{}} {
		require.NoError(t, db.Model(model).Count(&count).Error)
		require.Zero(t, count)
	}
	require.NoError(t, db.Model(&seodomain.StorefrontRouteCatalogEntry{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	var profile urlmanagementdomain.StorefrontURLSearchProfile
	require.NoError(t, db.First(&profile).Error)
	require.Equal(t, entries[0].ID, profile.RouteEntryID)
	require.Equal(t, "Customized search title", profile.DisplayTitle)
	require.NoError(t, db.Table("site_quality_targets").Where("enabled = TRUE OR next_scheduled_at IS NOT NULL").Count(&count).Error)
	require.Zero(t, count)
}

func TestCurrentStorefrontCheckOverwritesPreviousResultAndRollsBackOnProjectionFailure(t *testing.T) {
	db := openCurrentStorefrontRouteSnapshotTestDatabase(t)
	entry := seodomain.StorefrontRouteCatalogEntry{RouteKey: "manifest:current:en", Path: "/current", Locale: "en", SourceType: "static", EntryStatus: "active", LastSeenAt: time.Now().UTC()}
	require.NoError(t, db.Create(&entry).Error)
	repository := NewStorefrontRouteCatalogRepository(db)
	failedCheck := seodomain.StorefrontRouteCheckResult{RouteEntryID: entry.ID, Status: "not_found", HTTPStatus: 404}
	require.NoError(t, repository.SaveCheck(&failedCheck))
	issue := urlmanagementdomain.StorefrontURLIssue{RouteEntryID: entry.ID, IssueType: "not_found", Severity: "high", LatestCheckResultID: &failedCheck.ID}
	require.NoError(t, db.Create(&issue).Error)
	healthyCheck := seodomain.StorefrontRouteCheckResult{RouteEntryID: entry.ID, Status: "ok", HTTPStatus: 200}
	require.NoError(t, repository.SaveCheck(&healthyCheck))
	require.Equal(t, failedCheck.ID, healthyCheck.ID)
	checks, total, err := repository.ListChecks(entry.ID, 1, 20)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "ok", checks[0].Status)
	require.NoError(t, db.First(&issue, issue.ID).Error)
	require.Equal(t, healthyCheck.ID, *issue.LatestCheckResultID)

	require.NoError(t, db.Exec(`CREATE TRIGGER reject_catalog_check_update BEFORE UPDATE ON storefront_route_catalog_entries BEGIN SELECT RAISE(ABORT, 'projection unavailable'); END`).Error)
	nextCheck := seodomain.StorefrontRouteCheckResult{RouteEntryID: entry.ID, Status: "server_error", HTTPStatus: 503}
	require.ErrorContains(t, repository.SaveCheck(&nextCheck), "projection unavailable")
	checks, total, err = repository.ListChecks(entry.ID, 1, 20)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "ok", checks[0].Status)
	require.NoError(t, db.First(&entry, entry.ID).Error)
	require.Equal(t, "ok", entry.LastCheckStatus)
}

func TestCurrentStorefrontSnapshotPrunesExistingCheckHistoryAndUpdatesIssueEvidenceReference(t *testing.T) {
	db := openCurrentStorefrontRouteSnapshotTestDatabase(t)
	entry := seodomain.StorefrontRouteCatalogEntry{RouteKey: "manifest:current:en", Path: "/current", Locale: "en", SourceType: "static", EntryStatus: "active", LastSeenAt: time.Now().UTC()}
	require.NoError(t, db.Create(&entry).Error)
	checkedAt := time.Now().UTC().Add(-time.Hour)
	checks := []seodomain.StorefrontRouteCheckResult{
		{RouteEntryID: entry.ID, Status: "not_found", HTTPStatus: 404, CheckedAt: checkedAt},
		{RouteEntryID: entry.ID, Status: "server_error", HTTPStatus: 503, CheckedAt: checkedAt.Add(time.Minute)},
		{RouteEntryID: entry.ID, Status: "ok", HTTPStatus: 200, CheckedAt: checkedAt.Add(time.Minute)},
	}
	require.NoError(t, db.Create(&checks).Error)
	issue := urlmanagementdomain.StorefrontURLIssue{RouteEntryID: entry.ID, IssueType: "not_found", Severity: "high", LatestCheckResultID: &checks[0].ID}
	require.NoError(t, db.Create(&issue).Error)
	repository := NewStorefrontRouteCatalogRepository(db)
	require.NoError(t, repository.UpsertSnapshot([]seodomain.StorefrontRouteCatalogEntry{entry}, time.Now().UTC()))
	results, total, err := repository.ListChecks(entry.ID, 1, 20)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, checks[2].ID, results[0].ID)
	require.NoError(t, db.First(&issue, issue.ID).Error)
	require.Equal(t, checks[2].ID, *issue.LatestCheckResultID)
}

func TestCurrentStorefrontSnapshotRollsBackRouteAndRelatedRecordCleanupWhenDeletionFails(t *testing.T) {
	db := openCurrentStorefrontRouteSnapshotTestDatabase(t)
	entry := seodomain.StorefrontRouteCatalogEntry{RouteKey: "manifest:removed:en", Path: "/removed", Locale: "en", SourceType: "static", EntryStatus: "active", LastSeenAt: time.Now().UTC().Add(-time.Hour)}
	require.NoError(t, db.Create(&entry).Error)
	check := seodomain.StorefrontRouteCheckResult{RouteEntryID: entry.ID, Status: "not_found", HTTPStatus: 404}
	require.NoError(t, db.Create(&check).Error)
	issue := urlmanagementdomain.StorefrontURLIssue{RouteEntryID: entry.ID, IssueType: "not_found", Severity: "high", LatestCheckResultID: &check.ID}
	require.NoError(t, db.Create(&issue).Error)
	require.NoError(t, db.Create(&urlmanagementdomain.StorefrontURLIssueEvent{IssueID: issue.ID, EventType: "detected", Metadata: "{}", CreatedAt: time.Now().UTC()}).Error)
	require.NoError(t, db.Exec(`CREATE TRIGGER reject_removed_route_delete BEFORE DELETE ON storefront_route_catalog_entries BEGIN SELECT RAISE(ABORT, 'route cleanup unavailable'); END`).Error)

	require.ErrorContains(t, NewStorefrontRouteCatalogRepository(db).UpsertSnapshot(nil, time.Now().UTC()), "route cleanup unavailable")
	for _, model := range []interface{}{&seodomain.StorefrontRouteCatalogEntry{}, &seodomain.StorefrontRouteCheckResult{}, &urlmanagementdomain.StorefrontURLIssue{}, &urlmanagementdomain.StorefrontURLIssueEvent{}} {
		var count int64
		require.NoError(t, db.Model(model).Count(&count).Error)
		require.EqualValues(t, 1, count)
	}
}
