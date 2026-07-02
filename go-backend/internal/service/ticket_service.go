package service

import (
	"errors"
	"strings"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrCustomerServiceConversationAccessDenied = errors.New("conversation access denied")
	ErrCustomerServiceOwnerRequired            = errors.New("conversation owner is required")
)

type CustomerServiceOwner struct {
	UserID             *uint
	VisitorSessionHash string
}

type TicketService struct {
	ticketRepo *repository.TicketRepository
	userRepo   *repository.UserRepository
}

func NewTicketService(ticketRepo *repository.TicketRepository, userRepo *repository.UserRepository) *TicketService {
	return &TicketService{
		ticketRepo: ticketRepo,
		userRepo:   userRepo,
	}
}

// CreateTicket 创建工单
func (s *TicketService) CreateTicket(t *ticket.Ticket) error {
	// 设置默认状态
	t.Status = "open"
	t.Priority = "medium" // 默认优先级

	return s.ticketRepo.CreateTicket(t)
}

// GetTicket 获取工单详情
func (s *TicketService) GetTicket(id uint, userID uint, isStaff bool) (*ticket.Ticket, error) {
	t, err := s.ticketRepo.FindTicketByID(id)
	if err != nil {
		return nil, err
	}

	// 验证权限
	if !isStaff && t.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// 标记消息为已读
	if err := s.ticketRepo.MarkMessagesAsRead(id, isStaff); err != nil {
		return nil, err
	}

	return t, nil
}

// GetUserTickets 获取用户工单列表
func (s *TicketService) GetUserTickets(userID uint, page, pageSize int) ([]ticket.Ticket, int64, error) {
	return s.ticketRepo.FindTicketsByUserID(userID, page, pageSize)
}

// GetAllTickets 获取所有工单（管理员）
func (s *TicketService) GetAllTickets(page, pageSize int, status, priority string) ([]ticket.Ticket, int64, error) {
	return s.ticketRepo.FindAllTickets(page, pageSize, status, priority)
}

func (s *TicketService) GetCustomerServiceConversations(page, pageSize int) ([]ticket.Ticket, int64, error) {
	return s.ticketRepo.FindCustomerServiceConversations(page, pageSize)
}

func (s *TicketService) HasPublicCustomerServiceConversation(owner CustomerServiceOwner) (bool, string, uint, error) {
	t, err := s.findCustomerServiceConversationByOwner(owner)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, "", 0, nil
	}
	if err != nil {
		return false, "", 0, err
	}
	return true, ticketConversationID(t), t.AssignedTo, nil
}

func (s *TicketService) GetOrCreatePublicCustomerServiceConversation(owner CustomerServiceOwner, agentID uint) (*ticket.Ticket, error) {
	owner = normalizeCustomerServiceOwner(owner)
	if !owner.Valid() {
		return nil, ErrCustomerServiceOwnerRequired
	}

	t, err := s.findCustomerServiceConversationByOwner(owner)
	if err == nil {
		if err := s.updateCustomerServiceConversationOwner(t, owner, agentID); err != nil {
			return nil, err
		}
		return t, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	persistedUserID, err := s.customerServicePersistedUserID(owner.UserID, agentID)
	if err != nil {
		return nil, err
	}

	conversationID := uuid.NewString()
	t = &ticket.Ticket{
		UserID:             persistedUserID,
		CustomerUserID:     owner.UserID,
		ConversationID:     &conversationID,
		VisitorSessionHash: owner.VisitorSessionHash,
		Subject:            "Customer service chat",
		Category:           "customer_service",
		Priority:           "medium",
		Status:             "open",
		AssignedTo:         agentID,
		Tags:               customerServiceConversationTag(conversationID),
	}
	if err := s.CreateTicket(t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *TicketService) AddPublicCustomerServiceMessage(conversationID string, owner CustomerServiceOwner, message string, agentID uint) (*ticket.Ticket, *ticket.TicketMessage, error) {
	t, err := s.getOrCreateAccessibleCustomerServiceConversation(conversationID, owner, agentID)
	if err != nil {
		return nil, nil, err
	}

	persistedUserID := t.UserID
	if owner.UserID != nil && *owner.UserID > 0 {
		persistedUserID = *owner.UserID
	}

	msg := &ticket.TicketMessage{
		TicketID:   t.ID,
		UserID:     persistedUserID,
		IsStaff:    false,
		Content:    message,
		IsRead:     false,
		IsInternal: false,
	}
	if err := s.ticketRepo.CreateTicketMessage(msg); err != nil {
		return nil, nil, err
	}

	return t, msg, nil
}

// GetWelcomeMessage 获取欢迎语。如果24小时内未发送过，则会自动创建一条客服自动回复消息。
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

	// 24小时内未发送过，自动插入一条消息（UserID = 0 表示系统自动回复）
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

// MatchKeywordMessage 关键字匹配自动回复。如果匹配到，会自动插入一条客服自动回复消息。
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

	// 自动回复一条消息（UserID = 0 表示系统自动回复）
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

func (s *TicketService) GetPublicCustomerServiceMessages(conversationID string, owner CustomerServiceOwner, limit, offset int) ([]ticket.TicketMessage, error) {
	t, err := s.getAccessibleCustomerServiceConversation(conversationID, owner)
	if err != nil {
		return nil, err
	}
	messages, err := s.ticketRepo.FindMessagesByTicketID(t.ID)
	if err != nil {
		return nil, err
	}
	if offset < 0 {
		offset = 0
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset >= len(messages) {
		return []ticket.TicketMessage{}, nil
	}
	end := offset + limit
	if end > len(messages) {
		end = len(messages)
	}
	return messages[offset:end], nil
}

func (s *TicketService) CanAccessCustomerServiceConversation(conversationID string, owner CustomerServiceOwner) (bool, error) {
	_, err := s.getAccessibleCustomerServiceConversation(conversationID, owner)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrCustomerServiceConversationAccessDenied) || errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (s *TicketService) getOrCreateAccessibleCustomerServiceConversation(conversationID string, owner CustomerServiceOwner, agentID uint) (*ticket.Ticket, error) {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return s.GetOrCreatePublicCustomerServiceConversation(owner, agentID)
	}

	t, err := s.getAccessibleCustomerServiceConversation(conversationID, owner)
	if err != nil {
		return nil, err
	}
	if err := s.updateCustomerServiceConversationOwner(t, normalizeCustomerServiceOwner(owner), agentID); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TicketService) getAccessibleCustomerServiceConversation(conversationID string, owner CustomerServiceOwner) (*ticket.Ticket, error) {
	owner = normalizeCustomerServiceOwner(owner)
	if strings.TrimSpace(conversationID) == "" || !owner.Valid() {
		return nil, ErrCustomerServiceConversationAccessDenied
	}

	t, err := s.ticketRepo.FindCustomerServiceConversationByConversationID(strings.TrimSpace(conversationID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerServiceConversationAccessDenied
		}
		return nil, err
	}

	if customerServiceOwnerMatches(t, owner) {
		return t, nil
	}
	return nil, ErrCustomerServiceConversationAccessDenied
}

func (s *TicketService) findCustomerServiceConversationByOwner(owner CustomerServiceOwner) (*ticket.Ticket, error) {
	owner = normalizeCustomerServiceOwner(owner)
	if !owner.Valid() {
		return nil, gorm.ErrRecordNotFound
	}

	if owner.UserID != nil {
		t, err := s.ticketRepo.FindCustomerServiceConversationByOwner(owner.UserID, "")
		if err == nil {
			return t, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	if owner.VisitorSessionHash != "" {
		return s.ticketRepo.FindCustomerServiceConversationByOwner(nil, owner.VisitorSessionHash)
	}

	return nil, gorm.ErrRecordNotFound
}

func (s *TicketService) updateCustomerServiceConversationOwner(t *ticket.Ticket, owner CustomerServiceOwner, agentID uint) error {
	changed := false
	if t.ConversationID == nil || strings.TrimSpace(*t.ConversationID) == "" {
		conversationID := uuid.NewString()
		t.ConversationID = &conversationID
		t.Tags = customerServiceConversationTag(conversationID)
		changed = true
	}
	if t.CustomerUserID == nil && owner.UserID != nil {
		t.CustomerUserID = owner.UserID
		changed = true
	}
	if t.VisitorSessionHash == "" && owner.VisitorSessionHash != "" {
		t.VisitorSessionHash = owner.VisitorSessionHash
		changed = true
	}
	if agentID > 0 && t.AssignedTo != agentID {
		t.AssignedTo = agentID
		changed = true
	}
	if t.Status == "" || t.Status == "closed" || t.Status == "resolved" {
		t.Status = "open"
		changed = true
	}
	if t.UserID == 0 {
		persistedUserID, err := s.customerServicePersistedUserID(owner.UserID, agentID)
		if err != nil {
			return err
		}
		t.UserID = persistedUserID
		changed = true
	}
	if !changed {
		return nil
	}
	return s.ticketRepo.UpdateTicket(t)
}

func (s *TicketService) customerServicePersistedUserID(userID *uint, agentID uint) (uint, error) {
	if userID != nil && *userID > 0 {
		return *userID, nil
	}
	if agentID > 0 {
		return agentID, nil
	}

	agents, err := s.ListCustomerServiceAgentProfiles(1)
	if err != nil {
		return 0, err
	}
	if len(agents) == 0 {
		return 0, errors.New("no customer service agents configured")
	}
	if agents[0].UserID == nil {
		return 0, errors.New("customer service agent is not linked to a Go user")
	}
	return *agents[0].UserID, nil
}

func normalizeCustomerServiceOwner(owner CustomerServiceOwner) CustomerServiceOwner {
	owner.VisitorSessionHash = strings.TrimSpace(owner.VisitorSessionHash)
	if owner.UserID != nil && *owner.UserID == 0 {
		owner.UserID = nil
	}
	return owner
}

func (owner CustomerServiceOwner) Valid() bool {
	return owner.UserID != nil || strings.TrimSpace(owner.VisitorSessionHash) != ""
}

func customerServiceOwnerMatches(t *ticket.Ticket, owner CustomerServiceOwner) bool {
	if owner.UserID != nil && t.CustomerUserID != nil && *t.CustomerUserID == *owner.UserID {
		return true
	}
	return owner.VisitorSessionHash != "" && t.VisitorSessionHash == owner.VisitorSessionHash
}

func ticketConversationID(t *ticket.Ticket) string {
	if t == nil || t.ConversationID == nil {
		return ""
	}
	return strings.TrimSpace(*t.ConversationID)
}

func (s *TicketService) ListCustomerServiceAgents(limit int) ([]user.User, error) {
	return s.userRepo.FindCustomerServiceAgents(limit)
}

func (s *TicketService) ListCustomerServiceAgentProfiles(limit int) ([]user.AgentProfile, error) {
	return s.userRepo.FindCustomerServiceAgentProfiles(limit)
}

// GetAssignedTickets 获取分配给客服的工单
func (s *TicketService) GetAssignedTickets(assignedTo uint, page, pageSize int) ([]ticket.Ticket, int64, error) {
	return s.ticketRepo.FindTicketsByAssignedTo(assignedTo, page, pageSize)
}

// UpdateTicket 更新工单
func (s *TicketService) UpdateTicket(t *ticket.Ticket, userID uint, isStaff bool) error {
	existing, err := s.ticketRepo.FindTicketByID(t.ID)
	if err != nil {
		return err
	}

	// 验证权限
	if !isStaff && existing.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.ticketRepo.UpdateTicket(t)
}

// UpdateTicketStatus 更新工单状态
func (s *TicketService) UpdateTicketStatus(id uint, status string) error {
	// 验证状态
	validStatuses := []string{"open", "in_progress", "resolved", "closed"}
	isValid := false
	for _, s := range validStatuses {
		if s == status {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.New("invalid status")
	}

	return s.ticketRepo.UpdateTicketStatus(id, status)
}

// AssignTicket 分配工单
func (s *TicketService) AssignTicket(id, assignedTo uint) error {
	// 更新分配
	if err := s.ticketRepo.AssignTicket(id, assignedTo); err != nil {
		return err
	}

	// 更新状态为处理中
	return s.ticketRepo.UpdateTicketStatus(id, "in_progress")
}

// CloseTicket 关闭工单
func (s *TicketService) CloseTicket(id uint, userID uint, isStaff bool) error {
	t, err := s.ticketRepo.FindTicketByID(id)
	if err != nil {
		return err
	}

	// 验证权限
	if !isStaff && t.UserID != userID {
		return errors.New("unauthorized")
	}

	// 只有resolved状态可以关闭
	if t.Status != "resolved" {
		return errors.New("only resolved tickets can be closed")
	}

	return s.ticketRepo.UpdateTicketStatus(id, "closed")
}

// DeleteTicket 删除工单
func (s *TicketService) DeleteTicket(id uint, userID uint, isStaff bool) error {
	t, err := s.ticketRepo.FindTicketByID(id)
	if err != nil {
		return err
	}

	// 验证权限（只有管理员或工单所有者可以删除）
	if !isStaff && t.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.ticketRepo.DeleteTicket(id)
}

// GetTicketStats 获取工单统计
func (s *TicketService) GetTicketStats(userID uint) (map[string]int64, error) {
	return s.ticketRepo.GetTicketStats(userID)
}

// Message 相关方法

// AddMessage 添加消息
func (s *TicketService) AddMessage(m *ticket.TicketMessage, userID uint, isStaff bool) error {
	// 验证工单权限
	t, err := s.ticketRepo.FindTicketByID(m.TicketID)
	if err != nil {
		return err
	}

	if !isStaff && t.UserID != userID {
		return errors.New("unauthorized")
	}

	m.UserID = userID
	m.IsStaff = isStaff

	if err := s.ticketRepo.CreateTicketMessage(m); err != nil {
		return err
	}

	// 如果工单是关闭状态，重新打开
	if t.Status == "closed" {
		if err := s.ticketRepo.UpdateTicketStatus(t.ID, "open"); err != nil {
			return err
		}
	}

	return nil
}

// GetMessages 获取工单消息列表
func (s *TicketService) GetMessages(ticketID uint, userID uint, isStaff bool) ([]ticket.TicketMessage, error) {
	// 验证权限
	t, err := s.ticketRepo.FindTicketByID(ticketID)
	if err != nil {
		return nil, err
	}

	if !isStaff && t.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return s.ticketRepo.FindMessagesByTicketID(ticketID)
}

// DeleteMessage 删除消息
func (s *TicketService) DeleteMessage(id uint, userID uint, isStaff bool) error {
	m, err := s.ticketRepo.FindMessageByID(id)
	if err != nil {
		return err
	}

	// 验证权限（只有消息作者或管理员可以删除）
	if !isStaff && m.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.ticketRepo.DeleteTicketMessage(id)
}

// CountUnreadMessages 统计未读消息
func (s *TicketService) CountUnreadMessages(ticketID uint, isStaff bool) (int64, error) {
	return s.ticketRepo.CountUnreadMessages(ticketID, isStaff)
}

func (s *TicketService) MarkMessagesAsRead(ticketID uint, isStaff bool) error {
	return s.ticketRepo.MarkMessagesAsRead(ticketID, isStaff)
}

func customerServiceConversationTag(conversationID string) string {
	return "conversation_id:" + conversationID
}

// GetDashboard 获取客服仪表板数据
func (s *TicketService) GetDashboard() (map[string]interface{}, error) {
	dashboard := make(map[string]interface{})

	// 获取基础统计
	stats, err := s.ticketRepo.GetTicketStats(0)
	if err != nil {
		return nil, err
	}
	dashboard["stats"] = stats

	// 计算总工单数
	totalTickets := stats["open"] + stats["in_progress"] + stats["resolved"] + stats["closed"]
	dashboard["total_tickets"] = totalTickets

	// 计算待处理工单数（open + in_progress）
	pendingTickets := stats["open"] + stats["in_progress"]
	dashboard["pending_tickets"] = pendingTickets

	// 计算紧急工单数
	dashboard["urgent_tickets"] = stats["urgent"]

	// 计算解决率
	if totalTickets > 0 {
		resolvedRate := float64(stats["resolved"]+stats["closed"]) / float64(totalTickets) * 100
		dashboard["resolved_rate"] = resolvedRate
	} else {
		dashboard["resolved_rate"] = 0.0
	}

	return dashboard, nil
}

// GetRecentTickets 获取最近的工单
func (s *TicketService) GetRecentTickets(limit int) ([]ticket.Ticket, error) {
	tickets, _, err := s.ticketRepo.FindAllTickets(1, limit, "", "")
	return tickets, err
}
