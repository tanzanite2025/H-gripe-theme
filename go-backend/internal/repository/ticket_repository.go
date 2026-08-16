package repository

import (
	"commerce-platform/internal/domain/ticket"
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrTicketStatusVersionConflict = errors.New("ticket status version conflict")

type TicketRepository struct {
	db *gorm.DB
}

type CustomerServiceConversationFilters struct {
	AssignedTo      *uint
	GroupID         *uint
	RecipientUserID uint
	Status          string
	UnreadOnly      bool
	Identity        string
	Search          string
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

func (r *TicketRepository) WithTx(tx *gorm.DB) *TicketRepository {
	return &TicketRepository{db: tx}
}

// WithinTx is used by customer-service command paths that must persist a
// message and an Outbox event atomically.
func (r *TicketRepository) WithinTx(fn func(*TicketRepository, *gorm.DB) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(r.WithTx(tx), tx)
	})
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

// FindTicketByIDForUpdate locks the conversation row for a command mutation.
// It intentionally avoids preloading associations so the caller's transaction
// decides authorization and state transitions from one current ticket row.
func (r *TicketRepository) FindTicketByIDForUpdate(id uint) (*ticket.Ticket, error) {
	var t ticket.Ticket
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepository) FindCustomerServiceConversations(page, pageSize int, filters CustomerServiceConversationFilters) ([]ticket.Ticket, int64, error) {
	var tickets []ticket.Ticket
	var total int64

	baseQuery := func() *gorm.DB {
		query := r.db.Model(&ticket.Ticket{}).Where("tickets.category = ?", "customer_service")
		return applyCustomerServiceConversationFilters(query, filters)
	}

	countQuery := baseQuery()
	if err := countQuery.Distinct("tickets.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := baseQuery()
	if filters.RecipientUserID > 0 {
		query = query.Select(customerServiceConversationUnreadCountSelect())
	} else {
		// A raw JOIN used by search filters makes GORM enumerate model fields.
		// Keep the transient unread projection out of legacy, no-recipient reads.
		query = query.Select("tickets.*")
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
		Where("tickets.created_at < ?", end).
		Where(`
			tickets.created_at >= ?
			OR EXISTS (
				SELECT 1
				FROM ticket_messages customer_service_window_messages
				WHERE customer_service_window_messages.ticket_id = tickets.id
					AND customer_service_window_messages.created_at >= ?
					AND customer_service_window_messages.created_at < ?
			)
		`, start, start, end)
	query = applyCustomerServiceConversationFilters(query, filters)

	err := query.
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.
				Where(`
					(created_at >= ? AND created_at < ?)
					OR (
						created_at < ?
						AND is_internal = ?
						AND NOT EXISTS (
							SELECT 1
							FROM ticket_messages previous_customer_service_messages
							WHERE previous_customer_service_messages.ticket_id = ticket_messages.ticket_id
								AND previous_customer_service_messages.created_at < ?
								AND previous_customer_service_messages.is_internal = ?
								AND (
									previous_customer_service_messages.created_at > ticket_messages.created_at
									OR (
										previous_customer_service_messages.created_at = ticket_messages.created_at
										AND previous_customer_service_messages.id > ticket_messages.id
									)
								)
						)
					)
				`, start, end, start, false, start, false).
				Order("created_at ASC, id ASC")
		}).
		Order("tickets.created_at DESC").
		Find(&tickets).Error
	return tickets, err
}

func applyCustomerServiceConversationFilters(query *gorm.DB, filters CustomerServiceConversationFilters) *gorm.DB {
	if filters.RecipientUserID > 0 {
		query = query.Joins(
			"LEFT JOIN customer_service_inbox_states AS customer_service_inbox_state ON customer_service_inbox_state.ticket_id = tickets.id AND customer_service_inbox_state.recipient_user_id = ? AND customer_service_inbox_state.deleted_at IS NULL",
			filters.RecipientUserID,
		)
	}
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
		if filters.RecipientUserID > 0 {
			query = query.Where(customerServiceUnreadMessageExistsCondition())
		} else {
			// Compatibility fallback for internal callers that do not have a
			// backoffice recipient identity. Customer-service HTTP routes always
			// pass RecipientUserID and therefore do not rely on this legacy flag.
			query = query.Where(
				"EXISTS (SELECT 1 FROM ticket_messages customer_unread_messages WHERE customer_unread_messages.ticket_id = tickets.id AND customer_unread_messages.is_staff = ? AND customer_unread_messages.is_read = ?)",
				false,
				false,
			)
		}
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

func customerServiceConversationUnreadCountSelect() string {
	return `tickets.*, COALESCE((
		SELECT COUNT(*)
		FROM ticket_messages AS customer_service_unread_messages
		WHERE customer_service_unread_messages.ticket_id = tickets.id
			AND customer_service_unread_messages.is_staff = FALSE
			AND customer_service_unread_messages.id > COALESCE(customer_service_inbox_state.last_read_message_id, 0)
	), 0) AS customer_service_unread_count`
}

func customerServiceUnreadMessageExistsCondition() string {
	return `EXISTS (
		SELECT 1
		FROM ticket_messages AS customer_service_unread_messages
		WHERE customer_service_unread_messages.ticket_id = tickets.id
			AND customer_service_unread_messages.is_staff = FALSE
			AND customer_service_unread_messages.id > COALESCE(customer_service_inbox_state.last_read_message_id, 0)
	)`
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
		"status":         status,
		"status_version": gorm.Expr("COALESCE(status_version, 1) + ?", 1),
	}

	if status == "resolved" || status == "closed" {
		updates["resolved_at"] = gorm.Expr("NOW()")
	}

	return r.db.Model(&ticket.Ticket{}).
		Where("id = ?", id).
		Where("COALESCE(status, '') <> ?", status).
		Updates(updates).Error
}

// UpdateTicketStatusForUpdate applies a status transition to a row that was
// already locked by FindTicketByIDForUpdate. The version is part of the
// compare-and-update predicate so the emitted event identifies one committed
// mutation, even if another caller reaches this repository method later.
func (r *TicketRepository) UpdateTicketStatusForUpdate(current *ticket.Ticket, status string) (uint, bool, error) {
	if current == nil || current.ID == 0 {
		return 0, false, gorm.ErrInvalidData
	}
	if current.Status == status {
		return current.StatusVersion, false, nil
	}

	currentVersion := current.StatusVersion
	if currentVersion == 0 {
		currentVersion = 1
	}
	nextVersion := currentVersion + 1
	updates := map[string]interface{}{
		"status":         status,
		"status_version": nextVersion,
	}
	if status == "resolved" || status == "closed" {
		updates["resolved_at"] = gorm.Expr("NOW()")
	}

	result := r.db.Model(&ticket.Ticket{}).
		Where("id = ? AND COALESCE(status, '') = ? AND status_version = ?", current.ID, current.Status, currentVersion).
		Updates(updates)
	if result.Error != nil {
		return 0, false, result.Error
	}
	if result.RowsAffected != 1 {
		return 0, false, ErrTicketStatusVersionConflict
	}
	return nextVersion, true, nil
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

func (r *TicketRepository) TouchTicket(id uint, updatedAt time.Time) error {
	if id == 0 {
		return gorm.ErrInvalidData
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	return r.db.Model(&ticket.Ticket{}).Where("id = ?", id).Update("updated_at", updatedAt.UTC()).Error
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

// RecordCustomerServiceInboxCustomerMessage updates the assigned recipient's
// materialized unread count. The read cursor remains authoritative, so list
// queries reconcile the count from ticket_messages when necessary.
func (r *TicketRepository) RecordCustomerServiceInboxCustomerMessage(ticketID, recipientUserID, messageID uint) error {
	if ticketID == 0 || recipientUserID == 0 || messageID == 0 {
		return gorm.ErrInvalidData
	}

	now := time.Now().UTC()
	state := ticket.CustomerServiceInboxState{
		TicketID:          ticketID,
		RecipientUserID:   recipientUserID,
		UnreadCount:       1,
		AssignmentVersion: 1,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "recipient_user_id"},
			{Name: "ticket_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"unread_count": gorm.Expr(
				"CASE WHEN customer_service_inbox_states.last_read_message_id < ? THEN customer_service_inbox_states.unread_count + 1 ELSE customer_service_inbox_states.unread_count END",
				messageID,
			),
			"deleted_at": nil,
			"updated_at": now,
		}),
	}).Create(&state).Error
}

// MarkCustomerServiceInboxRead advances only the requesting staff member's
// cursor. It deliberately does not update ticket_messages.is_read because
// that global field cannot express multi-agent inbox state.
func (r *TicketRepository) MarkCustomerServiceInboxRead(ticketID, recipientUserID uint, readAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		_, _, err := r.WithTx(tx).AdvanceCustomerServiceInboxRead(ticketID, recipientUserID, readAt)
		return err
	})
}

// AdvanceCustomerServiceInboxRead advances a recipient's cursor on the
// caller's transaction and reports whether a durable read fact changed. The
// result lets the service create a matching Outbox event in the same commit.
func (r *TicketRepository) AdvanceCustomerServiceInboxRead(ticketID, recipientUserID uint, readAt time.Time) (ticket.CustomerServiceInboxState, bool, error) {
	if ticketID == 0 || recipientUserID == 0 {
		return ticket.CustomerServiceInboxState{}, false, gorm.ErrInvalidData
	}
	if readAt.IsZero() {
		readAt = time.Now().UTC()
	} else {
		readAt = readAt.UTC()
	}

	var latestCustomerMessageID uint
	if err := r.db.Model(&ticket.TicketMessage{}).
		Where("ticket_id = ? AND is_staff = ?", ticketID, false).
		Select("COALESCE(MAX(id), 0)").
		Scan(&latestCustomerMessageID).Error; err != nil {
		return ticket.CustomerServiceInboxState{}, false, err
	}
	if latestCustomerMessageID == 0 {
		return ticket.CustomerServiceInboxState{}, false, nil
	}

	var existing ticket.CustomerServiceInboxState
	err := r.db.Where("ticket_id = ? AND recipient_user_id = ?", ticketID, recipientUserID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return ticket.CustomerServiceInboxState{}, false, err
	}
	if err == nil && existing.LastReadMessageID >= latestCustomerMessageID && existing.UnreadCount == 0 {
		return existing, false, nil
	}

	state := ticket.CustomerServiceInboxState{
		TicketID:          ticketID,
		RecipientUserID:   recipientUserID,
		LastReadMessageID: latestCustomerMessageID,
		UnreadCount:       0,
		AssignmentVersion: 1,
		LastReadAt:        &readAt,
		CreatedAt:         readAt,
		UpdatedAt:         readAt,
	}
	if err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "recipient_user_id"},
			{Name: "ticket_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_read_message_id": latestCustomerMessageID,
			"unread_count":         0,
			"last_read_at":         readAt,
			"deleted_at":           nil,
			"updated_at":           readAt,
		}),
	}).Create(&state).Error; err != nil {
		return ticket.CustomerServiceInboxState{}, false, err
	}
	if err := r.db.Where("ticket_id = ? AND recipient_user_id = ?", ticketID, recipientUserID).First(&state).Error; err != nil {
		return ticket.CustomerServiceInboxState{}, false, err
	}
	return state, true, nil
}

// ResetCustomerServiceInboxAssignment gives a newly assigned support agent a
// fresh cursor over the conversation. Assignment history remains in the row
// through AssignmentVersion while old recipients keep their own read state.
func (r *TicketRepository) ResetCustomerServiceInboxAssignment(ticketID, recipientUserID uint, assignedAt time.Time) error {
	_, err := r.ResetCustomerServiceInboxAssignmentState(ticketID, recipientUserID, assignedAt)
	return err
}

// ResetCustomerServiceInboxAssignmentState resets one newly assigned agent's
// cursor and returns the persisted assignment version from the same
// transaction. That version forms a stable identity for an assignment event.
func (r *TicketRepository) ResetCustomerServiceInboxAssignmentState(ticketID, recipientUserID uint, assignedAt time.Time) (ticket.CustomerServiceInboxState, error) {
	if ticketID == 0 || recipientUserID == 0 {
		return ticket.CustomerServiceInboxState{}, gorm.ErrInvalidData
	}
	if assignedAt.IsZero() {
		assignedAt = time.Now().UTC()
	} else {
		assignedAt = assignedAt.UTC()
	}

	var unreadCount int64
	if err := r.db.Model(&ticket.TicketMessage{}).
		Where("ticket_id = ? AND is_staff = ?", ticketID, false).
		Count(&unreadCount).Error; err != nil {
		return ticket.CustomerServiceInboxState{}, err
	}

	state := ticket.CustomerServiceInboxState{
		TicketID:          ticketID,
		RecipientUserID:   recipientUserID,
		LastReadMessageID: 0,
		UnreadCount:       int(unreadCount),
		AssignmentVersion: 1,
		CreatedAt:         assignedAt,
		UpdatedAt:         assignedAt,
	}
	if err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "recipient_user_id"},
			{Name: "ticket_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_read_message_id": 0,
			"unread_count":         int(unreadCount),
			"assignment_version":   gorm.Expr("customer_service_inbox_states.assignment_version + 1"),
			"last_read_at":         nil,
			"deleted_at":           nil,
			"updated_at":           assignedAt,
		}),
	}).Create(&state).Error; err != nil {
		return ticket.CustomerServiceInboxState{}, err
	}
	if err := r.db.Where("ticket_id = ? AND recipient_user_id = ?", ticketID, recipientUserID).First(&state).Error; err != nil {
		return ticket.CustomerServiceInboxState{}, err
	}
	return state, nil
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
	return r.CreateAutoReplyMessageIfNotRecentWithTx(message, dedupeKey, content, since, nil)
}

// CreateAutoReplyMessageIfNotRecentWithTx keeps the cooldown check, automatic
// reply insert, and an optional dependent write in one database transaction.
// The callback only runs when this invocation inserted a new reply.
func (r *TicketRepository) CreateAutoReplyMessageIfNotRecentWithTx(
	message *ticket.TicketMessage,
	dedupeKey, content string,
	since time.Time,
	afterCreate func(*TicketRepository, *gorm.DB) error,
) (bool, error) {
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
		if afterCreate != nil {
			if err := afterCreate(r.WithTx(tx), tx); err != nil {
				return err
			}
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
