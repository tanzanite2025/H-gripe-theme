package service

import (
	"testing"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestStorefrontRedirectRuleServiceCanRepublishDisabledRule(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&urlmanagementdomain.StorefrontRedirectRule{},
		&seodomain.StorefrontRouteCatalogEntry{},
	))

	require.NoError(t, db.Create(&seodomain.StorefrontRouteCatalogEntry{
		RouteKey:      "manifest:canonical-target:en",
		Path:          "/replacement",
		Locale:        "en",
		SourceType:    seodomain.RouteSourceStatic,
		CanonicalPath: "/replacement",
		EntryStatus:   seodomain.RouteEntryStatusActive,
	}).Error)

	disabledAt := time.Now().UTC().Add(-time.Hour)
	rule := urlmanagementdomain.StorefrontRedirectRule{
		SourcePath: "/retired",
		TargetPath: "/replacement",
		StatusCode: 301,
		State:      urlmanagementdomain.RedirectRuleStateDisabled,
		DisabledAt: &disabledAt,
	}
	require.NoError(t, db.Create(&rule).Error)

	service := NewStorefrontRedirectRuleService(
		repository.NewStorefrontRedirectRuleRepository(db),
		repository.NewStorefrontRouteCatalogRepository(db),
	)
	republished, err := service.Publish(rule.ID, 12)
	require.NoError(t, err)
	require.Equal(t, urlmanagementdomain.RedirectRuleStatePublished, republished.State)
	require.Nil(t, republished.DisabledAt)
	require.NotNil(t, republished.PublishedAt)
	require.NotNil(t, republished.PublishedByID)
	require.Equal(t, uint(12), *republished.PublishedByID)
}
