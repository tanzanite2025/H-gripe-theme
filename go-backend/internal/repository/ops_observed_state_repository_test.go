package repository

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/ops"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsVPSBindingUpdateDoesNotOverwriteObservedState(t *testing.T) {
	db := newOpsObservedStateTestDB(t)
	repo := NewOpsVPSBindingRepository(db)

	observedAt := time.Date(2026, 8, 13, 8, 30, 0, 0, time.UTC)
	record := ops.VPSBinding{
		Name:               "Hostinger Production VPS",
		Provider:           ops.VPSProviderHostinger,
		Environment:        ops.VPSEnvironmentProduction,
		ProviderResourceID: "1834903",
		Hostname:           "srv1834903.hstgr.cloud",
		IPv4:               "2.25.85.201",
		Status:             ops.VPSStatusActive,
		ObservedStatus:     ops.VPSObservedUnknown,
		Enabled:            true,
	}
	require.NoError(t, repo.Create(&record))
	require.NoError(t, repo.UpdateObservedState(
		record.ID,
		ops.VPSObservedHealthy,
		"running",
		"hostinger:Production",
		"srv1834903.hstgr.cloud",
		"2.25.85.201",
		"Ubuntu 24.04 LTS",
		"KVM 2",
		"data_center:12",
		observedAt,
		"",
	))

	record.Status = ops.VPSStatusDrifted
	record.ObservedStatus = ops.VPSObservedOffline
	record.ObservedState = "stopped"
	record.ObservedSource = "manual"
	record.ObservedHostname = "manual.example.com"
	record.ObservedIPv4 = "198.51.100.1"
	record.ObservedOS = "Manual OS"
	record.ObservedPlan = "wrong"
	record.ObservedRegion = "wrong"
	record.LastError = "manual error"
	require.NoError(t, repo.Update(&record))

	refreshed, err := repo.FindByID(record.ID)
	require.NoError(t, err)
	require.Equal(t, ops.VPSStatusDrifted, refreshed.Status)
	require.Equal(t, ops.VPSObservedHealthy, refreshed.ObservedStatus)
	require.Equal(t, "running", refreshed.ObservedState)
	require.Equal(t, "hostinger:Production", refreshed.ObservedSource)
	require.Equal(t, "srv1834903.hstgr.cloud", refreshed.ObservedHostname)
	require.Equal(t, "2.25.85.201", refreshed.ObservedIPv4)
	require.Equal(t, "Ubuntu 24.04 LTS", refreshed.ObservedOS)
	require.Equal(t, "KVM 2", refreshed.ObservedPlan)
	require.Equal(t, "data_center:12", refreshed.ObservedRegion)
	require.Empty(t, refreshed.LastError)
	require.NotNil(t, refreshed.LastObservedAt)
	require.True(t, refreshed.LastObservedAt.Equal(observedAt))
}

func TestOpsProjectBindingUpdateDoesNotOverwriteObservedState(t *testing.T) {
	db := newOpsObservedStateTestDB(t)
	vpsRepo := NewOpsVPSBindingRepository(db)
	projectRepo := NewOpsProjectBindingRepository(db)

	vps := ops.VPSBinding{
		Name:               "Hostinger Production VPS",
		Provider:           ops.VPSProviderHostinger,
		Environment:        ops.VPSEnvironmentProduction,
		ProviderResourceID: "1834903",
		Status:             ops.VPSStatusActive,
		ObservedStatus:     ops.VPSObservedUnknown,
		Enabled:            true,
	}
	require.NoError(t, vpsRepo.Create(&vps))

	checkedAt := time.Date(2026, 8, 13, 9, 0, 0, 0, time.UTC)
	record := ops.ProjectBinding{
		Name:               "commerce-platform",
		VPSBindingID:       vps.ID,
		Environment:        ops.ProjectEnvironmentProduction,
		ComposeSource:      "compose.prod.yml",
		ComposeProjectName: "commerce-platform",
		Status:             ops.ProjectStatusActive,
		HealthStatus:       ops.ProjectHealthUnknown,
		Enabled:            true,
	}
	require.NoError(t, projectRepo.Create(&record))
	require.NoError(t, projectRepo.UpdateObservedState(
		record.ID,
		ops.ProjectHealthHealthy,
		"running",
		"hostinger:Production",
		6,
		6,
		5,
		checkedAt,
		"",
	))

	record.Status = ops.ProjectStatusDrifted
	record.HealthStatus = ops.ProjectHealthOffline
	record.ObservedState = "stopped"
	record.ObservedSource = "manual"
	record.ObservedContainerCount = 1
	record.ObservedRunningCount = 0
	record.ObservedHealthyCount = 0
	record.LastError = "manual error"
	require.NoError(t, projectRepo.Update(&record))

	refreshed, err := projectRepo.FindByID(record.ID)
	require.NoError(t, err)
	require.Equal(t, ops.ProjectStatusDrifted, refreshed.Status)
	require.Equal(t, ops.ProjectHealthHealthy, refreshed.HealthStatus)
	require.Equal(t, "running", refreshed.ObservedState)
	require.Equal(t, "hostinger:Production", refreshed.ObservedSource)
	require.Equal(t, 6, refreshed.ObservedContainerCount)
	require.Equal(t, 6, refreshed.ObservedRunningCount)
	require.Equal(t, 5, refreshed.ObservedHealthyCount)
	require.Empty(t, refreshed.LastError)
	require.NotNil(t, refreshed.LastCheckedAt)
	require.True(t, refreshed.LastCheckedAt.Equal(checkedAt))
}

func newOpsObservedStateTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&ops.VPSBinding{}, &ops.ProjectBinding{}))
	return db
}
