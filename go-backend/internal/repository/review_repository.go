package repository

import (
	"commerce-platform/internal/domain/review"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) CreateReview(rev *review.Review) error {
	return r.db.Create(rev).Error
}

func (r *ReviewRepository) FindReviewByID(id uint) (*review.Review, error) {
	var rev review.Review
	if err := r.db.Preload("User").First(&rev, id).Error; err != nil {
		return nil, err
	}
	return &rev, nil
}

func (r *ReviewRepository) FindReviewsByProductID(productID uint, page, pageSize int, status string) ([]review.Review, int64, error) {
	var reviews []review.Review
	var total int64

	query := r.db.Model(&review.Review{}).Where("product_id = ?", productID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&reviews).Error

	return reviews, total, err
}

type ReviewAdminListOptions struct {
	Status    string
	Search    string
	ProductID *uint
	Page      int
	PageSize  int
}

func (r *ReviewRepository) FindReviewsForAdmin(options ReviewAdminListOptions) ([]review.Review, int64, error) {
	var reviews []review.Review
	var total int64

	query := r.db.Model(&review.Review{}).
		Joins("LEFT JOIN products ON products.id = reviews.product_id").
		Joins("LEFT JOIN users ON users.id = reviews.user_id")

	if status := strings.TrimSpace(options.Status); status != "" {
		query = query.Where("reviews.status = ?", status)
	}
	if options.ProductID != nil {
		query = query.Where("reviews.product_id = ?", *options.ProductID)
	}
	if search := strings.TrimSpace(options.Search); search != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		query = query.Where(`
			LOWER(reviews.title) LIKE ?
			OR LOWER(reviews.content) LIKE ?
			OR LOWER(products.name) LIKE ?
			OR LOWER(products.sku) LIKE ?
			OR LOWER(users.username) LIKE ?
			OR LOWER(users.email) LIKE ?
		`, pattern, pattern, pattern, pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := options.Page
	if page < 1 {
		page = 1
	}
	pageSize := options.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	err := query.
		Preload("User").
		Preload("Product").
		Order("reviews.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&reviews).Error

	return reviews, total, err
}

func (r *ReviewRepository) FindReviewsByUserID(userID uint, page, pageSize int) ([]review.Review, int64, error) {
	var reviews []review.Review
	var total int64

	query := r.db.Model(&review.Review{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Product").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&reviews).Error

	return reviews, total, err
}

func (r *ReviewRepository) FindFeaturedReviews(limit int) ([]review.Review, error) {
	var reviews []review.Review
	err := r.db.Where("featured = ? AND status = ?", true, review.StatusApproved).
		Preload("User").
		Preload("Product").
		Order("created_at DESC").
		Limit(limit).
		Find(&reviews).Error
	return reviews, err
}

func (r *ReviewRepository) DeleteReview(id uint) error {
	if err := r.db.Where("review_id = ?", id).Delete(&review.ReviewHelpful{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&review.Review{}, id).Error
}

func (r *ReviewRepository) CheckUserReviewExists(userID, productID uint) (bool, error) {
	var count int64
	err := r.db.Model(&review.Review{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

func (r *ReviewRepository) CreateReviewHelpful(h *review.ReviewHelpful) error {
	return r.db.Create(h).Error
}

func (r *ReviewRepository) UpdateReviewHelpful(h *review.ReviewHelpful) error {
	return r.db.Save(h).Error
}

func (r *ReviewRepository) FindReviewHelpful(reviewID, userID uint) (*review.ReviewHelpful, error) {
	var h review.ReviewHelpful
	if err := r.db.Where("review_id = ? AND user_id = ?", reviewID, userID).First(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *ReviewRepository) DeleteReviewHelpful(reviewID, userID uint) error {
	return r.db.Where("review_id = ? AND user_id = ?", reviewID, userID).
		Delete(&review.ReviewHelpful{}).Error
}

func (r *ReviewRepository) CountReviewHelpful(reviewID uint, isHelpful bool) (int64, error) {
	var count int64
	err := r.db.Model(&review.ReviewHelpful{}).
		Where("review_id = ? AND helpful = ?", reviewID, isHelpful).
		Count(&count).Error
	return count, err
}

func (r *ReviewRepository) UpdateReviewHelpfulCounts(reviewID uint) error {
	helpfulCount, err := r.CountReviewHelpful(reviewID, true)
	if err != nil {
		return err
	}

	return r.db.Model(&review.Review{}).Where("id = ?", reviewID).
		Updates(map[string]interface{}{
			"helpful_count": helpfulCount,
		}).Error
}

func (r *ReviewRepository) GetOrCreateReviewSummary(productID uint) (*review.ReviewSummary, error) {
	return r.getOrCreateReviewSummary(r.db, productID)
}

func (r *ReviewRepository) UpdateReviewSummary(productID uint) error {
	return r.updateReviewSummary(r.db, productID)
}

func (r *ReviewRepository) updateReviewSummary(db *gorm.DB, productID uint) error {
	var stats struct {
		TotalReviews  int64   `gorm:"column:total_reviews"`
		AverageRating float64 `gorm:"column:average_rating"`
		Rating1Count  int64   `gorm:"column:rating_1_count"`
		Rating2Count  int64   `gorm:"column:rating_2_count"`
		Rating3Count  int64   `gorm:"column:rating_3_count"`
		Rating4Count  int64   `gorm:"column:rating_4_count"`
		Rating5Count  int64   `gorm:"column:rating_5_count"`
	}

	err := db.Model(&review.Review{}).
		Select(`
			COUNT(*) AS total_reviews,
			COALESCE(AVG(rating), 0) AS average_rating,
			SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END) AS rating_1_count,
			SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END) AS rating_2_count,
			SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END) AS rating_3_count,
			SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END) AS rating_4_count,
			SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END) AS rating_5_count
		`).
		Where("product_id = ? AND status = ?", productID, review.StatusApproved).
		Scan(&stats).Error
	if err != nil {
		return err
	}

	summary, err := r.getOrCreateReviewSummary(db, productID)
	if err != nil {
		return err
	}

	return db.Model(summary).
		Updates(map[string]interface{}{
			"total_reviews":  stats.TotalReviews,
			"average_rating": stats.AverageRating,
			"rating_1_count": stats.Rating1Count,
			"rating_2_count": stats.Rating2Count,
			"rating_3_count": stats.Rating3Count,
			"rating_4_count": stats.Rating4Count,
			"rating_5_count": stats.Rating5Count,
		}).Error
}

func (r *ReviewRepository) getOrCreateReviewSummary(db *gorm.DB, productID uint) (*review.ReviewSummary, error) {
	err := db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&review.ReviewSummary{ProductID: productID}).Error
	if err != nil {
		return nil, err
	}
	var summary review.ReviewSummary
	if err := db.Where("product_id = ?", productID).First(&summary).Error; err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *ReviewRepository) UpdateReviewModeration(
	id uint,
	status string,
	reason string,
	adminID uint,
) (*review.Review, error) {
	var updated review.Review
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current review.Review
		if err := tx.First(&current, id).Error; err != nil {
			return err
		}

		now := time.Now().UTC()
		updates := map[string]interface{}{
			"status":            status,
			"moderated_at":      &now,
			"moderated_by":      adminID,
			"moderation_reason": strings.TrimSpace(reason),
		}
		if status == review.StatusPending {
			updates["moderated_at"] = nil
			updates["moderated_by"] = nil
		}

		if err := tx.Model(&current).Updates(updates).Error; err != nil {
			return err
		}
		if err := r.updateReviewSummary(tx, current.ProductID); err != nil {
			return err
		}
		return tx.Preload("User").Preload("Product").First(&updated, id).Error
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *ReviewRepository) FindReviewSummaryByProductID(productID uint) (*review.ReviewSummary, error) {
	var summary review.ReviewSummary
	if err := r.db.Where("product_id = ?", productID).First(&summary).Error; err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *ReviewRepository) FindReviewSummariesByProductIDs(productIDs []uint) (map[uint]review.ReviewSummary, error) {
	result := make(map[uint]review.ReviewSummary, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}

	var summaries []review.ReviewSummary
	if err := r.db.Where("product_id IN ?", productIDs).Find(&summaries).Error; err != nil {
		return nil, err
	}
	for _, summary := range summaries {
		result[summary.ProductID] = summary
	}
	return result, nil
}
