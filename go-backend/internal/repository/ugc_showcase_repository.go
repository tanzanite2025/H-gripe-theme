package repository

import (
	"context"
	"time"

	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/domain/user"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UGCShowcaseRepository struct {
	db                    *gorm.DB
	validateImageBindings func(context.Context, *ugcshowcase.UGCShowcase, *ugcshowcase.UGCShowcase) error
	resolveImageObjectKey func(string) (string, error)
}

func NewUGCShowcaseRepository(db *gorm.DB) *UGCShowcaseRepository {
	return &UGCShowcaseRepository{db: db}
}

func (r *UGCShowcaseRepository) WithTx(tx *gorm.DB) *UGCShowcaseRepository {
	copy := *r
	copy.db = tx
	return &copy
}

func (r *UGCShowcaseRepository) Create(item *ugcshowcase.UGCShowcase) error {
	return r.WithTransaction(func(repo *UGCShowcaseRepository, tx *gorm.DB) error {
		if item == nil || repo.validateImageBindings == nil {
			return ErrUGCShowcaseImageOwnershipValidationUnavailable
		}
		if err := repo.validateImageBindings(tx.Statement.Context, nil, item); err != nil {
			return err
		}
		return tx.Create(item).Error
	})
}

func (r *UGCShowcaseRepository) WithTransaction(fn func(repo *UGCShowcaseRepository, tx *gorm.DB) error) error {
	return runWithDedicatedPublicObjectNamespaceTransactionLock(r.db.Statement.Context, r.db, dedicatedUGCShowcaseLifecycleNamespace, func(tx *gorm.DB) error {
		return fn(r.WithTx(tx), tx)
	})
}

func (r *UGCShowcaseRepository) LockUserForSubmissionLimit(userID uint) error {
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

func (r *UGCShowcaseRepository) GetByID(id uint) (*ugcshowcase.UGCShowcase, error) {
	var item ugcshowcase.UGCShowcase
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *UGCShowcaseRepository) List(kind string, status string, limit int, offset int) ([]ugcshowcase.UGCShowcase, error) {
	var items []ugcshowcase.UGCShowcase
	query := r.db.Model(&ugcshowcase.UGCShowcase{})
	if kind != "" {
		query = query.Where("kind = ?", kind)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&items).Error
	return items, err
}

func (r *UGCShowcaseRepository) Count(kind string, status string) (int64, error) {
	var count int64
	query := r.db.Model(&ugcshowcase.UGCShowcase{})
	if kind != "" {
		query = query.Where("kind = ?", kind)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *UGCShowcaseRepository) CountByUserAndStatus(userID uint, status string) (int64, error) {
	var count int64
	err := r.db.Model(&ugcshowcase.UGCShowcase{}).
		Where("user_id = ? AND status = ?", userID, status).
		Count(&count).Error
	return count, err
}

func (r *UGCShowcaseRepository) UpdateStatus(id uint, status string, reason string) error {
	_, err := r.mutateUGCShowcaseImagesAndStatusWithNamespaceAndCurrentRowLocks(id, "", nil, &status, reason)
	return err
}

func (r *UGCShowcaseRepository) UpdatePendingStatus(id uint, status string, reason string) (bool, error) {
	return r.mutateUGCShowcaseImagesAndStatusWithNamespaceAndCurrentRowLocks(id, ugcshowcase.StatusPending, nil, &status, reason)
}

func (r *UGCShowcaseRepository) UpdateImagesAndStatus(id uint, images datatypes.JSON, status string, reason string) error {
	_, err := r.mutateUGCShowcaseImagesAndStatusWithNamespaceAndCurrentRowLocks(id, "", &images, &status, reason)
	return err
}

func (r *UGCShowcaseRepository) UpdatePendingImagesAndStatus(id uint, images datatypes.JSON, status string, reason string) (bool, error) {
	return r.mutateUGCShowcaseImagesAndStatusWithNamespaceAndCurrentRowLocks(id, ugcshowcase.StatusPending, &images, &status, reason)
}

func (r *UGCShowcaseRepository) ListImageCleanupCandidates(cutoff time.Time, limit int) ([]ugcshowcase.UGCShowcase, error) {
	return r.listExactPendingUGCShowcaseCleanupCandidates(cutoff, limit)
}

func (r *UGCShowcaseRepository) UpdateImagesByStatus(id uint, status string, images datatypes.JSON) (bool, error) {
	return r.mutateUGCShowcaseImagesAndStatusWithNamespaceAndCurrentRowLocks(id, status, &images, nil, "")
}

func (r *UGCShowcaseRepository) CreateComment(comment *ugcshowcase.UGCShowcaseComment) error {
	return r.db.Create(comment).Error
}

func (r *UGCShowcaseRepository) ListComments(showcaseID uint, limit int, offset int) ([]ugcshowcase.UGCShowcaseComment, error) {
	var comments []ugcshowcase.UGCShowcaseComment
	err := r.db.Where("showcase_id = ? AND status = ?", showcaseID, ugcshowcase.StatusApproved).
		Order("created_at desc").
		Limit(limit).Offset(offset).
		Find(&comments).Error
	return comments, err
}

func (r *UGCShowcaseRepository) ListAllComments(showcaseID uint) ([]ugcshowcase.UGCShowcaseComment, error) {
	var comments []ugcshowcase.UGCShowcaseComment
	query := r.db.Model(&ugcshowcase.UGCShowcaseComment{})
	if showcaseID > 0 {
		query = query.Where("showcase_id = ?", showcaseID)
	}
	err := query.Order("created_at desc").Find(&comments).Error
	return comments, err
}

func (r *UGCShowcaseRepository) UpdateCommentStatus(id uint, status string) error {
	return r.db.Model(&ugcshowcase.UGCShowcaseComment{}).Where("id = ?", id).Update("status", status).Error
}
