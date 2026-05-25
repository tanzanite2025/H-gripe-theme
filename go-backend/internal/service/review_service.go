package service

import (
	"errors"
	"tanzanite/internal/domain/review"
	"tanzanite/internal/repository"
)

type ReviewService struct {
	reviewRepo *repository.ReviewRepository
}

func NewReviewService(reviewRepo *repository.ReviewRepository) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
	}
}

// CreateReview 创建评价
func (s *ReviewService) CreateReview(r *review.Review) error {
	// 检查用户是否已评价该产品
	exists, err := s.reviewRepo.CheckUserReviewExists(r.UserID, r.ProductID)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("you have already reviewed this product")
	}

	// 设置默认状态为待审核
	r.Status = "pending"

	if err := s.reviewRepo.CreateReview(r); err != nil {
		return err
	}

	// 更新产品评价摘要
	return s.reviewRepo.UpdateReviewSummary(r.ProductID)
}

// GetReview 获取评价详情
func (s *ReviewService) GetReview(id uint) (*review.Review, error) {
	return s.reviewRepo.FindReviewByID(id)
}

// GetProductReviews 获取产品评价列表
func (s *ReviewService) GetProductReviews(productID uint, page, pageSize int, status string) ([]review.Review, int64, error) {
	// 如果没有指定状态，默认只显示已审核的
	if status == "" {
		status = "approved"
	}

	return s.reviewRepo.FindReviewsByProductID(productID, page, pageSize, status)
}

// GetUserReviews 获取用户评价列表
func (s *ReviewService) GetUserReviews(userID uint, page, pageSize int) ([]review.Review, int64, error) {
	return s.reviewRepo.FindReviewsByUserID(userID, page, pageSize)
}

// GetFeaturedReviews 获取精选评价
func (s *ReviewService) GetFeaturedReviews(limit int) ([]review.Review, error) {
	return s.reviewRepo.FindFeaturedReviews(limit)
}

// GetPendingReviews 获取待审核评价（管理员）
func (s *ReviewService) GetPendingReviews(page, pageSize int) ([]review.Review, int64, error) {
	return s.reviewRepo.FindPendingReviews(page, pageSize)
}

// ApproveReview 审核通过评价
func (s *ReviewService) ApproveReview(id uint) error {
	r, err := s.reviewRepo.FindReviewByID(id)
	if err != nil {
		return err
	}

	if err := s.reviewRepo.UpdateReviewStatus(id, "approved"); err != nil {
		return err
	}

	// 更新产品评价摘要
	return s.reviewRepo.UpdateReviewSummary(r.ProductID)
}

// RejectReview 拒绝评价
func (s *ReviewService) RejectReview(id uint) error {
	return s.reviewRepo.UpdateReviewStatus(id, "rejected")
}

// SetFeatured 设置精选评价
func (s *ReviewService) SetFeatured(id uint, featured bool) error {
	r, err := s.reviewRepo.FindReviewByID(id)
	if err != nil {
		return err
	}

	r.Featured = featured
	return s.reviewRepo.UpdateReview(r)
}

// DeleteReview 删除评价
func (s *ReviewService) DeleteReview(id uint, userID uint) error {
	r, err := s.reviewRepo.FindReviewByID(id)
	if err != nil {
		return err
	}

	// 验证权限
	if r.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := s.reviewRepo.DeleteReview(id); err != nil {
		return err
	}

	// 更新产品评价摘要
	return s.reviewRepo.UpdateReviewSummary(r.ProductID)
}

// MarkHelpful 标记评价有用
func (s *ReviewService) MarkHelpful(reviewID, userID uint, isHelpful bool) error {
	// 检查是否已标记
	existing, _ := s.reviewRepo.FindReviewHelpful(reviewID, userID)
	if existing != nil {
		// 如果已存在且标记相同，则取消标记
		if existing.Helpful == isHelpful {
			if err := s.reviewRepo.DeleteReviewHelpful(reviewID, userID); err != nil {
				return err
			}
		} else {
			// 更新标记
			existing.Helpful = isHelpful
			if err := s.reviewRepo.UpdateReviewHelpful(existing); err != nil {
				return err
			}
		}
	} else {
		// 创建新标记
		helpful := &review.ReviewHelpful{
			ReviewID: reviewID,
			UserID:   userID,
			Helpful:  isHelpful,
		}
		if err := s.reviewRepo.CreateReviewHelpful(helpful); err != nil {
			return err
		}
	}

	// 更新评价的有用统计
	return s.reviewRepo.UpdateReviewHelpfulCounts(reviewID)
}

// GetReviewSummary 获取产品评价摘要
func (s *ReviewService) GetReviewSummary(productID uint) (*review.ReviewSummary, error) {
	summary, err := s.reviewRepo.FindReviewSummaryByProductID(productID)
	if err != nil {
		// 如果不存在，创建一个
		return s.reviewRepo.GetOrCreateReviewSummary(productID)
	}
	return summary, nil
}
