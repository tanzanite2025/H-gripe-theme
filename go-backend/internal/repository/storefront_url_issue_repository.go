package repository

import (
	"errors"
	"time"

	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StorefrontURLIssueRepository struct {
	db *gorm.DB
}

type StorefrontURLIssueListFilter struct {
	Page      int
	PageSize  int
	State     string
	Severity  string
	IssueType string
}

func NewStorefrontURLIssueRepository(db *gorm.DB) *StorefrontURLIssueRepository {
	return &StorefrontURLIssueRepository{db: db}
}

func (r *StorefrontURLIssueRepository) List(
	filter StorefrontURLIssueListFilter,
) ([]urlmanagementdomain.StorefrontURLIssue, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("storefront URL issue repository is unavailable")
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 200 {
		filter.PageSize = 50
	}

	query := r.db.Model(&urlmanagementdomain.StorefrontURLIssue{})
	switch filter.State {
	case "active":
		query = query.Where("state IN ?", []string{
			urlmanagementdomain.URLIssueStateOpen,
			urlmanagementdomain.URLIssueStateAcknowledged,
			urlmanagementdomain.URLIssueStateResolved,
		})
	case "", "all":
	default:
		query = query.Where("state = ?", filter.State)
	}
	if filter.Severity != "" {
		query = query.Where("severity = ?", filter.Severity)
	}
	if filter.IssueType != "" {
		query = query.Where("issue_type = ?", filter.IssueType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var issues []urlmanagementdomain.StorefrontURLIssue
	err := query.
		Preload("RouteEntry").
		Order("CASE severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END").
		Order("last_detected_at ASC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&issues).Error
	return issues, total, err
}

func (r *StorefrontURLIssueRepository) Stats() (urlmanagementdomain.StorefrontURLIssueStats, error) {
	if r == nil || r.db == nil {
		return urlmanagementdomain.StorefrontURLIssueStats{}, errors.New("storefront URL issue repository is unavailable")
	}
	var stats urlmanagementdomain.StorefrontURLIssueStats
	err := r.db.Model(&urlmanagementdomain.StorefrontURLIssue{}).
		Select(`
			COALESCE(SUM(CASE WHEN state IN ('open', 'acknowledged', 'resolved') THEN 1 ELSE 0 END), 0) AS active,
			COALESCE(SUM(CASE WHEN state = 'open' THEN 1 ELSE 0 END), 0) AS open,
			COALESCE(SUM(CASE WHEN state = 'acknowledged' THEN 1 ELSE 0 END), 0) AS acknowledged,
			COALESCE(SUM(CASE WHEN state = 'resolved' THEN 1 ELSE 0 END), 0) AS resolved,
			COALESCE(SUM(CASE WHEN state = 'verified' THEN 1 ELSE 0 END), 0) AS verified,
			COALESCE(SUM(CASE WHEN state = 'suppressed' THEN 1 ELSE 0 END), 0) AS suppressed,
			COALESCE(SUM(CASE WHEN severity = 'critical' AND state IN ('open', 'acknowledged', 'resolved') THEN 1 ELSE 0 END), 0) AS critical,
			COALESCE(SUM(CASE WHEN severity = 'high' AND state IN ('open', 'acknowledged', 'resolved') THEN 1 ELSE 0 END), 0) AS high
		`).
		Scan(&stats).Error
	return stats, err
}

func (r *StorefrontURLIssueRepository) FindByID(
	id uint,
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront URL issue repository is unavailable")
	}
	var issue urlmanagementdomain.StorefrontURLIssue
	if err := r.db.Preload("RouteEntry").First(&issue, id).Error; err != nil {
		return nil, err
	}
	return &issue, nil
}

func (r *StorefrontURLIssueRepository) ListEvents(
	issueID uint,
	page int,
	pageSize int,
) ([]urlmanagementdomain.StorefrontURLIssueEvent, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("storefront URL issue repository is unavailable")
	}
	if issueID == 0 {
		return nil, 0, errors.New("URL issue ID is required")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	query := r.db.Model(&urlmanagementdomain.StorefrontURLIssueEvent{}).Where("issue_id = ?", issueID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var events []urlmanagementdomain.StorefrontURLIssueEvent
	err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&events).Error
	return events, total, err
}

func (r *StorefrontURLIssueRepository) RecordDetection(
	routeEntryID uint,
	issueType string,
	severity string,
	latestCheckResultID *uint,
	detectedAt time.Time,
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront URL issue repository is unavailable")
	}
	if routeEntryID == 0 {
		return nil, errors.New("route entry ID is required")
	}
	if detectedAt.IsZero() {
		detectedAt = time.Now().UTC()
	}

	var result urlmanagementdomain.StorefrontURLIssue
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var issue urlmanagementdomain.StorefrontURLIssue
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("route_entry_id = ? AND issue_type = ?", routeEntryID, issueType).
			First(&issue).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			issue = urlmanagementdomain.StorefrontURLIssue{
				RouteEntryID:        routeEntryID,
				IssueType:           issueType,
				Severity:            severity,
				State:               urlmanagementdomain.URLIssueStateOpen,
				LatestCheckResultID: latestCheckResultID,
				FirstDetectedAt:     detectedAt,
				LastDetectedAt:      detectedAt,
			}
			if err := tx.Create(&issue).Error; err != nil {
				return err
			}
			if err := r.createEvent(tx, issue.ID, urlmanagementdomain.URLIssueEventDetected, 0, "", map[string]interface{}{
				"issue_type":             issueType,
				"severity":               severity,
				"latest_check_result_id": latestCheckResultID,
			}, detectedAt); err != nil {
				return err
			}
			result = issue
			return nil
		}
		if err != nil {
			return err
		}

		updates := map[string]interface{}{
			"severity":         severity,
			"last_detected_at": detectedAt,
			"updated_at":       detectedAt,
		}
		if latestCheckResultID != nil {
			updates["latest_check_result_id"] = *latestCheckResultID
		}
		reopened := issue.State == urlmanagementdomain.URLIssueStateVerified ||
			issue.State == urlmanagementdomain.URLIssueStateResolved ||
			(issue.State == urlmanagementdomain.URLIssueStateSuppressed &&
				issue.SuppressedUntil != nil && !issue.SuppressedUntil.After(detectedAt))
		if reopened {
			updates["state"] = urlmanagementdomain.URLIssueStateOpen
			updates["resolved_at"] = nil
			updates["verified_at"] = nil
			updates["suppressed_until"] = nil
			updates["suppression_reason"] = ""
			updates["resolution_type"] = ""
			updates["resolution_note"] = ""
		}
		if err := tx.Model(&urlmanagementdomain.StorefrontURLIssue{}).
			Where("id = ?", issue.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		if reopened {
			if err := r.createEvent(tx, issue.ID, urlmanagementdomain.URLIssueEventReopened, 0, "", map[string]interface{}{
				"latest_check_result_id": latestCheckResultID,
			}, detectedAt); err != nil {
				return err
			}
		}
		if err := tx.First(&result, issue.ID).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *StorefrontURLIssueRepository) UpdateWithEvent(
	id uint,
	updates map[string]interface{},
	eventType string,
	actorUserID uint,
	note string,
	metadata map[string]interface{},
) (*urlmanagementdomain.StorefrontURLIssue, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront URL issue repository is unavailable")
	}
	if id == 0 {
		return nil, errors.New("URL issue ID is required")
	}
	if updates == nil {
		updates = map[string]interface{}{}
	}

	var result urlmanagementdomain.StorefrontURLIssue
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var issue urlmanagementdomain.StorefrontURLIssue
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&issue, id).Error; err != nil {
			return err
		}
		updates["updated_at"] = time.Now().UTC()
		if len(updates) > 0 {
			if err := tx.Model(&urlmanagementdomain.StorefrontURLIssue{}).
				Where("id = ?", id).
				Updates(updates).Error; err != nil {
				return err
			}
		}
		if eventType != "" {
			if err := r.createEvent(tx, id, eventType, actorUserID, note, metadata, time.Now().UTC()); err != nil {
				return err
			}
		}
		return tx.Preload("RouteEntry").First(&result, id).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *StorefrontURLIssueRepository) createEvent(
	tx *gorm.DB,
	issueID uint,
	eventType string,
	actorUserID uint,
	note string,
	metadata map[string]interface{},
	createdAt time.Time,
) error {
	encodedMetadata, err := encodeStorefrontURLIssueEventMetadata(metadata)
	if err != nil {
		return err
	}
	return tx.Create(&urlmanagementdomain.StorefrontURLIssueEvent{
		IssueID:     issueID,
		EventType:   eventType,
		ActorUserID: actorUserID,
		Note:        note,
		Metadata:    encodedMetadata,
		CreatedAt:   createdAt,
	}).Error
}
