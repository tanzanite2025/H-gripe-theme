package repository

import (
	"commerce-platform/internal/domain/feedback"
	"database/sql"
	"strings"
	"time"

	"gorm.io/gorm"
)

type FeedbackRepository struct {
	db *gorm.DB
}

type FeedbackAdminListFilter struct {
	Status    string
	ThreadKey string
	PagePath  string
	Search    string
	Page      int
	PageSize  int
}

type FeedbackRiskSummary struct {
	PendingTotal       int64
	PendingOver24Hours int64
	WindowTotal        int64
	LastHourTotal      int64
	StatusCounts       []FeedbackStatusCount
	HotPages           []FeedbackRiskPage
	SourceBursts       []FeedbackRiskSource
}

type FeedbackStatusCount struct {
	Status string
	Count  int64
}

type FeedbackRiskPage struct {
	PagePath       string
	PageTitle      string
	ThreadKey      string
	FilterKind     string
	FilterValue    string
	FeedbackCount  int64
	PendingCount   int64
	LastFeedbackAt sql.NullString
}

type FeedbackRiskSource struct {
	SourceHash     string
	FeedbackCount  int64
	PageCount      int64
	PendingCount   int64
	LastFeedbackAt sql.NullString
}

func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

func (r *FeedbackRepository) List(threadKey, status, search string, page, pageSize int) ([]feedback.Feedback, int64, error) {
	var items []feedback.Feedback
	var total int64

	query := r.db.Model(&feedback.Feedback{}).Where("thread_key = ?", threadKey)
	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status = ?", "approved")
	}
	if search != "" {
		query = query.Where("content LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *FeedbackRepository) Create(item *feedback.Feedback) error {
	return r.db.Create(item).Error
}

func (r *FeedbackRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&feedback.Feedback{}).Where("id = ?", id).Update("status", status).Error
}

func (r *FeedbackRepository) ListAdmin(filter FeedbackAdminListFilter) ([]feedback.Feedback, int64, error) {
	var items []feedback.Feedback
	var total int64

	query := r.db.Model(&feedback.Feedback{})
	if filter.Status != "" && filter.Status != "all" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ThreadKey != "" {
		query = query.Where("thread_key = ?", filter.ThreadKey)
	}
	if filter.PagePath != "" {
		query = query.Where("page_path = ?", filter.PagePath)
	}
	if filter.Search != "" {
		searchLike := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where(
			"LOWER(content) LIKE ? OR LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(thread_key) LIKE ? OR LOWER(page_path) LIKE ? OR LOWER(page_title) LIKE ?",
			searchLike,
			searchLike,
			searchLike,
			searchLike,
			searchLike,
			searchLike,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&items).Error
	return items, total, err
}

func (r *FeedbackRepository) Get(id uint) (*feedback.Feedback, error) {
	var item feedback.Feedback
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *FeedbackRepository) UpdateAdmin(item *feedback.Feedback) error {
	return r.db.Model(&feedback.Feedback{}).
		Where("id = ?", item.ID).
		Updates(map[string]interface{}{
			"status":        item.Status,
			"reply_content": item.ReplyContent,
			"replied_at":    item.RepliedAt,
			"replied_by":    item.RepliedBy,
			"reviewed_at":   item.ReviewedAt,
			"reviewed_by":   item.ReviewedBy,
			"updated_at":    time.Now().UTC(),
		}).Error
}

func (r *FeedbackRepository) RiskSummary(windowStart, staleBefore, lastHourStart time.Time, hotLimit, sourceLimit int) (FeedbackRiskSummary, error) {
	if hotLimit <= 0 {
		hotLimit = 5
	}
	if sourceLimit <= 0 {
		sourceLimit = 5
	}

	var summary FeedbackRiskSummary
	if err := r.db.Model(&feedback.Feedback{}).
		Where("status = ?", "pending").
		Count(&summary.PendingTotal).Error; err != nil {
		return summary, err
	}
	if err := r.db.Model(&feedback.Feedback{}).
		Where("status = ? AND created_at <= ?", "pending", staleBefore).
		Count(&summary.PendingOver24Hours).Error; err != nil {
		return summary, err
	}
	if err := r.db.Model(&feedback.Feedback{}).
		Where("created_at >= ?", windowStart).
		Count(&summary.WindowTotal).Error; err != nil {
		return summary, err
	}
	if err := r.db.Model(&feedback.Feedback{}).
		Where("created_at >= ?", lastHourStart).
		Count(&summary.LastHourTotal).Error; err != nil {
		return summary, err
	}
	if err := r.db.Model(&feedback.Feedback{}).
		Select("status, COUNT(*) AS count").
		Group("status").
		Scan(&summary.StatusCounts).Error; err != nil {
		return summary, err
	}
	if err := r.db.Model(&feedback.Feedback{}).
		Select(`
			COALESCE(NULLIF(COALESCE(page_path, ''), ''), thread_key) AS page_path,
			COALESCE(MAX(page_title), '') AS page_title,
			MIN(thread_key) AS thread_key,
			CASE WHEN COALESCE(page_path, '') <> '' THEN 'page_path' ELSE 'thread_key' END AS filter_kind,
			COALESCE(NULLIF(COALESCE(page_path, ''), ''), thread_key) AS filter_value,
			COUNT(*) AS feedback_count,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) AS pending_count,
			CAST(MAX(created_at) AS TEXT) AS last_feedback_at
		`).
		Where("created_at >= ?", windowStart).
		Group("CASE WHEN COALESCE(page_path, '') <> '' THEN 'page_path' ELSE 'thread_key' END").
		Group("COALESCE(NULLIF(COALESCE(page_path, ''), ''), thread_key)").
		Order("feedback_count DESC, pending_count DESC, last_feedback_at DESC").
		Limit(hotLimit).
		Scan(&summary.HotPages).Error; err != nil {
		return summary, err
	}
	if err := r.db.Model(&feedback.Feedback{}).
		Select(`
			source_hash,
			COUNT(*) AS feedback_count,
			COUNT(DISTINCT (
				CASE WHEN COALESCE(page_path, '') <> '' THEN 'page_path' ELSE 'thread_key' END
				|| ':' ||
				COALESCE(NULLIF(COALESCE(page_path, ''), ''), thread_key)
			)) AS page_count,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) AS pending_count,
			CAST(MAX(created_at) AS TEXT) AS last_feedback_at
		`).
		Where("created_at >= ? AND source_hash <> ''", windowStart).
		Group("source_hash").
		Having(`COUNT(*) >= 2 OR COUNT(DISTINCT (
			CASE WHEN COALESCE(page_path, '') <> '' THEN 'page_path' ELSE 'thread_key' END
			|| ':' ||
			COALESCE(NULLIF(COALESCE(page_path, ''), ''), thread_key)
		)) >= 2`).
		Order("page_count DESC, feedback_count DESC, pending_count DESC, last_feedback_at DESC").
		Limit(sourceLimit).
		Scan(&summary.SourceBursts).Error; err != nil {
		return summary, err
	}

	return summary, nil
}
