package service

import (
	"errors"
	"strings"
	"tanzanite/internal/domain/suggestionfeedback"
	"tanzanite/internal/repository"
)

const defaultSuggestionAttachmentLevel = "silver"

var (
	ErrSuggestionFeedbackMissingMessage = errors.New("message is required")
	ErrSuggestionFeedbackInvalidStatus  = errors.New("invalid suggestion feedback status")
)

type SuggestionFeedbackService struct {
	suggestionRepo *repository.SuggestionFeedbackRepository
}

type SuggestionEligibility struct {
	LoggedIn      bool    `json:"loggedIn"`
	CanAttach     bool    `json:"canAttach"`
	RequiredLevel string  `json:"requiredLevel"`
	UserLevel     string  `json:"userLevel"`
	Reason        *string `json:"reason"`
}

func NewSuggestionFeedbackService(suggestionRepo *repository.SuggestionFeedbackRepository) *SuggestionFeedbackService {
	return &SuggestionFeedbackService{suggestionRepo: suggestionRepo}
}

func (s *SuggestionFeedbackService) GetEligibility(userID uint, loggedIn bool) (SuggestionEligibility, error) {
	requiredLevel := defaultSuggestionAttachmentLevel
	userLevel := ""
	if loggedIn {
		level, err := s.suggestionRepo.GetUserMemberLevelName(userID)
		if err != nil {
			return SuggestionEligibility{}, err
		}
		userLevel = strings.ToLower(strings.TrimSpace(level))
	}

	canAttach := loggedIn && memberLevelMeetsRequirement(userLevel, requiredLevel)
	var reason *string
	if !loggedIn {
		message := "Please sign in."
		reason = &message
	} else if !canAttach {
		message := "Your current membership level does not support image uploads."
		reason = &message
	}

	return SuggestionEligibility{
		LoggedIn:      loggedIn,
		CanAttach:     canAttach,
		RequiredLevel: requiredLevel,
		UserLevel:     userLevel,
		Reason:        reason,
	}, nil
}

func (s *SuggestionFeedbackService) Create(item *suggestionfeedback.SuggestionFeedback, attachments []suggestionfeedback.Attachment) error {
	item.FullName = strings.TrimSpace(item.FullName)
	item.Email = strings.TrimSpace(item.Email)
	item.Country = strings.TrimSpace(item.Country)
	item.OrderNumber = strings.TrimSpace(item.OrderNumber)
	item.ProductCategory = strings.TrimSpace(item.ProductCategory)
	item.RequestType = strings.TrimSpace(item.RequestType)
	item.Message = strings.TrimSpace(item.Message)

	if item.Message == "" {
		return ErrSuggestionFeedbackMissingMessage
	}
	if item.Status == "" {
		item.Status = "new"
	}
	if !validSuggestionFeedbackStatus(item.Status) {
		return ErrSuggestionFeedbackInvalidStatus
	}
	if item.MemberLevelRequired == "" {
		item.MemberLevelRequired = defaultSuggestionAttachmentLevel
	}
	if !item.MemberLevelMet {
		attachments = []suggestionfeedback.Attachment{}
	}
	item.Attachments = suggestionfeedback.JSONFromAttachments(cleanSuggestionAttachments(attachments))

	return s.suggestionRepo.Create(item)
}

func cleanSuggestionAttachments(attachments []suggestionfeedback.Attachment) []suggestionfeedback.Attachment {
	clean := make([]suggestionfeedback.Attachment, 0, len(attachments))
	for _, attachment := range attachments {
		if strings.TrimSpace(attachment.URL) == "" {
			continue
		}
		clean = append(clean, suggestionfeedback.Attachment{
			Name: strings.TrimSpace(attachment.Name),
			URL:  strings.TrimSpace(attachment.URL),
			Size: attachment.Size,
		})
		if len(clean) == 3 {
			break
		}
	}
	return clean
}

func memberLevelMeetsRequirement(userLevel, requiredLevel string) bool {
	if requiredLevel == "" {
		return true
	}
	if userLevel == "" {
		return false
	}

	hierarchy := []string{"bronze", "silver", "gold", "platinum"}
	userIndex := -1
	requiredIndex := -1
	for index, level := range hierarchy {
		if level == strings.ToLower(userLevel) {
			userIndex = index
		}
		if level == strings.ToLower(requiredLevel) {
			requiredIndex = index
		}
	}
	if requiredIndex == -1 {
		return true
	}
	if userIndex == -1 {
		return false
	}
	return userIndex >= requiredIndex
}

func validSuggestionFeedbackStatus(status string) bool {
	switch status {
	case "new", "in_review", "resolved", "archived":
		return true
	default:
		return false
	}
}
