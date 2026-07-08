package repository

import (
	"errors"
	"tanzanite/internal/domain/review"

	"gorm.io/gorm"
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
	err := r.db.Where("featured = ? AND status = ?", true, "approved").
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

	notHelpfulCount, err := r.CountReviewHelpful(reviewID, false)
	if err != nil {
		return err
	}

	return r.db.Model(&review.Review{}).Where("id = ?", reviewID).
		Updates(map[string]interface{}{
			"helpful_count":     helpfulCount,
			"not_helpful_count": notHelpfulCount,
		}).Error
}

func (r *ReviewRepository) GetOrCreateReviewSummary(productID uint) (*review.ReviewSummary, error) {
	var summary review.ReviewSummary
	err := r.db.Where("product_id = ?", productID).First(&summary).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		summary = review.ReviewSummary{ProductID: productID}
		if err := r.db.Create(&summary).Error; err != nil {
			return nil, err
		}
		return &summary, nil
	}
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *ReviewRepository) UpdateReviewSummary(productID uint) error {
	var stats struct {
		TotalCount    int64
		AverageRating float64
		Rating1Count  int64
		Rating2Count  int64
		Rating3Count  int64
		Rating4Count  int64
		Rating5Count  int64
	}

	approvedReviews := r.db.Model(&review.Review{}).Where("product_id = ? AND status = ?", productID, "approved")
	approvedReviews.Count(&stats.TotalCount)
	approvedReviews.Select("AVG(rating)").Scan(&stats.AverageRating)

	for rating := 1; rating <= 5; rating++ {
		var count int64
		r.db.Model(&review.Review{}).
			Where("product_id = ? AND status = ? AND rating = ?", productID, "approved", rating).
			Count(&count)

		switch rating {
		case 1:
			stats.Rating1Count = count
		case 2:
			stats.Rating2Count = count
		case 3:
			stats.Rating3Count = count
		case 4:
			stats.Rating4Count = count
		case 5:
			stats.Rating5Count = count
		}
	}

	return r.db.Model(&review.ReviewSummary{}).Where("product_id = ?", productID).
		Updates(map[string]interface{}{
			"total_count":    stats.TotalCount,
			"average_rating": stats.AverageRating,
			"rating_1_count": stats.Rating1Count,
			"rating_2_count": stats.Rating2Count,
			"rating_3_count": stats.Rating3Count,
			"rating_4_count": stats.Rating4Count,
			"rating_5_count": stats.Rating5Count,
		}).Error
}

func (r *ReviewRepository) FindReviewSummaryByProductID(productID uint) (*review.ReviewSummary, error) {
	var summary review.ReviewSummary
	if err := r.db.Where("product_id = ?", productID).First(&summary).Error; err != nil {
		return nil, err
	}
	return &summary, nil
}
