package repository

import (
	"tanzanite/internal/domain/review"

	"gorm.io/gorm"
)

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Review 相关方法

// CreateReview 创建评价
func (r *ReviewRepository) CreateReview(rev *review.Review) error {
	return r.db.Create(rev).Error
}

// FindReviewByID 根据ID查找评价
func (r *ReviewRepository) FindReviewByID(id uint) (*review.Review, error) {
	var rev review.Review
	err := r.db.Preload("User").First(&rev, id).Error
	if err != nil {
		return nil, err
	}
	return &rev, nil
}

// FindReviewsByProductID 查找产品的评价列表
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
	err := query.Preload("User").Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&reviews).Error

	return reviews, total, err
}

// FindReviewsByUserID 查找用户的评价列表
func (r *ReviewRepository) FindReviewsByUserID(userID uint, page, pageSize int) ([]review.Review, int64, error) {
	var reviews []review.Review
	var total int64

	query := r.db.Model(&review.Review{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Product").Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&reviews).Error

	return reviews, total, err
}

// FindFeaturedReviews 查找精选评价
func (r *ReviewRepository) FindFeaturedReviews(limit int) ([]review.Review, error) {
	var reviews []review.Review
	err := r.db.Where("featured = ? AND status = ?", true, "approved").
		Preload("User").Preload("Product").
		Order("created_at DESC").Limit(limit).Find(&reviews).Error
	return reviews, err
}

// FindPendingReviews 查找待审核评价
func (r *ReviewRepository) FindPendingReviews(page, pageSize int) ([]review.Review, int64, error) {
	var reviews []review.Review
	var total int64

	query := r.db.Model(&review.Review{}).Where("status = ?", "pending")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("User").Preload("Product").Order("created_at ASC").
		Offset(offset).Limit(pageSize).Find(&reviews).Error

	return reviews, total, err
}

// UpdateReview 更新评价
func (r *ReviewRepository) UpdateReview(rev *review.Review) error {
	return r.db.Save(rev).Error
}

// UpdateReviewStatus 更新评价状态
func (r *ReviewRepository) UpdateReviewStatus(id uint, status string) error {
	return r.db.Model(&review.Review{}).Where("id = ?", id).
		Update("status", status).Error
}

// DeleteReview 删除评价
func (r *ReviewRepository) DeleteReview(id uint) error {
	// 先删除关联的有用记录
	if err := r.db.Where("review_id = ?", id).Delete(&review.ReviewHelpful{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&review.Review{}, id).Error
}

// CheckUserReviewExists 检查用户是否已评价该产品
func (r *ReviewRepository) CheckUserReviewExists(userID, productID uint) (bool, error) {
	var count int64
	err := r.db.Model(&review.Review{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

// ReviewHelpful 相关方法

// CreateReviewHelpful 创建有用记录
func (r *ReviewRepository) CreateReviewHelpful(h *review.ReviewHelpful) error {
	return r.db.Create(h).Error
}

// UpdateReviewHelpful 更新有用标记
func (r *ReviewRepository) UpdateReviewHelpful(h *review.ReviewHelpful) error {
	return r.db.Save(h).Error
}

// FindReviewHelpful 查找有用记录
func (r *ReviewRepository) FindReviewHelpful(reviewID, userID uint) (*review.ReviewHelpful, error) {
	var h review.ReviewHelpful
	err := r.db.Where("review_id = ? AND user_id = ?", reviewID, userID).First(&h).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// DeleteReviewHelpful 删除有用记录
func (r *ReviewRepository) DeleteReviewHelpful(reviewID, userID uint) error {
	return r.db.Where("review_id = ? AND user_id = ?", reviewID, userID).
		Delete(&review.ReviewHelpful{}).Error
}

// CountReviewHelpful 统计评价的有用数
func (r *ReviewRepository) CountReviewHelpful(reviewID uint, isHelpful bool) (int64, error) {
	var count int64
	err := r.db.Model(&review.ReviewHelpful{}).
		Where("review_id = ? AND helpful = ?", reviewID, isHelpful).
		Count(&count).Error
	return count, err
}

// UpdateReviewHelpfulCounts 更新评价的有用统计
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

// ReviewSummary 相关方法

// GetOrCreateReviewSummary 获取或创建评价摘要
func (r *ReviewRepository) GetOrCreateReviewSummary(productID uint) (*review.ReviewSummary, error) {
	var summary review.ReviewSummary
	err := r.db.Where("product_id = ?", productID).First(&summary).Error

	if err == gorm.ErrRecordNotFound {
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

// UpdateReviewSummary 更新评价摘要
func (r *ReviewRepository) UpdateReviewSummary(productID uint) error {
	// 计算评价统计
	var stats struct {
		TotalCount    int64
		AverageRating float64
		Rating1Count  int64
		Rating2Count  int64
		Rating3Count  int64
		Rating4Count  int64
		Rating5Count  int64
	}

	// 总数和平均分
	r.db.Model(&review.Review{}).
		Where("product_id = ? AND status = ?", productID, "approved").
		Count(&stats.TotalCount)

	r.db.Model(&review.Review{}).
		Where("product_id = ? AND status = ?", productID, "approved").
		Select("AVG(rating)").Scan(&stats.AverageRating)

	// 各星级数量
	r.db.Model(&review.Review{}).
		Where("product_id = ? AND status = ? AND rating = ?", productID, "approved", 1).
		Count(&stats.Rating1Count)

	r.db.Model(&review.Review{}).
		Where("product_id = ? AND status = ? AND rating = ?", productID, "approved", 2).
		Count(&stats.Rating2Count)

	r.db.Model(&review.Review{}).
		Where("product_id = ? AND status = ? AND rating = ?", productID, "approved", 3).
		Count(&stats.Rating3Count)

	r.db.Model(&review.Review{}).
		Where("product_id = ? AND status = ? AND rating = ?", productID, "approved", 4).
		Count(&stats.Rating4Count)

	r.db.Model(&review.Review{}).
		Where("product_id = ? AND status = ? AND rating = ?", productID, "approved", 5).
		Count(&stats.Rating5Count)

	// 更新摘要
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

// FindReviewSummaryByProductID 根据产品ID查找评价摘要
func (r *ReviewRepository) FindReviewSummaryByProductID(productID uint) (*review.ReviewSummary, error) {
	var summary review.ReviewSummary
	err := r.db.Where("product_id = ?", productID).First(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}
