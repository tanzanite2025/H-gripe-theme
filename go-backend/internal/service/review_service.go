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

func (s *ReviewService) CreateReview(r *review.Review) error {
	exists, err := s.reviewRepo.CheckUserReviewExists(r.UserID, r.ProductID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("you have already reviewed this product")
	}

	r.Status = "pending"
	if err := s.reviewRepo.CreateReview(r); err != nil {
		return err
	}

	return s.reviewRepo.UpdateReviewSummary(r.ProductID)
}

func (s *ReviewService) GetReview(id uint) (*review.Review, error) {
	return s.reviewRepo.FindReviewByID(id)
}

func (s *ReviewService) GetProductReviews(productID uint, page, pageSize int, status string) ([]review.Review, int64, error) {
	if status == "" {
		status = "approved"
	}
	return s.reviewRepo.FindReviewsByProductID(productID, page, pageSize, status)
}

func (s *ReviewService) GetUserReviews(userID uint, page, pageSize int) ([]review.Review, int64, error) {
	return s.reviewRepo.FindReviewsByUserID(userID, page, pageSize)
}

func (s *ReviewService) GetFeaturedReviews(limit int) ([]review.Review, error) {
	return s.reviewRepo.FindFeaturedReviews(limit)
}

func (s *ReviewService) DeleteReview(id uint, userID uint) error {
	r, err := s.reviewRepo.FindReviewByID(id)
	if err != nil {
		return err
	}
	if r.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := s.reviewRepo.DeleteReview(id); err != nil {
		return err
	}

	return s.reviewRepo.UpdateReviewSummary(r.ProductID)
}

func (s *ReviewService) MarkHelpful(reviewID, userID uint, isHelpful bool) error {
	existing, _ := s.reviewRepo.FindReviewHelpful(reviewID, userID)
	if existing != nil {
		if existing.Helpful == isHelpful {
			if err := s.reviewRepo.DeleteReviewHelpful(reviewID, userID); err != nil {
				return err
			}
		} else {
			existing.Helpful = isHelpful
			if err := s.reviewRepo.UpdateReviewHelpful(existing); err != nil {
				return err
			}
		}
	} else {
		helpful := &review.ReviewHelpful{
			ReviewID: reviewID,
			UserID:   userID,
			Helpful:  isHelpful,
		}
		if err := s.reviewRepo.CreateReviewHelpful(helpful); err != nil {
			return err
		}
	}

	return s.reviewRepo.UpdateReviewHelpfulCounts(reviewID)
}

func (s *ReviewService) GetReviewSummary(productID uint) (*review.ReviewSummary, error) {
	summary, err := s.reviewRepo.FindReviewSummaryByProductID(productID)
	if err != nil {
		return s.reviewRepo.GetOrCreateReviewSummary(productID)
	}
	return summary, nil
}
