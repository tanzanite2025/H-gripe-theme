package service

import (
	"errors"
	"sort"
	"strings"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/repository"
)

const siteQualityTargetSelectionPageSize = 200

type SiteQualityTargetOption struct {
	URL        string `json:"url"`
	Path       string `json:"path"`
	Title      string `json:"title"`
	Locale     string `json:"locale"`
	SourceType string `json:"source_type"`
	IsHome     bool   `json:"is_home"`
}

type SiteQualityTargetOptions struct {
	DefaultURL string                    `json:"default_url"`
	Items      []SiteQualityTargetOption `json:"items"`
}

func (s *SiteQualityEngineService) ListTargetOptions() (*SiteQualityTargetOptions, error) {
	if s == nil {
		return nil, errors.New("SiteQuality quality engine is unavailable")
	}

	fallbackURL, err := s.siteQualityFallbackTargetURL()
	if err != nil {
		return nil, err
	}

	entries, err := s.siteQualityTargetEntries()
	if err != nil {
		return nil, err
	}

	options := make([]SiteQualityTargetOption, 0, len(entries)+1)
	seenURLs := make(map[string]struct{}, len(entries)+1)
	defaultURL := fallbackURL
	homeFound := false

	for _, entry := range entries {
		targetURL, err := s.canonicalURLForRoute(entry)
		if err != nil {
			continue
		}
		if _, seen := seenURLs[targetURL]; seen {
			continue
		}
		seenURLs[targetURL] = struct{}{}

		option := SiteQualityTargetOption{
			URL:        targetURL,
			Path:       strings.TrimSpace(entry.Path),
			Title:      strings.TrimSpace(entry.Title),
			Locale:     strings.TrimSpace(entry.Locale),
			SourceType: strings.TrimSpace(entry.SourceType),
			IsHome:     siteQualityRouteIsHome(entry),
		}
		if option.IsHome && !homeFound {
			defaultURL = option.URL
			homeFound = true
		}
		options = append(options, option)
	}

	if !homeFound {
		fallbackOption := siteQualityFallbackTargetOption(fallbackURL)
		if _, seen := seenURLs[fallbackOption.URL]; !seen {
			options = append([]SiteQualityTargetOption{fallbackOption}, options...)
		} else {
			for idx := range options {
				if options[idx].URL == fallbackOption.URL {
					options[idx].IsHome = true
					break
				}
			}
		}
		defaultURL = fallbackOption.URL
	}

	sort.SliceStable(options, func(i, j int) bool {
		iHome := options[i].IsHome
		jHome := options[j].IsHome
		if iHome != jHome {
			return iHome
		}
		if options[i].Path != options[j].Path {
			return options[i].Path < options[j].Path
		}
		if options[i].Locale != options[j].Locale {
			return options[i].Locale < options[j].Locale
		}
		if options[i].Title != options[j].Title {
			return options[i].Title < options[j].Title
		}
		return options[i].URL < options[j].URL
	})

	if len(options) == 0 {
		options = append(options, siteQualityFallbackTargetOption(fallbackURL))
		defaultURL = fallbackURL
	}

	return &SiteQualityTargetOptions{
		DefaultURL: defaultURL,
		Items:      options,
	}, nil
}

func (s *SiteQualityEngineService) siteQualityTargetEntries() ([]seodomain.StorefrontRouteCatalogEntry, error) {
	if s == nil || s.routeCatalog == nil {
		return nil, nil
	}

	entries := make([]seodomain.StorefrontRouteCatalogEntry, 0, siteQualityTargetSelectionPageSize)
	page := 1
	for {
		batch, total, err := s.routeCatalog.List(repository.StorefrontRouteCatalogListFilter{
			Page:         page,
			PageSize:     siteQualityTargetSelectionPageSize,
			EntryStatus:  seodomain.RouteEntryStatusActive,
			ExcludeAlias: true,
		})
		if err != nil {
			return nil, err
		}
		for _, entry := range batch {
			if !entry.IsCheckable || entry.IsAlias || entry.EntryStatus != seodomain.RouteEntryStatusActive {
				continue
			}
			entries = append(entries, entry)
		}
		if int64(page*siteQualityTargetSelectionPageSize) >= total || len(batch) == 0 {
			break
		}
		page++
	}

	sort.SliceStable(entries, func(i, j int) bool {
		iHome := siteQualityRouteIsHome(entries[i])
		jHome := siteQualityRouteIsHome(entries[j])
		if iHome != jHome {
			return iHome
		}
		if entries[i].Path != entries[j].Path {
			return entries[i].Path < entries[j].Path
		}
		if entries[i].Locale != entries[j].Locale {
			return entries[i].Locale < entries[j].Locale
		}
		if entries[i].Title != entries[j].Title {
			return entries[i].Title < entries[j].Title
		}
		return entries[i].ID < entries[j].ID
	})

	return entries, nil
}

func (s *SiteQualityEngineService) siteQualityFallbackTargetURL() (string, error) {
	if s != nil && s.lighthouseRunner != nil {
		fallback := strings.TrimSpace(s.lighthouseRunner.defaultTargetURL)
		if fallback != "" {
			if normalized, err := canonicalizeAbsoluteSiteQualityURL(fallback); err == nil {
				return normalized, nil
			}
			return fallback, nil
		}
	}
	if s != nil {
		if normalized, err := canonicalizeAbsoluteSiteQualityURL(s.cfg.BaseURL); err == nil && normalized != "" {
			return normalized, nil
		}
	}
	return "", errors.New("SiteQuality default target URL is unavailable")
}

func (s *SiteQualityEngineService) defaultSiteQualityTargetURL() (string, error) {
	options, err := s.ListTargetOptions()
	if err != nil {
		return "", err
	}
	return options.DefaultURL, nil
}

func siteQualityRouteIsHome(entry seodomain.StorefrontRouteCatalogEntry) bool {
	path := strings.TrimSpace(entry.CanonicalPath)
	if path == "" {
		path = strings.TrimSpace(entry.Path)
	}
	return path == "/"
}

func siteQualityFallbackTargetOption(url string) SiteQualityTargetOption {
	return SiteQualityTargetOption{
		URL:    strings.TrimSpace(url),
		Path:   "/",
		IsHome: true,
	}
}
