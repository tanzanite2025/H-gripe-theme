package repository

import (
	"tanzanite/internal/domain/ticket"
	"time"

	"gorm.io/gorm"
)

type TicketRepository struct {
	db *gorm.DB
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

// FindTicketsByUserID 查找用户的工单列表
func (r *TicketRepository) FindTicketsByUserID(userID uint, page, pageSize int) ([]ticket.Ticket, int64, error) {
	var tickets []ticket.Ticket
	var total int64

	query := r.db.Model(&ticket.Ticket{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&tickets).Error

	return tickets, total, err
}

// FindAllTickets 查找所有工单（管理员）
func (r *TicketRepository) FindAllTickets(page, pageSize int, status, priority string) ([]ticket.Ticket, int64, error) {
	var tickets []ticket.Ticket
	var total int64

	query := r.db.Model(&ticket.Ticket{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("User").Order("updated_at DESC").
		Offset(offset).Limit(pageSize).Find(&tickets).Error

	return tickets, total, err
}

func (r *TicketRepository) FindCustomerServiceConversations(page, pageSize int) ([]ticket.Ticket, int64, error) {
	var tickets []ticket.Ticket
	var total int64

	query := r.db.Model(&ticket.Ticket{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("User").Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&tickets).Error

	return tickets, total, err
}

func (r *TicketRepository) FindCustomerServiceConversationByTag(tag string) (*ticket.Ticket, error) {
	var t ticket.Ticket
	err := r.db.Preload("User").Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("category = ? AND tags = ?", "customer_service", tag).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FindTicketsByAssignedTo 查找分配给某客服的工单
func (r *TicketRepository) FindTicketsByAssignedTo(assignedTo uint, page, pageSize int) ([]ticket.Ticket, int64, error) {
	var tickets []ticket.Ticket
	var total int64

	query := r.db.Model(&ticket.Ticket{}).Where("assigned_to = ?", assignedTo)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("User").Order("updated_at DESC").
		Offset(offset).Limit(pageSize).Find(&tickets).Error

	return tickets, total, err
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

// DeleteTicket 删除工单
func (r *TicketRepository) DeleteTicket(id uint) error {
	// 先删除关联的消息
	if err := r.db.Where("ticket_id = ?", id).Delete(&ticket.TicketMessage{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&ticket.Ticket{}, id).Error
}

// GetTicketStats 获取工单统计
func (r *TicketRepository) GetTicketStats(userID uint) (map[string]int64, error) {
	stats := make(map[string]int64)

	query := r.db.Model(&ticket.Ticket{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// 统计各状态工单数量
	statuses := []string{"open", "in_progress", "resolved", "closed"}
	for _, status := range statuses {
		var count int64
		if err := query.Where("status = ?", status).Count(&count).Error; err != nil {
			return nil, err
		}
		stats[status] = count
	}

	// 统计各优先级工单数量
	priorities := []string{"low", "medium", "high", "urgent"}
	for _, priority := range priorities {
		var count int64
		if err := query.Where("priority = ?", priority).Count(&count).Error; err != nil {
			return nil, err
		}
		stats[priority] = count
	}

	return stats, nil
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

// FindMessageByID 根据ID查找消息
func (r *TicketRepository) FindMessageByID(id uint) (*ticket.TicketMessage, error) {
	var m ticket.TicketMessage
	err := r.db.Preload("User").First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// UpdateTicketMessage 更新消息
func (r *TicketRepository) UpdateTicketMessage(m *ticket.TicketMessage) error {
	return r.db.Save(m).Error
}

// DeleteTicketMessage 删除消息
func (r *TicketRepository) DeleteTicketMessage(id uint) error {
	return r.db.Delete(&ticket.TicketMessage{}, id).Error
}

// CountUnreadMessages 统计未读消息数
func (r *TicketRepository) CountUnreadMessages(ticketID uint, isStaff bool) (int64, error) {
	var count int64
	err := r.db.Model(&ticket.TicketMessage{}).
		Where("ticket_id = ? AND is_staff = ? AND is_read = ?", ticketID, !isStaff, false).
		Count(&count).Error
	return count, err
}

// MarkMessagesAsRead 标记消息为已读
func (r *TicketRepository) MarkMessagesAsRead(ticketID uint, isStaff bool) error {
	return r.db.Model(&ticket.TicketMessage{}).
		Where("ticket_id = ? AND is_staff = ?", ticketID, !isStaff).
		Update("is_read", true).Error
}

// GetStats 获取工单统计
func (r *TicketRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总工单数
	var total int64
	if err := r.db.Model(&ticket.Ticket{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&ticket.Ticket{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusStats).Error; err != nil {
		return nil, err
	}

	for _, stat := range statusStats {
		stats[stat.Status] = stat.Count
	}

	// 按优先级统计
	var priorityStats []struct {
		Priority string
		Count    int64
	}
	if err := r.db.Model(&ticket.Ticket{}).Select("priority, COUNT(*) as count").Group("priority").Scan(&priorityStats).Error; err != nil {
		return nil, err
	}

	priorityMap := make(map[string]int64)
	for _, stat := range priorityStats {
		priorityMap[stat.Priority] = stat.Count
	}
	stats["by_priority"] = priorityMap

	return stats, nil
}

// FindRecent 获取最近工单
func (r *TicketRepository) FindRecent(limit int) ([]ticket.Ticket, error) {
	var tickets []ticket.Ticket
	err := r.db.Order("created_at DESC").Limit(limit).Find(&tickets).Error
	return tickets, err
}

// GetActiveAutoReplyRules 获取激活的自动回复规则
func (r *TicketRepository) GetActiveAutoReplyRules(ruleType string) ([]ticket.AutoReplyRule, error) {
	var rules []ticket.AutoReplyRule
	query := r.db.Model(&ticket.AutoReplyRule{}).Where("type = ? AND is_active = ?", ruleType, true)
	if ruleType == "welcome" {
		err := query.Order("created_at DESC").Limit(1).Find(&rules).Error
		return rules, err
	}
	err := query.Order("priority DESC, created_at DESC").Find(&rules).Error
	return rules, err
}

// GetLastWelcomeMessageTime 获取特定工单下最新一条欢迎自动回复的发送时间
func (r *TicketRepository) GetLastWelcomeMessageTime(ticketID uint, content string) (time.Time, error) {
	var m ticket.TicketMessage
	err := r.db.Model(&ticket.TicketMessage{}).
		Where("ticket_id = ? AND user_id = ? AND content = ?", ticketID, 0, content).
		Order("created_at DESC").First(&m).Error
	if err != nil {
		return time.Time{}, err
	}
	return m.CreatedAt, nil
}
