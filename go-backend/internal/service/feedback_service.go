package service

import (
	"commerce-platform/internal/domain/feedback"
	"commerce-platform/internal/pkg/ugc"
	"commerce-platform/internal/repository"
	"errors"
	"strings"
)

var (
	ErrFeedbackMissingThread  = errors.New("thread is required")
	ErrFeedbackMissingContent = errors.New("content is required")
	ErrFeedbackContentTooLong = errors.New("feedback content is too long")
	ErrFeedbackNameTooLong    = errors.New("feedback name is too long")
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

func (s *FeedbackService) ListPublic(threadKey, search string, page, pageSize int) ([]feedback.Feedback, int64, error) {
	items, total, err := s.List(threadKey, "approved", search, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for index := range items {
		items[index].Name = normalizeFeedbackPublicText(items[index].Name)
		items[index].Content = normalizeFeedbackPublicText(items[index].Content)
	}
	return items, total, nil
}

func (s *FeedbackService) Create(item *feedback.Feedback) error {
	item.ThreadKey = strings.TrimSpace(item.ThreadKey)
	content, err := ugc.PlainText(item.Content, 3000)
	if errors.Is(err, ugc.ErrTextTooLong) {
		return ErrFeedbackContentTooLong
	}
	if err != nil {
		return err
	}
	name, err := ugc.PlainText(item.Name, 120)
	if errors.Is(err, ugc.ErrTextTooLong) {
		return ErrFeedbackNameTooLong
	}
	if err != nil {
		return err
	}
	item.Content = content
	item.Name = name
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

func normalizeFeedbackPublicText(value string) string {
	normalized, err := ugc.PlainText(value, 0)
	if err != nil {
		return ""
	}
	return normalized
}
