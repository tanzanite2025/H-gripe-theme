package service

import (
	"context"
	"errors"
	"time"

	"commerce-platform/internal/repository"
)

const (
	SiteQualityRunRetention         = 7 * 24 * time.Hour
	AfterSalesEventArchiveRetention = 90 * 24 * time.Hour
	defaultHotDataArchiveBatch      = 500
)

type HotDataRetentionAndArchiveResult struct {
	SiteQualityRunCutoff          time.Time `json:"site_quality_run_cutoff"`
	AfterSalesEventArchiveCutoff  time.Time `json:"after_sales_event_archive_cutoff"`
	SiteQualityRunsPurged         int       `json:"site_quality_runs_purged"`
	ArchivedSiteQualityRunsPurged int       `json:"archived_site_quality_runs_purged"`
	AfterSalesEventsArchived      int       `json:"after_sales_events_archived"`
}

type HotDataArchiveService struct {
	repository *repository.HotDataArchiveRepository
	batchLimit int
}

func NewHotDataArchiveService(
	archiveRepository *repository.HotDataArchiveRepository,
	batchLimit int,
) *HotDataArchiveService {
	if batchLimit <= 0 {
		batchLimit = defaultHotDataArchiveBatch
	}
	return &HotDataArchiveService{
		repository: archiveRepository,
		batchLimit: batchLimit,
	}
}

func (s *HotDataArchiveService) ProcessExpiredTerminalData(
	ctx context.Context,
	now time.Time,
) (HotDataRetentionAndArchiveResult, error) {
	if s == nil || s.repository == nil {
		return HotDataRetentionAndArchiveResult{}, errors.New("hot data archive service is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	result := HotDataRetentionAndArchiveResult{
		SiteQualityRunCutoff:         now.Add(-SiteQualityRunRetention),
		AfterSalesEventArchiveCutoff: now.Add(-AfterSalesEventArchiveRetention),
	}
	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		count, err := s.repository.PurgeExpiredTerminalSiteQualityRuns(ctx, result.SiteQualityRunCutoff, s.batchLimit)
		if err != nil {
			return result, err
		}
		result.SiteQualityRunsPurged += count
		if count == 0 {
			break
		}
	}
	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		count, err := s.repository.PurgeExpiredArchivedSiteQualityRuns(ctx, result.SiteQualityRunCutoff, s.batchLimit)
		if err != nil {
			return result, err
		}
		result.ArchivedSiteQualityRunsPurged += count
		if count == 0 {
			break
		}
	}
	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		count, err := s.repository.ArchiveTerminalAfterSalesEvents(ctx, result.AfterSalesEventArchiveCutoff, s.batchLimit)
		if err != nil {
			return result, err
		}
		result.AfterSalesEventsArchived += count
		if count == 0 {
			break
		}
	}
	return result, nil
}
