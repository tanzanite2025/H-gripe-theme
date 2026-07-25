package service

import (
	"testing"
	"time"

	"tanzanite/internal/domain/visitor"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestVisitorProfileTouchBindsCartAndCustomerServiceVisitor(t *testing.T) {
	db, visitorProfileService := newTestVisitorProfileService(t)

	firstSeen := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	profile, err := visitorProfileService.Touch(VisitorProfileTouchInput{
		CartSessionID: "cart-session-1",
		Locale:        "en-US,en;q=0.9",
		LocaleSource:  "accept_language",
		SeenAt:        firstSeen,
	})
	require.NoError(t, err)
	require.NotNil(t, profile)
	assert.Equal(t, "cart-session-1", profile.CartSessionID)
	assert.Equal(t, "en-us", profile.Locale)

	updated, err := visitorProfileService.Touch(VisitorProfileTouchInput{
		CustomerServiceVisitorHash: "visitor-hash-1",
		CartSessionID:              "cart-session-1",
		Email:                      "Customer@Example.COM",
		EmailSource:                "public_chat",
		CountryCode:                "us",
		Region:                     "California",
		City:                       "Los Angeles",
		IPAddress:                  "203.0.113.10",
		UserAgent:                  "Mozilla/5.0 Test",
		SeenAt:                     firstSeen.Add(time.Hour),
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, profile.ID, updated.ID)
	assert.Equal(t, "visitor-hash-1", updated.CustomerServiceVisitorHash)
	assert.Equal(t, "customer@example.com", updated.Email)
	assert.Equal(t, "public_chat", updated.EmailSource)
	assert.Equal(t, "US", updated.CountryCode)
	assert.NotEmpty(t, updated.IPHash)
	assert.NotContains(t, updated.IPHash, "203.0.113.10")
	assert.NotEmpty(t, updated.UserAgentHash)

	var count int64
	require.NoError(t, db.Model(&visitor.Profile{}).Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

func TestVisitorProfileAdminListAndStats(t *testing.T) {
	_, visitorProfileService := newTestVisitorProfileService(t)
	userID := uint(12)
	now := time.Now().UTC()

	_, err := visitorProfileService.Touch(VisitorProfileTouchInput{
		UserID:                     &userID,
		CustomerServiceVisitorHash: "member-visitor-hash",
		CartSessionID:              "member-cart-session",
		Email:                      "member@example.test",
		EmailSource:                "account",
		Locale:                     "en-US",
		CountryCode:                "US",
		Region:                     "California",
		City:                       "Los Angeles",
		IPAddress:                  "203.0.113.15",
		UserAgent:                  "Member Test Browser",
		SeenAt:                     now,
	})
	require.NoError(t, err)

	_, err = visitorProfileService.Touch(VisitorProfileTouchInput{
		CustomerServiceVisitorHash: "anonymous-visitor-hash",
		CartSessionID:              "anonymous-cart-session",
		Locale:                     "de-DE",
		CountryCode:                "DE",
		SeenAt:                     now.Add(-48 * time.Hour),
	})
	require.NoError(t, err)

	accountProfiles, total, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{Identity: "account"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, accountProfiles, 1)
	assert.Equal(t, "account", accountProfiles[0].Identity)
	assert.Equal(t, "member@example.test", accountProfiles[0].Email)
	assert.NotContains(t, accountProfiles[0].CustomerServiceVisitorHashPreview, "member-visitor-hash")
	assert.True(t, accountProfiles[0].HasIPFingerprint)
	assert.True(t, accountProfiles[0].HasUserAgentFingerprint)

	anonymousProfiles, total, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{Identity: "anonymous", Email: "missing"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, anonymousProfiles, 1)
	assert.Equal(t, "anonymous", anonymousProfiles[0].Identity)
	assert.False(t, anonymousProfiles[0].HasEmail)

	cartProfiles, total, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{Search: "anonymous-cart-session"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, cartProfiles, 1)
	assert.Equal(t, "anonymous-cart-session", cartProfiles[0].CartSessionID)

	recentProfiles, total, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{LastSeen: "24h"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, recentProfiles, 1)
	assert.Equal(t, "account", recentProfiles[0].Identity)

	stats, err := visitorProfileService.GetStats()
	require.NoError(t, err)
	assert.EqualValues(t, 2, stats.Total)
	assert.EqualValues(t, 1, stats.AccountCount)
	assert.EqualValues(t, 1, stats.AnonymousCount)
	assert.EqualValues(t, 1, stats.EmailCount)
	assert.EqualValues(t, 2, stats.CartLinkedCount)
	assert.EqualValues(t, 2, stats.CustomerServiceCount)
	assert.EqualValues(t, 2, stats.RegionCount)
}

func newTestVisitorProfileService(t *testing.T) (*gorm.DB, *VisitorProfileService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(&visitor.Profile{}))

	return db, NewVisitorProfileService(repository.NewVisitorProfileRepository(db))
}
