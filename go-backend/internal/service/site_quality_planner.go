package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/repository"
)

type SiteQualityPlanResult struct {
	EnqueuedJobs int `json:"enqueued_jobs"`
	SkippedJobs  int `json:"skipped_jobs"`
}

// PlanDueWork creates durable scheduled jobs for targets that are already due.
// Route-catalog reconciliation is deliberately separate so a ledger update
// never performs provider work in the same step.
func (s *SiteQualityEngineService) PlanDueWork(now time.Time, limit int) (SiteQualityPlanResult, error) {
	if s == nil || s.targets == nil || s.jobs == nil {
		return SiteQualityPlanResult{}, errors.New("SiteQuality quality planner is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	targets, err := s.targets.ListDue(now, limit)
	if err != nil {
		return SiteQualityPlanResult{}, err
	}

	var result SiteQualityPlanResult
	for _, target := range targets {
		dueAt := now
		if target.NextScheduledAt != nil && !target.NextScheduledAt.IsZero() {
			dueAt = target.NextScheduledAt.UTC()
		}
		interval := siteQualityTargetInterval(target)
		nextScheduledAt := dueAt.Add(interval)
		for !nextScheduledAt.After(now) {
			nextScheduledAt = nextScheduledAt.Add(interval)
		}

		job := sitequalitydomain.SiteQualityJob{
			TargetID:              target.ID,
			Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
			Kind:                  sitequalitydomain.SiteQualityJobKindScheduled,
			Status:                sitequalitydomain.SiteQualityJobStatusQueued,
			IdempotencyKey:        siteQualityScheduledJobKey(target.ID, dueAt),
			SampleCount:           s.cfg.SampleCount,
			RequiredConfirmations: s.cfg.RequiredConfirmations,
			MaxAttempts:           s.cfg.MaxAttempts,
			AvailableAt:           now,
			ReleaseID:             s.cfg.ReleaseID,
		}
		_, created, err := s.jobs.Enqueue(job)
		if err != nil {
			return result, err
		}
		if created {
			result.EnqueuedJobs++
		} else {
			result.SkippedJobs++
		}
		if err := s.targets.MarkScheduled(target.ID, now, nextScheduledAt); err != nil {
			return result, err
		}
	}
	return result, nil
}

func siteQualityTargetInterval(target sitequalitydomain.SiteQualityTarget) time.Duration {
	if target.SamplingIntervalSeconds > 0 {
		return time.Duration(target.SamplingIntervalSeconds) * time.Second
	}
	return 7 * 24 * time.Hour
}

func siteQualityScheduledJobKey(targetID uint, dueAt time.Time) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%d:%d", targetID, dueAt.UnixNano())))
	return fmt.Sprintf("site-quality:%s:%s", sitequalitydomain.SiteQualityJobKindScheduled, hex.EncodeToString(sum[:]))
}

// SyncTargetsFromRouteCatalog reconciles target identity only. It never
// schedules or enqueues a Lighthouse inspection.
func (s *SiteQualityEngineService) SyncTargetsFromRouteCatalog(now time.Time, limit int) (int, error) {
	if s == nil || s.targets == nil || s.routeCatalog == nil {
		return 0, errors.New("SiteQuality quality engine is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	page := 1
	synced := 0
	for {
		entries, total, err := s.routeCatalog.List(repository.StorefrontRouteCatalogListFilter{
			Page:     page,
			PageSize: limit,
		})
		if err != nil {
			return synced, err
		}
		for _, entry := range entries {
			if err := s.syncTargetFromRouteEntry(entry, now, "", now); err != nil {
				return synced, err
			}
			if entry.IsCheckable && !entry.IsAlias && entry.EntryStatus == seodomain.RouteEntryStatusActive {
				synced++
			}
		}
		if int64(page*limit) >= total || len(entries) == 0 {
			break
		}
		page++
	}
	return synced, nil
}

func (s *SiteQualityEngineService) SyncTargetFromRouteEntry(
	ctx context.Context,
	routeEntryID uint,
	marker string,
	syncedAt ...time.Time,
) error {
	if s == nil || s.targets == nil || s.routeCatalog == nil {
		return errors.New("SiteQuality quality engine is unavailable")
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	entry, err := s.routeCatalog.FindByID(routeEntryID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	eventAt := now
	if len(syncedAt) > 0 && !syncedAt[0].IsZero() {
		eventAt = syncedAt[0].UTC()
	}
	return s.syncTargetFromRouteEntry(*entry, now, marker, eventAt)
}

func (s *SiteQualityEngineService) syncTargetFromRouteEntry(
	entry seodomain.StorefrontRouteCatalogEntry,
	now time.Time,
	marker string,
	syncedAt time.Time,
) error {
	syncMarker := strings.TrimSpace(entry.ManifestVersion)
	if strings.TrimSpace(marker) != "" {
		syncMarker = strings.TrimSpace(marker)
	}
	ledgerSyncedAt := now
	if !syncedAt.IsZero() {
		ledgerSyncedAt = syncedAt.UTC()
	}
	if !entry.IsCheckable || entry.IsAlias || entry.EntryStatus != seodomain.RouteEntryStatusActive {
		return s.targets.DisableByRouteEntryID(
			entry.ID,
			"route catalog target is no longer active or checkable",
			syncMarker,
			now,
			ledgerSyncedAt,
		)
	}
	canonicalURL, err := s.canonicalURLForRoute(entry)
	if err != nil {
		return s.targets.DisableByRouteEntryID(entry.ID, err.Error(), syncMarker, now, ledgerSyncedAt)
	}
	tier, interval := siteQualityTier(entry)
	_, err = s.targets.Upsert(sitequalitydomain.SiteQualityTargetInput{
		RouteEntryID:            &entry.ID,
		CanonicalURL:            canonicalURL,
		Locale:                  entry.Locale,
		Source:                  sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
		SourceType:              entry.SourceType,
		Title:                   entry.Title,
		SamplingTier:            tier,
		SamplingIntervalSeconds: int(interval.Seconds()),
		Enabled:                 true,
		LedgerSynced:            true,
		LedgerSyncMarker:        syncMarker,
		LedgerSyncedAt:          &ledgerSyncedAt,
	}, now)
	return err
}

func (s *SiteQualityEngineService) canonicalURLForRoute(entry seodomain.StorefrontRouteCatalogEntry) (string, error) {
	path := strings.TrimSpace(entry.CanonicalPath)
	if path == "" {
		path = strings.TrimSpace(entry.Path)
	}
	if path == "" {
		return "", errors.New("route path is required")
	}
	base := strings.TrimRight(strings.TrimSpace(s.cfg.BaseURL), "/")
	if base == "" {
		return "", errors.New("storefront base URL is required")
	}
	parsedBase, err := url.Parse(base)
	if err != nil || parsedBase.Scheme == "" || parsedBase.Host == "" {
		return "", errors.New("storefront base URL must be absolute")
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return canonicalizeAbsoluteSiteQualityURL(path)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	parsedBase.Path = strings.TrimRight(parsedBase.Path, "/") + path
	parsedBase.RawQuery = ""
	parsedBase.Fragment = ""
	return canonicalizeAbsoluteSiteQualityURL(parsedBase.String())
}

func canonicalizeAbsoluteSiteQualityURL(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("canonical URL must be absolute")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	host := strings.ToLower(parsed.Hostname())
	port := parsed.Port()
	if (parsed.Scheme == "https" && port == "443") || (parsed.Scheme == "http" && port == "80") {
		port = ""
	}
	parsed.Host = host
	if port != "" {
		parsed.Host = host + ":" + port
	}
	parsed.Fragment = ""
	parsed.RawQuery = ""
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	if parsed.Path != "/" {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
		if parsed.Path == "" {
			parsed.Path = "/"
		}
	}
	return parsed.String(), nil
}

func siteQualityTier(entry seodomain.StorefrontRouteCatalogEntry) (string, time.Duration) {
	path := strings.ToLower(strings.TrimSpace(entry.Path))
	sourceType := strings.ToLower(strings.TrimSpace(entry.SourceType))
	if path == "/" ||
		strings.Contains(path, "/checkout") ||
		strings.Contains(path, "/account") ||
		strings.Contains(path, "/support") ||
		sourceType == seodomain.RouteSourceStatic {
		return sitequalitydomain.SiteQualityTargetTierCritical, 24 * time.Hour
	}
	if sourceType == seodomain.RouteSourceProduct {
		return sitequalitydomain.SiteQualityTargetTierStandard, 7 * 24 * time.Hour
	}
	return sitequalitydomain.SiteQualityTargetTierBackground, 30 * 24 * time.Hour
}
