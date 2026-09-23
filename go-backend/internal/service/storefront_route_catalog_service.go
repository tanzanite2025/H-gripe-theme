package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/repository"
)

type StorefrontRouteCatalogService struct {
	repository      *repository.StorefrontRouteCatalogRepository
	postService     *PostService
	productService  *ProductService
	baseURL         string
	internalBaseURL string
	httpClient      *http.Client
	issueReconciler storefrontRouteCatalogIssueReconciler
	tasksMu         sync.RWMutex
	checkTasks      map[string]*storefrontRouteCatalogCheckTask
}

type storefrontRouteCatalogIssueReconciler interface {
	ReconcileCatalog(ctx context.Context) error
	ReconcileEntry(ctx context.Context, routeEntryID uint, latestCheckResultID *uint) error
}

func NewStorefrontRouteCatalogService(
	routeRepository *repository.StorefrontRouteCatalogRepository,
	postService *PostService,
	productService *ProductService,
	baseURL string,
	internalBaseURL string,
) *StorefrontRouteCatalogService {
	publicOrigin := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	privateOrigin := strings.TrimRight(strings.TrimSpace(internalBaseURL), "/")

	return &StorefrontRouteCatalogService{
		repository:      routeRepository,
		postService:     postService,
		productService:  productService,
		baseURL:         publicOrigin,
		internalBaseURL: privateOrigin,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		checkTasks: make(map[string]*storefrontRouteCatalogCheckTask),
	}
}

func (s *StorefrontRouteCatalogService) ConfigureIssueReconciler(
	reconciler storefrontRouteCatalogIssueReconciler,
) {
	if s == nil {
		return
	}
	s.issueReconciler = reconciler
}

type StorefrontRouteCatalogSyncSummary struct {
	ManifestVersion string `json:"manifest_version"`
	Entries         int    `json:"entries"`
	StaticEntries   int    `json:"static_entries"`
	ProductEntries  int    `json:"product_entries"`
	BlogEntries     int    `json:"blog_entries"`
	AliasEntries    int    `json:"alias_entries"`
	Duplicates      int    `json:"duplicates"`
}

type StorefrontRouteCatalogCheckSummary struct {
	Checked       int `json:"checked"`
	Eligible      int `json:"eligible"`
	Remaining     int `json:"remaining"`
	OK            int `json:"ok"`
	Redirects     int `json:"redirects"`
	NotFound      int `json:"not_found"`
	ServerErrors  int `json:"server_errors"`
	CanonicalMiss int `json:"canonical_mismatch"`
	Errors        int `json:"errors"`
}

const (
	StorefrontRouteCatalogCheckTaskQueued    = "queued"
	StorefrontRouteCatalogCheckTaskRunning   = "running"
	StorefrontRouteCatalogCheckTaskCompleted = "completed"
	StorefrontRouteCatalogCheckTaskFailed    = "failed"
)

// StorefrontRouteCatalogCheckTask is the process-local task projection used
// by the admin polling API. Check evidence remains persisted in the catalog
// repository, so a task restart cannot rewrite route history.
type StorefrontRouteCatalogCheckTask struct {
	ID        string                             `json:"task_id"`
	Status    string                             `json:"status"`
	StartedAt time.Time                          `json:"started_at"`
	UpdatedAt time.Time                          `json:"updated_at"`
	EndedAt   *time.Time                         `json:"ended_at,omitempty"`
	Locale    string                             `json:"locale,omitempty"`
	Checked   int                                `json:"checked"`
	Eligible  int                                `json:"eligible"`
	Remaining int                                `json:"remaining"`
	Summary   StorefrontRouteCatalogCheckSummary `json:"summary"`
	Error     string                             `json:"error,omitempty"`
}

type storefrontRouteCatalogCheckTask struct {
	mu   sync.RWMutex
	data StorefrontRouteCatalogCheckTask
}

func (s *StorefrontRouteCatalogService) List(filter repository.StorefrontRouteCatalogListFilter) ([]seodomain.StorefrontRouteCatalogEntry, int64, error) {
	if s == nil || s.repository == nil {
		return nil, 0, errors.New("storefront route catalog service is unavailable")
	}
	return s.repository.List(filter)
}

func (s *StorefrontRouteCatalogService) Stats() (seodomain.StorefrontRouteCatalogStats, error) {
	return s.StatsForLocale("")
}

func (s *StorefrontRouteCatalogService) StatsForLocale(locale string) (seodomain.StorefrontRouteCatalogStats, error) {
	return s.StatsForLocaleAndScope(locale, "")
}

func (s *StorefrontRouteCatalogService) StatsForLocaleAndScope(locale, problemScope string) (seodomain.StorefrontRouteCatalogStats, error) {
	if s == nil || s.repository == nil {
		return seodomain.StorefrontRouteCatalogStats{}, errors.New("storefront route catalog service is unavailable")
	}
	return s.repository.StatsForLocaleAndScope(locale, problemScope)
}

func (s *StorefrontRouteCatalogService) Get(id uint) (*seodomain.StorefrontRouteCatalogEntry, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("storefront route catalog service is unavailable")
	}
	if id == 0 {
		return nil, errors.New("route entry ID is required")
	}
	return s.repository.FindByID(id)
}

func (s *StorefrontRouteCatalogService) ListChecks(id uint, page, pageSize int) ([]seodomain.StorefrontRouteCheckResult, int64, error) {
	if s == nil || s.repository == nil {
		return nil, 0, errors.New("storefront route catalog service is unavailable")
	}
	if id == 0 {
		return nil, 0, errors.New("route entry ID is required")
	}
	if _, err := s.repository.FindByID(id); err != nil {
		return nil, 0, err
	}
	return s.repository.ListChecks(id, page, pageSize)
}
