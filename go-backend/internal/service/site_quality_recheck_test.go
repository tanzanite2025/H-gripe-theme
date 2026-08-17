package service

import (
	"context"
	"testing"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestEnqueueRecheckFindingBindsTheExactFinding(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&sitequalitydomain.SiteQualityTarget{},
		&sitequalitydomain.SiteQualityJob{},
	))

	targetRepo := repository.NewSiteQualityTargetRepository(db)
	jobRepo := repository.NewSiteQualityJobRepository(db)
	now := time.Now().UTC()
	target, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/recheck",
		Source:                  sitequalitydomain.SiteQualityTargetSourceOperator,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
		SamplingIntervalSeconds: 604800,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)

	engine := &SiteQualityEngineService{
		targets: targetRepo,
		jobs:    jobRepo,
		cfg: SiteQualityEngineConfig{
			SampleCount:           3,
			RequiredConfirmations: 2,
			MaxAttempts:           4,
		},
	}
	finding := &sitequalitydomain.SiteQualityFinding{
		ID:        321,
		TargetID:  &target.ID,
		TargetURL: target.CanonicalURL,
		Strategy:  sitequalitydomain.SiteQualityStrategyDesktop,
	}

	job, err := engine.EnqueueRecheckFinding(finding, 7)
	require.NoError(t, err)
	require.Equal(t, sitequalitydomain.SiteQualityJobKindRecheck, job.Kind)
	require.NotNil(t, job.FindingID)
	require.Equal(t, finding.ID, *job.FindingID)
	require.Equal(t, target.ID, job.TargetID)
	require.Equal(t, sitequalitydomain.SiteQualityStrategyDesktop, job.Strategy)
}

func TestEnqueueManualTargetRejectsDisabledLedgerTarget(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&sitequalitydomain.SiteQualityTarget{},
		&sitequalitydomain.SiteQualityJob{},
	))

	targetRepo := repository.NewSiteQualityTargetRepository(db)
	jobRepo := repository.NewSiteQualityJobRepository(db)
	now := time.Now().UTC()
	routeEntryID := uint(77)
	_, err = targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		RouteEntryID:            &routeEntryID,
		CanonicalURL:            "https://example.com/retired",
		Source:                  sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
		SamplingIntervalSeconds: 604800,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)
	require.NoError(t, targetRepo.DisableByRouteEntryID(routeEntryID, "route retired", "manifest-2", now.Add(time.Minute)))

	engine := &SiteQualityEngineService{
		targets: targetRepo,
		jobs:    jobRepo,
		lighthouseRunner: NewLighthouseRunnerService(
			nil,
			nil,
			LighthouseRunnerConfig{StorefrontBaseURL: "https://example.com"},
		),
		cfg: SiteQualityEngineConfig{
			SampleCount:           3,
			RequiredConfirmations: 2,
			MaxAttempts:           4,
		},
	}

	job, err := engine.EnqueueManualTarget(
		context.Background(),
		"https://example.com/retired",
		sitequalitydomain.SiteQualityStrategyMobile,
		7,
		sitequalitydomain.SiteQualityJobKindManual,
	)
	require.ErrorIs(t, err, ErrInvalidSiteQualityRun)
	require.Nil(t, job)

	var count int64
	require.NoError(t, db.Model(&sitequalitydomain.SiteQualityJob{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestPlanDueWorkCreatesScheduledJobAndAdvancesCadence(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&sitequalitydomain.SiteQualityTarget{},
		&sitequalitydomain.SiteQualityJob{},
	))

	targetRepo := repository.NewSiteQualityTargetRepository(db)
	jobRepo := repository.NewSiteQualityJobRepository(db)
	now := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)
	target, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/scheduled",
		Source:                  sitequalitydomain.SiteQualityTargetSourceOperator,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierCritical,
		SamplingIntervalSeconds: 24 * 60 * 60,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)

	engine := &SiteQualityEngineService{
		targets: targetRepo,
		jobs:    jobRepo,
		cfg: SiteQualityEngineConfig{
			SampleCount:           3,
			RequiredConfirmations: 2,
			MaxAttempts:           4,
			ReleaseID:             "test-release",
		},
	}

	result, err := engine.PlanDueWork(now, 10)
	require.NoError(t, err)
	require.Equal(t, 1, result.EnqueuedJobs)
	require.Zero(t, result.SkippedJobs)

	var job sitequalitydomain.SiteQualityJob
	require.NoError(t, db.Where("target_id = ?", target.ID).First(&job).Error)
	require.Equal(t, sitequalitydomain.SiteQualityJobKindScheduled, job.Kind)
	require.Equal(t, sitequalitydomain.SiteQualityStrategyMobile, job.Strategy)
	require.Equal(t, sitequalitydomain.SiteQualityJobStatusQueued, job.Status)
	require.Equal(t, "test-release", job.ReleaseID)

	updated, err := targetRepo.FindByID(target.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.NextScheduledAt)
	require.Equal(t, now.Add(24*time.Hour), *updated.NextScheduledAt)
}
