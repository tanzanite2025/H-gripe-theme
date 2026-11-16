package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	aftersalesdomain "commerce-platform/internal/domain/aftersales"
	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPurgeExpiredTerminalSiteQualityRunsDeletesReportsInBoundedBatches(t *testing.T) {
	db := newHotDataArchiveTestDB(t)
	now := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
	oldRun := sitequalitydomain.SiteQualityRun{
		TargetURL:       "https://example.com/old",
		Strategy:        sitequalitydomain.SiteQualityStrategyMobile,
		Status:          sitequalitydomain.SiteQualityRunStatusSuccess,
		RawResponseJSON: strings.Repeat("x", 512*1024),
		CreatedAt:       now.Add(-8 * 24 * time.Hour),
		UpdatedAt:       now.Add(-8 * 24 * time.Hour),
	}
	secondExpiredRun := sitequalitydomain.SiteQualityRun{
		TargetURL: "https://example.com/second-old",
		Strategy:  sitequalitydomain.SiteQualityStrategyDesktop,
		Status:    sitequalitydomain.SiteQualityRunStatusFailed,
		CreatedAt: now.Add(-9 * 24 * time.Hour),
		UpdatedAt: now.Add(-9 * 24 * time.Hour),
	}
	recentRun := sitequalitydomain.SiteQualityRun{
		TargetURL: "https://example.com/recent",
		Strategy:  sitequalitydomain.SiteQualityStrategyMobile,
		Status:    sitequalitydomain.SiteQualityRunStatusSuccess,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}
	require.NoError(t, db.Create(&oldRun).Error)
	require.NoError(t, db.Create(&secondExpiredRun).Error)
	require.NoError(t, db.Create(&recentRun).Error)

	repository := NewHotDataArchiveRepository(db)
	cutoff := now.Add(-7 * 24 * time.Hour)
	purged, err := repository.PurgeExpiredTerminalSiteQualityRuns(context.Background(), cutoff, 1)
	require.NoError(t, err)
	require.Equal(t, 1, purged)

	var hotCount int64
	require.NoError(t, db.Model(&sitequalitydomain.SiteQualityRun{}).Count(&hotCount).Error)
	require.Equal(t, int64(2), hotCount)

	var archiveCount int64
	require.NoError(t, db.Model(&sitequalitydomain.SiteQualityRunArchive{}).Count(&archiveCount).Error)
	require.Zero(t, archiveCount)

	purged, err = repository.PurgeExpiredTerminalSiteQualityRuns(context.Background(), cutoff, 10)
	require.NoError(t, err)
	require.Equal(t, 1, purged)
	require.NoError(t, db.Model(&sitequalitydomain.SiteQualityRun{}).Count(&hotCount).Error)
	require.Equal(t, int64(1), hotCount)

	runRepo := NewSiteQualityRunRepository(db)
	runs, total, err := runRepo.List(SiteQualityRunListFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, runs, 1)
	require.Equal(t, recentRun.ID, runs[0].ID)

	latestSuccessAt, err := runRepo.LatestSuccessfulAt()
	require.NoError(t, err)
	require.NotNil(t, latestSuccessAt)
	require.True(t, latestSuccessAt.Equal(recentRun.CreatedAt))
}

func TestPurgeExpiredArchivedSiteQualityRunsDeletesReportsAfterRetentionCutoff(t *testing.T) {
	db := newHotDataArchiveTestDB(t)
	now := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
	expiredRun := sitequalitydomain.SiteQualityRunArchive{SiteQualityRun: sitequalitydomain.SiteQualityRun{
		TargetURL: "https://example.com/expired-archive",
		Strategy:  sitequalitydomain.SiteQualityStrategyMobile,
		Status:    sitequalitydomain.SiteQualityRunStatusSuccess,
		CreatedAt: now.Add(-8 * 24 * time.Hour),
		UpdatedAt: now.Add(-8 * 24 * time.Hour),
	}}
	recentRun := sitequalitydomain.SiteQualityRunArchive{SiteQualityRun: sitequalitydomain.SiteQualityRun{
		TargetURL: "https://example.com/recent-archive",
		Strategy:  sitequalitydomain.SiteQualityStrategyDesktop,
		Status:    sitequalitydomain.SiteQualityRunStatusSuccess,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}}
	require.NoError(t, db.Create(&expiredRun).Error)
	require.NoError(t, db.Create(&recentRun).Error)

	repository := NewHotDataArchiveRepository(db)
	purged, err := repository.PurgeExpiredArchivedSiteQualityRuns(
		context.Background(),
		now.Add(-7*24*time.Hour),
		10,
	)
	require.NoError(t, err)
	require.Equal(t, 1, purged)

	var remaining int64
	require.NoError(t, db.Model(&sitequalitydomain.SiteQualityRunArchive{}).Count(&remaining).Error)
	require.Equal(t, int64(1), remaining)
}

func TestArchiveTerminalAfterSalesEventsKeepsCaseTimeline(t *testing.T) {
	db := newHotDataArchiveTestDB(t)
	now := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
	closedAt := now.Add(-91 * 24 * time.Hour)
	caseRecord := aftersalesdomain.AfterSalesCase{
		OrderID:   1,
		Type:      aftersalesdomain.TypeRefundOnly,
		Status:    aftersalesdomain.StatusCompleted,
		Reason:    "Damaged item",
		ClosedAt:  &closedAt,
		CreatedAt: closedAt.Add(-time.Hour),
		UpdatedAt: closedAt,
	}
	require.NoError(t, db.Create(&caseRecord).Error)
	oldEvent := aftersalesdomain.AfterSalesCaseEvent{
		CaseID:     caseRecord.ID,
		ToStatus:   aftersalesdomain.StatusRequested,
		Resolution: "Created",
		CreatedAt:  closedAt.Add(-time.Hour),
	}
	terminalEvent := aftersalesdomain.AfterSalesCaseEvent{
		CaseID:     caseRecord.ID,
		FromStatus: aftersalesdomain.StatusResolving,
		ToStatus:   aftersalesdomain.StatusCompleted,
		Resolution: "Completed",
		CreatedAt:  closedAt,
	}
	require.NoError(t, db.Create(&oldEvent).Error)
	require.NoError(t, db.Create(&terminalEvent).Error)

	archiveRepo := NewHotDataArchiveRepository(db)
	archived, err := archiveRepo.ArchiveTerminalAfterSalesEvents(context.Background(), now.Add(-90*24*time.Hour), 10)
	require.NoError(t, err)
	require.Equal(t, 2, archived)

	var hotCount int64
	require.NoError(t, db.Model(&aftersalesdomain.AfterSalesCaseEvent{}).Count(&hotCount).Error)
	require.Zero(t, hotCount)
	var archiveCount int64
	require.NoError(t, db.Model(&aftersalesdomain.AfterSalesCaseEventArchive{}).Count(&archiveCount).Error)
	require.Equal(t, int64(2), archiveCount)

	caseRepo := NewAfterSalesCaseRepository(db)
	record, err := caseRepo.FindByID(caseRecord.ID)
	require.NoError(t, err)
	require.Len(t, record.Events, 2)
	require.Equal(t, oldEvent.ID, record.Events[0].ID)
	require.Equal(t, terminalEvent.ID, record.Events[1].ID)
}

func newHotDataArchiveTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&sitequalitydomain.SiteQualityRun{},
		&sitequalitydomain.SiteQualityRunArchive{},
		&aftersalesdomain.AfterSalesCase{},
		&aftersalesdomain.AfterSalesCaseItem{},
		&aftersalesdomain.AfterSalesCaseEvent{},
		&aftersalesdomain.AfterSalesCaseEventArchive{},
		&aftersalesdomain.AfterSalesCaseAttachment{},
		&aftersalesdomain.AfterSalesRefundReview{},
	))
	return db
}
