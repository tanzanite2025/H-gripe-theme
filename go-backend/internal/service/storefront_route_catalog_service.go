package service

import (
	"errors"
	"net/http"
	"strings"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/repository"
)

type StorefrontRouteCatalogService struct {
	repository     *repository.StorefrontRouteCatalogRepository
	postService    *PostService
	productService *ProductService
	baseURL        string
	httpClient     *http.Client
}

func NewStorefrontRouteCatalogService(
	routeRepository *repository.StorefrontRouteCatalogRepository,
	postService *PostService,
	productService *ProductService,
	baseURL string,
) *StorefrontRouteCatalogService {
	return &StorefrontRouteCatalogService{
		repository:     routeRepository,
		postService:    postService,
		productService: productService,
		baseURL:        strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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
	OK            int `json:"ok"`
	Redirects     int `json:"redirects"`
	NotFound      int `json:"not_found"`
	ServerErrors  int `json:"server_errors"`
	CanonicalMiss int `json:"canonical_mismatch"`
	Errors        int `json:"errors"`
}

func (s *StorefrontRouteCatalogService) List(filter repository.StorefrontRouteCatalogListFilter) ([]seodomain.StorefrontRouteCatalogEntry, int64, error) {
	if s == nil || s.repository == nil {
		return nil, 0, errors.New("storefront route catalog service is unavailable")
	}
	return s.repository.List(filter)
}

func (s *StorefrontRouteCatalogService) Stats() (seodomain.StorefrontRouteCatalogStats, error) {
	if s == nil || s.repository == nil {
		return seodomain.StorefrontRouteCatalogStats{}, errors.New("storefront route catalog service is unavailable")
	}
	return s.repository.Stats()
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
