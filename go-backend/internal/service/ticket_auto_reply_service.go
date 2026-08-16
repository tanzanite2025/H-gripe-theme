package service

import (
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/ticket"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrInvalidAutoReplyRule  = errors.New("invalid automatic-reply rule")
	ErrAutoReplyRuleNotFound = errors.New("automatic-reply rule not found")
)

type AutoReplyRuleInput struct {
	Type            string `json:"type"`
	TriggerKeyword  string `json:"trigger_keyword"`
	ReplyMessage    string `json:"reply_message"`
	AgentID         string `json:"agent_id"`
	GroupID         *uint  `json:"group_id"`
	Locale          string `json:"locale"`
	MessageType     string `json:"message_type"`
	Metadata        string `json:"metadata"`
	Attachments     string `json:"attachments"`
	IsActive        bool   `json:"is_active"`
	Priority        int    `json:"priority"`
	MatchType       string `json:"match_type"`
	CooldownSeconds int    `json:"cooldown_seconds"`
}

func (s *TicketService) GetWelcomeMessage(conversationID string, owner CustomerServiceOwner, agentID uint, locale string) (string, bool, *ticket.TicketMessage, error) {
	locale = resolveAutoReplyLocale(locale)
	if locale == "" {
		return "", false, nil, nil
	}
	groupIDs, err := s.customerServiceAgentGroupIDs(agentID)
	if err != nil {
		return "", false, nil, err
	}
	rules, err := s.ticketRepo.GetActiveAutoReplyRules("welcome", locale, agentID, groupIDs)
	if err != nil {
		return "", false, nil, err
	}
	if len(rules) == 0 {
		return "", false, nil, nil
	}

	t, err := s.getOrCreateAccessibleCustomerServiceConversation(conversationID, owner, agentID)
	if err != nil {
		return "", false, nil, err
	}

	rule := selectAutoReplyRule(rules, locale, agentID, groupIDs)
	if rule == nil {
		return "", false, nil, nil
	}

	if err := s.validateAutoReplyFAQReference(*rule); err != nil {
		return "", false, nil, err
	}

	return s.deliverAutoReply(t, *rule, locale, "welcome")
}

func (s *TicketService) MatchKeywordMessage(conversationID, message string, owner CustomerServiceOwner, agentID uint, locale string) (string, uint, *ticket.TicketMessage, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return "", 0, nil, nil
	}

	locale = resolveAutoReplyLocale(locale)
	if locale == "" {
		return "", 0, nil, nil
	}
	groupIDs, err := s.customerServiceAgentGroupIDs(agentID)
	if err != nil {
		return "", 0, nil, err
	}
	rules, err := s.ticketRepo.GetActiveAutoReplyRules("keyword", locale, agentID, groupIDs)
	if err != nil {
		return "", 0, nil, err
	}

	var matchedRule *ticket.AutoReplyRule
	for index := range rules {
		rule := &rules[index]
		keyword := strings.TrimSpace(rule.TriggerKeyword)
		if keyword == "" {
			continue
		}

		isMatch := false
		if strings.EqualFold(strings.TrimSpace(rule.MatchType), "contains") {
			isMatch = strings.Contains(strings.ToLower(message), strings.ToLower(keyword))
		} else {
			isMatch = strings.EqualFold(message, keyword)
		}

		if isMatch {
			if matchedRule == nil || autoReplyRulePriority(*rule, locale, agentID, groupIDs) > autoReplyRulePriority(*matchedRule, locale, agentID, groupIDs) {
				matchedRule = rule
			}
		}
	}

	if matchedRule == nil {
		return "", 0, nil, nil
	}

	if err := s.validateAutoReplyFAQReference(*matchedRule); err != nil {
		return "", 0, nil, err
	}

	t, err := s.getOrCreateAccessibleCustomerServiceConversation(conversationID, owner, agentID)
	if err != nil {
		return "", 0, nil, err
	}

	reply, _, msg, err := s.deliverAutoReply(t, *matchedRule, locale, "keyword")
	if err != nil {
		return "", 0, nil, err
	}
	return reply, matchedRule.ID, msg, nil
}

func (s *TicketService) deliverAutoReply(t *ticket.Ticket, rule ticket.AutoReplyRule, locale, trigger string) (string, bool, *ticket.TicketMessage, error) {
	if t == nil {
		return "", false, nil, errors.New("customer-service conversation is required")
	}
	if s.customerServiceRealtimeOutbox == nil {
		return "", false, nil, errors.New("customer-service realtime outbox is unavailable")
	}
	if err := s.validateAutoReplyFAQReference(rule); err != nil {
		return "", false, nil, err
	}

	cooldown := rule.CooldownSeconds
	if cooldown <= 0 {
		if rule.Type == "welcome" {
			cooldown = 24 * 60 * 60
		} else {
			cooldown = 30
		}
	}

	replyMessage := strings.TrimSpace(rule.ReplyMessage)
	dedupeKey := autoReplyDedupeKey(rule)
	metadata := buildAutoReplyMetadata(rule, locale, trigger, dedupeKey)
	since := time.Now().UTC().Add(-time.Duration(cooldown) * time.Second)

	userID, err := s.autoReplyStaffUserID(t)
	if err != nil {
		return "", false, nil, err
	}

	msg := &ticket.TicketMessage{
		TicketID:    t.ID,
		UserID:      userID,
		IsStaff:     true,
		Content:     replyMessage,
		MessageType: normalizeAutoReplyMessageType(rule.MessageType),
		Metadata:    metadata,
		Attachments: normalizeAutoReplyAttachments(rule.Attachments),
		IsRead:      false,
		IsInternal:  false,
	}
	created, err := s.ticketRepo.CreateAutoReplyMessageIfNotRecentWithTx(
		msg,
		dedupeKey,
		replyMessage,
		since,
		func(ticketRepo *repository.TicketRepository, tx *gorm.DB) error {
			if err := ticketRepo.TouchTicket(t.ID, time.Now().UTC()); err != nil {
				return err
			}
			return enqueueCustomerServiceMessageCreatedOutboxEvent(
				s.customerServiceRealtimeOutbox.WithTx(tx),
				t,
				msg,
				CustomerServiceRealtimeActor{Kind: "system"},
			)
		},
	)
	if err != nil {
		return "", false, nil, err
	}
	if !created {
		return replyMessage, true, nil, nil
	}

	return replyMessage, false, msg, nil
}

func (s *TicketService) autoReplyStaffUserID(t *ticket.Ticket) (uint, error) {
	if t == nil {
		return 0, errors.New("customer-service conversation is required")
	}
	if s.userRepo == nil {
		return 0, errors.New("customer-service user repository is unavailable")
	}
	if t.AssignedTo > 0 {
		assignedID := strconv.FormatUint(uint64(t.AssignedTo), 10)
		if err := s.validateAutoReplyAgentID(assignedID); err == nil {
			return t.AssignedTo, nil
		}
	}

	// Older customer-service rows may have stored the staff user in
	// tickets.user_id. Only reuse it after checking the role; authenticated
	// public visitors can also occupy that column.
	if t.UserID > 0 {
		if candidate, err := s.userRepo.FindByID(t.UserID); err == nil &&
			candidate != nil &&
			strings.TrimSpace(candidate.Status) == "active" &&
			auth.IsCustomerServiceAgentRole(candidate.Role) {
			return candidate.ID, nil
		}
	}

	userID, err := s.customerServicePersistedUserID(nil, 0)
	if err != nil {
		return 0, fmt.Errorf("automatic reply has no valid staff user: %w", err)
	}
	return userID, nil
}

func (s *TicketService) ListAutoReplyRules() ([]ticket.AutoReplyRule, error) {
	return s.ticketRepo.ListAutoReplyRules()
}

func (s *TicketService) GetAutoReplyRule(id uint) (*ticket.AutoReplyRule, error) {
	rule, err := s.ticketRepo.FindAutoReplyRuleByID(id)
	if err != nil {
		if IsRecordNotFound(err) {
			return nil, ErrAutoReplyRuleNotFound
		}
		return nil, err
	}
	return rule, nil
}

func (s *TicketService) CreateAutoReplyRule(input AutoReplyRuleInput) (*ticket.AutoReplyRule, error) {
	if err := s.validateAutoReplyAgentID(input.AgentID); err != nil {
		return nil, err
	}
	if err := s.validateAutoReplyGroupID(input.GroupID); err != nil {
		return nil, err
	}
	rule, err := normalizeAutoReplyRuleInput(input, nil)
	if err != nil {
		return nil, err
	}
	if err := s.validateAutoReplyFAQReference(*rule); err != nil {
		return nil, err
	}
	if err := s.ticketRepo.CreateAutoReplyRule(rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *TicketService) UpdateAutoReplyRule(id uint, input AutoReplyRuleInput) (*ticket.AutoReplyRule, error) {
	if err := s.validateAutoReplyAgentID(input.AgentID); err != nil {
		return nil, err
	}
	if err := s.validateAutoReplyGroupID(input.GroupID); err != nil {
		return nil, err
	}
	rule, err := s.GetAutoReplyRule(id)
	if err != nil {
		return nil, err
	}
	if _, err := normalizeAutoReplyRuleInput(input, rule); err != nil {
		return nil, err
	}
	if err := s.validateAutoReplyFAQReference(*rule); err != nil {
		return nil, err
	}
	if err := s.ticketRepo.UpdateAutoReplyRule(rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *TicketService) DeleteAutoReplyRule(id uint) error {
	if _, err := s.GetAutoReplyRule(id); err != nil {
		return err
	}
	return s.ticketRepo.DeleteAutoReplyRule(id)
}

func normalizeAutoReplyRuleInput(input AutoReplyRuleInput, existing *ticket.AutoReplyRule) (*ticket.AutoReplyRule, error) {
	rule := existing
	if rule == nil {
		rule = &ticket.AutoReplyRule{}
	}

	rule.Type = strings.ToLower(strings.TrimSpace(input.Type))
	rule.TriggerKeyword = strings.TrimSpace(input.TriggerKeyword)
	rule.ReplyMessage = strings.TrimSpace(input.ReplyMessage)
	rule.AgentID = strings.TrimSpace(input.AgentID)
	rule.GroupID = normalizeAutoReplyGroupID(input.GroupID)
	locale, err := requireAutoReplyLocale(input.Locale)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAutoReplyRule, err)
	}
	rule.Locale = locale
	messageType, err := parseAutoReplyMessageType(input.MessageType)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAutoReplyRule, err)
	}
	rule.MessageType = messageType
	rule.Metadata = strings.TrimSpace(input.Metadata)
	rule.Attachments = strings.TrimSpace(input.Attachments)
	rule.IsActive = input.IsActive
	rule.Priority = input.Priority
	rule.MatchType = strings.ToLower(strings.TrimSpace(input.MatchType))
	rule.CooldownSeconds = input.CooldownSeconds

	if rule.Type != "welcome" && rule.Type != "keyword" {
		return nil, fmt.Errorf("%w: type must be welcome or keyword", ErrInvalidAutoReplyRule)
	}
	if rule.Type == "keyword" && rule.TriggerKeyword == "" {
		return nil, fmt.Errorf("%w: keyword rules require trigger_keyword", ErrInvalidAutoReplyRule)
	}
	if rule.Type == "welcome" {
		rule.TriggerKeyword = ""
	}
	if rule.ReplyMessage == "" {
		return nil, fmt.Errorf("%w: reply_message is required", ErrInvalidAutoReplyRule)
	}
	if rule.MatchType == "" {
		rule.MatchType = "exact"
	}
	if rule.MatchType != "exact" && rule.MatchType != "contains" {
		return nil, fmt.Errorf("%w: match_type must be exact or contains", ErrInvalidAutoReplyRule)
	}
	if rule.Priority < 0 || rule.CooldownSeconds < 0 {
		return nil, fmt.Errorf("%w: priority and cooldown_seconds cannot be negative", ErrInvalidAutoReplyRule)
	}
	if err := validateAutoReplyMetadata(rule.MessageType, rule.Metadata); err != nil {
		return nil, err
	}
	if err := validateAutoReplyAttachments(rule.Attachments); err != nil {
		return nil, err
	}
	rule.Attachments = normalizeAutoReplyAttachments(rule.Attachments)
	if rule.MessageType == "image" && rule.Attachments == "" {
		return nil, fmt.Errorf("%w: image replies require at least one attachment", ErrInvalidAutoReplyRule)
	}
	return rule, nil
}

func normalizeAutoReplyMessageType(value string) string {
	messageType, err := parseAutoReplyMessageType(value)
	if err != nil {
		return "text"
	}
	return messageType
}

func parseAutoReplyMessageType(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "text":
		return "text", nil
	case "image", "link", "product", "order", "faq":
		return value, nil
	default:
		return "", errors.New("message_type is unsupported")
	}
}

func normalizeAutoReplyAttachments(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var items []string
	if err := json.Unmarshal([]byte(value), &items); err != nil {
		return ""
	}
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			normalized = append(normalized, item)
		}
	}
	if len(normalized) == 0 {
		return ""
	}
	payload, _ := json.Marshal(normalized)
	return string(payload)
}

func validateAutoReplyMetadata(messageType, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if messageType == "link" || messageType == "product" || messageType == "order" || messageType == "faq" {
			return fmt.Errorf("%w: %s replies require metadata", ErrInvalidAutoReplyRule, messageType)
		}
		return nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return fmt.Errorf("%w: metadata must be a JSON object", ErrInvalidAutoReplyRule)
	}
	switch messageType {
	case "link":
		if err := validateStructuredURL(payload, "url", true, true); err != nil {
			return fmt.Errorf("%w: link url is invalid: %v", ErrInvalidAutoReplyRule, err)
		}
		if err := validateStructuredURL(payload, "thumbnail", true, false); err != nil {
			return fmt.Errorf("%w: link thumbnail is invalid: %v", ErrInvalidAutoReplyRule, err)
		}
	case "product":
		if err := validateStructuredURL(payload, "url", true, false); err != nil {
			return fmt.Errorf("%w: product url is invalid: %v", ErrInvalidAutoReplyRule, err)
		}
		if err := validateStructuredURL(payload, "thumbnail", true, false); err != nil {
			return fmt.Errorf("%w: product thumbnail is invalid: %v", ErrInvalidAutoReplyRule, err)
		}
	case "order":
		if err := validateStructuredURL(payload, "url", true, false); err != nil {
			return fmt.Errorf("%w: order url is invalid: %v", ErrInvalidAutoReplyRule, err)
		}
	case "faq":
		if err := validateStructuredID(payload, "faq_id"); err != nil {
			return fmt.Errorf("%w: faq_id is invalid: %v", ErrInvalidAutoReplyRule, err)
		}
		if err := validateStructuredString(payload, "question", true); err != nil {
			return fmt.Errorf("%w: faq question is invalid: %v", ErrInvalidAutoReplyRule, err)
		}
		if err := validateStructuredURL(payload, "url", true, true); err != nil {
			return fmt.Errorf("%w: faq url is invalid: %v", ErrInvalidAutoReplyRule, err)
		}
		if err := validateStructuredURL(payload, "answer_image_url", true, false); err != nil {
			return fmt.Errorf("%w: faq answer_image_url is invalid: %v", ErrInvalidAutoReplyRule, err)
		}
	}
	return nil
}

func validateStructuredID(payload map[string]interface{}, key string) error {
	value, exists := payload[key]
	if !exists {
		return errors.New("id is required")
	}
	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return errors.New("id is required")
		}
	case float64:
		if typed <= 0 {
			return errors.New("id must be positive")
		}
	default:
		return errors.New("id must be a string or number")
	}
	return nil
}

func validateStructuredString(payload map[string]interface{}, key string, required bool) error {
	value, exists := payload[key]
	if !exists {
		if required {
			return errors.New("value is required")
		}
		return nil
	}
	raw, ok := value.(string)
	if !ok {
		return errors.New("value must be a string")
	}
	if strings.TrimSpace(raw) == "" && required {
		return errors.New("value is required")
	}
	return nil
}

func validateStructuredURL(payload map[string]interface{}, key string, allowRelative, required bool) error {
	value, exists := payload[key]
	if !exists {
		if required {
			return errors.New("url is required")
		}
		return nil
	}
	raw, ok := value.(string)
	if !ok {
		return errors.New("url must be a string")
	}
	raw = strings.TrimSpace(raw)
	if raw == "" && !required {
		return nil
	}
	return validateAutoReplyURL(raw, allowRelative)
}

func validateAutoReplyAttachments(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return fmt.Errorf("%w: attachments must be a JSON array", ErrInvalidAutoReplyRule)
	}
	for _, item := range items {
		if err := validateAutoReplyURL(item, true); err != nil {
			return fmt.Errorf("%w: attachment url is invalid: %v", ErrInvalidAutoReplyRule, err)
		}
	}
	return nil
}

func validateAutoReplyURL(value string, allowRelative bool) error {
	value = strings.TrimSpace(value)
	if value == "" || strings.IndexFunc(value, func(r rune) bool { return r < 0x20 }) >= 0 {
		return errors.New("url is empty or contains control characters")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return err
	}
	if allowRelative && strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") {
		return nil
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("only http and https urls are allowed")
	}
	if parsed.Host == "" {
		return errors.New("url host is required")
	}
	return nil
}

func selectAutoReplyRule(rules []ticket.AutoReplyRule, locale string, agentID uint, scopedGroupIDs ...[]uint) *ticket.AutoReplyRule {
	groupIDs := []uint{}
	if len(scopedGroupIDs) > 0 {
		groupIDs = scopedGroupIDs[0]
	}
	var selected *ticket.AutoReplyRule
	for index := range rules {
		candidate := &rules[index]
		if selected == nil || autoReplyRulePriority(*candidate, locale, agentID, groupIDs) > autoReplyRulePriority(*selected, locale, agentID, groupIDs) {
			selected = candidate
		}
	}
	return selected
}

func autoReplyRulePriority(rule ticket.AutoReplyRule, locale string, agentID uint, groupIDs []uint) int {
	score := rule.Priority
	if resolveAutoReplyLocale(rule.Locale) == locale {
		score += 300000
	}
	if agentID > 0 && rule.AgentID == strconv.FormatUint(uint64(agentID), 10) {
		score += 10000
	} else if rule.GroupID != nil && containsUint(groupIDs, *rule.GroupID) {
		score += 5000
	} else if strings.TrimSpace(rule.AgentID) == "" {
		score += 1000
	}
	score++
	return score
}

func (s *TicketService) customerServiceAgentGroupIDs(agentID uint) ([]uint, error) {
	if agentID == 0 {
		return []uint{}, nil
	}
	if s.userRepo == nil {
		return nil, fmt.Errorf("%w: customer-service user repository is unavailable", ErrInvalidAutoReplyRule)
	}
	ids, err := s.userRepo.FindCustomerServiceAgentProfileGroupIDsByUserID(agentID)
	if err != nil {
		if IsRecordNotFound(err) {
			return []uint{}, nil
		}
		return nil, err
	}
	return ids, nil
}

func (s *TicketService) validateAutoReplyGroupID(groupID *uint) error {
	if groupID == nil || *groupID == 0 {
		return nil
	}
	if s.userRepo == nil {
		return fmt.Errorf("%w: customer-service user repository is unavailable", ErrInvalidAutoReplyRule)
	}
	group, err := s.userRepo.FindCustomerServiceAgentGroupByID(*groupID)
	if err != nil || group == nil || strings.TrimSpace(group.Status) != "active" {
		return fmt.Errorf("%w: group_id is invalid", ErrInvalidAutoReplyRule)
	}
	return nil
}

func normalizeAutoReplyGroupID(groupID *uint) *uint {
	if groupID == nil || *groupID == 0 {
		return nil
	}
	value := *groupID
	return &value
}

func containsUint(values []uint, target uint) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func buildAutoReplyMetadata(rule ticket.AutoReplyRule, locale, trigger, dedupeKey string) string {
	payload := map[string]interface{}{}
	if strings.TrimSpace(rule.Metadata) != "" {
		_ = json.Unmarshal([]byte(rule.Metadata), &payload)
	}
	payload["_source"] = "auto_reply"
	payload["_rule_id"] = rule.ID
	payload["_dedupe_key"] = dedupeKey
	payload["_trigger"] = trigger
	payload["_locale"] = locale
	result, _ := json.Marshal(payload)
	return string(result)
}

func autoReplyDedupeKey(rule ticket.AutoReplyRule) string {
	if rule.ID > 0 {
		return "autoreply:rule:" + strconv.FormatUint(uint64(rule.ID), 10)
	}
	parts := []string{
		strings.ToLower(strings.TrimSpace(rule.Type)),
		resolveAutoReplyLocale(rule.Locale),
		strings.TrimSpace(rule.AgentID),
		strings.ToLower(strings.TrimSpace(rule.TriggerKeyword)),
		strings.TrimSpace(rule.ReplyMessage),
	}
	digest := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "autoreply:rule:fallback:" + hex.EncodeToString(digest[:])
}
