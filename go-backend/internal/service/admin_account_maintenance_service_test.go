package service

import (
	"errors"
	"testing"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/user"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAdminAccountMaintenanceServiceCreatesBackofficeAccount(t *testing.T) {
	db := newAdminAccountMaintenanceTestDB(t)
	accountService := NewAdminAccountMaintenanceService(db)

	result, err := accountService.EnsureBackofficeAccount(AdminAccountMaintenanceInput{
		Email:       "Ops.Admin@Example.com",
		Password:    "N3w-Admin-Secret!",
		Operator:    "ops@example.com",
		AuditMethod: "HTTP",
		AuditPath:   "/api/admin/ops/admin-accounts/ensure",
	})

	require.NoError(t, err)
	require.True(t, result.Created)
	require.Equal(t, "ops.admin@example.com", result.Email)
	require.Equal(t, "ops.admin", result.Username)
	require.Equal(t, "admin", result.Role)
	require.Equal(t, "active", result.Status)

	var stored user.User
	require.NoError(t, db.First(&stored, result.UserID).Error)
	require.NotEqual(t, "N3w-Admin-Secret!", stored.Password)
	require.True(t, stored.CheckPassword("N3w-Admin-Secret!"))

	var logs []audit.AuditLog
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1)
	require.Equal(t, "create", logs[0].Action)
	require.Equal(t, "user", logs[0].Resource)
	require.Equal(t, result.UserID, logs[0].ResourceID)
	require.Equal(t, "HTTP", logs[0].Method)
	require.Equal(t, "/api/admin/ops/admin-accounts/ensure", logs[0].Path)
	require.NotContains(t, logs[0].NewValue, "N3w-Admin-Secret!")
}

func TestAdminAccountMaintenanceServiceListsBackofficeAccounts(t *testing.T) {
	db := newAdminAccountMaintenanceTestDB(t)
	accountService := NewAdminAccountMaintenanceService(db)

	require.NoError(t, db.Create(&user.User{
		Email:    "admin@example.com",
		Username: "primary-admin",
		Role:     "admin",
		Status:   "active",
		Locale:   "en",
	}).Error)
	require.NoError(t, db.Create(&user.User{
		Email:    "customer@example.com",
		Username: "customer",
		Role:     "user",
		Status:   "active",
		Locale:   "en",
	}).Error)

	accounts, err := accountService.ListBackofficeAccounts("primary")

	require.NoError(t, err)
	require.Len(t, accounts, 1)
	require.Equal(t, "admin@example.com", accounts[0].Email)
	require.Equal(t, "primary-admin", accounts[0].Username)
}

func TestAdminAccountMaintenanceServiceResetsExistingAccount(t *testing.T) {
	db := newAdminAccountMaintenanceTestDB(t)
	existing := user.User{
		Email:    "admin@example.com",
		Username: "old-admin",
		Role:     "support",
		Status:   "inactive",
		Locale:   "en",
	}
	require.NoError(t, existing.HashPassword("0ld-Admin-Secret!"))
	require.NoError(t, db.Create(&existing).Error)

	accountService := NewAdminAccountMaintenanceService(db)
	result, err := accountService.EnsureBackofficeAccount(AdminAccountMaintenanceInput{
		Email:    "admin@example.com",
		Username: "primary-admin",
		Password: "N3w-Admin-Secret!",
		Role:     "manager",
		Operator: "ops@example.com",
	})

	require.NoError(t, err)
	require.False(t, result.Created)
	require.Equal(t, existing.ID, result.UserID)
	require.Equal(t, "primary-admin", result.Username)
	require.Equal(t, "manager", result.Role)
	require.Equal(t, "active", result.Status)

	var stored user.User
	require.NoError(t, db.First(&stored, existing.ID).Error)
	require.True(t, stored.CheckPassword("N3w-Admin-Secret!"))
	require.False(t, stored.CheckPassword("0ld-Admin-Secret!"))

	var auditLog audit.AuditLog
	require.NoError(t, db.Last(&auditLog).Error)
	require.Equal(t, "reset_password", auditLog.Action)
	require.NotContains(t, auditLog.NewValue, "N3w-Admin-Secret!")
	require.NotContains(t, auditLog.OldValue, "0ld-Admin-Secret!")
}

func TestAdminAccountMaintenanceServicePreservesExistingProfileDefaults(t *testing.T) {
	db := newAdminAccountMaintenanceTestDB(t)
	existing := user.User{
		Email:     "admin@example.com",
		Username:  "primary-admin",
		FirstName: "Existing",
		LastName:  "Operator",
		Role:      "support",
		Status:    "inactive",
		Locale:    "fr",
	}
	require.NoError(t, existing.HashPassword("0ld-Admin-Secret!"))
	require.NoError(t, db.Create(&existing).Error)

	accountService := NewAdminAccountMaintenanceService(db)
	result, err := accountService.EnsureBackofficeAccount(AdminAccountMaintenanceInput{
		Email:    "admin@example.com",
		Password: "N3w-Admin-Secret!",
	})

	require.NoError(t, err)
	require.False(t, result.Created)
	require.Equal(t, "primary-admin", result.Username)
	require.Equal(t, "support", result.Role)

	var stored user.User
	require.NoError(t, db.First(&stored, existing.ID).Error)
	require.Equal(t, "Existing", stored.FirstName)
	require.Equal(t, "Operator", stored.LastName)
	require.Equal(t, "fr", stored.Locale)
	require.Equal(t, "active", stored.Status)
	require.True(t, stored.CheckPassword("N3w-Admin-Secret!"))
}

func TestAdminAccountMaintenanceServiceRejectsWeakPassword(t *testing.T) {
	db := newAdminAccountMaintenanceTestDB(t)
	accountService := NewAdminAccountMaintenanceService(db)

	_, err := accountService.EnsureBackofficeAccount(AdminAccountMaintenanceInput{
		Email:    "admin@example.com",
		Password: "short1!",
	})

	require.True(t, errors.Is(err, ErrAdminAccountWeakPassword))
}

func TestAdminAccountMaintenanceServiceRejectsNonBackofficeRole(t *testing.T) {
	db := newAdminAccountMaintenanceTestDB(t)
	accountService := NewAdminAccountMaintenanceService(db)

	_, err := accountService.EnsureBackofficeAccount(AdminAccountMaintenanceInput{
		Email:    "admin@example.com",
		Password: "N3w-Admin-Secret!",
		Role:     "user",
	})

	require.True(t, errors.Is(err, ErrAdminAccountRoleForbidden))
}

func TestAdminAccountMaintenanceServiceRejectsUsernameConflict(t *testing.T) {
	db := newAdminAccountMaintenanceTestDB(t)
	conflicting := user.User{
		Email:    "other@example.com",
		Username: "primary-admin",
		Role:     "admin",
		Status:   "active",
		Locale:   "en",
	}
	require.NoError(t, conflicting.HashPassword("0ld-Admin-Secret!"))
	require.NoError(t, db.Create(&conflicting).Error)

	accountService := NewAdminAccountMaintenanceService(db)
	_, err := accountService.EnsureBackofficeAccount(AdminAccountMaintenanceInput{
		Email:    "admin@example.com",
		Username: "primary-admin",
		Password: "N3w-Admin-Secret!",
	})

	require.True(t, errors.Is(err, ErrUsernameExists))
}

func newAdminAccountMaintenanceTestDB(t *testing.T) *gorm.DB {
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

	require.NoError(t, db.AutoMigrate(&user.User{}, &audit.AuditLog{}))
	return db
}
