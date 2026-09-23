package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

func (s *SiteQualityEngineService) GetJob(id uint) (*sitequalitydomain.SiteQualityJob, error) {
	if s == nil || s.jobs == nil {
		return nil, errors.New("SiteQuality quality engine is unavailable")
	}
	job, err := s.jobs.FindByID(id)
	if err != nil {
		return nil, err
	}
	if s.runs != nil {
		latestRunID, latestErr := s.runs.LatestIDForJob(id)
		if latestErr != nil {
			return nil, latestErr
		}
		job.LatestRunID = latestRunID
	}
	return job, nil
}

func (s *SiteQualityEngineService) CancelJob(id uint) (*sitequalitydomain.SiteQualityJob, error) {
	if s == nil || s.jobs == nil {
		return nil, errors.New("SiteQuality quality engine is unavailable")
	}
	job, changed, err := s.jobs.Cancel(id, time.Now().UTC(), "cancelled by operator")
	if err != nil {
		return nil, err
	}
	if !changed && job.Status != sitequalitydomain.SiteQualityJobStatusCancelled {
		return nil, fmt.Errorf("%w: current status is %s", ErrSiteQualityJobNotCancellable, job.Status)
	}
	s.cancelActiveJob(id)
	s.notifyWorker()
	return s.GetJob(id)
}

func (s *SiteQualityEngineService) EnqueueManualTarget(
	ctx context.Context,
	targetURL string,
	strategy string,
	actorUserID uint,
	kind string,
	auditScopes ...string,
) (*sitequalitydomain.SiteQualityJob, error) {
	if s == nil || s.targets == nil || s.jobs == nil || s.lighthouseRunner == nil {
		return nil, errors.New("site quality engine is unavailable")
	}
	if kind == "" {
		kind = sitequalitydomain.SiteQualityJobKindManual
	}
	auditScope := sitequalitydomain.SiteQualityAuditScopeFull
	if len(auditScopes) > 0 {
		auditScope = strings.TrimSpace(auditScopes[0])
	}
	if auditScope == "" {
		auditScope = sitequalitydomain.SiteQualityAuditScopeFull
	}
	switch auditScope {
	case sitequalitydomain.SiteQualityAuditScopeFull, sitequalitydomain.SiteQualityAuditScopeHeadings, sitequalitydomain.SiteQualityAuditScopeSchema, sitequalitydomain.SiteQualityAuditScopeLinkText:
	default:
		return nil, fmt.Errorf("%w: audit_scope must be full, headings, schema, or link_text", ErrInvalidSiteQualityRun)
	}
	normalizedURL, normalizedStrategy, err := s.lighthouseRunner.normalizeRunInput(LighthouseRunnerRunInput{
		URL:      targetURL,
		Strategy: strategy,
	})
	if err != nil {
		return nil, err
	}
	normalizedURL, err = canonicalizeAbsoluteSiteQualityURL(normalizedURL)
	if err != nil {
		return nil, err
	}
	_ = ctx
	now := time.Now().UTC()
	target, err := s.targets.FindByCanonicalURL(normalizedURL)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		target, err = s.targets.Upsert(sitequalitydomain.SiteQualityTargetInput{
			CanonicalURL:            normalizedURL,
			Source:                  sitequalitydomain.SiteQualityTargetSourceOperator,
			SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
			SamplingIntervalSeconds: int((7 * 24 * time.Hour).Seconds()),
			Enabled:                 true,
		}, now)
	}
	if err != nil {
		return nil, err
	}
	if !target.Enabled {
		return nil, fmt.Errorf("%w: Site Quality target is disabled", ErrInvalidSiteQualityRun)
	}
	return s.enqueuePriorityJob(*target, normalizedStrategy, actorUserID, kind, nil, auditScope, now)
}

func (s *SiteQualityEngineService) EnqueueRecheckFinding(
	finding *sitequalitydomain.SiteQualityFinding,
	actorUserID uint,
) (*sitequalitydomain.SiteQualityJob, error) {
	if finding == nil {
		return nil, errors.New("SiteQuality finding is required")
	}
	if finding.ID == 0 {
		return nil, errors.New("SiteQuality finding ID is required")
	}
	if finding.TargetID != nil && *finding.TargetID != 0 {
		target, err := s.targets.FindByID(*finding.TargetID)
		if err != nil {
			return nil, err
		}
		if !target.Enabled {
			return nil, fmt.Errorf("%w: Site Quality target is disabled", ErrInvalidSiteQualityRun)
		}
		findingID := finding.ID
		return s.enqueuePriorityJob(*target, finding.Strategy, actorUserID, sitequalitydomain.SiteQualityJobKindRecheck, &findingID, sitequalitydomain.SiteQualityAuditScopeFull, time.Now().UTC())
	}
	return nil, errors.New("SiteQuality finding is not bound to a target")
}

func (s *SiteQualityEngineService) ProcessReady(ctx context.Context, now time.Time, limit int) (SiteQualityProcessResult, error) {
	var result SiteQualityProcessResult
	if s == nil || s.jobs == nil {
		return result, errors.New("SiteQuality quality engine is unavailable")
	}
	result.WorkerID = s.workerID
	if ctx == nil {
		ctx = context.Background()
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if limit <= 0 {
		limit = s.cfg.WorkerBatchLimit
	}

	var (
		jobs []sitequalitydomain.SiteQualityJob
		err  error
	)
	if s.cfg.AutoScanEnabled {
		jobs, err = s.jobs.ClaimReady(now, s.workerID, limit, s.cfg.LeaseTimeout)
	} else {
		jobs, err = s.jobs.ClaimReadyInteractive(now, s.workerID, limit, s.cfg.LeaseTimeout)
	}
	if err != nil {
		return result, err
	}
	result.Claimed = len(jobs)
	for _, job := range jobs {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if err := s.processJob(ctx, job); err != nil {
			if errors.Is(err, ErrSiteQualityJobCancelled) {
				result.Cancelled++
				continue
			}
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			failedAt := time.Now().UTC()
			next := siteQualityRetryAt(failedAt, job.Attempts)
			if errors.Is(err, repository.ErrSiteQualityLeaseLost) {
				if s.jobWasCancelled(job.ID) {
					result.Cancelled++
				}
				continue
			}
			if markErr := s.jobs.MarkFailed(job, s.workerID, err.Error(), next, failedAt); markErr != nil {
				if errors.Is(markErr, repository.ErrSiteQualityLeaseLost) {
					continue
				}
				return result, markErr
			}
			if job.Attempts >= job.MaxAttempts {
				result.DeadLetter++
			} else {
				result.Failed++
			}
			continue
		}
		result.Succeeded++
	}
	return result, nil
}

func (s *SiteQualityEngineService) processJob(
	ctx context.Context,
	job sitequalitydomain.SiteQualityJob,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	jobCtx, cancel := context.WithCancel(ctx)
	s.registerActiveJob(job.ID, cancel)
	defer s.unregisterActiveJob(job.ID)
	defer cancel()

	if err := s.jobs.Heartbeat(job.ID, s.workerID, job.LeaseGeneration, time.Now().UTC(), s.cfg.LeaseTimeout); err != nil {
		if s.jobWasCancelled(job.ID) {
			return ErrSiteQualityJobCancelled
		}
		return err
	}
	if err := s.jobs.UpdateProgress(
		job.ID,
		s.workerID,
		job.LeaseGeneration,
		0,
		sitequalitydomain.SiteQualityJobProgressStageStarting,
		time.Now().UTC(),
		s.cfg.LeaseTimeout,
	); err != nil {
		if s.jobWasCancelled(job.ID) {
			return ErrSiteQualityJobCancelled
		}
		return err
	}
	target, err := s.targets.FindByID(job.TargetID)
	if err != nil {
		return err
	}
	runViews := make([]LighthouseRunnerRunView, 0, job.SampleCount)
	for sample := 0; sample < job.SampleCount; sample++ {
		if err := jobCtx.Err(); err != nil {
			if s.jobWasCancelled(job.ID) {
				return ErrSiteQualityJobCancelled
			}
			return err
		}
		if err := s.jobs.UpdateProgress(
			job.ID,
			s.workerID,
			job.LeaseGeneration,
			sample,
			sitequalitydomain.SiteQualityJobProgressStageWaitingForSlot,
			time.Now().UTC(),
			s.cfg.LeaseTimeout,
		); err != nil {
			if s.jobWasCancelled(job.ID) {
				return ErrSiteQualityJobCancelled
			}
			return err
		}
		run, err := s.captureWithProviderSlot(jobCtx, LighthouseRunnerCaptureInput{
			LighthouseRunnerRunInput: LighthouseRunnerRunInput{
				URL:               target.CanonicalURL,
				Strategy:          job.Strategy,
				InitiatedByUserID: job.InitiatedByUserID,
			},
			TargetID:         &target.ID,
			JobID:            &job.ID,
			LeaseWorkerID:    s.workerID,
			LeaseGeneration:  job.LeaseGeneration,
			CanonicalURL:     target.CanonicalURL,
			ReleaseID:        job.ReleaseID,
			TargetSource:     target.Source,
			TargetSourceType: target.SourceType,
			TargetTitle:      target.Title,
			TargetLocale:     target.Locale,
			AuditScope:       job.AuditScope,
		}, func() error {
			return s.jobs.UpdateProgress(
				job.ID,
				s.workerID,
				job.LeaseGeneration,
				sample,
				sitequalitydomain.SiteQualityJobProgressStageCapturing,
				time.Now().UTC(),
				s.cfg.LeaseTimeout,
			)
		})
		if run != nil {
			runViews = append(runViews, *run)
			if err := s.jobs.UpdateProgress(
				job.ID,
				s.workerID,
				job.LeaseGeneration,
				sample+1,
				sitequalitydomain.SiteQualityJobProgressStageCapturing,
				time.Now().UTC(),
				s.cfg.LeaseTimeout,
			); err != nil {
				if s.jobWasCancelled(job.ID) {
					return ErrSiteQualityJobCancelled
				}
				return err
			}
		}
		if err != nil {
			if s.jobWasCancelled(job.ID) {
				return ErrSiteQualityJobCancelled
			}
			return err
		}
	}
	if err := s.jobs.UpdateProgress(
		job.ID,
		s.workerID,
		job.LeaseGeneration,
		len(runViews),
		sitequalitydomain.SiteQualityJobProgressStageEvaluating,
		time.Now().UTC(),
		s.cfg.LeaseTimeout,
	); err != nil {
		if s.jobWasCancelled(job.ID) {
			return ErrSiteQualityJobCancelled
		}
		return err
	}
	if err := s.applyJobEvaluation(job, *target, runViews); err != nil {
		if s.jobWasCancelled(job.ID) {
			return ErrSiteQualityJobCancelled
		}
		return err
	}
	return nil
}

func (s *SiteQualityEngineService) captureWithProviderSlot(
	ctx context.Context,
	input LighthouseRunnerCaptureInput,
	onSlotAcquired func() error,
) (*LighthouseRunnerRunView, error) {
	slot, err := s.waitForProviderSlot(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = s.jobs.ReleaseProviderSlot(
			slot.Provider,
			slot.Slot,
			s.workerID,
			time.Now().UTC(),
			s.cfg.ProviderRequestInterval,
		)
	}()
	if onSlotAcquired != nil {
		if err := onSlotAcquired(); err != nil {
			return nil, err
		}
	}
	if s.lighthouseRunner == nil {
		return nil, errors.New("site quality runner is unavailable")
	}
	return s.lighthouseRunner.Capture(ctx, input)
}

func (s *SiteQualityEngineService) waitForProviderSlot(
	ctx context.Context,
) (*sitequalitydomain.SiteQualityProviderSlot, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		now := time.Now().UTC()
		slot, retryAt, err := s.jobs.AcquireProviderSlot(
			"lighthouse_runner",
			s.cfg.ProviderConcurrency,
			s.workerID,
			now,
			2*time.Minute,
		)
		if err != nil || slot != nil {
			return slot, err
		}
		wait := 5 * time.Second
		if retryAt != nil && retryAt.After(now) {
			wait = retryAt.Sub(now)
		}
		if wait > 30*time.Second {
			wait = 30 * time.Second
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
}

func (s *SiteQualityEngineService) enqueuePriorityJob(
	target sitequalitydomain.SiteQualityTarget,
	strategy string,
	actorUserID uint,
	kind string,
	findingID *uint,
	auditScope string,
	now time.Time,
) (*sitequalitydomain.SiteQualityJob, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if strings.TrimSpace(auditScope) == "" {
		auditScope = sitequalitydomain.SiteQualityAuditScopeFull
	}
	sampleCount := s.cfg.SampleCount
	requiredConfirmations := s.cfg.RequiredConfirmations
	if auditScope != sitequalitydomain.SiteQualityAuditScopeFull {
		sampleCount = 1
		requiredConfirmations = 1
	}
	job := sitequalitydomain.SiteQualityJob{
		TargetID:              target.ID,
		FindingID:             findingID,
		Strategy:              strategy,
		AuditScope:            auditScope,
		Kind:                  kind,
		Status:                sitequalitydomain.SiteQualityJobStatusQueued,
		IdempotencyKey:        siteQualityManualJobKey(target.ID, strategy, kind, findingID, now),
		SampleCount:           sampleCount,
		RequiredConfirmations: requiredConfirmations,
		MaxAttempts:           s.cfg.MaxAttempts,
		AvailableAt:           now,
		InitiatedByUserID:     actorUserID,
		ReleaseID:             s.cfg.ReleaseID,
	}
	createdJob, _, err := s.jobs.Enqueue(job)
	if err == nil {
		s.notifyWorker()
	}
	return createdJob, err
}

func siteQualityManualJobKey(targetID uint, strategy string, kind string, findingID *uint, now time.Time) string {
	boundFinding := uint(0)
	if findingID != nil {
		boundFinding = *findingID
	}
	sum := sha1.Sum([]byte(fmt.Sprintf("%d:%s:%s:%d:%d", targetID, strategy, kind, boundFinding, now.UnixNano())))
	return fmt.Sprintf("site-quality:%s:%s", kind, hex.EncodeToString(sum[:]))
}

func siteQualityRetryAt(now time.Time, attempts int) time.Time {
	if attempts < 1 {
		attempts = 1
	}
	factor := 1 << uint(min(attempts, 6))
	delay := time.Duration(factor) * 30 * time.Second
	if delay > 15*time.Minute {
		delay = 15 * time.Minute
	}
	jitter := time.Duration(rand.Int63n(int64(15 * time.Second)))
	return now.Add(delay + jitter)
}
