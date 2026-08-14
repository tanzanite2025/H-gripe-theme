package repository

import (
	"strings"
	"time"

	"commerce-platform/internal/domain/showcase"
	"commerce-platform/internal/domain/user"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ShowcaseRepository struct {
	db *gorm.DB
}

func NewShowcaseRepository(db *gorm.DB) *ShowcaseRepository {
	return &ShowcaseRepository{db: db}
}

func (r *ShowcaseRepository) WithTx(tx *gorm.DB) *ShowcaseRepository {
	return &ShowcaseRepository{db: tx}
}

func (r *ShowcaseRepository) Create(item *showcase.Showcase) error {
	return r.db.Create(item).Error
}

func (r *ShowcaseRepository) WithTransaction(fn func(repo *ShowcaseRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(r.WithTx(tx))
	})
}

func (r *ShowcaseRepository) LockUserForSubmissionLimit(userID uint) error {
	if userID == 0 {
		return gorm.ErrRecordNotFound
	}
	switch r.db.Dialector.Name() {
	case "postgres", "mysql", "sqlserver":
		var item user.User
		return r.db.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id").
			First(&item, userID).Error
	default:
		result := r.db.
			Model(&user.User{}).
			Where("id = ?", userID).
			Update("updated_at", time.Now().UTC())
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}
}

func (r *ShowcaseRepository) GetByID(id uint) (*showcase.Showcase, error) {
	var item showcase.Showcase
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ShowcaseRepository) List(kind string, status string, limit int, offset int) ([]showcase.Showcase, error) {
	var items []showcase.Showcase
	query := r.db.Model(&showcase.Showcase{})
	if kind != "" {
		query = query.Where("kind = ?", kind)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&items).Error
	return items, err
}

func (r *ShowcaseRepository) Count(kind string, status string) (int64, error) {
	var count int64
	query := r.db.Model(&showcase.Showcase{})
	if kind != "" {
		query = query.Where("kind = ?", kind)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *ShowcaseRepository) CountByUserAndStatus(userID uint, status string) (int64, error) {
	var count int64
	err := r.db.Model(&showcase.Showcase{}).
		Where("user_id = ? AND status = ?", userID, status).
		Count(&count).Error
	return count, err
}

func (r *ShowcaseRepository) ListByImageStorageKeyCandidate(key string) ([]showcase.Showcase, error) {
	var items []showcase.Showcase
	if key == "" {
		return items, nil
	}
	escapedKey := escapeShowcaseImageSearchPattern(key)
	imagesLikeClause := showcaseImagesLikeClause(r.db.Dialector.Name())
	err := r.db.Model(&showcase.Showcase{}).
		Where(imagesLikeClause, "%"+escapedKey+"%").
		Limit(25).
		Find(&items).Error
	return items, err
}

func showcaseImagesLikeClause(dialect string) string {
	switch dialect {
	case "postgres":
		return "CAST(images AS TEXT) LIKE ? ESCAPE '\\'"
	case "mysql":
		return "CAST(images AS CHAR) LIKE ?"
	default:
		return "images LIKE ? ESCAPE '\\'"
	}
}

func escapeShowcaseImageSearchPattern(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	return strings.ReplaceAll(value, `_`, `\_`)
}

func (r *ShowcaseRepository) UpdateStatus(id uint, status string, reason string) error {
	updates := map[string]interface{}{
		"status":          status,
		"rejected_reason": reason,
	}
	if status == showcase.StatusApproved {
		approvedAt := time.Now()
		updates["approved_at"] = &approvedAt
	}
	return r.db.Model(&showcase.Showcase{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ShowcaseRepository) UpdatePendingStatus(id uint, status string, reason string) (bool, error) {
	updates := map[string]interface{}{
		"status":          status,
		"rejected_reason": reason,
	}
	if status == showcase.StatusApproved {
		approvedAt := time.Now()
		updates["approved_at"] = &approvedAt
	}
	result := r.db.Model(&showcase.Showcase{}).
		Where("id = ? AND status = ?", id, showcase.StatusPending).
		Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *ShowcaseRepository) UpdateImagesAndStatus(id uint, images datatypes.JSON, status string, reason string) error {
	updates := map[string]interface{}{
		"images":          images,
		"status":          status,
		"rejected_reason": reason,
	}
	if status == showcase.StatusApproved {
		approvedAt := time.Now()
		updates["approved_at"] = &approvedAt
	}
	return r.db.Model(&showcase.Showcase{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ShowcaseRepository) UpdatePendingImagesAndStatus(id uint, images datatypes.JSON, status string, reason string) (bool, error) {
	updates := map[string]interface{}{
		"images":          images,
		"status":          status,
		"rejected_reason": reason,
	}
	if status == showcase.StatusApproved {
		approvedAt := time.Now()
		updates["approved_at"] = &approvedAt
	}
	result := r.db.Model(&showcase.Showcase{}).
		Where("id = ? AND status = ?", id, showcase.StatusPending).
		Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *ShowcaseRepository) ListImageCleanupCandidates(cutoff time.Time, limit int) ([]showcase.Showcase, error) {
	var items []showcase.Showcase
	if limit <= 0 {
		return items, nil
	}
	imagesLikeClause := showcaseImagesLikeClause(r.db.Dialector.Name())
	err := r.db.Model(&showcase.Showcase{}).
		Where(
			"(status = ? AND created_at < ?) OR (status = ? AND "+imagesLikeClause+")",
			showcase.StatusPending,
			cutoff,
			showcase.StatusRejected,
			"%showcase/pending/%",
		).
		Order("created_at asc").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *ShowcaseRepository) UpdateImagesByStatus(id uint, status string, images datatypes.JSON) (bool, error) {
	result := r.db.Model(&showcase.Showcase{}).
		Where("id = ? AND status = ?", id, status).
		Update("images", images)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *ShowcaseRepository) CreateComment(comment *showcase.Comment) error {
	return r.db.Create(comment).Error
}

func (r *ShowcaseRepository) ListComments(showcaseID uint, limit int, offset int) ([]showcase.Comment, error) {
	var comments []showcase.Comment
	err := r.db.Where("showcase_id = ? AND status = ?", showcaseID, showcase.StatusApproved).
		Order("created_at desc").
		Limit(limit).Offset(offset).
		Find(&comments).Error
	return comments, err
}

func (r *ShowcaseRepository) ListAllComments(showcaseID uint) ([]showcase.Comment, error) {
	var comments []showcase.Comment
	query := r.db.Model(&showcase.Comment{})
	if showcaseID > 0 {
		query = query.Where("showcase_id = ?", showcaseID)
	}
	err := query.Order("created_at desc").Find(&comments).Error
	return comments, err
}

func (r *ShowcaseRepository) UpdateCommentStatus(id uint, status string) error {
	return r.db.Model(&showcase.Comment{}).Where("id = ?", id).Update("status", status).Error
}
