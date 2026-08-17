package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/repository"
)

const (
	defaultSiteQualitySampleCount           = 3
	defaultSiteQualityRequiredConfirmations = 2
	defaultSiteQualityRequiredCleanRuns     = 2
	defaultSiteQualityLeaseTimeout          = 15 * time.Minute
	defaultSiteQualityWorkerBatchLimit      = 2
	defaultSiteQualityProviderInterval      = 5 * time.Second
)

type SiteQualityEngineConfig struct {
	BaseURL                  string
	WorkerEnabled            bool
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
}

type SiteQualityProcessResult struct {
	Claimed    int    `json:"claimed"`
	Succeeded  int    `json:"succeeded"`
	Failed     int    `json:"failed"`
	DeadLetter int    `json:"dead_letter"`
	WorkerID   string `json:"worker_id"`
}

type SiteQualityOperationalSummary struct {
	GeneratedAt                    time.Time                                 `json:"generated_at"`
	Status                         string                                    `json:"status"`
	Warnings                       []string                                  `json:"warnings,omitempty"`
	WorkerEnabled                  bool                                      `json:"worker_enabled"`
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
	AuditID        string                                                `json:"audit_id"`
	Kind           string                                                `json:"kind"`
	RuleVersion    string                                                `json:"rule_version"`
	Title          string                                                `json:"title"`
	Description    string                                                `json:"description,omitempty"`
	Severity       string                                                `json:"severity"`
	Confirmations  int                                                   `json:"confirmations"`
	SampleCount    int                                                   `json:"sample_count"`
	Confidence     float64                                               `json:"confidence"`
	MedianScore    *float64                                              `json:"median_score,omitempty"`
	MedianMS       *float64                                              `json:"median_savings_ms,omitempty"`
	MedianBytes    *int64                                                `json:"median_savings_bytes,omitempty"`
	Resources      []LighthouseRunnerResource                            `json:"resources,omitempty"`
	Headings       []sitequalitydomain.SiteQualityHeadingEvidence        `json:"headings,omitempty"`
	StructuredData []sitequalitydomain.SiteQualityStructuredDataEvidence `json:"structured_data,omitempty"`
	DisplayValue   string                                                `json:"display_value,omitempty"`
	NumericValue   *float64                                              `json:"numeric_value,omitempty"`
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
	}
}
