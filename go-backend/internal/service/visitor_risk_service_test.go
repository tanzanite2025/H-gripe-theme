package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestVisitorRiskServiceAggregatesRequestsIntoDailyFact(t *testing.T) {
	db, visitorRiskService := newTestVisitorRiskService(t, true)
	now := time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC)

	visitorRiskService.RecordRequest(VisitorRiskRecordInput{
		IPAddress:       "203.0.113.20",
		UserAgent:       "Mozilla/5.0",
		Path:            "/api/v1/products/42?utm=ignored",
		HasCookieHeader: false,
		StatusCode:      200,
		OccurredAt:      now,
	})
	visitorRiskService.RecordRequest(VisitorRiskRecordInput{
		IPAddress:        "203.0.113.20",
		UserAgent:        "Mozilla/5.0",
		Path:             "/api/v1/cart/add",
		SessionID:        "session-1",
		HasCookieHeader:  true,
		StatusCode:       200,
		OccurredAt:       now.Add(time.Minute),
		MeaningfulAction: true,
	})
	visitorRiskService.RecordRequest(VisitorRiskRecordInput{
		IPAddress:       "203.0.113.20",
		UserAgent:       "Mozilla/5.0",
		Path:            "/api/v1/auth/login",
		HasCookieHeader: true,
		StatusCode:      401,
		OccurredAt:      now.Add(2 * time.Minute),
	})

	result, err := visitorRiskService.Flush(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, result.FlushedFacts)

	var facts []visitor.RiskDailyFact
	require.NoError(t, db.Find(&facts).Error)
	require.Len(t, facts, 1)
	fact := facts[0]
	assert.Equal(t, 3, fact.RequestCount)
	assert.Equal(t, 3, fact.UniquePathCount)
	assert.Equal(t, 1, fact.UniqueSessionCount)
	assert.Equal(t, 1, fact.InvalidRequestCount)
	assert.Equal(t, 1, fact.AuthFailureCount)
	assert.Equal(t, 1, fact.NoCookieRequestCount)
	assert.Equal(t, 1, fact.MeaningfulActionCount)
	assert.Equal(t, visitor.RiskLevelNormal, fact.RiskLevel)
	assert.NotContains(t, fact.IPHash, "203.0.113.20")
	assert.JSONEq(t, `["/api/v1/products/42","/api/v1/cart/add","/api/v1/auth/login"]`, string(fact.SamplePaths))
}

func TestVisitorRiskServiceDisabledDoesNotAccumulate(t *testing.T) {
	_, visitorRiskService := newTestVisitorRiskService(t, false)

	visitorRiskService.RecordRequest(VisitorRiskRecordInput{
		IPAddress:  "203.0.113.30",
		UserAgent:  "curl/8.0",
		Path:       "/api/v1/products",
		StatusCode: 404,
	})

	assert.Equal(t, 0, visitorRiskService.PendingCount())
	result, err := visitorRiskService.Flush(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, result.FlushedFacts)
}

func TestVisitorRiskServiceCleanupDeletesFactsOlderThanRetention(t *testing.T) {
	db, visitorRiskService := newTestVisitorRiskService(t, true)
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)

	require.NoError(t, db.Create(&[]visitor.RiskDailyFact{
		{
			Day:         now.AddDate(0, 0, -366),
			IPHash:      "old-ip-hash",
			FirstSeenAt: now.AddDate(0, 0, -366),
			LastSeenAt:  now.AddDate(0, 0, -366),
			RiskLevel:   visitor.RiskLevelNormal,
			SamplePaths: []byte(`[]`),
		},
		{
			Day:         now.AddDate(0, 0, -365),
			IPHash:      "fresh-ip-hash",
			FirstSeenAt: now.AddDate(0, 0, -365),
			LastSeenAt:  now.AddDate(0, 0, -365),
			RiskLevel:   visitor.RiskLevelNormal,
			SamplePaths: []byte(`[]`),
		},
	}).Error)

	result, err := visitorRiskService.CleanupExpiredFacts(now)
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.DeletedFacts)

	var count int64
	require.NoError(t, db.Model(&visitor.RiskDailyFact{}).Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

func TestVisitorRiskServiceCleanupDeletesOnlyOldExpiredDecisions(t *testing.T) {
	db, visitorRiskService := newTestVisitorRiskService(t, true)
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	oldExpiry := now.AddDate(0, 0, -366)
	recentExpiry := now.AddDate(0, 0, -30)

	require.NoError(t, db.Create(&[]visitor.RiskDecision{
		{
			Scope:     visitor.RiskDecisionScopeIPHash,
			ValueHash: "old-expired",
			Action:    visitor.RiskDecisionActionTemporaryBlock,
			Reason:    "Old expired decision",
			ExpiresAt: &oldExpiry,
		},
		{
			Scope:     visitor.RiskDecisionScopeIPHash,
			ValueHash: "recent-expired",
			Action:    visitor.RiskDecisionActionTemporaryBlock,
			Reason:    "Recent expired decision",
			ExpiresAt: &recentExpiry,
		},
		{
			Scope:     visitor.RiskDecisionScopeIPHash,
			ValueHash: "indefinite",
			Action:    visitor.RiskDecisionActionWatch,
			Reason:    "Indefinite review decision",
		},
	}).Error)

	result, err := visitorRiskService.CleanupExpiredFacts(now)
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.DeletedDecisions)

	var count int64
	require.NoError(t, db.Model(&visitor.RiskDecision{}).Count(&count).Error)
	assert.EqualValues(t, 2, count)
}

func TestVisitorRiskServiceAcceptsFiveHundredUnicodeReasonCharacters(t *testing.T) {
	db, visitorRiskService := newTestVisitorRiskService(t, true)
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	fact := visitor.RiskDailyFact{
		Day:         now,
		IPHash:      "ip-hash",
		FirstSeenAt: now,
		LastSeenAt:  now,
		SamplePaths: []byte(`[]`),
	}
	require.NoError(t, db.Create(&fact).Error)

	_, err := visitorRiskService.CreateDecision(fact.ID, VisitorRiskDecisionInput{
		Action: visitor.RiskDecisionActionWatch,
		Reason: strings.Repeat("看", 500),
	}, 42)
	require.NoError(t, err)
}

func TestVisitorRiskServiceListsFactsAndStats(t *testing.T) {
	db, visitorRiskService := newTestVisitorRiskService(t, true)
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

	require.NoError(t, db.Create(&[]visitor.RiskDailyFact{
		{
			Day:                   now,
			IPHash:                "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			UserAgentHash:         "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			RequestCount:          10,
			InvalidRequestCount:   2,
			MeaningfulActionCount: 1,
			RiskScore:             21,
			RiskLevel:             visitor.RiskLevelWatch,
			FirstSeenAt:           now,
			LastSeenAt:            now.Add(time.Hour),
			SamplePaths:           []byte(`["/api/v1/products"]`),
		},
	}).Error)

	facts, total, err := visitorRiskService.ListFacts(1, 20, VisitorRiskFactListInput{
		RiskLevel: visitor.RiskLevelWatch,
		DayRange:  "all",
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, facts, 1)
	assert.Equal(t, "12345678...90abcdef", facts[0].IPHashPreview)
	assert.Equal(t, []string{"/api/v1/products"}, facts[0].SamplePaths)

	stats, err := visitorRiskService.GetStats("all")
	require.NoError(t, err)
	assert.EqualValues(t, 1, stats.TotalFacts)
	assert.EqualValues(t, 1, stats.WatchCount)
	assert.EqualValues(t, 10, stats.RequestCount)
	assert.EqualValues(t, 2, stats.InvalidRequestCount)
	assert.EqualValues(t, 1, stats.MeaningfulActionCount)
}

func TestVisitorRiskServiceCreatesAndResolvesCurrentDecision(t *testing.T) {
	db, visitorRiskService := newTestVisitorRiskService(t, true)
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	fact := visitor.RiskDailyFact{
		Day:           now,
		IPHash:        "ip-hash",
		UserAgentHash: "ua-hash",
		FirstSeenAt:   now,
		LastSeenAt:    now,
		RiskLevel:     visitor.RiskLevelSuspicious,
		SamplePaths:   []byte(`[]`),
	}
	require.NoError(t, db.Create(&fact).Error)

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	decision, err := visitorRiskService.CreateDecision(fact.ID, VisitorRiskDecisionInput{
		Action:    visitor.RiskDecisionActionTemporaryBlock,
		Reason:    "Repeated checkout failures require a short review window.",
		ExpiresAt: &expiresAt,
	}, 42)
	require.NoError(t, err)
	assert.Equal(t, visitor.RiskDecisionScopeIPUAHash, decision.Scope)
	assert.Equal(t, visitor.RiskDecisionIPUAValueHash(fact.IPHash, fact.UserAgentHash), decisionValueHash(t, db, decision.ID))
	assert.Equal(t, visitor.RiskDecisionActionTemporaryBlock, decision.Action)
	assert.Equal(t, uint(42), *decision.CreatedBy)

	current, err := visitorRiskService.GetDecision(fact.ID)
	require.NoError(t, err)
	require.NotNil(t, current)
	assert.Equal(t, decision.ID, current.ID)

	facts, _, err := visitorRiskService.ListFacts(1, 20, VisitorRiskFactListInput{})
	require.NoError(t, err)
	require.Len(t, facts, 1)
	require.NotNil(t, facts[0].Decision)
	assert.Equal(t, decision.ID, facts[0].Decision.ID)
}

func TestVisitorRiskServiceExpiresDecisionAndFallsBackToIPScope(t *testing.T) {
	db, visitorRiskService := newTestVisitorRiskService(t, true)
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	fact := visitor.RiskDailyFact{
		Day:           now,
		IPHash:        "ip-hash",
		UserAgentHash: "ua-hash",
		FirstSeenAt:   now,
		LastSeenAt:    now,
		RiskLevel:     visitor.RiskLevelWatch,
		SamplePaths:   []byte(`[]`),
	}
	require.NoError(t, db.Create(&fact).Error)

	expiredAt := now.Add(-time.Minute)
	require.NoError(t, db.Create(&visitor.RiskDecision{
		Scope:     visitor.RiskDecisionScopeIPUAHash,
		ValueHash: visitor.RiskDecisionIPUAValueHash(fact.IPHash, fact.UserAgentHash),
		Action:    visitor.RiskDecisionActionWatch,
		Reason:    "Expired exact decision",
		ExpiresAt: &expiredAt,
	}).Error)

	require.NoError(t, db.Create(&visitor.RiskDecision{
		Scope:     visitor.RiskDecisionScopeIPHash,
		ValueHash: fact.IPHash,
		Action:    visitor.RiskDecisionActionIgnore,
		Reason:    "Trusted IP review",
	}).Error)

	current, err := visitorRiskService.GetDecision(fact.ID)
	require.NoError(t, err)
	require.NotNil(t, current)
	assert.Equal(t, visitor.RiskDecisionScopeIPHash, current.Scope)
	assert.Equal(t, visitor.RiskDecisionActionIgnore, current.Action)
}

func TestVisitorRiskServiceDoesNotReuseIPUADecisionAcrossIPs(t *testing.T) {
	db, visitorRiskService := newTestVisitorRiskService(t, true)
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	facts := []visitor.RiskDailyFact{
		{
			Day:           now,
			IPHash:        "ip-hash-a",
			UserAgentHash: "same-ua-hash",
			FirstSeenAt:   now,
			LastSeenAt:    now,
			SamplePaths:   []byte(`[]`),
		},
		{
			Day:           now,
			IPHash:        "ip-hash-b",
			UserAgentHash: "same-ua-hash",
			FirstSeenAt:   now,
			LastSeenAt:    now,
			SamplePaths:   []byte(`[]`),
		},
	}
	require.NoError(t, db.Create(&facts).Error)

	_, err := visitorRiskService.CreateDecision(facts[0].ID, VisitorRiskDecisionInput{
		Action: visitor.RiskDecisionActionWatch,
		Reason: "Review only this IP and UA pair.",
	}, 42)
	require.NoError(t, err)

	decision, err := visitorRiskService.GetDecision(facts[1].ID)
	require.NoError(t, err)
	assert.Nil(t, decision)
}

func TestVisitorRiskServiceRequiresExpiryForTemporaryBlock(t *testing.T) {
	db, visitorRiskService := newTestVisitorRiskService(t, true)
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	fact := visitor.RiskDailyFact{
		Day:         now,
		IPHash:      "ip-hash",
		FirstSeenAt: now,
		LastSeenAt:  now,
		SamplePaths: []byte(`[]`),
	}
	require.NoError(t, db.Create(&fact).Error)

	_, err := visitorRiskService.CreateDecision(fact.ID, VisitorRiskDecisionInput{
		Action: visitor.RiskDecisionActionTemporaryBlock,
		Reason: "Needs a bounded review.",
	}, 42)
	require.ErrorIs(t, err, ErrVisitorRiskDecisionInvalid)
}

func decisionValueHash(t *testing.T, db *gorm.DB, id uint) string {
	t.Helper()
	var decision visitor.RiskDecision
	require.NoError(t, db.First(&decision, id).Error)
	return decision.ValueHash
}

func newTestVisitorRiskService(t *testing.T, enabled bool) (*gorm.DB, *VisitorRiskService) {
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

	require.NoError(t, db.AutoMigrate(&visitor.RiskDailyFact{}, &visitor.RiskDecision{}))

	return db, NewVisitorRiskService(
		repository.NewVisitorRiskFactRepository(db),
		config.VisitorRiskConfig{
			Enabled:              enabled,
			HashSalt:             "unit-test-risk-salt",
			FlushIntervalSeconds: 60,
			MaxPendingFacts:      100,
			SamplePathLimit:      8,
		},
		"",
	)
}
