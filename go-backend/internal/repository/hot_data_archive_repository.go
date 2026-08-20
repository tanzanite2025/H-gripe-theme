package repository

import (
	"context"
	"errors"
	"time"

	aftersalesdomain "commerce-platform/internal/domain/aftersales"
	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	siteQualityRunArchiveTable     = "site_quality_runs_archive"
	afterSalesEventArchiveTable    = "after_sales_case_events_archive"
	defaultHotDataArchiveBatchSize = 500
)

type HotDataArchiveRepository struct {
	db *gorm.DB
}

func NewHotDataArchiveRepository(db *gorm.DB) *HotDataArchiveRepository {
	return &HotDataArchiveRepository{db: db}
}

func (r *HotDataArchiveRepository) ArchiveTerminalSiteQualityRuns(
	ctx context.Context,
	cutoff time.Time,
	limit int,
) (int, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("hot data archive repository is unavailable")
	}
	if limit <= 0 {
		limit = defaultHotDataArchiveBatchSize
	}

	var archived int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var runs []sitequalitydomain.SiteQualityRun
		if err := tx.Model(&sitequalitydomain.SiteQualityRun{}).
			Where("created_at < ?", cutoff.UTC()).
			Where("status IN ?", []string{
				sitequalitydomain.SiteQualityRunStatusSuccess,
				sitequalitydomain.SiteQualityRunStatusFailed,
			}).
			Order("created_at ASC, id ASC").
			Limit(limit).
			Find(&runs).Error; err != nil {
			return err
		}
		if len(runs) == 0 {
			return nil
		}

		if err := tx.Table(siteQualityRunArchiveTable).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&runs).Error; err != nil {
			return err
		}
		ids := make([]uint, 0, len(runs))
		for _, run := range runs {
			ids = append(ids, run.ID)
		}
		if err := tx.Where("id IN ?", ids).Delete(&sitequalitydomain.SiteQualityRun{}).Error; err != nil {
			return err
		}
		archived = len(runs)
		return nil
	})
	return archived, err
}

func (r *HotDataArchiveRepository) ArchiveTerminalAfterSalesEvents(
	ctx context.Context,
	cutoff time.Time,
	limit int,
) (int, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("hot data archive repository is unavailable")
	}
	if limit <= 0 {
		limit = defaultHotDataArchiveBatchSize
	}

	var archived int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var events []aftersalesdomain.AfterSalesCaseEvent
		if err := tx.Table("after_sales_case_events AS events").
			Select("events.*").
			Joins("JOIN after_sales_cases AS cases ON cases.id = events.case_id").
			Where("cases.status IN ?", []string{
				aftersalesdomain.StatusCompleted,
				aftersalesdomain.StatusRejected,
				aftersalesdomain.StatusCancelled,
			}).
			Where("COALESCE(cases.closed_at, cases.updated_at) < ?", cutoff.UTC()).
			Where("events.created_at < ?", cutoff.UTC()).
			Order("events.created_at ASC, events.id ASC").
			Limit(limit).
			Find(&events).Error; err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}

		if err := tx.Table(afterSalesEventArchiveTable).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&events).Error; err != nil {
			return err
		}
		ids := make([]uint, 0, len(events))
		for _, event := range events {
			ids = append(ids, event.ID)
		}
		if err := tx.Where("id IN ?", ids).Delete(&aftersalesdomain.AfterSalesCaseEvent{}).Error; err != nil {
			return err
		}
		archived = len(events)
		return nil
	})
	return archived, err
}
