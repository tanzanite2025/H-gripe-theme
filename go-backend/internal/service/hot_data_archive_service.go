package service

import (
	"context"
	"errors"
	"time"

	"commerce-platform/internal/repository"
)

const (
	HotDataArchiveRetention    = 90 * 24 * time.Hour
	defaultHotDataArchiveBatch = 500
)

type HotDataArchiveResult struct {
	Cutoff                   time.Time `json:"cutoff"`
	SiteQualityRunsArchived  int       `json:"site_quality_runs_archived"`
	AfterSalesEventsArchived int       `json:"after_sales_events_archived"`
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

func (s *HotDataArchiveService) ArchiveExpiredTerminalData(
	ctx context.Context,
	now time.Time,
) (HotDataArchiveResult, error) {
	if s == nil || s.repository == nil {
		return HotDataArchiveResult{}, errors.New("hot data archive service is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	result := HotDataArchiveResult{
		Cutoff: now.Add(-HotDataArchiveRetention),
	}
	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		count, err := s.repository.ArchiveTerminalSiteQualityRuns(ctx, result.Cutoff, s.batchLimit)
		if err != nil {
			return result, err
		}
		result.SiteQualityRunsArchived += count
		if count == 0 {
			break
		}
	}
	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		count, err := s.repository.ArchiveTerminalAfterSalesEvents(ctx, result.Cutoff, s.batchLimit)
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
