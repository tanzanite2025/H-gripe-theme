package service

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/security"
	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/repository"

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
	assert.Equal(t, visitor.ProfileStatusActive, profile.ProfileStatus)
	assert.Equal(t, VisitorProfileQualityCartAction, profile.ProfileQualityScore)
	assert.Equal(t, VisitorProfileActionCart, profile.LastMeaningfulAction)
	require.NotNil(t, profile.FirstMeaningfulSeenAt)
	require.NotNil(t, profile.LastMeaningfulSeenAt)
	assert.Equal(t, firstSeen, *profile.FirstMeaningfulSeenAt)
	assert.Equal(t, firstSeen, *profile.LastMeaningfulSeenAt)
	require.NotNil(t, profile.RetentionUntil)

	updated, err := visitorProfileService.Touch(VisitorProfileTouchInput{
		CustomerServiceVisitorHash: "visitor-hash-1",
		CartSessionID:              "cart-session-1",
		Email:                      "Customer@Example.COM",
		EmailSource:                "public_chat",
		CountryCode:                "us",
		Region:                     "California",
		City:                       "Los Angeles",
		Timezone:                   "America/Los_Angeles",
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
	assert.Equal(t, "America/Los_Angeles", updated.Timezone)
	assert.NotEmpty(t, updated.IPHash)
	assert.NotContains(t, updated.IPHash, "203.0.113.10")
	assert.NotEmpty(t, updated.UserAgentHash)
	assert.Equal(t, VisitorProfileQualityCartAction+VisitorProfileQualityEmailCapture, updated.ProfileQualityScore)
	assert.Equal(t, VisitorProfileActionEmailCapture, updated.LastMeaningfulAction)
	require.NotNil(t, updated.FirstMeaningfulSeenAt)
	require.NotNil(t, updated.LastMeaningfulSeenAt)
	assert.Equal(t, firstSeen, *updated.FirstMeaningfulSeenAt)
	assert.Equal(t, firstSeen.Add(time.Hour), *updated.LastMeaningfulSeenAt)

	var count int64
	require.NoError(t, db.Model(&visitor.Profile{}).Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

func TestVisitorProfileIPBlockUsesRetainedIPAndProfileScopedSource(t *testing.T) {
	db, visitorProfileService, globalIPBlockService := newTestVisitorProfileIPBlockService(t)
	profile := &visitor.Profile{
		CustomerServiceVisitorHash: "visitor-with-ip",
		IPAddress:                  "203.0.113.44",
		IPHash:                     "retained-ip-hash",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 time.Now().UTC(),
	}
	require.NoError(t, db.Create(profile).Error)

	commercialRule, err := globalIPBlockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:            "203.0.113.44",
		Source:          security.IPBlockRuleSourceCommercialBot,
		SourceReference: "crawler-1",
		Reason:          "commercial crawler detection",
	})
	require.NoError(t, err)

	profileRule, err := visitorProfileService.BlockProfileIP(
		context.Background(),
		profile.ID,
		IPBlockRuleInput{Reason: "manual visitor review"},
		7,
	)
	require.NoError(t, err)
	assert.Equal(t, "203.0.113.44/32", profileRule.CIDR)
	assert.Equal(t, security.IPBlockRuleSourceVisitorProfile, profileRule.Source)
	assert.Equal(t, "1", profileRule.SourceReference)
	assert.Equal(t, uint(7), *profileRule.CreatedBy)

	blocked, match, err := globalIPBlockService.IsBlocked(context.Background(), "203.0.113.44", time.Now())
	require.NoError(t, err)
	require.True(t, blocked)
	require.NotNil(t, match)
	assert.Equal(t, profileRule.ID, match.ID)

	_, err = visitorProfileService.UnblockProfileIP(context.Background(), profile.ID, 8)
	require.NoError(t, err)

	blocked, match, err = globalIPBlockService.IsBlocked(context.Background(), "203.0.113.44", time.Now())
	require.NoError(t, err)
	require.True(t, blocked)
	require.NotNil(t, match)
	assert.Equal(t, commercialRule.ID, match.ID)
}

func TestVisitorProfileIPBlockRequiresRetainedIP(t *testing.T) {
	db, visitorProfileService, _ := newTestVisitorProfileIPBlockService(t)
	profile := &visitor.Profile{
		CustomerServiceVisitorHash: "visitor-without-ip",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 time.Now().UTC(),
	}
	require.NoError(t, db.Create(profile).Error)

	_, err := visitorProfileService.BlockProfileIP(
		context.Background(),
		profile.ID,
		IPBlockRuleInput{Reason: "manual visitor review"},
		7,
	)
	require.ErrorIs(t, err, ErrVisitorProfileIPUnavailable)

	var ruleCount int64
	require.NoError(t, db.Model(&security.IPBlockRule{}).Count(&ruleCount).Error)
	assert.Zero(t, ruleCount)
}

func TestVisitorProfileIPBlockKeepsHistoricalTargetsAndBatchUnblocksAfterIPDrift(t *testing.T) {
	db, visitorProfileService, globalIPBlockService := newTestVisitorProfileIPBlockService(t)
	profile := &visitor.Profile{
		CustomerServiceVisitorHash: "visitor-ip-drift",
		IPAddress:                  "198.51.100.10",
		IPHash:                     "first-ip-hash",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 time.Now().UTC(),
	}
	require.NoError(t, db.Create(profile).Error)

	firstRule, err := visitorProfileService.BlockProfileIP(
		context.Background(),
		profile.ID,
		IPBlockRuleInput{Reason: "first retained IP"},
		7,
	)
	require.NoError(t, err)

	profile.IPAddress = "198.51.100.11"
	profile.IPHash = "second-ip-hash"
	require.NoError(t, db.Save(profile).Error)

	secondRule, err := visitorProfileService.BlockProfileIP(
		context.Background(),
		profile.ID,
		IPBlockRuleInput{Reason: "second retained IP"},
		7,
	)
	require.NoError(t, err)
	require.NotEqual(t, firstRule.ID, secondRule.ID)

	activeRules, err := visitorProfileService.ActiveProfileIPBlocks(context.Background(), profile.ID)
	require.NoError(t, err)
	require.Len(t, activeRules, 2)
	assert.Equal(t, secondRule.ID, activeRules[0].ID)

	profiles, total, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{
		Status: visitor.ProfileStatusActive,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, profiles, 1)
	assert.Equal(t, "198.51.100.xxx", profiles[0].IPAddress)
	require.NotNil(t, profiles[0].IPBlock)
	assert.Equal(t, secondRule.ID, profiles[0].IPBlock.ID)
	require.Len(t, profiles[0].IPBlockRules, 2)
	assert.Equal(t, secondRule.ID, profiles[0].IPBlockRules[0].ID)
	require.NotNil(t, profiles[0].IPBlockMatch)
	assert.Equal(t, secondRule.ID, profiles[0].IPBlockMatch.ID)

	profile.IPAddress = "198.51.100.10"
	require.NoError(t, db.Save(profile).Error)
	profiles, _, err = visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{
		Status: visitor.ProfileStatusActive,
	})
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	require.NotNil(t, profiles[0].IPBlockMatch)
	assert.Equal(t, firstRule.ID, profiles[0].IPBlockMatch.ID)

	disabledRules, err := visitorProfileService.UnblockProfileIPs(context.Background(), profile.ID, 8)
	require.NoError(t, err)
	require.Len(t, disabledRules, 2)

	for _, ip := range []string{"198.51.100.10", "198.51.100.11"} {
		blocked, match, err := globalIPBlockService.IsBlocked(context.Background(), ip, time.Now())
		require.NoError(t, err)
		assert.False(t, blocked, ip)
		assert.Nil(t, match, ip)
	}
}

func TestVisitorProfileListMasksIPAddressButUsesRetainedAddressForMatch(t *testing.T) {
	db, visitorProfileService, _ := newTestVisitorProfileIPBlockService(t)
	profile := &visitor.Profile{
		CustomerServiceVisitorHash: "visitor-masked-ip",
		IPAddress:                  "203.0.113.44",
		IPHash:                     "retained-ip-hash",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 time.Now().UTC(),
	}
	require.NoError(t, db.Create(profile).Error)

	rule, err := visitorProfileService.BlockProfileIP(
		context.Background(),
		profile.ID,
		IPBlockRuleInput{Reason: "verify raw lookup remains internal"},
		7,
	)
	require.NoError(t, err)

	profiles, _, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{
		Status: visitor.ProfileStatusActive,
	})
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	assert.Equal(t, "203.0.113.xxx", profiles[0].IPAddress)
	assert.NotContains(t, profiles[0].IPAddress, "203.0.113.44")
	require.NotNil(t, profiles[0].IPBlockMatch)
	assert.Equal(t, rule.ID, profiles[0].IPBlockMatch.ID)
}

func TestVisitorProfileSearchDoesNotUseRawIPAddress(t *testing.T) {
	db, visitorProfileService := newTestVisitorProfileService(t)
	require.NoError(t, db.Create(&visitor.Profile{
		CustomerServiceVisitorHash: "visitor-search-ip",
		IPAddress:                  "203.0.113.55",
		IPHash:                     "search-ip-hash",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 time.Now().UTC(),
	}).Error)

	profiles, total, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{
		Search: "203.0.113.55",
		Status: "all",
	})
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, profiles)
}

func TestVisitorProfileIPAddressRetentionCleanupClearsAddressAndKeepsHash(t *testing.T) {
	db, visitorProfileService := newTestVisitorProfileService(t)
	visitorProfileService.ConfigureIPAddressRetentionDays(30)
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)

	oldProfile := &visitor.Profile{
		CustomerServiceVisitorHash: "old-ip-profile",
		IPAddress:                  "198.51.100.24",
		IPHash:                     "old-ip-hash",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 now.Add(-31 * 24 * time.Hour),
	}
	freshProfile := &visitor.Profile{
		CustomerServiceVisitorHash: "fresh-ip-profile",
		IPAddress:                  "203.0.113.10",
		IPHash:                     "fresh-ip-hash",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 now.Add(-29 * 24 * time.Hour),
	}
	require.NoError(t, db.Create([]*visitor.Profile{oldProfile, freshProfile}).Error)

	result, err := visitorProfileService.CleanupExpiredProfiles(now)
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.ClearedIPAddresses)
	assert.EqualValues(t, 1, result.TotalChanged)

	var storedOld visitor.Profile
	require.NoError(t, db.First(&storedOld, oldProfile.ID).Error)
	assert.Empty(t, storedOld.IPAddress)
	assert.Equal(t, "old-ip-hash", storedOld.IPHash)

	var storedFresh visitor.Profile
	require.NoError(t, db.First(&storedFresh, freshProfile.ID).Error)
	assert.Equal(t, "203.0.113.10", storedFresh.IPAddress)
	assert.Equal(t, "fresh-ip-hash", storedFresh.IPHash)
}

func TestMaskVisitorProfileIPAddress(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "ipv4", input: "203.0.113.44", want: "203.0.113.xxx"},
		{name: "ipv6", input: "2001:db8:1234:5678:90ab:cdef:1111:2222", want: "2001:db8:1234:5678:xxxx:xxxx:xxxx:xxxx"},
		{name: "invalid", input: "not-an-ip", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, maskVisitorProfileIPAddress(test.input))
		})
	}
}

func TestVisitorProfileListRetainsProfileRuleWhenCurrentIPIsCleared(t *testing.T) {
	db, visitorProfileService, _ := newTestVisitorProfileIPBlockService(t)
	profile := &visitor.Profile{
		CustomerServiceVisitorHash: "visitor-cleared-ip",
		IPAddress:                  "203.0.113.77",
		IPHash:                     "retained-ip-hash",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 time.Now().UTC(),
	}
	require.NoError(t, db.Create(profile).Error)

	rule, err := visitorProfileService.BlockProfileIP(
		context.Background(),
		profile.ID,
		IPBlockRuleInput{Reason: "rule must remain addressable"},
		7,
	)
	require.NoError(t, err)

	profile.IPAddress = ""
	require.NoError(t, db.Save(profile).Error)

	profiles, _, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{
		Status: visitor.ProfileStatusActive,
	})
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	assert.Empty(t, profiles[0].IPAddress)
	require.NotNil(t, profiles[0].IPBlock)
	assert.Equal(t, rule.ID, profiles[0].IPBlock.ID)
	assert.Nil(t, profiles[0].IPBlockMatch)

	_, err = visitorProfileService.UnblockProfileIP(context.Background(), profile.ID, 8)
	require.NoError(t, err)
}

func TestVisitorProfileListSeparatesProfileRuleFromActualSourceMatch(t *testing.T) {
	db, visitorProfileService, globalIPBlockService := newTestVisitorProfileIPBlockService(t)
	profile := &visitor.Profile{
		CustomerServiceVisitorHash: "visitor-multiple-sources",
		IPAddress:                  "192.0.2.44",
		IPHash:                     "retained-ip-hash",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 time.Now().UTC(),
	}
	require.NoError(t, db.Create(profile).Error)

	profileRule, err := visitorProfileService.BlockProfileIP(
		context.Background(),
		profile.ID,
		IPBlockRuleInput{Reason: "profile source"},
		7,
	)
	require.NoError(t, err)
	commercialRule, err := globalIPBlockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:            "192.0.2.44",
		Source:          security.IPBlockRuleSourceCommercialBot,
		SourceReference: "crawler-after-profile",
		Reason:          "commercial source",
	})
	require.NoError(t, err)

	profiles, _, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{
		Status: visitor.ProfileStatusActive,
	})
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	require.NotNil(t, profiles[0].IPBlock)
	require.NotNil(t, profiles[0].IPBlockMatch)
	assert.Equal(t, profileRule.ID, profiles[0].IPBlock.ID)
	assert.Equal(t, commercialRule.ID, profiles[0].IPBlockMatch.ID)
}

func TestVisitorProfilePassiveSeenDoesNotCreateAndDoesNotIncreaseQuality(t *testing.T) {
	db, visitorProfileService := newTestVisitorProfileService(t)
	now := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)

	passive, err := visitorProfileService.TouchPassiveSeen(VisitorProfileTouchInput{
		CartSessionID: "passive-cart-session",
		SeenAt:        now,
	})
	require.NoError(t, err)
	assert.Nil(t, passive)
	assertTableCount(t, db, &visitor.Profile{}, 0)

	active, err := visitorProfileService.TouchMeaningfulAction(VisitorProfileTouchInput{
		CartSessionID:     "active-cart-session",
		SeenAt:            now,
		MeaningfulAction:  VisitorProfileActionCart,
		QualityScoreDelta: VisitorProfileQualityCartAction,
	})
	require.NoError(t, err)
	require.NotNil(t, active)

	passiveUpdate, err := visitorProfileService.TouchPassiveSeen(VisitorProfileTouchInput{
		CartSessionID: "active-cart-session",
		Locale:        "fr-FR",
		SeenAt:        now.Add(2 * time.Hour),
	})
	require.NoError(t, err)
	require.NotNil(t, passiveUpdate)
	assert.Equal(t, VisitorProfileQualityCartAction, passiveUpdate.ProfileQualityScore)
	assert.Equal(t, VisitorProfileActionCart, passiveUpdate.LastMeaningfulAction)
	assert.Equal(t, "fr-fr", passiveUpdate.Locale)
	assert.Equal(t, now.Add(2*time.Hour), passiveUpdate.LastSeenAt)
	require.NotNil(t, passiveUpdate.LastMeaningfulSeenAt)
	assert.Equal(t, now, *passiveUpdate.LastMeaningfulSeenAt)
}

func TestVisitorProfileAdminListAndStats(t *testing.T) {
	db, visitorProfileService := newTestVisitorProfileService(t)
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

	require.NoError(t, db.Create(&visitor.Profile{
		CustomerServiceVisitorHash: "candidate-visitor-hash",
		ProfileStatus:              visitor.ProfileStatusCandidate,
		ProfileQualityScore:        1,
		LastSeenAt:                 now,
	}).Error)

	accountProfiles, total, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{Identity: "account"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, accountProfiles, 1)
	assert.Equal(t, "account", accountProfiles[0].Identity)
	assert.Equal(t, "member@example.test", accountProfiles[0].Email)
	assert.NotContains(t, accountProfiles[0].CustomerServiceVisitorHashPreview, "member-visitor-hash")
	assert.True(t, accountProfiles[0].HasIPFingerprint)
	assert.True(t, accountProfiles[0].HasUserAgentFingerprint)
	assert.Equal(t, visitor.ProfileStatusActive, accountProfiles[0].ProfileStatus)
	assert.GreaterOrEqual(t, accountProfiles[0].ProfileQualityScore, VisitorProfileQualityAccount)

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

	candidateProfiles, total, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{Status: visitor.ProfileStatusCandidate})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, candidateProfiles, 1)
	assert.Equal(t, visitor.ProfileStatusCandidate, candidateProfiles[0].ProfileStatus)

	allProfiles, total, err := visitorProfileService.ListProfiles(1, 20, VisitorProfileListInput{Status: "all"})
	require.NoError(t, err)
	assert.EqualValues(t, 3, total)
	require.Len(t, allProfiles, 3)

	stats, err := visitorProfileService.GetStats()
	require.NoError(t, err)
	assert.EqualValues(t, 3, stats.Total)
	assert.EqualValues(t, 1, stats.AccountCount)
	assert.EqualValues(t, 2, stats.AnonymousCount)
	assert.EqualValues(t, 1, stats.EmailCount)
	assert.EqualValues(t, 2, stats.CartLinkedCount)
	assert.EqualValues(t, 3, stats.CustomerServiceCount)
	assert.EqualValues(t, 2, stats.RegionCount)
	assert.EqualValues(t, 2, stats.ActiveCount)
	assert.EqualValues(t, 1, stats.CandidateCount)
}

func TestVisitorProfileRetentionCleanupDeletesCandidatesAndArchivesAnonymous(t *testing.T) {
	db, visitorProfileService := newTestVisitorProfileService(t)
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	userID := uint(88)

	require.NoError(t, db.Create(&[]visitor.Profile{
		{
			CustomerServiceVisitorHash: "expired-candidate",
			ProfileStatus:              visitor.ProfileStatusCandidate,
			RetentionUntil:             timePtr(now.Add(-time.Hour)),
			LastSeenAt:                 now.Add(-48 * time.Hour),
		},
		{
			CustomerServiceVisitorHash: "fresh-candidate",
			ProfileStatus:              visitor.ProfileStatusCandidate,
			RetentionUntil:             timePtr(now.Add(24 * time.Hour)),
			LastSeenAt:                 now,
		},
		{
			CustomerServiceVisitorHash: "expired-anonymous-active",
			ProfileStatus:              visitor.ProfileStatusActive,
			ProfileQualityScore:        VisitorProfileQualityCustomerService,
			RetentionUntil:             timePtr(now.Add(-time.Hour)),
			LastSeenAt:                 now.Add(-24 * time.Hour),
		},
		{
			UserID:                     &userID,
			CustomerServiceVisitorHash: "expired-account-active",
			ProfileStatus:              visitor.ProfileStatusActive,
			ProfileQualityScore:        VisitorProfileQualityAccount,
			RetentionUntil:             timePtr(now.Add(-time.Hour)),
			LastSeenAt:                 now.Add(-24 * time.Hour),
		},
	}).Error)

	result, err := visitorProfileService.CleanupExpiredProfiles(now)
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.DeletedCandidates)
	assert.EqualValues(t, 1, result.ArchivedAnonymous)
	assert.EqualValues(t, 2, result.TotalChanged)

	assert.EqualValues(t, 1, countVisitorProfilesWhere(t, db, "profile_status = ?", visitor.ProfileStatusCandidate))
	assert.EqualValues(t, 1, countVisitorProfilesWhere(t, db, "profile_status = ?", visitor.ProfileStatusActive))
	assert.EqualValues(t, 1, countVisitorProfilesWhere(t, db, "profile_status = ?", visitor.ProfileStatusArchived))

	var deletedCandidate visitor.Profile
	require.NoError(t, db.Unscoped().Where("customer_service_visitor_hash = ?", "expired-candidate").First(&deletedCandidate).Error)
	assert.True(t, deletedCandidate.DeletedAt.Valid)

	var accountProfile visitor.Profile
	require.NoError(t, db.Where("customer_service_visitor_hash = ?", "expired-account-active").First(&accountProfile).Error)
	assert.Equal(t, visitor.ProfileStatusActive, accountProfile.ProfileStatus)
	assert.Equal(t, userID, *accountProfile.UserID)
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

func newTestVisitorProfileIPBlockService(t *testing.T) (*gorm.DB, *VisitorProfileService, *GlobalIPBlockService) {
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

	require.NoError(t, db.AutoMigrate(&visitor.Profile{}, &security.IPBlockRule{}))

	globalIPBlockService := NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))
	visitorProfileService := NewVisitorProfileService(repository.NewVisitorProfileRepository(db))
	visitorProfileService.ConfigureGlobalIPBlockService(globalIPBlockService)
	return db, visitorProfileService, globalIPBlockService
}

func assertTableCount(t *testing.T, db *gorm.DB, model any, expected int64) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(model).Count(&count).Error)
	assert.EqualValues(t, expected, count)
}

func countVisitorProfilesWhere(t *testing.T, db *gorm.DB, query string, args ...interface{}) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&visitor.Profile{}).Where(query, args...).Count(&count).Error)
	return count
}

func timePtr(value time.Time) *time.Time {
	return &value
}
