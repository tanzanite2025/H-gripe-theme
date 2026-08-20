package repository

import (
	"testing"

	"commerce-platform/internal/domain/order"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOrderPolicyDisclosureRepositoryDistinguishesMissingFromDatabaseError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&order.PolicyDisclosure{}))

	repo := NewOrderPolicyDisclosureRepository(db)
	_, err = repo.FindByOrderID(42)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	require.NoError(t, db.Migrator().DropTable(&order.PolicyDisclosure{}))
	_, err = repo.FindByOrderID(42)
	require.Error(t, err)
	require.False(t, IsRecordNotFound(err))
}
