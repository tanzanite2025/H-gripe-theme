package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/loyalty"
	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/user"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestReferralProgramCreateVersionUsesOptimisticLock(t *testing.T) {
	db := newReferralRepositoryTestDB(t)
	require.NoError(t, db.AutoMigrate(&loyalty.ReferralProgramConfig{}))
	require.NoError(t, db.Create(referralProgramConfigFixture(1)).Error)

	repo := NewReferralProgramRepository(db)
	next := referralProgramConfigFixture(0)
	next.ReferrerRewardPoints = 1200
	require.NoError(t, repo.CreateVersion(next, 1))

	active, err := repo.FindActive()
	require.NoError(t, err)
	assert.Equal(t, 2, active.Version)
	assert.Equal(t, 1200, active.ReferrerRewardPoints)

	stale := referralProgramConfigFixture(0)
	err = repo.CreateVersion(stale, 1)
	assert.ErrorIs(t, err, ErrReferralProgramVersionConflict)
}

func TestReferralRecordStateUpdateRejectsStaleVersion(t *testing.T) {
	db := newReferralRepositoryTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&loyalty.ReferralProgramConfig{},
		&loyalty.ReferralIdentity{},
		&loyalty.ReferralRecord{},
	))
	config := referralProgramConfigFixture(1)
	require.NoError(t, db.Create(config).Error)
	identity := &loyalty.ReferralIdentity{UserID: 11, ReferralCode: "ABCD2345", IsActive: true}
	require.NoError(t, db.Create(identity).Error)
	record := &loyalty.ReferralRecord{
		ReferralIdentityID:   identity.ID,
		ProgramConfigID:      config.ID,
		ReferrerID:           identity.UserID,
		ReferralCodeSnapshot: identity.ReferralCode,
		AttributionSource:    "link",
		Currency:             "USD",
		Status:               loyalty.ReferralStatusPending,
		RecordVersion:        1,
		ExpiresAt:            time.Now().UTC().Add(30 * 24 * time.Hour),
	}
	require.NoError(t, db.Create(record).Error)

	repo := NewReferralRepository(db)
	require.NoError(t, repo.UpdateRecordState(record.ID, loyalty.ReferralStatusPending, 1, map[string]any{
		"status": loyalty.ReferralStatusOrdered,
	}))
	err := repo.UpdateRecordState(record.ID, loyalty.ReferralStatusPending, 1, map[string]any{
		"status": loyalty.ReferralStatusExpired,
	})
	assert.ErrorIs(t, err, ErrReferralRecordVersionConflict)

	stored, err := repo.FindRecordByIDForUpdate(record.ID)
	require.NoError(t, err)
	assert.Equal(t, loyalty.ReferralStatusOrdered, stored.Status)
	assert.Equal(t, 2, stored.RecordVersion)
}

func TestReferralIdentityLookupNormalizesCode(t *testing.T) {
	db := newReferralRepositoryTestDB(t)
	require.NoError(t, db.AutoMigrate(&loyalty.ReferralIdentity{}))
	identity := &loyalty.ReferralIdentity{UserID: 20, ReferralCode: "RACE2345", IsActive: true}
	require.NoError(t, db.Create(identity).Error)

	stored, err := NewReferralRepository(db).FindActiveIdentityByCode(" race2345 ")
	require.NoError(t, err)
	assert.Equal(t, identity.ID, stored.ID)

	_, err = NewReferralRepository(db).FindActiveIdentityByCode("invalid-code")
	assert.True(t, errors.Is(err, loyalty.ErrInvalidReferralCode))
}

func TestReferralAdminLedgerFiltersAndStats(t *testing.T) {
	db := newReferralRepositoryTestDB(t)
	require.NoError(t, db.AutoMigrate(&user.User{}, &orderdomain.Order{}, &loyalty.ReferralProgramConfig{}, &loyalty.ReferralIdentity{}, &loyalty.ReferralRecord{}))
	config := referralProgramConfigFixture(1)
	config.ReferrerRewardPoints = 1250
	config.Enabled = true
	require.NoError(t, db.Create(config).Error)
	referrer := &user.User{Email: "alex@example.test", Username: "alex", Password: "unused", Status: "active", Role: "user"}
	referee := &user.User{Email: "david@example.test", Username: "david", Password: "unused", Status: "active", Role: "user"}
	require.NoError(t, db.Create(referrer).Error)
	require.NoError(t, db.Create(referee).Error)
	identity := &loyalty.ReferralIdentity{UserID: referrer.ID, ReferralCode: "ALEX2345", IsActive: true}
	require.NoError(t, db.Create(identity).Error)
	flags, err := json.Marshal([]map[string]any{{"type": "shipping_address_match", "level": "high"}})
	require.NoError(t, err)
	record := &loyalty.ReferralRecord{
		ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &referee.ID,
		ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", Currency: "USD", OrderAmountMinor: 24000,
		Status: loyalty.ReferralStatusVesting, RecordVersion: 1, ExpiresAt: time.Now().UTC().Add(24 * time.Hour), RiskFlags: flags,
	}
	require.NoError(t, db.Create(record).Error)
	repo := NewReferralRepository(db)
	items, total, err := repo.ListAdminRecords(ReferralAdminFilters{Status: loyalty.ReferralStatusVesting, Keyword: "alex@example"}, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, record.ID, items[0].ID)

	stats, err := repo.AdminStats()
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.TotalReferrals)
	assert.Equal(t, int64(1), stats.ConvertedOrders)
	assert.Equal(t, int64(24000), stats.AttributedGMVMinor)
	assert.Equal(t, int64(1250), stats.PendingVestingPoints)
	assert.Equal(t, int64(1), stats.FraudBlockedCount)
}

func newReferralRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:referral-%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	return db
}

func referralProgramConfigFixture(version int) *loyalty.ReferralProgramConfig {
	return &loyalty.ReferralProgramConfig{
		Version:                 version,
		Status:                  "active",
		Enabled:                 false,
		Currency:                "USD",
		MinOrderAmountMinor:     20000,
		ReferrerRewardPoints:    1000,
		RefereeBenefitType:      loyalty.ReferralBenefitNone,
		VestingPeriodDays:       30,
		UndeliveredFallbackDays: 45,
		AttributionTTLDays:      30,
		MonthlyCapPerReferrer:   10,
		AntiFraudMode:           loyalty.ReferralFraudModeMonitor,
	}
}
