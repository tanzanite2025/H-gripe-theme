package service

import (
	"strings"
	"tanzanite/internal/domain/ticket"
	"time"
)

func (s *TicketService) GetWelcomeMessage(conversationID string, owner CustomerServiceOwner, agentID uint) (string, bool, error) {
	rules, err := s.ticketRepo.GetActiveAutoReplyRules("welcome")
	if err != nil {
		return "", false, err
	}
	if len(rules) == 0 {
		return "", false, nil
	}
	welcomeRule := rules[0]

	t, err := s.getOrCreateAccessibleCustomerServiceConversation(conversationID, owner, agentID)
	if err != nil {
		return "", false, err
	}

	lastSent, err := s.ticketRepo.GetLastWelcomeMessageTime(t.ID, welcomeRule.ReplyMessage)
	if err == nil && !lastSent.IsZero() && time.Since(lastSent) < 24*time.Hour {
		return welcomeRule.ReplyMessage, true, nil
	}

	msg := &ticket.TicketMessage{
		TicketID:   t.ID,
		UserID:     0,
		IsStaff:    true,
		Content:    welcomeRule.ReplyMessage,
		IsRead:     false,
		IsInternal: false,
	}
	if err := s.ticketRepo.CreateTicketMessage(msg); err != nil {
		return "", false, err
	}

	return welcomeRule.ReplyMessage, false, nil
}

func (s *TicketService) MatchKeywordMessage(conversationID, message string, owner CustomerServiceOwner, agentID uint) (string, uint, error) {
	rules, err := s.ticketRepo.GetActiveAutoReplyRules("keyword")
	if err != nil {
		return "", 0, err
	}
	if len(rules) == 0 {
		return "", 0, nil
	}

	var matchedRule *ticket.AutoReplyRule
	for _, rule := range rules {
		keyword := strings.TrimSpace(rule.TriggerKeyword)
		if keyword == "" {
			continue
		}

		isMatch := false
		if rule.MatchType == "contains" {
			isMatch = strings.Contains(strings.ToLower(message), strings.ToLower(keyword))
		} else {
			isMatch = strings.EqualFold(strings.TrimSpace(message), keyword)
		}

		if isMatch {
			matchedRule = &rule
			break
		}
	}

	if matchedRule == nil {
		return "", 0, nil
	}

	t, err := s.getOrCreateAccessibleCustomerServiceConversation(conversationID, owner, agentID)
	if err != nil {
		return "", 0, err
	}

	msg := &ticket.TicketMessage{
		TicketID:   t.ID,
		UserID:     0,
		IsStaff:    true,
		Content:    matchedRule.ReplyMessage,
		IsRead:     false,
		IsInternal: false,
	}
	if err := s.ticketRepo.CreateTicketMessage(msg); err != nil {
		return "", 0, err
	}

	return matchedRule.ReplyMessage, matchedRule.ID, nil
}
