package service

import "commerce-platform/internal/repository"

// Missing summaries have the same empty public counters as the ordinary
// create-on-read path, without persisting a row. Other query failures propagate.
// The recovery adapter must also enforce read-only SQL and allow only GETs.
func NewReviewServiceForReadOnlyRecoveryVerification(reviewRepository *repository.ReviewRepository) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepository, readOnlyRecoveryVerification: true}
}
