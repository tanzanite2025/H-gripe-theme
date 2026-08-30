package service

import (
	"testing"
	"time"

	seoDomain "commerce-platform/internal/domain/seo"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestStorefrontURLSearchProfileServicePublicIndexUsesRouteCatalogFallbacks(t *testing.T) {
	db := newStorefrontURLSearchProfileTestDB(t)
	routeRepo := repository.NewStorefrontRouteCatalogRepository(db)
	profileRepo := repository.NewStorefrontURLSearchProfileRepository(db)

	now := time.Now().UTC()
	entries := []seoDomain.StorefrontRouteCatalogEntry{
		{
			RouteKey:      "static:about",
			Path:          "/zh_cn/about",
			Locale:        "zh_cn",
			SourceType:    seoDomain.RouteSourceStatic,
			Title:         "About",
			Summary:       "About summary",
			CanonicalPath: "/zh_cn/about",
			IsSearchable:  true,
			IsIndexable:   true,
			EntryStatus:   seoDomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
		{
			RouteKey:      "static:help",
			Path:          "/zh_cn/help/faq",
			Locale:        "zh_cn",
			SourceType:    seoDomain.RouteSourceStatic,
			Title:         "FAQ",
			Summary:       "Frequently asked questions",
			CanonicalPath: "/zh_cn/help/faq",
			IsSearchable:  true,
			IsIndexable:   true,
			EntryStatus:   seoDomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
		{
			RouteKey:      "static:hidden",
			Path:          "/zh_cn/hidden",
			Locale:        "zh_cn",
			SourceType:    seoDomain.RouteSourceStatic,
			Title:         "Hidden",
			Summary:       "Hidden summary",
			CanonicalPath: "/zh_cn/hidden",
			IsSearchable:  false,
			IsIndexable:   true,
			EntryStatus:   seoDomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
		{
			RouteKey:      "static:guides",
			Path:          "/zh_cn/guides",
			Locale:        "zh_cn",
			SourceType:    seoDomain.RouteSourceStatic,
			Title:         "Guides",
			Summary:       "Guides summary",
			CanonicalPath: "/zh_cn/guides",
			IsSearchable:  true,
			IsIndexable:   true,
			EntryStatus:   seoDomain.RouteEntryStatusActive,
			LastSeenAt:    now,
		},
	}
	require.NoError(t, db.Create(&entries).Error)
	require.NoError(t, db.Model(&seoDomain.StorefrontRouteCatalogEntry{}).
		Where("id = ?", entries[2].ID).
		Update("is_searchable", false).Error)

	service := NewStorefrontURLSearchProfileService(profileRepo, routeRepo)
	_, err := service.Upsert(entries[1].ID, urlmanagementdomain.StorefrontURLSearchProfileInput{
		Enabled:        true,
		SearchWeight:   200,
		Keywords:       []string{"faq", "support"},
		DisplayTitle:   "Help Center",
		DisplaySummary: "FAQ summary override",
	})
	require.NoError(t, err)
	_, err = service.Upsert(entries[3].ID, urlmanagementdomain.StorefrontURLSearchProfileInput{
		Enabled:        false,
		SearchWeight:   300,
		Keywords:       []string{"guide"},
		DisplayTitle:   "Hidden Guides",
		DisplaySummary: "Should not surface",
	})
	require.NoError(t, err)

	items, err := service.PublicIndex("zh_cn")
	require.NoError(t, err)
	require.Len(t, items, 2)

	itemsByPath := make(map[string]seoDomain.StorefrontRouteCatalogEntry, len(items))
	for _, item := range items {
		require.NotNil(t, item.RouteEntry)
		itemsByPath[item.RouteEntry.Path] = *item.RouteEntry
	}

	faq, ok := itemsByPath["/zh_cn/help/faq"]
	require.True(t, ok, "expected explicit profile-backed URL in public index")
	assert.Equal(t, "FAQ", faq.Title)
	assert.Equal(t, "Frequently asked questions", faq.Summary)

	about, ok := itemsByPath["/zh_cn/about"]
	require.True(t, ok, "expected searchable route catalog fallback in public index")
	assert.Equal(t, "About", about.Title)
	assert.Equal(t, "About summary", about.Summary)

	_, hiddenOK := itemsByPath["/zh_cn/hidden"]
	assert.False(t, hiddenOK, "non-searchable route should stay excluded")
	_, guidesOK := itemsByPath["/zh_cn/guides"]
	assert.False(t, guidesOK, "disabled search profile should suppress fallback")
}

func newStorefrontURLSearchProfileTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seoDomain.StorefrontRouteCatalogEntry{},
		&urlmanagementdomain.StorefrontURLSearchProfile{},
	))
	return db
}
