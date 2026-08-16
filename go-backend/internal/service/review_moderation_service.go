package service

import (
	"errors"
	"strings"

	"commerce-platform/internal/domain/review"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrReviewModerationNotFound = gorm.ErrRecordNotFound
	ErrInvalidReviewStatus      = errors.New("invalid review status")
	ErrInvalidReviewTransition  = errors.New("invalid review status transition")
	ErrReviewModerationReason   = errors.New("moderation reason is required when rejecting a review")
)

type ReviewModerationService struct {
	reviewRepo *repository.ReviewRepository
}

func NewReviewModerationService(reviewRepo *repository.ReviewRepository) *ReviewModerationService {
	return &ReviewModerationService{reviewRepo: reviewRepo}
}

func (s *ReviewModerationService) List(
	status string,
	search string,
	productID *uint,
	page int,
	pageSize int,
) ([]review.Review, int64, error) {
	normalizedStatus := normalizeReviewStatusFilter(status)
	return s.reviewRepo.FindReviewsForAdmin(repository.ReviewAdminListOptions{
		Status:    normalizedStatus,
		Search:    search,
		ProductID: productID,
		Page:      page,
		PageSize:  pageSize,
	})
}

func (s *ReviewModerationService) Get(id uint) (*review.Review, error) {
	return s.reviewRepo.FindReviewByID(id)
}

func (s *ReviewModerationService) UpdateStatus(
	id uint,
	nextStatus string,
	reason string,
	adminID uint,
) (*review.Review, error) {
	nextStatus = strings.ToLower(strings.TrimSpace(nextStatus))
	if !isReviewStatus(nextStatus) {
		return nil, ErrInvalidReviewStatus
	}
	if nextStatus == review.StatusRejected && strings.TrimSpace(reason) == "" {
		return nil, ErrReviewModerationReason
	}

	current, err := s.reviewRepo.FindReviewByID(id)
	if err != nil {
		return nil, err
	}
	if !canTransitionReviewStatus(current.Status, nextStatus) {
		return nil, ErrInvalidReviewTransition
	}

	return s.reviewRepo.UpdateReviewModeration(id, nextStatus, reason, adminID)
}

func normalizeReviewStatusFilter(status string) string {
	switch value := strings.ToLower(strings.TrimSpace(status)); value {
	case "", "all", "*":
		return ""
	case review.StatusPending, review.StatusApproved, review.StatusRejected:
		return value
	default:
		return review.StatusPending
	}
}

func isReviewStatus(status string) bool {
	switch status {
	case review.StatusPending, review.StatusApproved, review.StatusRejected:
		return true
	default:
		return false
	}
}

func canTransitionReviewStatus(current, next string) bool {
	if current == next {
		return true
	}

	switch current {
	case review.StatusPending:
		return next == review.StatusApproved || next == review.StatusRejected
	case review.StatusApproved, review.StatusRejected:
		return next == review.StatusPending
	default:
		return false
	}
}
