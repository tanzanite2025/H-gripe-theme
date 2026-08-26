package repository

import (
	"encoding/json"
	"testing"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSiteQualityRunListAlignsHotAndArchiveColumns(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	const columns = `
		id INTEGER PRIMARY KEY,
		target_id INTEGER,
		job_id INTEGER,
		target_url TEXT NOT NULL,
		canonical_url TEXT NOT NULL DEFAULT '',
		final_url TEXT NOT NULL DEFAULT '',
		strategy TEXT NOT NULL,
		status TEXT NOT NULL,
		initiated_by_user_id INTEGER NOT NULL DEFAULT 0,
		provider TEXT NOT NULL DEFAULT '',
		lighthouse_version TEXT NOT NULL DEFAULT '',
		environment_json TEXT NOT NULL DEFAULT '{}',
		release_id TEXT NOT NULL DEFAULT '',
		performance_score INTEGER,
		accessibility_score INTEGER,
		best_practices_score INTEGER,
		seo_score INTEGER,
		first_contentful_paint_ms REAL,
		largest_contentful_paint_ms REAL,
		interaction_to_next_paint_ms REAL,
		cumulative_layout_shift REAL,
		total_blocking_time_ms REAL,
		speed_index_ms REAL,
		issues_json TEXT NOT NULL DEFAULT '[]',
		raw_response_json TEXT NOT NULL DEFAULT '{}',
		error_message TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	`
	require.NoError(t, db.Exec("CREATE TABLE site_quality_runs ("+
		"id INTEGER PRIMARY KEY, target_url TEXT NOT NULL, final_url TEXT NOT NULL, strategy TEXT NOT NULL, "+
		"status TEXT NOT NULL, initiated_by_user_id INTEGER NOT NULL DEFAULT 0, performance_score INTEGER, "+
		"accessibility_score INTEGER, best_practices_score INTEGER, seo_score INTEGER, "+
		"first_contentful_paint_ms REAL, largest_contentful_paint_ms REAL, interaction_to_next_paint_ms REAL, "+
		"cumulative_layout_shift REAL, total_blocking_time_ms REAL, speed_index_ms REAL, issues_json TEXT NOT NULL, "+
		"raw_response_json TEXT NOT NULL, error_message TEXT NOT NULL, created_at DATETIME NOT NULL, "+
		"updated_at DATETIME NOT NULL, target_id INTEGER, job_id INTEGER, canonical_url TEXT NOT NULL, "+
		"provider TEXT NOT NULL, lighthouse_version TEXT NOT NULL, environment_json TEXT NOT NULL, release_id TEXT NOT NULL"+
		")").Error)
	require.NoError(t, db.Exec("CREATE TABLE site_quality_runs_archive ("+columns+")").Error)

	now := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Exec(`
		INSERT INTO site_quality_runs (
			id, target_url, final_url, strategy, status, initiated_by_user_id,
			performance_score, accessibility_score, best_practices_score, seo_score,
			first_contentful_paint_ms, largest_contentful_paint_ms, interaction_to_next_paint_ms,
			cumulative_layout_shift, total_blocking_time_ms, speed_index_ms, issues_json,
			raw_response_json, error_message, created_at, updated_at, target_id, job_id,
			canonical_url, provider, lighthouse_version, environment_json, release_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, 1, "https://example.com/hot", "https://example.com/hot", "mobile", "success", 0,
		90, 90, 90, 90, 100, 200, 20, 0.1, 30, 300, "[]", "{}", "", now, now, nil, nil,
		"https://example.com/hot", "lighthouse_runner", "", "{}", "").Error)
	require.NoError(t, db.Exec(`
		INSERT INTO site_quality_runs_archive (
			id, target_id, job_id, target_url, canonical_url, final_url, strategy, status,
			initiated_by_user_id, provider, lighthouse_version, environment_json, release_id,
			performance_score, accessibility_score, best_practices_score, seo_score,
			first_contentful_paint_ms, largest_contentful_paint_ms, interaction_to_next_paint_ms,
			cumulative_layout_shift, total_blocking_time_ms, speed_index_ms, issues_json,
			raw_response_json, error_message, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, 2, nil, nil, "https://example.com/archive", "https://example.com/archive",
		"https://example.com/archive", "desktop", "success", 0, "lighthouse_runner", "", "{}", "",
		80, 80, 80, 80, 110, 210, 30, 0.2, 40, 310, "[]", "{}", "", now.Add(-time.Hour), now.Add(-time.Hour)).Error)

	runs, total, err := NewSiteQualityRunRepository(db).List(SiteQualityRunListFilter{
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, runs, 2)
	require.Equal(t, "https://example.com/hot", runs[0].TargetURL)
	require.Equal(t, "https://example.com/archive", runs[1].TargetURL)
}

func TestSiteQualityJobClaimIsSingleConsumerAndRecoversStaleLease(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	jobRepo := NewSiteQualityJobRepository(db)
	now := time.Now().UTC()
	target, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/",
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierCritical,
		SamplingIntervalSeconds: 86400,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)

	_, created, err := jobRepo.Enqueue(sitequalitydomain.SiteQualityJob{
		TargetID:              target.ID,
		Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
		Kind:                  sitequalitydomain.SiteQualityJobKindManual,
		Status:                sitequalitydomain.SiteQualityJobStatusQueued,
		IdempotencyKey:        "job-once",
		SampleCount:           3,
		RequiredConfirmations: 2,
		MaxAttempts:           4,
		AvailableAt:           now,
	})
	require.NoError(t, err)
	require.True(t, created)

	firstClaim, err := jobRepo.ClaimReady(now, "worker-a", 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, firstClaim, 1)
	require.Equal(t, "worker-a", firstClaim[0].LockedBy)

	secondClaim, err := jobRepo.ClaimReady(now, "worker-b", 1, time.Minute)
	require.NoError(t, err)
	require.Empty(t, secondClaim)

	recoveredClaim, err := jobRepo.ClaimReady(now.Add(2*time.Minute), "worker-b", 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, recoveredClaim, 1)
	require.Equal(t, "worker-b", recoveredClaim[0].LockedBy)
	require.Equal(t, 2, recoveredClaim[0].Attempts)
}

func TestSiteQualityJobClaimConsumesScheduledManualAndRecheckJobs(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	jobRepo := NewSiteQualityJobRepository(db)
	now := time.Now().UTC()
	target, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/manual-only",
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
		SamplingIntervalSeconds: 604800,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)

	for _, job := range []sitequalitydomain.SiteQualityJob{
		{
			TargetID:              target.ID,
			Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
			Kind:                  sitequalitydomain.SiteQualityJobKindScheduled,
			Status:                sitequalitydomain.SiteQualityJobStatusQueued,
			IdempotencyKey:        "scheduled-inspection",
			SampleCount:           3,
			RequiredConfirmations: 2,
			MaxAttempts:           4,
			AvailableAt:           now,
		},
		{
			TargetID:              target.ID,
			Strategy:              sitequalitydomain.SiteQualityStrategyDesktop,
			Kind:                  sitequalitydomain.SiteQualityJobKindManual,
			Status:                sitequalitydomain.SiteQualityJobStatusQueued,
			IdempotencyKey:        "manual-inspection",
			SampleCount:           3,
			RequiredConfirmations: 2,
			MaxAttempts:           4,
			AvailableAt:           now,
		},
		{
			TargetID:              target.ID,
			Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
			Kind:                  sitequalitydomain.SiteQualityJobKindRecheck,
			Status:                sitequalitydomain.SiteQualityJobStatusQueued,
			IdempotencyKey:        "finding-recheck",
			SampleCount:           3,
			RequiredConfirmations: 2,
			MaxAttempts:           4,
			AvailableAt:           now,
		},
	} {
		require.NoError(t, db.Create(&job).Error)
	}

	claimed, err := jobRepo.ClaimReady(now, "worker-all-kinds", 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, claimed, 3)
	require.Equal(t, sitequalitydomain.SiteQualityJobKindRecheck, claimed[0].Kind)
	require.Equal(t, sitequalitydomain.SiteQualityJobKindManual, claimed[1].Kind)
	require.Equal(t, sitequalitydomain.SiteQualityJobKindScheduled, claimed[2].Kind)
}

func TestSiteQualityJobClaimSkipsJobsForDisabledTargets(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	jobRepo := NewSiteQualityJobRepository(db)
	now := time.Now().UTC()
	target, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/disabled",
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
		SamplingIntervalSeconds: 604800,
		Enabled:                 false,
	}, now)
	require.NoError(t, err)

	require.NoError(t, db.Create(&sitequalitydomain.SiteQualityJob{
		TargetID:              target.ID,
		Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
		Kind:                  sitequalitydomain.SiteQualityJobKindScheduled,
		Status:                sitequalitydomain.SiteQualityJobStatusQueued,
		IdempotencyKey:        "disabled-scheduled",
		SampleCount:           3,
		RequiredConfirmations: 2,
		MaxAttempts:           4,
		AvailableAt:           now,
	}).Error)

	claimed, err := jobRepo.ClaimReady(now, "worker-disabled-target", 10, time.Minute)
	require.NoError(t, err)
	require.Empty(t, claimed)

	stats, err := jobRepo.Stats(now, time.Minute)
	require.NoError(t, err)
	require.Equal(t, int64(1), stats.Total)
	require.Equal(t, int64(0), stats.Claimable)
}

func TestSiteQualityProviderSlotSerializesAcrossWorkers(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	jobRepo := NewSiteQualityJobRepository(db)
	now := time.Now().UTC()

	slot, retryAt, err := jobRepo.AcquireProviderSlot(
		"lighthouse_runner",
		1,
		"worker-a",
		now,
		time.Minute,
	)
	require.NoError(t, err)
	require.NotNil(t, slot)
	require.Nil(t, retryAt)

	blocked, retryAt, err := jobRepo.AcquireProviderSlot(
		"lighthouse_runner",
		1,
		"worker-b",
		now,
		time.Minute,
	)
	require.NoError(t, err)
	require.Nil(t, blocked)
	require.NotNil(t, retryAt)

	require.NoError(t, jobRepo.ReleaseProviderSlot(
		"lighthouse_runner",
		slot.Slot,
		"worker-a",
		now,
		5*time.Second,
	))

	blocked, retryAt, err = jobRepo.AcquireProviderSlot(
		"lighthouse_runner",
		1,
		"worker-b",
		now.Add(time.Second),
		time.Minute,
	)
	require.NoError(t, err)
	require.Nil(t, blocked)
	require.NotNil(t, retryAt)

	claimed, retryAt, err := jobRepo.AcquireProviderSlot(
		"lighthouse_runner",
		1,
		"worker-b",
		now.Add(6*time.Second),
		time.Minute,
	)
	require.NoError(t, err)
	require.NotNil(t, claimed)
	require.Nil(t, retryAt)
}

func TestSiteQualityTargetSchedulingKeepsCadenceIndependentOfCompletion(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	now := time.Date(2026, time.August, 17, 9, 0, 0, 0, time.UTC)
	interval := 24 * time.Hour

	target, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/support/warranty",
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierCritical,
		SamplingIntervalSeconds: int(interval.Seconds()),
		Enabled:                 true,
	}, now)
	require.NoError(t, err)
	require.NotNil(t, target.NextScheduledAt)
	require.Equal(t, now, *target.NextScheduledAt)

	nextScheduledAt := now.Add(interval)
	require.NoError(t, targetRepo.MarkScheduled(target.ID, now, nextScheduledAt))
	require.NoError(t, targetRepo.MarkCompleted(target.ID, now.Add(5*time.Minute)))

	updated, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            target.CanonicalURL,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierCritical,
		SamplingIntervalSeconds: int(interval.Seconds()),
		Enabled:                 true,
	}, now.Add(10*time.Minute))
	require.NoError(t, err)
	require.NotNil(t, updated.NextScheduledAt)
	require.Equal(t, nextScheduledAt, *updated.NextScheduledAt)
	require.NotNil(t, updated.LastCompletedAt)
	require.Equal(t, now.Add(5*time.Minute), *updated.LastCompletedAt)
}

func TestSiteQualityTargetMatchesRouteEntryBeforeCanonicalURLAndMigratesURL(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	now := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)
	routeEntryID := uint(42)

	original, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		RouteEntryID:            &routeEntryID,
		CanonicalURL:            "https://example.com/legacy",
		Source:                  sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
		SamplingIntervalSeconds: 604800,
		Enabled:                 true,
		LedgerSyncMarker:        "manifest-1",
	}, now)
	require.NoError(t, err)

	migrated, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		RouteEntryID:            &routeEntryID,
		CanonicalURL:            "https://example.com/current",
		Source:                  sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierCritical,
		SamplingIntervalSeconds: 86400,
		Enabled:                 true,
		LedgerSyncMarker:        "manifest-2",
	}, now.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, original.ID, migrated.ID)
	require.Equal(t, "https://example.com/current", migrated.CanonicalURL)
	require.Equal(t, sitequalitydomain.SiteQualityTargetSourceRouteCatalog, migrated.Source)
	require.True(t, migrated.LedgerSynced)
	require.Equal(t, "manifest-2", migrated.LedgerSyncMarker)
}

func TestSiteQualityTargetCanonicalMigrationRejectsAnotherRouteIdentity(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	now := time.Now().UTC()
	firstRouteID := uint(10)
	secondRouteID := uint(11)

	first, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		RouteEntryID:            &firstRouteID,
		CanonicalURL:            "https://example.com/first",
		Source:                  sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
		SamplingIntervalSeconds: 604800,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)
	second, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		RouteEntryID:            &secondRouteID,
		CanonicalURL:            "https://example.com/second",
		Source:                  sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
		SamplingIntervalSeconds: 604800,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)

	_, err = targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		RouteEntryID:            &firstRouteID,
		CanonicalURL:            second.CanonicalURL,
		Source:                  sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierCritical,
		SamplingIntervalSeconds: 86400,
		Enabled:                 true,
	}, now.Add(time.Minute))
	require.ErrorIs(t, err, ErrSiteQualityTargetCanonicalConflict)

	unchangedFirst, err := targetRepo.FindByID(first.ID)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/first", unchangedFirst.CanonicalURL)
	unchangedSecond, err := targetRepo.FindByID(second.ID)
	require.NoError(t, err)
	require.Equal(t, secondRouteID, *unchangedSecond.RouteEntryID)
}

func TestSiteQualityTargetDisablePreservesTargetIdentity(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	routeEntryID := uint(99)
	now := time.Now().UTC()

	target, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		RouteEntryID:            &routeEntryID,
		CanonicalURL:            "https://example.com/retired",
		Source:                  sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
		SamplingIntervalSeconds: 604800,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)
	require.NoError(t, targetRepo.DisableByRouteEntryID(routeEntryID, "route became stale", "manifest-3", now.Add(time.Minute)))

	disabled, err := targetRepo.FindByID(target.ID)
	require.NoError(t, err)
	require.False(t, disabled.Enabled)
	require.Equal(t, target.ID, disabled.ID)
	require.Equal(t, "route became stale", disabled.DisableReason)
	require.Equal(t, "manifest-3", disabled.LedgerSyncMarker)
}

func TestSiteQualityEvaluationScopesRecheckToBoundFinding(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	findingRepo := NewSiteQualityFindingRepository(db)
	now := time.Now().UTC()
	target, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/recheck-scope",
		SamplingIntervalSeconds: 86400,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)

	targetID := target.ID
	first, second := sitequalitydomain.SiteQualityFinding{
		TargetID:        &targetID,
		TargetURL:       target.CanonicalURL,
		Strategy:        sitequalitydomain.SiteQualityStrategyMobile,
		AuditID:         "audit-one",
		Title:           "First finding",
		Severity:        "medium",
		State:           sitequalitydomain.SiteQualityFindingStateResolved,
		FirstDetectedAt: now,
		LastDetectedAt:  now,
		LatestRunID:     1,
		LatestEvidence:  "{}",
	}, sitequalitydomain.SiteQualityFinding{
		TargetID:        &targetID,
		TargetURL:       target.CanonicalURL,
		Strategy:        sitequalitydomain.SiteQualityStrategyMobile,
		AuditID:         "audit-two",
		Title:           "Second finding",
		Severity:        "low",
		State:           sitequalitydomain.SiteQualityFindingStateResolved,
		FirstDetectedAt: now,
		LastDetectedAt:  now,
		LatestRunID:     1,
		LatestEvidence:  "{}",
	}
	require.NoError(t, db.Create(&first).Error)
	require.NoError(t, db.Create(&second).Error)

	findingID := first.ID
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return findingRepo.ApplyEvaluation(tx, sitequalitydomain.SiteQualityFindingEvaluationInput{
			TargetID:                 target.ID,
			FindingID:                &findingID,
			TargetURL:                target.CanonicalURL,
			Strategy:                 sitequalitydomain.SiteQualityStrategyMobile,
			CleanAuditIDs:            []string{"audit-one", "audit-two"},
			RequiredCleanEvaluations: 2,
			LatestRunID:              1,
			DecidedAt:                now.Add(time.Minute),
		})
	}))

	updatedFirst, err := findingRepo.FindByID(first.ID)
	require.NoError(t, err)
	updatedSecond, err := findingRepo.FindByID(second.ID)
	require.NoError(t, err)
	require.Equal(t, 1, updatedFirst.ConsecutiveClean)
	require.Equal(t, 0, updatedSecond.ConsecutiveClean)
}

func TestSiteQualityFindingReadNormalizesHistoricalRuleIdentity(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	now := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	finding := sitequalitydomain.SiteQualityFinding{
		TargetURL:       "https://example.com/blog",
		Strategy:        sitequalitydomain.SiteQualityStrategyMobile,
		AuditID:         "link-text",
		RuleID:          "",
		ProviderAuditID: "",
		FindingKind:     "links",
		Title:           "Links do not have descriptive text",
		Severity:        "medium",
		State:           sitequalitydomain.SiteQualityFindingStateOpen,
		FirstDetectedAt: now,
		LastDetectedAt:  now,
		LatestRunID:     1,
		LatestEvidence:  `{"title":"legacy evidence","rule_id":"link-text","provider_audit_id":"link_descriptive_text"}`,
	}
	require.NoError(t, db.Create(&finding).Error)

	normalized, err := NewSiteQualityFindingRepository(db).FindByID(finding.ID)
	require.NoError(t, err)
	require.Equal(t, sitequalitydomain.SiteQualityRuleIDDescriptiveLinkText, normalized.RuleID)
	require.Equal(t, sitequalitydomain.SiteQualityProviderAuditIDLinkText, normalized.ProviderAuditID)

	var evidence map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(normalized.LatestEvidence), &evidence))
	require.Equal(t, sitequalitydomain.SiteQualityRuleIDDescriptiveLinkText, evidence["rule_id"])
	require.Equal(t, sitequalitydomain.SiteQualityProviderAuditIDLinkText, evidence["provider_audit_id"])
	require.Equal(t, "legacy evidence", evidence["title"])
}

func TestSiteQualityLeaseCompareAndSetRejectsPreviousWorker(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	jobRepo := NewSiteQualityJobRepository(db)
	now := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)

	target, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/lease",
		SamplingIntervalSeconds: 86400,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)
	_, created, err := jobRepo.Enqueue(sitequalitydomain.SiteQualityJob{
		TargetID:              target.ID,
		Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
		Kind:                  sitequalitydomain.SiteQualityJobKindManual,
		IdempotencyKey:        "lease-cas",
		SampleCount:           3,
		RequiredConfirmations: 2,
		MaxAttempts:           4,
		AvailableAt:           now,
	})
	require.NoError(t, err)
	require.True(t, created)

	first, err := jobRepo.ClaimReady(now, "worker-a", 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, first, 1)
	second, err := jobRepo.ClaimReady(now.Add(2*time.Minute), "worker-b", 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, second, 1)
	require.Equal(t, first[0].LeaseGeneration+1, second[0].LeaseGeneration)

	err = jobRepo.MarkSucceeded(first[0].ID, "worker-a", first[0].LeaseGeneration, now.Add(2*time.Minute))
	require.ErrorIs(t, err, ErrSiteQualityLeaseLost)

	current, err := jobRepo.FindByID(first[0].ID)
	require.NoError(t, err)
	require.Equal(t, sitequalitydomain.SiteQualityJobStatusProcessing, current.Status)
	require.Equal(t, "worker-b", current.LockedBy)
	require.Equal(t, second[0].LeaseGeneration, current.LeaseGeneration)
}

func TestSiteQualityLeaseCompareAndSetRejectsExpiredCurrentLease(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	jobRepo := NewSiteQualityJobRepository(db)
	now := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)

	target, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/expired-lease",
		SamplingIntervalSeconds: 86400,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)
	_, created, err := jobRepo.Enqueue(sitequalitydomain.SiteQualityJob{
		TargetID:              target.ID,
		Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
		Kind:                  sitequalitydomain.SiteQualityJobKindManual,
		IdempotencyKey:        "expired-lease-cas",
		SampleCount:           3,
		RequiredConfirmations: 2,
		MaxAttempts:           4,
		AvailableAt:           now,
	})
	require.NoError(t, err)
	require.True(t, created)

	claimed, err := jobRepo.ClaimReady(now, "worker-a", 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, claimed, 1)

	err = jobRepo.MarkSucceeded(claimed[0].ID, "worker-a", claimed[0].LeaseGeneration, now.Add(2*time.Minute))
	require.ErrorIs(t, err, ErrSiteQualityLeaseLost)
	err = jobRepo.MarkFailed(claimed[0], "worker-a", "late failure", now.Add(3*time.Minute), now.Add(2*time.Minute))
	require.ErrorIs(t, err, ErrSiteQualityLeaseLost)

	current, err := jobRepo.FindByID(claimed[0].ID)
	require.NoError(t, err)
	require.Equal(t, sitequalitydomain.SiteQualityJobStatusProcessing, current.Status)
	require.Equal(t, "worker-a", current.LockedBy)
}

func TestSiteQualityStatsExposeScheduledManualAndRecheckJobs(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	targetRepo := NewSiteQualityTargetRepository(db)
	jobRepo := NewSiteQualityJobRepository(db)
	now := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)

	critical, err := targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/critical",
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierCritical,
		SamplingIntervalSeconds: 86400,
		Enabled:                 true,
	}, now.Add(-time.Hour))
	require.NoError(t, err)
	_, err = targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/standard",
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
		SamplingIntervalSeconds: 604800,
		Enabled:                 true,
	}, now.Add(time.Hour))
	require.NoError(t, err)
	_, err = targetRepo.Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            "https://example.com/background",
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierBackground,
		SamplingIntervalSeconds: 2592000,
		Enabled:                 false,
	}, now.Add(-time.Hour))
	require.NoError(t, err)

	targetStats, err := targetRepo.Stats(now)
	require.NoError(t, err)
	require.Equal(t, int64(3), targetStats.Total)
	require.Equal(t, int64(2), targetStats.Enabled)
	require.Equal(t, int64(1), targetStats.Due)
	require.Equal(t, int64(1), targetStats.Critical)
	require.Equal(t, int64(1), targetStats.Standard)
	require.Equal(t, int64(1), targetStats.Background)

	queuedAvailableAt := now.Add(-5 * time.Minute)
	processingLockedAt := now.Add(-3 * time.Minute)
	finishedAt := now.Add(-2 * time.Minute)
	require.NoError(t, db.Create(&sitequalitydomain.SiteQualityJob{
		TargetID:              critical.ID,
		Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
		Kind:                  sitequalitydomain.SiteQualityJobKindScheduled,
		Status:                sitequalitydomain.SiteQualityJobStatusQueued,
		IdempotencyKey:        "stats-queued",
		SampleCount:           3,
		RequiredConfirmations: 2,
		MaxAttempts:           4,
		AvailableAt:           queuedAvailableAt,
	}).Error)
	require.NoError(t, db.Create(&sitequalitydomain.SiteQualityJob{
		TargetID:              critical.ID,
		Strategy:              sitequalitydomain.SiteQualityStrategyDesktop,
		Kind:                  sitequalitydomain.SiteQualityJobKindRecheck,
		Status:                sitequalitydomain.SiteQualityJobStatusProcessing,
		IdempotencyKey:        "stats-processing",
		SampleCount:           3,
		RequiredConfirmations: 2,
		Attempts:              1,
		MaxAttempts:           4,
		AvailableAt:           now.Add(-time.Hour),
		LockedAt:              &processingLockedAt,
		LockedBy:              "worker-a",
	}).Error)
	require.NoError(t, db.Create(&sitequalitydomain.SiteQualityJob{
		TargetID:              critical.ID,
		Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
		Kind:                  sitequalitydomain.SiteQualityJobKindManual,
		Status:                sitequalitydomain.SiteQualityJobStatusSucceeded,
		IdempotencyKey:        "stats-succeeded",
		SampleCount:           3,
		RequiredConfirmations: 2,
		MaxAttempts:           4,
		AvailableAt:           now.Add(-time.Hour),
		FinishedAt:            &finishedAt,
	}).Error)

	stats, err := jobRepo.Stats(now, time.Minute)
	require.NoError(t, err)
	require.Equal(t, int64(3), stats.Total)
	require.Equal(t, int64(1), stats.Queued)
	require.Equal(t, int64(1), stats.Processing)
	require.Equal(t, int64(1), stats.Succeeded)
	require.Equal(t, int64(2), stats.Claimable)
	require.Equal(t, int64(1), stats.StaleLeases)
	require.NotNil(t, stats.OldestQueuedAt)
	require.Equal(t, queuedAvailableAt, *stats.OldestQueuedAt)
	require.NotNil(t, stats.OldestProcessingAt)
	require.Equal(t, processingLockedAt, *stats.OldestProcessingAt)
	require.NotNil(t, stats.LatestSuccessAt)
	require.Equal(t, finishedAt, *stats.LatestSuccessAt)
}

func TestSiteQualityProviderSlotStatsIncludeStaleAndUninitializedCapacity(t *testing.T) {
	db := newSiteQualityRepositoryTestDB(t)
	jobRepo := NewSiteQualityJobRepository(db)
	now := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)
	staleLockedAt := now.Add(-3 * time.Minute)
	freshLockedAt := now.Add(-10 * time.Second)

	require.NoError(t, db.Create(&sitequalitydomain.SiteQualityProviderSlot{
		Provider:    "lighthouse_runner",
		Slot:        1,
		AvailableAt: now.Add(-time.Minute),
		LockedAt:    &staleLockedAt,
		LockedBy:    "worker-a",
	}).Error)
	require.NoError(t, db.Create(&sitequalitydomain.SiteQualityProviderSlot{
		Provider:    "lighthouse_runner",
		Slot:        2,
		AvailableAt: now,
		LockedAt:    &freshLockedAt,
		LockedBy:    "worker-b",
	}).Error)

	stats, err := jobRepo.ProviderSlotStats("lighthouse_runner", 3, now, time.Minute)
	require.NoError(t, err)
	require.Equal(t, "lighthouse_runner", stats.Provider)
	require.Equal(t, 3, stats.Configured)
	require.Equal(t, int64(2), stats.Total)
	require.Equal(t, int64(2), stats.Locked)
	require.Equal(t, int64(1), stats.StaleLocked)
	require.Equal(t, int64(2), stats.Available)
	require.NotNil(t, stats.NextAvailableAt)
	require.Equal(t, now, *stats.NextAvailableAt)
}

func newSiteQualityRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&sitequalitydomain.SiteQualityTarget{},
		&sitequalitydomain.SiteQualityJob{},
		&sitequalitydomain.SiteQualityProviderSlot{},
		&sitequalitydomain.SiteQualityEvaluation{},
		&sitequalitydomain.SiteQualityRun{},
		&sitequalitydomain.SiteQualityFinding{},
		&sitequalitydomain.SiteQualityFindingEvent{},
	))
	return db
}
