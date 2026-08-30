package service

import (
	"errors"
	"strings"
	"time"

	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/domain/seo"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const storefrontURLSearchPublicPageSize = 200

type StorefrontURLSearchProfileService struct {
	profiles *repository.StorefrontURLSearchProfileRepository
	catalog  *repository.StorefrontRouteCatalogRepository
}

func NewStorefrontURLSearchProfileService(
	profiles *repository.StorefrontURLSearchProfileRepository,
	catalog *repository.StorefrontRouteCatalogRepository,
) *StorefrontURLSearchProfileService {
	return &StorefrontURLSearchProfileService{
		profiles: profiles,
		catalog:  catalog,
	}
}

func (s *StorefrontURLSearchProfileService) List(locale string) ([]urlmanagementdomain.StorefrontURLSearchProfile, error) {
	profiles, err := s.list(false)
	if err != nil {
		return nil, err
	}
	return filterURLSearchProfiles(profiles, locale, false), nil
}

func (s *StorefrontURLSearchProfileService) PublicIndex(locale string) ([]urlmanagementdomain.StorefrontURLSearchProfile, error) {
	profiles, err := s.list(false)
	if err != nil {
		return nil, err
	}

	catalogEntries, err := s.listPublicCatalogEntries(locale)
	if err != nil {
		return nil, err
	}

	profileByRouteEntryID := make(map[uint]urlmanagementdomain.StorefrontURLSearchProfile, len(profiles))
	for _, profile := range profiles {
		if profile.RouteEntry == nil {
			continue
		}
		profileByRouteEntryID[profile.RouteEntryID] = profile
	}

	merged := make([]urlmanagementdomain.StorefrontURLSearchProfile, 0, len(catalogEntries))
	for _, entry := range catalogEntries {
		if profile, ok := profileByRouteEntryID[entry.ID]; ok {
			if !isPublicURLSearchProfile(profile) {
				continue
			}
			merged = append(merged, profile)
			continue
		}

		entryCopy := entry
		merged = append(merged, urlmanagementdomain.StorefrontURLSearchProfile{
			RouteEntryID:   entryCopy.ID,
			Enabled:        true,
			SearchWeight:   0,
			Keywords:       datatypes.JSONSlice[string]{},
			DisplayTitle:   entryCopy.Title,
			DisplaySummary: entryCopy.Summary,
			RouteEntry:     &entryCopy,
		})
	}

	return merged, nil
}

func (s *StorefrontURLSearchProfileService) Get(routeEntryID uint) (*urlmanagementdomain.StorefrontURLSearchProfile, error) {
	if s == nil || s.catalog == nil {
		return nil, errors.New("storefront URL search profile service is unavailable")
	}
	if routeEntryID == 0 {
		return nil, errors.New("route entry ID is required")
	}

	entry, err := s.catalog.FindByID(routeEntryID)
	if err != nil {
		return nil, err
	}

	profile, err := s.profiles.FindByRouteEntryID(routeEntryID)
	if err == nil {
		return profile, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &urlmanagementdomain.StorefrontURLSearchProfile{
		RouteEntryID:    routeEntryID,
		Enabled:         true,
		SearchWeight:    100,
		Keywords:        datatypes.JSONSlice[string]{},
		DisplayTitle:    entry.Title,
		DisplaySummary:  entry.Summary,
		RouteEntry:      entry,
		CreatedAt:       time.Time{},
		UpdatedAt:       time.Time{},
	}, nil
}

func (s *StorefrontURLSearchProfileService) Upsert(
	routeEntryID uint,
	input urlmanagementdomain.StorefrontURLSearchProfileInput,
) (*urlmanagementdomain.StorefrontURLSearchProfile, error) {
	if s == nil || s.profiles == nil || s.catalog == nil {
		return nil, errors.New("storefront URL search profile service is unavailable")
	}
	if routeEntryID == 0 {
		return nil, errors.New("route entry ID is required")
	}
	if _, err := s.catalog.FindByID(routeEntryID); err != nil {
		return nil, err
	}

	keywords := normalizeURLSearchKeywords(input.Keywords)
	profile := &urlmanagementdomain.StorefrontURLSearchProfile{
		RouteEntryID:   routeEntryID,
		Enabled:        input.Enabled,
		SearchWeight:   normalizeURLSearchWeight(input.SearchWeight),
		Keywords:       keywords,
		DisplayTitle:   strings.TrimSpace(input.DisplayTitle),
		DisplaySummary: strings.TrimSpace(input.DisplaySummary),
	}
	if err := s.profiles.Upsert(profile); err != nil {
		return nil, err
	}
	return s.profiles.FindByRouteEntryID(routeEntryID)
}

func (s *StorefrontURLSearchProfileService) list(publicOnly bool) ([]urlmanagementdomain.StorefrontURLSearchProfile, error) {
	if s == nil || s.profiles == nil {
		return nil, errors.New("storefront URL search profile service is unavailable")
	}
	profiles, err := s.profiles.List()
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func filterURLSearchProfiles(
	profiles []urlmanagementdomain.StorefrontURLSearchProfile,
	locale string,
	publicOnly bool,
) []urlmanagementdomain.StorefrontURLSearchProfile {
	normalizedLocale := strings.TrimSpace(locale)
	filtered := make([]urlmanagementdomain.StorefrontURLSearchProfile, 0, len(profiles))
	for _, profile := range profiles {
		if profile.RouteEntry == nil {
			continue
		}
		if normalizedLocale != "" && profile.RouteEntry.Locale != normalizedLocale {
			continue
		}
		if publicOnly {
			if !profile.Enabled {
				continue
			}
			if profile.RouteEntry.EntryStatus != seo.RouteEntryStatusActive {
				continue
			}
			if profile.RouteEntry.IsAlias || !profile.RouteEntry.IsIndexable {
				continue
			}
		}
		filtered = append(filtered, profile)
	}
	return filtered
}

func (s *StorefrontURLSearchProfileService) listPublicCatalogEntries(locale string) ([]seo.StorefrontRouteCatalogEntry, error) {
	if s == nil || s.catalog == nil {
		return nil, errors.New("storefront URL search profile service is unavailable")
	}

	normalizedLocale := strings.TrimSpace(locale)
	entries := make([]seo.StorefrontRouteCatalogEntry, 0)
	for page := 1; ; page++ {
		batch, total, err := s.catalog.List(repository.StorefrontRouteCatalogListFilter{
			Page:        page,
			PageSize:    storefrontURLSearchPublicPageSize,
			Locale:      normalizedLocale,
			EntryStatus: seo.RouteEntryStatusActive,
			Searchable:  boolValuePtr(true),
			Indexable:   boolValuePtr(true),
			ExcludeAlias: true,
		})
		if err != nil {
			return nil, err
		}

		for _, entry := range batch {
			if !isPublicURLSearchCatalogEntry(entry) {
				continue
			}
			entries = append(entries, entry)
		}
		if len(batch) == 0 || len(entries) >= int(total) || len(batch) < storefrontURLSearchPublicPageSize {
			break
		}
	}

	return entries, nil
}

func isPublicURLSearchProfile(profile urlmanagementdomain.StorefrontURLSearchProfile) bool {
	if !profile.Enabled {
		return false
	}
	if profile.RouteEntry == nil {
		return false
	}
	if profile.RouteEntry.EntryStatus != seo.RouteEntryStatusActive {
		return false
	}
	if profile.RouteEntry.IsAlias || !profile.RouteEntry.IsSearchable || !profile.RouteEntry.IsIndexable {
		return false
	}
	return true
}

func isPublicURLSearchCatalogEntry(entry seo.StorefrontRouteCatalogEntry) bool {
	if entry.EntryStatus != seo.RouteEntryStatusActive {
		return false
	}
	if entry.IsAlias || !entry.IsSearchable || !entry.IsIndexable {
		return false
	}
	return true
}

func boolValuePtr(value bool) *bool {
	return &value
}

func normalizeURLSearchKeywords(values []string) datatypes.JSONSlice[string] {
	seen := make(map[string]struct{}, len(values))
	keywords := make([]string, 0, len(values))
	for _, value := range values {
		keyword := strings.TrimSpace(value)
		if keyword == "" {
			continue
		}
		lookupKey := strings.ToLower(keyword)
		if _, exists := seen[lookupKey]; exists {
			continue
		}
		seen[lookupKey] = struct{}{}
		keywords = append(keywords, keyword)
	}
	return datatypes.JSONSlice[string](keywords)
}

func normalizeURLSearchWeight(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
