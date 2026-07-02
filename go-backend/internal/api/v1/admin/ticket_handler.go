package admin

import (
	"errors"
	"strconv"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketHandler struct {
	ticketService *service.TicketService
}

func NewTicketHandler(ticketService *service.TicketService) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
	}
}

// ListTickets 获取工单列表
// GET /api/admin/tickets
func (h *TicketHandler) ListTickets(c *gin.Context) {
	params := pagination.ParsePagination(c)
	status := c.Query("status")
	priority := c.Query("priority")

	tickets, total, err := h.ticketService.GetAllTickets(params.Page, params.PageSize, status, priority)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize

	response.Success(c, gin.H{
		"tickets": tickets,
		"pagination": gin.H{
			"page":        params.Page,
			"page_size":   params.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetTicket 获取工单详情
// GET /api/admin/tickets/:id
func (h *TicketHandler) GetTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid ticket ID")
		return
	}

	ticketItem, err := h.ticketService.GetTicket(uint(id), 0, true)
	if err != nil {
		apierror.RespondNotFound(c, "Ticket")
		return
	}

	response.Success(c, gin.H{
		"ticket": ticketItem,
	})
}

// UpdateTicketStatus 更新工单状态
// PATCH /api/admin/tickets/:id/status
func (h *TicketHandler) UpdateTicketStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid ticket ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=open in_progress resolved closed"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.ticketService.UpdateTicketStatus(uint(id), req.Status); err != nil {
		if errors.Is(err, service.ErrInvalidTicketStatus) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Ticket status updated successfully", nil)
}

// AssignTicket 分配工单
// PATCH /api/admin/tickets/:id/assign
func (h *TicketHandler) AssignTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid ticket ID")
		return
	}

	var req struct {
		AssignedTo uint `json:"assigned_to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.ticketService.AssignTicket(uint(id), req.AssignedTo); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Ticket assigned successfully", nil)
}

// UpdateTicket 更新工单
// PUT /api/admin/tickets/:id
func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid ticket ID")
		return
	}

	var req struct {
		Subject    string `json:"subject"`
		Priority   string `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
		Status     string `json:"status" binding:"omitempty,oneof=open in_progress resolved closed"`
		AssignedTo *uint  `json:"assigned_to"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	updatedTicket, err := h.ticketService.UpdateAdminTicket(uint(id), service.TicketAdminUpdateInput{
		Subject:    req.Subject,
		Priority:   req.Priority,
		Status:     req.Status,
		AssignedTo: req.AssignedTo,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidTicketPriority) || errors.Is(err, service.ErrInvalidTicketStatus) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierror.RespondNotFound(c, "Ticket")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Ticket updated successfully", gin.H{
		"ticket": updatedTicket,
	})
}

// DeleteTicket 删除工单
// DELETE /api/admin/tickets/:id
func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid ticket ID")
		return
	}

	if err := h.ticketService.DeleteTicket(uint(id), 0, true); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierror.RespondNotFound(c, "Ticket")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Ticket deleted successfully", nil)
}

// GetTicketStats 获取工单统计
// GET /api/admin/tickets/stats
func (h *TicketHandler) GetTicketStats(c *gin.Context) {
	stats, err := h.ticketService.GetAdminTicketStats()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, stats)
}

// CreateMessage 创建工单消息
// POST /api/admin/tickets/:id/messages
func (h *TicketHandler) CreateMessage(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid ticket ID")
		return
	}

	var req struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	newMessage := &ticket.TicketMessage{
		TicketID: uint(ticketID),
		UserID:   userID.(uint),
		Content:  req.Message,
		IsStaff:  true, // 管理员发送的消息
	}

	if err := h.ticketService.AddMessage(newMessage, userID.(uint), true); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Created(c, gin.H{
		"message": "Message created successfully",
		"data":    newMessage,
	})
}

// GetMessages 获取工单消息列表
// GET /api/admin/tickets/:id/messages
func (h *TicketHandler) GetMessages(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid ticket ID")
		return
	}

	messages, err := h.ticketService.GetMessages(uint(ticketID), 0, true)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"messages": messages,
	})
}

// MarkMessagesAsRead 标记消息为已读
// POST /api/admin/tickets/:id/messages/mark-read
func (h *TicketHandler) MarkMessagesAsRead(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid ticket ID")
		return
	}

	if err := h.ticketService.MarkMessagesAsRead(uint(ticketID), true); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Messages marked as read", nil)
}
