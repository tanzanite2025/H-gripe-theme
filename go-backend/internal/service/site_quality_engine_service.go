package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/repository"
)

var ErrSiteQualityJobCancelled = errors.New("site quality job was cancelled")
var ErrSiteQualityJobNotCancellable = errors.New("site quality job is no longer cancellable")

const (
	defaultSiteQualitySampleCount           = 3
	defaultSiteQualityRequiredConfirmations = 2
	defaultSiteQualityRequiredCleanRuns     = 2
	defaultSiteQualityLeaseTimeout          = 15 * time.Minute
	defaultSiteQualityWorkerBatchLimit      = 1
	defaultSiteQualityProviderInterval      = 5 * time.Second
)

type SiteQualityEngineConfig struct {
	BaseURL                  string
	WorkerEnabled            bool
	AutoScanEnabled          bool
	WorkerInterval           time.Duration
	SampleCount              int
	RequiredConfirmations    int
	RequiredCleanEvaluations int
	WorkerBatchLimit         int
	LeaseTimeout             time.Duration
	ProviderConcurrency      int
	ProviderRequestInterval  time.Duration
	MaxAttempts              int
	ReleaseID                string
}

type SiteQualityEngineService struct {
	targets          *repository.SiteQualityTargetRepository
	jobs             *repository.SiteQualityJobRepository
	runs             *repository.SiteQualityRunRepository
	findings         *repository.SiteQualityFindingRepository
	routeCatalog     *repository.StorefrontRouteCatalogRepository
	lighthouseRunner *LighthouseRunnerService
	cfg              SiteQualityEngineConfig
	workerID         string
	wake             chan struct{}
	activeMu         sync.Mutex
	activeCancels    map[uint]context.CancelFunc
}

type SiteQualityProcessResult struct {
	Claimed    int    `json:"claimed"`
	Succeeded  int    `json:"succeeded"`
	Failed     int    `json:"failed"`
	DeadLetter int    `json:"dead_letter"`
	Cancelled  int    `json:"cancelled"`
	WorkerID   string `json:"worker_id"`
}

type SiteQualityOperationalSummary struct {
	GeneratedAt                    time.Time                                 `json:"generated_at"`
	Status                         string                                    `json:"status"`
	Warnings                       []string                                  `json:"warnings,omitempty"`
	WorkerEnabled                  bool                                      `json:"worker_enabled"`
	AutoScanEnabled                bool                                      `json:"auto_scan_enabled"`
	WorkerIntervalSeconds          int                                       `json:"worker_interval_seconds"`
	RunnerConfigured               bool                                      `json:"runner_configured"`
	DefaultURL                     string                                    `json:"default_url,omitempty"`
	ReleaseID                      string                                    `json:"release_id,omitempty"`
	SampleCount                    int                                       `json:"sample_count"`
	RequiredConfirmations          int                                       `json:"required_confirmations"`
	RequiredCleanEvaluations       int                                       `json:"required_clean_evaluations"`
	WorkerBatchLimit               int                                       `json:"worker_batch_limit"`
	LeaseTimeoutSeconds            int                                       `json:"lease_timeout_seconds"`
	ProviderConcurrency            int                                       `json:"provider_concurrency"`
	ProviderRequestIntervalSeconds int                                       `json:"provider_request_interval_seconds"`
	RunCount                       int64                                     `json:"run_count"`
	LatestRun                      *LighthouseRunnerRunView                  `json:"latest_run,omitempty"`
	LatestSuccessAt                *time.Time                                `json:"latest_success_at,omitempty"`
	Targets                        repository.SiteQualityTargetStats         `json:"targets"`
	Jobs                           repository.SiteQualityJobStats            `json:"jobs"`
	ProviderSlots                  repository.SiteQualityProviderSlotStats   `json:"provider_slots"`
	Findings                       sitequalitydomain.SiteQualityFindingStats `json:"findings"`
}

type siteQualityDecision struct {
	AuditID         string                                                `json:"audit_id"`
	RuleID          string                                                `json:"rule_id"`
	ProviderAuditID string                                                `json:"provider_audit_id,omitempty"`
	Kind            string                                                `json:"kind"`
	RuleVersion     string                                                `json:"rule_version"`
	Title           string                                                `json:"title"`
	Description     string                                                `json:"description,omitempty"`
	Severity        string                                                `json:"severity"`
	Confirmations   int                                                   `json:"confirmations"`
	SampleCount     int                                                   `json:"sample_count"`
	Confidence      float64                                               `json:"confidence"`
	MedianScore     *float64                                              `json:"median_score,omitempty"`
	MedianMS        *float64                                              `json:"median_savings_ms,omitempty"`
	MedianBytes     *int64                                                `json:"median_savings_bytes,omitempty"`
	Resources       []LighthouseRunnerResource                            `json:"resources,omitempty"`
	Links           []sitequalitydomain.SiteQualityLinkEvidence           `json:"links,omitempty"`
	Headings        []sitequalitydomain.SiteQualityHeadingEvidence        `json:"headings,omitempty"`
	StructuredData  []sitequalitydomain.SiteQualityStructuredDataEvidence `json:"structured_data,omitempty"`
	Runtime         []sitequalitydomain.SiteQualityRuntimeEvidence        `json:"runtime,omitempty"`
	DisplayValue    string                                                `json:"display_value,omitempty"`
	NumericValue    *float64                                              `json:"numeric_value,omitempty"`
}

type siteQualityEvaluationDecision struct {
	Confirmed []siteQualityDecision `json:"confirmed"`
	Clean     []string              `json:"clean"`
	Observed  []string              `json:"observed"`
	Runs      []uint                `json:"runs"`
}

func NewSiteQualityEngineService(
	targets *repository.SiteQualityTargetRepository,
	jobs *repository.SiteQualityJobRepository,
	runs *repository.SiteQualityRunRepository,
	findings *repository.SiteQualityFindingRepository,
	routeCatalog *repository.StorefrontRouteCatalogRepository,
	lighthouseRunner *LighthouseRunnerService,
	cfg SiteQualityEngineConfig,
) *SiteQualityEngineService {
	if cfg.SampleCount <= 0 {
		cfg.SampleCount = defaultSiteQualitySampleCount
	}
	if cfg.RequiredConfirmations <= 0 || cfg.RequiredConfirmations > cfg.SampleCount {
		cfg.RequiredConfirmations = defaultSiteQualityRequiredConfirmations
	}
	if cfg.RequiredConfirmations > cfg.SampleCount {
		cfg.RequiredConfirmations = cfg.SampleCount
	}
	if cfg.RequiredCleanEvaluations <= 0 {
		cfg.RequiredCleanEvaluations = defaultSiteQualityRequiredCleanRuns
	}
	if cfg.WorkerBatchLimit <= 0 {
		cfg.WorkerBatchLimit = defaultSiteQualityWorkerBatchLimit
	}
	if cfg.LeaseTimeout <= 0 {
		cfg.LeaseTimeout = defaultSiteQualityLeaseTimeout
	}
	if cfg.WorkerInterval <= 0 {
		cfg.WorkerInterval = 30 * time.Second
	}
	if cfg.ProviderRequestInterval <= 0 {
		cfg.ProviderRequestInterval = defaultSiteQualityProviderInterval
	}
	if cfg.ProviderConcurrency <= 0 {
		cfg.ProviderConcurrency = 1
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 4
	}
	if cfg.ReleaseID == "" {
		cfg.ReleaseID = strings.TrimSpace(os.Getenv("APP_RELEASE_ID"))
	}
	if lighthouseRunner != nil {
		lighthouseRunner.ConfigureJobRepository(jobs)
	}

	return &SiteQualityEngineService{
		targets:          targets,
		jobs:             jobs,
		runs:             runs,
		findings:         findings,
		routeCatalog:     routeCatalog,
		lighthouseRunner: lighthouseRunner,
		cfg:              cfg,
		workerID:         fmt.Sprintf("site-quality-%d-%d", os.Getpid(), time.Now().UnixNano()),
		wake:             make(chan struct{}, 1),
		activeCancels:    make(map[uint]context.CancelFunc),
	}
}

func (s *SiteQualityEngineService) WorkerWakeChannel() <-chan struct{} {
	if s == nil {
		return nil
	}
	return s.wake
}

func (s *SiteQualityEngineService) notifyWorker() {
	if s == nil || s.wake == nil {
		return
	}
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *SiteQualityEngineService) registerActiveJob(jobID uint, cancel context.CancelFunc) {
	if s == nil || jobID == 0 || cancel == nil {
		return
	}
	s.activeMu.Lock()
	if s.activeCancels == nil {
		s.activeCancels = make(map[uint]context.CancelFunc)
	}
	s.activeCancels[jobID] = cancel
	s.activeMu.Unlock()
}

func (s *SiteQualityEngineService) unregisterActiveJob(jobID uint) {
	if s == nil || jobID == 0 {
		return
	}
	s.activeMu.Lock()
	delete(s.activeCancels, jobID)
	s.activeMu.Unlock()
}

func (s *SiteQualityEngineService) cancelActiveJob(jobID uint) {
	if s == nil || jobID == 0 {
		return
	}
	s.activeMu.Lock()
	cancel := s.activeCancels[jobID]
	s.activeMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *SiteQualityEngineService) jobWasCancelled(jobID uint) bool {
	if s == nil || s.jobs == nil || jobID == 0 {
		return false
	}
	job, err := s.jobs.FindByID(jobID)
	return err == nil && job.Status == sitequalitydomain.SiteQualityJobStatusCancelled
}
