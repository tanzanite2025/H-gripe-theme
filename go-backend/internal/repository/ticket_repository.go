package repository

import (
	"commerce-platform/internal/domain/ticket"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TicketRepository struct {
	db *gorm.DB
}

type CustomerServiceConversationFilters struct {
	AssignedTo *uint
	GroupID    *uint
	Status     string
	UnreadOnly bool
	Identity   string
	Search     string
}

type DisputeCommunicationFilter struct {
	UserID      uint
	Emails      []string
	OrderNumber string
	Limit       int
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// Ticket 相关方法

// CreateTicket 创建工单
func (r *TicketRepository) CreateTicket(t *ticket.Ticket) error {
	return r.db.Create(t).Error
}

// FindTicketByID 根据ID查找工单
func (r *TicketRepository) FindTicketByID(id uint) (*ticket.Ticket, error) {
	var t ticket.Ticket
	err := r.db.Preload("Messages").Preload("User").First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepository) FindCustomerServiceConversations(page, pageSize int, filters CustomerServiceConversationFilters) ([]ticket.Ticket, int64, error) {
	var tickets []ticket.Ticket
	var total int64

	query := r.db.Model(&ticket.Ticket{}).Where("tickets.category = ?", "customer_service")
	query = applyCustomerServiceConversationFilters(query, filters)

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Distinct("tickets.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("User").Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Order("tickets.updated_at DESC").Offset(offset).Limit(pageSize).Find(&tickets).Error

	return tickets, total, err
}

func (r *TicketRepository) FindCustomerServiceConversationsInWindow(start, end time.Time, filters CustomerServiceConversationFilters) ([]ticket.Ticket, error) {
	var tickets []ticket.Ticket

	query := r.db.Model(&ticket.Ticket{}).
		Where("tickets.category = ?", "customer_service").
		Where("tickets.created_at >= ? AND tickets.created_at < ?", start, end)
	query = applyCustomerServiceConversationFilters(query, filters)

	err := query.Order("tickets.created_at DESC").Find(&tickets).Error
	return tickets, err
}

func applyCustomerServiceConversationFilters(query *gorm.DB, filters CustomerServiceConversationFilters) *gorm.DB {
	if filters.AssignedTo != nil {
		query = query.Where("tickets.assigned_to = ?", *filters.AssignedTo)
	}
	if filters.GroupID != nil {
		query = query.Where(`
			EXISTS (
				SELECT 1
				FROM customer_service_agent_profiles group_agent_profiles
				JOIN customer_service_agent_group_members group_members
					ON group_members.agent_profile_id = group_agent_profiles.id
				WHERE group_agent_profiles.user_id = tickets.assigned_to
					AND group_members.group_id = ?
			)
		`, *filters.GroupID)
	}

	switch strings.ToLower(strings.TrimSpace(filters.Status)) {
	case "pending", "open":
		query = query.Where("(tickets.status = ? OR tickets.status = '')", "open")
	case "active", "in_progress":
		query = query.Where("tickets.status = ?", "in_progress")
	case "closed":
		query = query.Where("tickets.status IN ?", []string{"resolved", "closed"})
	case "resolved":
		query = query.Where("tickets.status = ?", "resolved")
	}

	if filters.UnreadOnly {
		query = query.Where(
			"EXISTS (SELECT 1 FROM ticket_messages customer_unread_messages WHERE customer_unread_messages.ticket_id = tickets.id AND customer_unread_messages.is_staff = ? AND customer_unread_messages.is_read = ?)",
			false,
			false,
		)
	}

	switch strings.ToLower(strings.TrimSpace(filters.Identity)) {
	case "account", "member", "user":
		query = query.Where("tickets.customer_user_id IS NOT NULL")
	case "anonymous", "visitor", "guest":
		query = query.Where("tickets.customer_user_id IS NULL")
	}

	search := customerServiceConversationSearchPattern(filters.Search)
	if search != "" {
		query = query.
			Joins("LEFT JOIN users AS customer_users ON customer_users.id = tickets.customer_user_id AND customer_users.deleted_at IS NULL").
			Joins("LEFT JOIN visitor_profiles AS conversation_visitors ON conversation_visitors.customer_service_visitor_hash = tickets.visitor_session_hash AND conversation_visitors.deleted_at IS NULL").
			Where(
				`LOWER(COALESCE(tickets.ticket_number, '')) LIKE ?
				OR LOWER(COALESCE(tickets.conversation_id, '')) LIKE ?
				OR LOWER(COALESCE(tickets.subject, '')) LIKE ?
				OR LOWER(COALESCE(tickets.visitor_session_hash, '')) LIKE ?
				OR LOWER(COALESCE(customer_users.email, '')) LIKE ?
				OR LOWER(COALESCE(customer_users.username, '')) LIKE ?
				OR LOWER(COALESCE(customer_users.first_name, '')) LIKE ?
				OR LOWER(COALESCE(customer_users.last_name, '')) LIKE ?
				OR LOWER(COALESCE(conversation_visitors.email, '')) LIKE ?
				OR LOWER(COALESCE(conversation_visitors.cart_session_id, '')) LIKE ?
				OR EXISTS (
					SELECT 1
					FROM ticket_messages searched_customer_service_messages
					WHERE searched_customer_service_messages.ticket_id = tickets.id
					AND LOWER(COALESCE(searched_customer_service_messages.content, '')) LIKE ?
				)`,
				search,
				search,
				search,
				search,
				search,
				search,
				search,
				search,
				search,
				search,
				search,
			)
	}

	return query
}

func customerServiceConversationSearchPattern(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) > 120 {
		value = string(runes[:120])
	}
	return "%" + value + "%"
}

func (r *TicketRepository) FindCustomerServiceConversationByConversationID(conversationID string) (*ticket.Ticket, error) {
	var t ticket.Ticket
	err := r.db.Preload("User").Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("category = ? AND conversation_id = ?", "customer_service", conversationID).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepository) FindCustomerServiceConversationByOwner(userID *uint, visitorSessionHash string) (*ticket.Ticket, error) {
	var t ticket.Ticket
	query := r.db.Preload("User").Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("category = ?", "customer_service")

	if userID != nil {
		query = query.Where("customer_user_id = ?", *userID)
	} else {
		query = query.Where("visitor_session_hash = ?", visitorSessionHash)
	}

	err := query.Order("updated_at DESC").First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTicket 更新工单
func (r *TicketRepository) UpdateTicket(t *ticket.Ticket) error {
	return r.db.Save(t).Error
}

// UpdateTicketStatus 更新工单状态
func (r *TicketRepository) UpdateTicketStatus(id uint, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == "resolved" || status == "closed" {
		updates["resolved_at"] = gorm.Expr("NOW()")
	}

	return r.db.Model(&ticket.Ticket{}).Where("id = ?", id).Updates(updates).Error
}

// AssignTicket 分配工单
func (r *TicketRepository) AssignTicket(id, assignedTo uint) error {
	return r.db.Model(&ticket.Ticket{}).Where("id = ?", id).
		Update("assigned_to", assignedTo).Error
}

// TicketMessage 相关方法

// CreateTicketMessage 创建工单消息
func (r *TicketRepository) CreateTicketMessage(m *ticket.TicketMessage) error {
	return r.db.Create(m).Error
}

// FindMessagesByTicketID 查找工单的消息列表
func (r *TicketRepository) FindMessagesByTicketID(ticketID uint) ([]ticket.TicketMessage, error) {
	var messages []ticket.TicketMessage
	err := r.db.Where("ticket_id = ?", ticketID).
		Preload("User").Order("created_at ASC").Find(&messages).Error
	return messages, err
}

func (r *TicketRepository) FindDisputeCandidateMessages(filter DisputeCommunicationFilter) ([]ticket.TicketMessage, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 80
	}

	conditions := make([]string, 0, 4)
	args := make([]interface{}, 0, 8)
	if filter.UserID > 0 {
		conditions = append(conditions, "(tickets.user_id = ? OR tickets.customer_user_id = ? OR ticket_messages.user_id = ?)")
		args = append(args, filter.UserID, filter.UserID, filter.UserID)
	}

	terms := disputeCommunicationTerms(filter)
	for _, term := range terms {
		like := "%" + strings.ToLower(term) + "%"
		conditions = append(conditions, "(LOWER(COALESCE(tickets.subject, '')) LIKE ? OR LOWER(COALESCE(ticket_messages.content, '')) LIKE ? OR LOWER(COALESCE(message_users.email, '')) LIKE ?)")
		args = append(args, like, like, like)
	}
	if len(conditions) == 0 {
		return []ticket.TicketMessage{}, nil
	}

	var messages []ticket.TicketMessage
	err := r.db.Model(&ticket.TicketMessage{}).
		Joins("JOIN tickets ON tickets.id = ticket_messages.ticket_id").
		Joins("LEFT JOIN users AS message_users ON message_users.id = ticket_messages.user_id").
		Where("ticket_messages.is_internal = ?", false).
		Where("("+strings.Join(conditions, " OR ")+")", args...).
		Preload("User").
		Order("ticket_messages.created_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func disputeCommunicationTerms(filter DisputeCommunicationFilter) []string {
	seen := map[string]bool{}
	terms := make([]string, 0, len(filter.Emails)+1)
	add := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		terms = append(terms, value)
	}
	add(filter.OrderNumber)
	for _, email := range filter.Emails {
		add(email)
	}
	return terms
}

// MarkMessagesAsRead 标记消息为已读
func (r *TicketRepository) MarkMessagesAsRead(ticketID uint, isStaff bool) error {
	return r.db.Model(&ticket.TicketMessage{}).
		Where("ticket_id = ? AND is_staff = ?", ticketID, !isStaff).
		Update("is_read", true).Error
}

// GetActiveAutoReplyRules returns active rules for one canonical locale and
// public agent. Locale fallback policy belongs to the service layer; this
// repository intentionally performs an exact locale lookup only.
func (r *TicketRepository) GetActiveAutoReplyRules(ruleType, locale string, agentID uint, groupIDs []uint) ([]ticket.AutoReplyRule, error) {
	var rules []ticket.AutoReplyRule
	query := r.db.Model(&ticket.AutoReplyRule{}).Where("type = ? AND is_active = ?", ruleType, true)

	locale = strings.TrimSpace(locale)
	if locale == "" {
		return rules, nil
	}
	query = query.Where("locale = ?", locale)

	if agentID > 0 {
		agent := strconv.FormatUint(uint64(agentID), 10)
		query = query.Where("(agent_id = ? OR agent_id = '' OR agent_id IS NULL)", agent)
	} else {
		query = query.Where("(agent_id = '' OR agent_id IS NULL)")
	}
	if len(groupIDs) > 0 {
		query = query.Where("(group_id IN ? OR group_id IS NULL)", groupIDs)
	} else {
		query = query.Where("group_id IS NULL")
	}

	err := query.Order("priority DESC, created_at DESC").Find(&rules).Error
	return rules, err
}

// FindRecentAutoReplyMessage checks whether an equivalent automatic reply was
// already persisted in the cooldown window. The content fallback preserves
// idempotency for legacy rows created before rule metadata existed.
func (r *TicketRepository) FindRecentAutoReplyMessage(ticketID uint, dedupeKey, content string, since time.Time) (bool, error) {
	var m ticket.TicketMessage
	err := r.db.Model(&ticket.TicketMessage{}).
		Where("ticket_id = ? AND is_staff = ? AND created_at >= ?", ticketID, true, since).
		Where(
			"(metadata LIKE ? OR (? <> '' AND metadata LIKE ?) OR (COALESCE(metadata, '') = '' AND content = ?))",
			autoReplyDedupeLikePattern(dedupeKey),
			autoReplyLegacyRuleLikePattern(dedupeKey),
			autoReplyLegacyRuleLikePattern(dedupeKey),
			content,
		).
		Order("created_at DESC").First(&m).Error
	if err == nil {
		return true, nil
	}
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return false, err
}

// CreateAutoReplyMessageIfNotRecent atomically applies the cooldown check and
// inserts an automatic reply. PostgreSQL locks the conversation row so two
// concurrent clients cannot both pass the check for the same conversation.
func (r *TicketRepository) CreateAutoReplyMessageIfNotRecent(message *ticket.TicketMessage, dedupeKey, content string, since time.Time) (bool, error) {
	if message == nil {
		return false, gorm.ErrInvalidData
	}

	created := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if r.db.Dialector.Name() == "postgres" {
			var conversation ticket.Ticket
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("id").
				First(&conversation, message.TicketID).Error; err != nil {
				return err
			}
		}

		var existing ticket.TicketMessage
		err := tx.Model(&ticket.TicketMessage{}).
			Where("ticket_id = ? AND is_staff = ? AND created_at >= ?", message.TicketID, true, since).
			Where(
				"(metadata LIKE ? OR (? <> '' AND metadata LIKE ?) OR (COALESCE(metadata, '') = '' AND content = ?))",
				autoReplyDedupeLikePattern(dedupeKey),
				autoReplyLegacyRuleLikePattern(dedupeKey),
				autoReplyLegacyRuleLikePattern(dedupeKey),
				content,
			).
			Order("created_at DESC").
			First(&existing).Error
		if err == nil {
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		created = true
		return nil
	})
	return created, err
}

func autoReplyDedupeLikePattern(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return `%"_dedupe_key":""%`
	}
	return `%"_dedupe_key":"` + value + `"%`
}

func autoReplyLegacyRuleLikePattern(dedupeKey string) string {
	const prefix = "autoreply:rule:"
	rawID := strings.TrimPrefix(strings.TrimSpace(dedupeKey), prefix)
	if rawID == dedupeKey || rawID == "" || strings.Contains(rawID, ":") {
		return ""
	}
	if _, err := strconv.ParseUint(rawID, 10, 64); err != nil {
		return ""
	}
	return `%"_rule_id":` + rawID + `%`
}

// ListAutoReplyRules returns all rules for the admin management surface.
func (r *TicketRepository) ListAutoReplyRules() ([]ticket.AutoReplyRule, error) {
	var rules []ticket.AutoReplyRule
	err := r.db.Order("type ASC, priority DESC, created_at DESC").Find(&rules).Error
	return rules, err
}

// FindAutoReplyRuleByID returns one rule for admin editing.
func (r *TicketRepository) FindAutoReplyRuleByID(id uint) (*ticket.AutoReplyRule, error) {
	var rule ticket.AutoReplyRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreateAutoReplyRule persists an admin-authored rule.
func (r *TicketRepository) CreateAutoReplyRule(rule *ticket.AutoReplyRule) error {
	return r.db.Create(rule).Error
}

// UpdateAutoReplyRule persists an admin-authored rule update.
func (r *TicketRepository) UpdateAutoReplyRule(rule *ticket.AutoReplyRule) error {
	return r.db.Save(rule).Error
}

// DeleteAutoReplyRule deletes one rule.
func (r *TicketRepository) DeleteAutoReplyRule(id uint) error {
	return r.db.Delete(&ticket.AutoReplyRule{}, id).Error
}
