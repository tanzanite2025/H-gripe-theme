package service

import (
	"testing"

	"tanzanite/internal/domain/showcase"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShowcaseListPublicOnlyReturnsApprovedItems(t *testing.T) {
	db, showcaseService := newTestShowcaseService(t)

	require.NoError(t, db.Create(&showcase.Showcase{
		UserID: 1,
		Kind:   showcase.KindUser,
		Title:  "approved",
		Status: showcase.StatusApproved,
	}).Error)
	require.NoError(t, db.Create(&showcase.Showcase{
		UserID: 2,
		Kind:   showcase.KindUser,
		Title:  "pending",
		Status: showcase.StatusPending,
	}).Error)

	items, err := showcaseService.ListPublic(showcase.KindUser, 1, 20)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "approved", items[0].Title)
}

func newTestShowcaseService(t *testing.T) (*gorm.DB, *ShowcaseService) {
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

	require.NoError(t, db.AutoMigrate(&showcase.Showcase{}, &showcase.Comment{}))
	return db, NewShowcaseService(repository.NewShowcaseRepository(db), nil)
}
