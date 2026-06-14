package service

import (
	"errors"
	"strings"
	"tanzanite/internal/domain/feedback"
	"tanzanite/internal/repository"
)

var (
	ErrFeedbackMissingThread  = errors.New("thread is required")
	ErrFeedbackMissingContent = errors.New("content is required")
	ErrFeedbackInvalidStatus  = errors.New("invalid feedback status")
)

type FeedbackService struct {
	feedbackRepo *repository.FeedbackRepository
}

func NewFeedbackService(feedbackRepo *repository.FeedbackRepository) *FeedbackService {
	return &FeedbackService{feedbackRepo: feedbackRepo}
}

func (s *FeedbackService) List(threadKey, status, search string, page, pageSize int) ([]feedback.Feedback, int64, error) {
	threadKey = strings.TrimSpace(threadKey)
	if threadKey == "" {
		return nil, 0, ErrFeedbackMissingThread
	}
	if status != "" && !validFeedbackStatus(status) {
		return nil, 0, ErrFeedbackInvalidStatus
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.feedbackRepo.List(threadKey, status, strings.TrimSpace(search), page, pageSize)
}

func (s *FeedbackService) Create(item *feedback.Feedback) error {
	item.ThreadKey = strings.TrimSpace(item.ThreadKey)
	item.Content = strings.TrimSpace(item.Content)
	item.Name = strings.TrimSpace(item.Name)
	item.Email = strings.TrimSpace(item.Email)
	item.Locale = strings.TrimSpace(item.Locale)

	if item.ThreadKey == "" {
		return ErrFeedbackMissingThread
	}
	if item.Content == "" {
		return ErrFeedbackMissingContent
	}
	if item.Status == "" {
		item.Status = "pending"
	}
	if !validFeedbackStatus(item.Status) {
		return ErrFeedbackInvalidStatus
	}

	return s.feedbackRepo.Create(item)
}

func (s *FeedbackService) UpdateStatus(id uint, status string) error {
	status = strings.TrimSpace(status)
	if !validFeedbackStatus(status) {
		return ErrFeedbackInvalidStatus
	}
	return s.feedbackRepo.UpdateStatus(id, status)
}

func validFeedbackStatus(status string) bool {
	switch status {
	case "pending", "approved", "rejected", "hidden":
		return true
	default:
		return false
	}
}
