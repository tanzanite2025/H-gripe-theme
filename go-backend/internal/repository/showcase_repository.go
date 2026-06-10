package repository

import (
	"tanzanite/internal/domain/showcase"

	"gorm.io/gorm"
)

type ShowcaseRepository struct {
	db *gorm.DB
}

func NewShowcaseRepository(db *gorm.DB) *ShowcaseRepository {
	return &ShowcaseRepository{db: db}
}

func (r *ShowcaseRepository) Create(item *showcase.Showcase) error {
	return r.db.Create(item).Error
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

func (r *ShowcaseRepository) UpdateStatus(id uint, status string, reason string) error {
	updates := map[string]interface{}{
		"status": status,
		"rejected_reason": reason,
	}
	if status == showcase.StatusApproved {
		updates["approved_at"] = gorm.Expr("NOW()")
	}
	return r.db.Model(&showcase.Showcase{}).Where("id = ?", id).Updates(updates).Error
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
