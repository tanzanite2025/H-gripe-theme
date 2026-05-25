package service

import (
	"errors"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/repository"
)

type TicketService struct {
	ticketRepo *repository.TicketRepository
}

func NewTicketService(ticketRepo *repository.TicketRepository) *TicketService {
	return &TicketService{
		ticketRepo: ticketRepo,
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
