package service

import (
	"testing"

	spokedomain "tanzanite/internal/domain/spoke"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSpokeServiceListUserHistoryFiltersByOwner(t *testing.T) {
	db, spokeService := newTestSpokeService(t)

	userID := uint(7)
	otherUserID := uint(8)
	rimModel := "R 460"
	otherRimModel := "Hidden"

	require.NoError(t, db.Create(&spokedomain.History{
		UserID:   &userID,
		RimModel: &rimModel,
	}).Error)
	require.NoError(t, db.Create(&spokedomain.History{
		UserID:   &otherUserID,
		RimModel: &otherRimModel,
	}).Error)

	items, total, err := spokeService.ListUserHistory(userID, "", 1, 10)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, items, 1)
	require.NotNil(t, items[0].RimModel)
	assert.Equal(t, rimModel, *items[0].RimModel)
}

func newTestSpokeService(t *testing.T) (*gorm.DB, *SpokeService) {
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

	require.NoError(t, db.AutoMigrate(&spokedomain.History{}))

	return db, NewSpokeService(repository.NewSpokeRepository(db))
}
