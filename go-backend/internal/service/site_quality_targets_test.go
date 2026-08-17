package service

import (
	"testing"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSiteQualityTargetOptionsPreferHomepageRoute(t *testing.T) {
	db := newSiteQualityTargetOptionsTestDB(t)
	routeRepo := repository.NewStorefrontRouteCatalogRepository(db)
	now := time.Now().UTC()
	entries := []seodomain.StorefrontRouteCatalogEntry{
		{
			RouteKey:      "static:about",
			Path:          "/about",
			Locale:        "en-US",
			SourceType:    seodomain.RouteSourceStatic,
			Title:         "About",
			CanonicalPath: "/about",
			IsCheckable:   true,
			IsIndexable:   true,
			EntryStatus:   seodomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
		{
			RouteKey:      "static:home",
			Path:          "/",
			Locale:        "en-US",
			SourceType:    seodomain.RouteSourceStatic,
			Title:         "Home",
			CanonicalPath: "/",
			IsCheckable:   true,
			IsIndexable:   true,
			EntryStatus:   seodomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
		{
			RouteKey:      "static:private",
			Path:          "/private",
			Locale:        "en-US",
			SourceType:    seodomain.RouteSourceStatic,
			Title:         "Private",
			CanonicalPath: "/private",
			IsCheckable:   false,
			IsIndexable:   false,
			EntryStatus:   seodomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
	}
	require.NoError(t, db.Select("*").Create(&entries).Error)
	require.NoError(t, db.Model(&seodomain.StorefrontRouteCatalogEntry{}).
		Where("route_key = ?", "static:private").
		Update("is_checkable", false).Error)

	engine := NewSiteQualityEngineService(
		nil,
		nil,
		nil,
		nil,
		routeRepo,
		nil,
		SiteQualityEngineConfig{BaseURL: "https://example.com"},
	)
	options, err := engine.ListTargetOptions()
	require.NoError(t, err)
	require.Equal(t, "https://example.com/", options.DefaultURL)
	require.GreaterOrEqual(t, len(options.Items), 2)
	require.True(t, options.Items[0].IsHome)
	require.Equal(t, "/", options.Items[0].Path)
	require.Equal(t, "https://example.com/", options.Items[0].URL)
	require.Equal(t, "/about", options.Items[1].Path)
	for _, option := range options.Items {
		require.NotEqual(t, "/private", option.Path)
	}
}

func TestSiteQualityTargetOptionsFallbackToStorefrontRoot(t *testing.T) {
	engine := NewSiteQualityEngineService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		SiteQualityEngineConfig{BaseURL: "https://example.com"},
	)
	options, err := engine.ListTargetOptions()
	require.NoError(t, err)
	require.Equal(t, "https://example.com/", options.DefaultURL)
	require.Len(t, options.Items, 1)
	require.True(t, options.Items[0].IsHome)
	require.Equal(t, "https://example.com/", options.Items[0].URL)
}

func TestLighthouseRunnerDefaultURLUsesStorefrontRoot(t *testing.T) {
	db := newLighthouseRunnerTestDB(t)
	runner := NewLighthouseRunnerService(
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		LighthouseRunnerConfig{StorefrontBaseURL: "https://example.com"},
	)
	list, err := runner.List(repository.SiteQualityRunListFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, "https://example.com/", list.DefaultURL)
}

func newSiteQualityTargetOptionsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&seodomain.StorefrontRouteCatalogEntry{}))
	return db
}
