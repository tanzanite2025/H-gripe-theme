package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	outboxdomain "commerce-platform/internal/domain/outbox"
	seodomain "commerce-platform/internal/domain/seo"
	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSiteQualityRouteCatalogOutboxSyncMigratesAndDisablesTarget(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&sitequalitydomain.SiteQualityTarget{},
	))

	routeRepo := repository.NewStorefrontRouteCatalogRepository(db)
	targetRepo := repository.NewSiteQualityTargetRepository(db)
	now := time.Now().UTC()
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey:        "static:support",
		Path:            "/support",
		Locale:          "en-US",
		SourceType:      seodomain.RouteSourceStatic,
		CanonicalPath:   "/support",
		IsCheckable:     true,
		IsIndexable:     true,
		EntryStatus:     seodomain.RouteEntryStatusActive,
		ManifestVersion: "manifest-1",
		LastSeenAt:      now,
	}
	require.NoError(t, db.Create(&entry).Error)

	engine := NewSiteQualityEngineService(
		targetRepo,
		nil,
		nil,
		nil,
		routeRepo,
		nil,
		SiteQualityEngineConfig{BaseURL: "https://example.com"},
	)
	handler := NewSiteQualityRouteCatalogOutboxHandler(engine)
	syncEvent := func(marker string, seenAt time.Time) error {
		payload, marshalErr := json.Marshal(map[string]interface{}{
			"route_entry_id": entry.ID,
			"marker":         marker,
			"seen_at":        seenAt,
		})
		if marshalErr != nil {
			return marshalErr
		}
		return handler.Handle(context.Background(), outboxdomain.Event{
			Payload: datatypes.JSON(payload),
		})
	}

	firstSeenAt := now
	require.NoError(t, syncEvent("manifest-1", firstSeenAt))
	original, err := targetRepo.FindByCanonicalURL("https://example.com/support")
	require.NoError(t, err)
	require.Equal(t, entry.ID, *original.RouteEntryID)
	require.Equal(t, sitequalitydomain.SiteQualityTargetSourceRouteCatalog, original.Source)
	require.True(t, original.Enabled)
	require.Equal(t, "manifest-1", original.LedgerSyncMarker)

	entry.Path = "/help"
	entry.CanonicalPath = "/help"
	entry.ManifestVersion = "manifest-2"
	require.NoError(t, db.Save(&entry).Error)
	secondSeenAt := now.Add(time.Minute)
	require.NoError(t, syncEvent("manifest-2", secondSeenAt))

	migrated, err := targetRepo.FindByID(original.ID)
	require.NoError(t, err)
	require.Equal(t, original.ID, migrated.ID)
	require.Equal(t, "https://example.com/help", migrated.CanonicalURL)
	require.Equal(t, "manifest-2", migrated.LedgerSyncMarker)

	require.NoError(t, syncEvent("manifest-1", firstSeenAt))
	stillCurrent, err := targetRepo.FindByID(original.ID)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/help", stillCurrent.CanonicalURL)
	require.Equal(t, "manifest-2", stillCurrent.LedgerSyncMarker)

	entry.EntryStatus = seodomain.RouteEntryStatusStale
	entry.ManifestVersion = "manifest-3"
	require.NoError(t, db.Save(&entry).Error)
	require.NoError(t, syncEvent("manifest-3", now.Add(2*time.Minute)))

	disabled, err := targetRepo.FindByID(original.ID)
	require.NoError(t, err)
	require.False(t, disabled.Enabled)
	require.NotNil(t, disabled.DisabledAt)
	require.Equal(t, "manifest-3", disabled.LedgerSyncMarker)
}

func TestSiteQualityFullReconciliationDisablesStaleTargetWithCurrentObservationTime(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&sitequalitydomain.SiteQualityTarget{},
	))

	routeRepo := repository.NewStorefrontRouteCatalogRepository(db)
	targetRepo := repository.NewSiteQualityTargetRepository(db)
	observedAt := time.Date(2026, time.August, 17, 10, 0, 0, 0, time.UTC)
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey:        "static:retired",
		Path:            "/retired",
		Locale:          "en-US",
		SourceType:      seodomain.RouteSourceStatic,
		CanonicalPath:   "/retired",
		IsCheckable:     true,
		IsIndexable:     true,
		EntryStatus:     seodomain.RouteEntryStatusActive,
		ManifestVersion: "manifest-1",
		LastSeenAt:      observedAt.Add(-time.Hour),
	}
	require.NoError(t, db.Create(&entry).Error)

	engine := NewSiteQualityEngineService(
		targetRepo,
		nil,
		nil,
		nil,
		routeRepo,
		nil,
		SiteQualityEngineConfig{BaseURL: "https://example.com"},
	)
	_, err = engine.SyncTargetsFromRouteCatalog(observedAt, 10)
	require.NoError(t, err)

	entry.EntryStatus = seodomain.RouteEntryStatusStale
	require.NoError(t, db.Save(&entry).Error)
	reconciledAt := observedAt.Add(5 * time.Minute)
	_, err = engine.SyncTargetsFromRouteCatalog(reconciledAt, 10)
	require.NoError(t, err)

	target, err := targetRepo.FindByRouteEntryID(entry.ID)
	require.NoError(t, err)
	require.False(t, target.Enabled)
	require.NotNil(t, target.DisabledAt)
	require.Equal(t, reconciledAt, *target.LedgerSyncedAt)
}
