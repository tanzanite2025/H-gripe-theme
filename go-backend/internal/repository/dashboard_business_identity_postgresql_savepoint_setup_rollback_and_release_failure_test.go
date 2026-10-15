package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createMockPostgreSQLDashboardBusinessIdentityTransaction(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	mock.ExpectBegin()
	tx := db.Begin()
	require.NoError(t, tx.Error)
	return tx, mock
}

func TestDashboardBusinessIdentityReadDoesNotStartWhenSavepointSetupFails(t *testing.T) {
	db, mock := createMockPostgreSQLDashboardBusinessIdentityTransaction(t)
	setupError := errors.New("savepoint setup connection failure")
	mock.ExpectExec(`SAVEPOINT dashboard_business_identity_[0-9]+`).WillReturnError(setupError)
	readStarted := false
	err := readDashboardBusinessIdentitiesWithPostgreSQLSavepointIsolation(db, func(*gorm.DB) error {
		readStarted = true
		return nil
	})
	require.ErrorIs(t, err, setupError)
	require.False(t, readStarted)
	mock.ExpectRollback()
	require.NoError(t, db.Rollback().Error)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardBusinessIdentityReadPreservesDomainAndSavepointRollbackErrors(t *testing.T) {
	db, mock := createMockPostgreSQLDashboardBusinessIdentityTransaction(t)
	domainError := errors.New("business identity query failed")
	rollbackError := errors.New("savepoint rollback connection failure")
	mock.ExpectExec(`SAVEPOINT dashboard_business_identity_[0-9]+`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`ROLLBACK TO SAVEPOINT dashboard_business_identity_[0-9]+`).WillReturnError(rollbackError)
	err := readDashboardBusinessIdentitiesWithPostgreSQLSavepointIsolation(db, func(*gorm.DB) error { return domainError })
	require.ErrorIs(t, err, domainError)
	require.ErrorIs(t, err, rollbackError)
	mock.ExpectRollback()
	require.NoError(t, db.Rollback().Error)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardBusinessIdentityReadReportsSavepointReleaseFailureAfterSuccessfulRead(t *testing.T) {
	db, mock := createMockPostgreSQLDashboardBusinessIdentityTransaction(t)
	releaseError := errors.New("savepoint release connection failure")
	mock.ExpectExec(`SAVEPOINT dashboard_business_identity_[0-9]+`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`RELEASE SAVEPOINT dashboard_business_identity_[0-9]+`).WillReturnError(releaseError)
	err := readDashboardBusinessIdentitiesWithPostgreSQLSavepointIsolation(db, func(*gorm.DB) error { return nil })
	require.ErrorIs(t, err, releaseError)
	mock.ExpectRollback()
	require.NoError(t, db.Rollback().Error)
	require.NoError(t, mock.ExpectationsWereMet())
}
