package service

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProcessExpiredTerminalDataPurgesSevenDayReportsAndKeepsAfterSalesArchivePolicy(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&sitequality.SiteQualityRun{},
		&sitequality.SiteQualityRunArchive{},
		&aftersales.AfterSalesCase{},
		&aftersales.AfterSalesCaseItem{},
		&aftersales.AfterSalesCaseEvent{},
		&aftersales.AfterSalesCaseEventArchive{},
		&aftersales.AfterSalesCaseAttachment{},
		&aftersales.AfterSalesRefundReview{},
	))

	now := time.Date(2026, time.November, 16, 12, 0, 0, 0, time.UTC)
	expiredReport := sitequality.SiteQualityRun{
		TargetURL: "https://example.com/expired",
		Strategy:  sitequality.SiteQualityStrategyMobile,
		Status:    sitequality.SiteQualityRunStatusSuccess,
		CreatedAt: now.Add(-8 * 24 * time.Hour),
		UpdatedAt: now.Add(-8 * 24 * time.Hour),
	}
	recentReport := sitequality.SiteQualityRun{
		TargetURL: "https://example.com/recent",
		Strategy:  sitequality.SiteQualityStrategyDesktop,
		Status:    sitequality.SiteQualityRunStatusSuccess,
		CreatedAt: now.Add(-6 * 24 * time.Hour),
		UpdatedAt: now.Add(-6 * 24 * time.Hour),
	}
	expiredArchivedReport := sitequality.SiteQualityRunArchive{SiteQualityRun: sitequality.SiteQualityRun{
		TargetURL: "https://example.com/expired-archive",
		Strategy:  sitequality.SiteQualityStrategyMobile,
		Status:    sitequality.SiteQualityRunStatusFailed,
		CreatedAt: now.Add(-100 * 24 * time.Hour),
		UpdatedAt: now.Add(-100 * 24 * time.Hour),
	}}
	require.NoError(t, db.Create(&expiredReport).Error)
	require.NoError(t, db.Create(&recentReport).Error)
	require.NoError(t, db.Create(&expiredArchivedReport).Error)

	closedAt := now.Add(-100 * 24 * time.Hour)
	oldCase := aftersales.AfterSalesCase{
		OrderID:   1,
		Type:      aftersales.TypeRefundOnly,
		Status:    aftersales.StatusCompleted,
		Reason:    "Damaged item",
		ClosedAt:  &closedAt,
		CreatedAt: closedAt.Add(-time.Hour),
		UpdatedAt: closedAt,
	}
	require.NoError(t, db.Create(&oldCase).Error)
	oldEvent := aftersales.AfterSalesCaseEvent{
		CaseID:     oldCase.ID,
		ToStatus:   aftersales.StatusRequested,
		Resolution: "Created",
		CreatedAt:  closedAt.Add(-time.Hour),
	}
	require.NoError(t, db.Create(&oldEvent).Error)

	maintenance := NewHotDataArchiveService(repository.NewHotDataArchiveRepository(db), 100)
	result, err := maintenance.ProcessExpiredTerminalData(context.Background(), now)
	require.NoError(t, err)
	require.Equal(t, 1, result.SiteQualityRunsPurged)
	require.Equal(t, 1, result.ArchivedSiteQualityRunsPurged)
	require.Equal(t, 1, result.AfterSalesEventsArchived)
	require.True(t, result.SiteQualityRunCutoff.Equal(now.Add(-7*24*time.Hour)))
	require.True(t, result.AfterSalesEventArchiveCutoff.Equal(now.Add(-90*24*time.Hour)))

	var remainingReports int64
	require.NoError(t, db.Model(&sitequality.SiteQualityRun{}).Count(&remainingReports).Error)
	require.Equal(t, int64(1), remainingReports)
	var remainingArchivedReports int64
	require.NoError(t, db.Model(&sitequality.SiteQualityRunArchive{}).Count(&remainingArchivedReports).Error)
	require.Zero(t, remainingArchivedReports)
	var archivedAfterSalesEvents int64
	require.NoError(t, db.Model(&aftersales.AfterSalesCaseEventArchive{}).Count(&archivedAfterSalesEvents).Error)
	require.Equal(t, int64(1), archivedAfterSalesEvents)
}
