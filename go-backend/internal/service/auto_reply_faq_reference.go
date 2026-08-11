package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/ticket"
)

func (s *TicketService) validateAutoReplyFAQReference(rule ticket.AutoReplyRule) error {
	if strings.ToLower(strings.TrimSpace(rule.MessageType)) != "faq" {
		return nil
	}
	if s.faqRepo == nil {
		return fmt.Errorf("%w: FAQ repository is unavailable", ErrInvalidAutoReplyRule)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(rule.Metadata), &payload); err != nil {
		return fmt.Errorf("%w: FAQ metadata is invalid", ErrInvalidAutoReplyRule)
	}

	faqID, err := autoReplyFAQID(payload["faq_id"])
	if err != nil {
		return fmt.Errorf("%w: FAQ reference is invalid: %v", ErrInvalidAutoReplyRule, err)
	}

	item, err := s.faqRepo.FindByID(faqID)
	if err != nil {
		if IsRecordNotFound(err) {
			return fmt.Errorf("%w: FAQ %d does not exist", ErrInvalidAutoReplyRule, faqID)
		}
		return err
	}

	if item == nil || strings.ToLower(strings.TrimSpace(item.Status)) != "published" {
		return fmt.Errorf("%w: FAQ %d is not published", ErrInvalidAutoReplyRule, faqID)
	}

	ruleLocale := resolveAutoReplyLocale(rule.Locale)
	faqLocale := resolveAutoReplyLocale(item.Locale)
	if ruleLocale == "" || faqLocale == "" || ruleLocale != faqLocale {
		return fmt.Errorf("%w: FAQ %d locale does not match automatic-reply locale", ErrInvalidAutoReplyRule, faqID)
	}

	if value := strings.TrimSpace(stringValue(payload["locale"])); value != "" &&
		resolveAutoReplyLocale(value) != ruleLocale {
		return fmt.Errorf("%w: FAQ metadata locale does not match automatic-reply locale", ErrInvalidAutoReplyRule)
	}
	if value := strings.TrimSpace(stringValue(payload["page_id"])); value != "" && value != item.PageID {
		return fmt.Errorf("%w: FAQ page does not match the referenced FAQ", ErrInvalidAutoReplyRule)
	}
	if value := strings.TrimSpace(stringValue(payload["category"])); value != "" && value != item.Category {
		return fmt.Errorf("%w: FAQ category does not match the referenced FAQ", ErrInvalidAutoReplyRule)
	}

	return nil
}

func autoReplyFAQID(value interface{}) (uint, error) {
	switch typed := value.(type) {
	case string:
		id, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 32)
		if err != nil || id == 0 {
			return 0, errors.New("faq_id must be a positive integer")
		}
		return uint(id), nil
	case float64:
		if typed <= 0 || typed != float64(uint(typed)) {
			return 0, errors.New("faq_id must be a positive integer")
		}
		return uint(typed), nil
	default:
		return 0, errors.New("faq_id must be a positive integer")
	}
}

func stringValue(value interface{}) string {
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}
