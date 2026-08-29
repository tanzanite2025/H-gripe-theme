package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/security"
	"commerce-platform/internal/repository"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGlobalIPBlockServiceMatchesExactIPAndCIDRWithMostSpecificRule(t *testing.T) {
	db, blockService := newTestGlobalIPBlockService(t)

	broad, err := blockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:            "203.0.113.0/24",
		Source:          security.IPBlockRuleSourceCommercialBot,
		SourceReference: "crawler-1",
		Reason:          "commercial crawler probe",
	})
	require.NoError(t, err)
	narrow, err := blockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:            "203.0.113.19",
		Source:          security.IPBlockRuleSourceVisitorProfile,
		SourceReference: "42",
		Reason:          "manual review",
	})
	require.NoError(t, err)
	require.NotEqual(t, broad.ID, narrow.ID)

	match, err := blockService.FindMatch(context.Background(), "203.0.113.19", time.Now())
	require.NoError(t, err)
	require.NotNil(t, match)
	assert.Equal(t, narrow.ID, match.ID)
	assert.Equal(t, "203.0.113.19/32", match.CIDR)

	match, err = blockService.FindMatch(context.Background(), "203.0.113.20", time.Now())
	require.NoError(t, err)
	require.NotNil(t, match)
	assert.Equal(t, broad.ID, match.ID)

	match, err = blockService.FindMatch(context.Background(), "198.51.100.10", time.Now())
	require.NoError(t, err)
	assert.Nil(t, match)

	var count int64
	require.NoError(t, db.Model(&security.IPBlockRule{}).Count(&count).Error)
	assert.EqualValues(t, 2, count)
}

func TestGlobalIPBlockServiceUpsertsSameSourceIdentityAndDisablesRule(t *testing.T) {
	_, blockService := newTestGlobalIPBlockService(t)

	first, err := blockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:            "2001:db8::1",
		Source:          security.IPBlockRuleSourceVisitorProfile,
		SourceReference: "profile-9",
		Reason:          "first reason",
	})
	require.NoError(t, err)
	second, err := blockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:            "2001:db8::1/128",
		Source:          security.IPBlockRuleSourceVisitorProfile,
		SourceReference: "profile-9",
		Reason:          "updated reason",
	})
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, "updated reason", second.Reason)

	match, err := blockService.FindMatch(context.Background(), "2001:db8::1", time.Now())
	require.NoError(t, err)
	require.NotNil(t, match)
	assert.Equal(t, first.ID, match.ID)

	disabled, err := blockService.DisableBySourceReference(
		context.Background(),
		security.IPBlockRuleSourceVisitorProfile,
		"profile-9",
		7,
	)
	require.NoError(t, err)
	assert.Equal(t, security.IPBlockRuleStatusDisabled, disabled.Status)
	assert.False(t, disabled.Enabled)

	match, err = blockService.FindMatch(context.Background(), "2001:db8::1", time.Now())
	require.NoError(t, err)
	assert.Nil(t, match)
}

func TestGlobalIPBlockServiceBlockWithPreviousReturnsUpdatedRuleState(t *testing.T) {
	_, blockService := newTestGlobalIPBlockService(t)

	first, err := blockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:   "198.51.100.42",
		Reason: "initial review",
	})
	require.NoError(t, err)

	before, after, err := blockService.BlockWithPrevious(context.Background(), IPBlockRuleInput{
		CIDR:   "198.51.100.42/32",
		Reason: "updated review",
	})
	require.NoError(t, err)
	require.NotNil(t, before)
	assert.Equal(t, first.ID, before.ID)
	assert.Equal(t, first.ID, after.ID)
	assert.Equal(t, "initial review", before.Reason)
	assert.Equal(t, "updated review", after.Reason)
	assert.Equal(t, security.IPBlockRuleStatusActive, before.Status)
	assert.Equal(t, security.IPBlockRuleStatusActive, after.Status)
}

func TestGlobalIPBlockServiceCreateRollsBackWhenAuditWriteFails(t *testing.T) {
	db, blockService := newTestGlobalIPBlockService(t)
	auditErr := errors.New("audit database is down")
	blockService.ConfigureAuditRecorderFactory(func(_ *gorm.DB) IPBlockAuditRecorder {
		return failingIPBlockAuditRecorder{err: auditErr}
	})

	_, after, err := blockService.BlockWithPreviousAndAudit(
		context.Background(),
		IPBlockRuleInput{
			CIDR:   "198.51.100.61",
			Reason: "audit rollback test",
		},
		func(_ *IPBlockRuleSnapshot, _ IPBlockRuleSnapshot) (*audit.AuditLog, error) {
			return &audit.AuditLog{Action: "create", Resource: "global_ip_block_rule"}, nil
		},
	)

	require.ErrorIs(t, err, ErrIPBlockAuditWrite)
	require.ErrorIs(t, err, auditErr)
	assert.NotZero(t, after.ID)

	var count int64
	require.NoError(t, db.Model(&security.IPBlockRule{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestGlobalIPBlockServiceUpdateRollsBackWhenAuditWriteFails(t *testing.T) {
	db, blockService := newTestGlobalIPBlockService(t)
	first, err := blockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:   "198.51.100.62",
		Reason: "original reason",
	})
	require.NoError(t, err)

	auditErr := errors.New("audit database is down")
	blockService.ConfigureAuditRecorderFactory(func(_ *gorm.DB) IPBlockAuditRecorder {
		return failingIPBlockAuditRecorder{err: auditErr}
	})

	_, after, err := blockService.BlockWithPreviousAndAudit(
		context.Background(),
		IPBlockRuleInput{
			CIDR:   "198.51.100.62/32",
			Reason: "updated reason",
		},
		func(_ *IPBlockRuleSnapshot, _ IPBlockRuleSnapshot) (*audit.AuditLog, error) {
			return &audit.AuditLog{Action: "update", Resource: "global_ip_block_rule"}, nil
		},
	)

	require.ErrorIs(t, err, ErrIPBlockAuditWrite)
	require.ErrorIs(t, err, auditErr)
	assert.Equal(t, first.ID, after.ID)

	var stored security.IPBlockRule
	require.NoError(t, db.First(&stored, first.ID).Error)
	assert.Equal(t, "original reason", stored.Reason)
	assert.True(t, stored.Enabled)
}

func TestGlobalIPBlockServiceDisableRollsBackWhenAuditWriteFails(t *testing.T) {
	db, blockService := newTestGlobalIPBlockService(t)
	rule, err := blockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:   "198.51.100.63",
		Reason: "disable audit rollback test",
	})
	require.NoError(t, err)

	auditErr := errors.New("audit database is down")
	blockService.ConfigureAuditRecorderFactory(func(_ *gorm.DB) IPBlockAuditRecorder {
		return failingIPBlockAuditRecorder{err: auditErr}
	})

	_, after, err := blockService.DisableWithPreviousAndAudit(
		context.Background(),
		rule.ID,
		7,
		func(_ *IPBlockRuleSnapshot, _ IPBlockRuleSnapshot) (*audit.AuditLog, error) {
			return &audit.AuditLog{Action: "delete", Resource: "global_ip_block_rule"}, nil
		},
	)

	require.ErrorIs(t, err, ErrIPBlockAuditWrite)
	require.ErrorIs(t, err, auditErr)
	assert.Equal(t, rule.ID, after.ID)
	assert.False(t, after.Enabled)

	var stored security.IPBlockRule
	require.NoError(t, db.First(&stored, rule.ID).Error)
	assert.True(t, stored.Enabled)
	assert.Nil(t, stored.DisabledAt)
	assert.Nil(t, stored.DisabledBy)
}

func TestGlobalIPBlockServiceInvalidatesStaleCacheAfterCrossInstanceNoopDisable(t *testing.T) {
	db, firstService := newTestGlobalIPBlockService(t)
	secondService := NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))

	rule, err := firstService.Block(context.Background(), IPBlockRuleInput{
		CIDR:   "203.0.113.88",
		Reason: "cross-instance cache test",
	})
	require.NoError(t, err)

	match, err := secondService.FindMatch(context.Background(), "203.0.113.88", time.Now())
	require.NoError(t, err)
	require.NotNil(t, match)
	assert.Equal(t, rule.ID, match.ID)

	_, _, err = firstService.DisableWithPrevious(context.Background(), rule.ID, 11)
	require.NoError(t, err)

	_, _, err = secondService.DisableWithPrevious(context.Background(), rule.ID, 12)
	require.NoError(t, err)

	blocked, match, err := secondService.IsBlocked(context.Background(), "203.0.113.88", time.Now())
	require.NoError(t, err)
	assert.False(t, blocked)
	assert.Nil(t, match)
}

func TestGlobalIPBlockServicePublishesCacheInvalidationAcrossInstances(t *testing.T) {
	db, firstService := newTestGlobalIPBlockService(t)
	secondService := NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))
	redisServer := miniredis.RunT(t)
	publisher := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	subscriber := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() {
		_ = publisher.Close()
		_ = subscriber.Close()
	})

	firstService.ConfigureCacheInvalidation(publisher)
	secondService.ConfigureCacheInvalidation(subscriber)
	secondService.StartCacheInvalidationListener(context.Background())
	t.Cleanup(secondService.StopCacheInvalidationListener)

	require.Eventually(t, func() bool {
		return redisServer.PubSubNumSub(globalIPBlockCacheInvalidationChannel)[globalIPBlockCacheInvalidationChannel] == 1
	}, time.Second, 5*time.Millisecond)

	rule, err := firstService.Block(context.Background(), IPBlockRuleInput{
		CIDR:   "203.0.113.89",
		Reason: "cross-instance pub/sub test",
	})
	require.NoError(t, err)

	match, err := secondService.FindMatch(context.Background(), "203.0.113.89", time.Now())
	require.NoError(t, err)
	require.NotNil(t, match)
	assert.Equal(t, rule.ID, match.ID)

	_, _, err = firstService.DisableWithPrevious(context.Background(), rule.ID, 11)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		blocked, match, findErr := secondService.IsBlocked(
			context.Background(),
			"203.0.113.89",
			time.Now(),
		)
		return findErr == nil && !blocked && match == nil
	}, time.Second, 5*time.Millisecond)
}

func TestGlobalIPBlockServiceReusesExpiredEnabledIdentity(t *testing.T) {
	db, blockService := newTestGlobalIPBlockService(t)
	expiredAt := time.Now().UTC().Add(-time.Hour)
	previous := &security.IPBlockRule{
		CIDR:            "198.51.100.24/32",
		Source:          security.IPBlockRuleSourceManual,
		SourceReference: "admin",
		Reason:          "previous window",
		ExpiresAt:       &expiredAt,
		Enabled:         true,
	}
	require.NoError(t, db.Create(previous).Error)

	reblocked, err := blockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:            "198.51.100.24",
		Source:          security.IPBlockRuleSourceManual,
		SourceReference: "admin",
		Reason:          "new window",
	})
	require.NoError(t, err)
	assert.Equal(t, previous.ID, reblocked.ID)
	assert.Equal(t, "new window", reblocked.Reason)

	var count int64
	require.NoError(t, db.Model(&security.IPBlockRule{}).Count(&count).Error)
	assert.EqualValues(t, 1, count)

	match, err := blockService.FindMatch(context.Background(), "198.51.100.24", time.Now())
	require.NoError(t, err)
	require.NotNil(t, match)
	assert.Equal(t, previous.ID, match.ID)
}

func TestGlobalIPBlockServiceFailsClosedWhenRefreshCannotCompileRule(t *testing.T) {
	db, blockService := newTestGlobalIPBlockService(t)
	valid, err := blockService.Block(context.Background(), IPBlockRuleInput{
		CIDR:   "203.0.113.0/24",
		Reason: "valid rule",
	})
	require.NoError(t, err)

	require.NoError(t, db.Create(&security.IPBlockRule{
		CIDR:            "not-a-cidr",
		Source:          security.IPBlockRuleSourceManual,
		SourceReference: "corrupt",
		Reason:          "invalid stored rule",
		Enabled:         true,
	}).Error)
	blockService.Invalidate()

	match, err := blockService.FindMatch(context.Background(), "203.0.113.19", time.Now())
	require.ErrorIs(t, err, ErrIPBlockCacheUnavailable)
	assert.Nil(t, match)
	assert.Contains(t, err.Error(), "not-a-cidr")
	assert.NotZero(t, valid.ID)
}

func TestGlobalIPBlockServiceReturnsUnavailableWithoutRepository(t *testing.T) {
	blockService := NewGlobalIPBlockService(nil)

	blocked, match, err := blockService.IsBlocked(context.Background(), "203.0.113.19", time.Now())
	require.ErrorIs(t, err, ErrIPBlockCacheUnavailable)
	assert.False(t, blocked)
	assert.Nil(t, match)
}

func TestGlobalIPBlockServiceLoadsCacheBeforeIgnoringInvalidIP(t *testing.T) {
	db, blockService := newTestGlobalIPBlockService(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	match, err := blockService.FindMatch(context.Background(), "not-an-ip", time.Now())
	require.ErrorIs(t, err, ErrIPBlockCacheUnavailable)
	assert.Nil(t, match)
}

func TestGlobalIPBlockServiceSerializesConcurrentSameIdentityCreation(t *testing.T) {
	db, blockService := newTestGlobalIPBlockService(t)
	const callers = 8

	type result struct {
		snapshot IPBlockRuleSnapshot
		err      error
	}
	results := make(chan result, callers)
	var waitGroup sync.WaitGroup
	waitGroup.Add(callers)
	for index := 0; index < callers; index++ {
		go func() {
			defer waitGroup.Done()
			snapshot, err := blockService.Block(context.Background(), IPBlockRuleInput{
				CIDR:            "192.0.2.55",
				Source:          security.IPBlockRuleSourceCommercialBot,
				SourceReference: "crawler-duplicate-check",
				Reason:          "concurrent request",
			})
			results <- result{snapshot: snapshot, err: err}
		}()
	}
	waitGroup.Wait()
	close(results)

	var firstID uint
	for item := range results {
		require.NoError(t, item.err)
		if firstID == 0 {
			firstID = item.snapshot.ID
		}
		assert.Equal(t, firstID, item.snapshot.ID)
	}

	var count int64
	require.NoError(t, db.Model(&security.IPBlockRule{}).Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

func TestNormalizeIPOrCIDRCanonicalizesIPv4AndRejectsInvalidInput(t *testing.T) {
	cases := map[string]string{
		"203.0.113.10":    "203.0.113.10/32",
		"203.0.113.10/24": "203.0.113.0/24",
		"2001:db8::1":     "2001:db8::1/128",
		"2001:db8::1/64":  "2001:db8::/64",
		" 203.0.113.10  ": "203.0.113.10/32",
	}
	for input, expected := range cases {
		actual, err := NormalizeIPOrCIDR(input)
		require.NoError(t, err, input)
		assert.Equal(t, expected, actual)
	}

	_, err := NormalizeIPOrCIDR("not-an-ip")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIPBlockRuleInvalid)
}

func newTestGlobalIPBlockService(t *testing.T) (*gorm.DB, *GlobalIPBlockService) {
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

	require.NoError(t, db.AutoMigrate(&security.IPBlockRule{}))
	return db, NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))
}

type failingIPBlockAuditRecorder struct {
	err error
}

func (r failingIPBlockAuditRecorder) CreateAuditLog(_ *audit.AuditLog) error {
	return r.err
}
