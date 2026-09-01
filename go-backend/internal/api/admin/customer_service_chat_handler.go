package admin

import (
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListCustomerServiceConversations returns the backoffice chat inbox conversation list.
// The inbox is scoped to the current backend account so transferred conversations
// move cleanly between agents instead of pooling every support record together.
func (h *TicketHandler) ListCustomerServiceConversations(c *gin.Context) {
	params := pagination.ParsePagination(c)

	filters, ok := parseAdminCustomerServiceConversationFilters(c)
	if !ok {
		return
	}
	filters, agentUserID, canViewAll := scopeAdminCustomerServiceConversationFilters(c, filters)

	tickets, total, err := h.ticketService.ListCustomerServiceConversationsForAgent(params.Page, params.PageSize, agentUserID, canViewAll, filters)
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	items := make([]gin.H, 0, len(tickets))
	for _, item := range tickets {
		var summary *service.CustomerServiceConversationSummary
		if h.customerServiceContext != nil {
			itemSummary := h.customerServiceContext.ConversationListSummary(item)
			summary = &itemSummary
		}
		items = append(items, adminCustomerServiceConversationResponse(item, summary))
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize
	response.Success(c, gin.H{
		"conversations": items,
		"pagination": gin.H{
			"page":        params.Page,
			"page_size":   params.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
		"filters": adminCustomerServiceConversationFilterResponse(filters),
	})
}

// ListCustomerServiceAgents returns assignable public chat staff profiles for the admin inbox.
func (h *TicketHandler) ListCustomerServiceAgents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit < 1 || limit > 500 {
		limit = 100
	}

	agents, err := h.ticketService.ListCustomerServiceAgentProfiles(limit)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	groups, err := h.ticketService.ListCustomerServiceAgentGroups(500, false)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	items := make([]gin.H, 0, len(agents))
	for _, agent := range agents {
		if agent.UserID == nil {
			continue
		}
		items = append(items, gin.H{
			"id":            agent.ID,
			"user_id":       *agent.UserID,
			"agent_id":      agent.AgentID,
			"name":          agent.DisplayName(),
			"email":         agent.PublicEmail(),
			"avatar":        agent.Avatar,
			"whatsapp":      agent.WhatsApp,
			"online_status": agent.OnlineStatus,
			"status":        agent.Status,
			"group_ids":     adminCustomerServiceAgentGroupIDs(agent.Groups),
			"groups":        adminCustomerServiceAgentGroupsResponse(agent.Groups),
			"primary_group": adminCustomerServicePrimaryAgentGroup(agent.Groups),
		})
	}

	response.Success(c, gin.H{
		"agents": items,
		"groups": adminCustomerServiceGroupsResponse(groups),
	})
}

// ListCustomerServiceGroups returns active groups for inbox filters and routing selectors.
func (h *TicketHandler) ListCustomerServiceGroups(c *gin.Context) {
	groups, err := h.ticketService.ListCustomerServiceAgentGroups(500, false)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"groups": adminCustomerServiceGroupsResponse(groups)})
}

// GetCustomerServiceConversationContext returns the customer snapshot beside one chat.
func (h *TicketHandler) GetCustomerServiceConversationContext(c *gin.Context) {
	if h.customerServiceContext == nil {
		apierror.RespondInternalError(c, errors.New("customer service context service is not configured"))
		return
	}

	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	context, err := h.customerServiceContext.GetConversationContextForAgent(ticketID, agentUserID, canViewAll)
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"context": context})
}

// GetCustomerServiceConversationMessages returns messages for one conversation.
func (h *TicketHandler) GetCustomerServiceConversationMessages(c *gin.Context) {
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	messages, err := h.ticketService.GetCustomerServiceMessagesForAgent(ticketID, agentUserID, canViewAll)
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	items := make([]gin.H, 0, len(messages))
	for _, item := range messages {
		items = append(items, adminCustomerServiceMessageResponse(item))
	}

	response.Success(c, gin.H{"messages": items})
}

// CreateCustomerServiceConversationMessage sends a staff reply from the admin chat inbox.
func (h *TicketHandler) CreateCustomerServiceConversationMessage(c *gin.Context) {
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}

	var req struct {
		Message       string      `json:"message" binding:"required"`
		MessageType   string      `json:"message_type"`
		Metadata      interface{} `json:"metadata"`
		AttachmentURL string      `json:"attachment_url"`
		Attachments   []string    `json:"attachments"`
	}
	limitAdminCustomerServiceJSONBody(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAdminJSONBindError(c, err)
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	metadata, err := marshalAdminCustomerServiceMessageMetadata(req.Metadata)
	if err != nil {
		respondAdminJSONBindError(c, err)
		return
	}
	attachments, err := h.sanitizeAdminCustomerServiceAttachments(req.AttachmentURL, req.Attachments)
	if err != nil {
		respondAdminAttachmentError(c, err)
		return
	}
	attachmentsJSON, _ := json.Marshal(attachments)

	msg := &ticket.TicketMessage{
		TicketID:    ticketID,
		UserID:      agentUserID,
		IsStaff:     true,
		Content:     strings.TrimSpace(req.Message),
		MessageType: normalizeAdminCustomerServiceMessageType(req.MessageType),
		Metadata:    metadata,
		Attachments: string(attachmentsJSON),
		IsRead:      false,
		IsInternal:  false,
	}
	if err := h.ticketService.AddCustomerServiceAgentMessage(msg, agentUserID, canViewAll); err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	messagePayload := adminCustomerServiceMessageResponse(*msg)
	if conversation, err := h.ticketService.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll); err == nil {
		h.publishAdminCustomerServiceMessageCreated(
			conversation,
			msg.ID,
			msg.CreatedAt,
			adminCustomerServiceRealtimeActor(agentUserID),
		)
	}

	response.Created(c, gin.H{"message": messagePayload})
}

// MarkCustomerServiceConversationMessagesRead marks customer messages as read in one conversation.
func (h *TicketHandler) MarkCustomerServiceConversationMessagesRead(c *gin.Context) {
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	mutation, err := h.ticketService.MarkCustomerServiceMessagesReadForAgentWithRealtimeEvent(ticketID, agentUserID, canViewAll)
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	if mutation != nil && h.customerServiceEvents != nil {
		h.customerServiceEvents.Publish(mutation.Event)
	}

	response.SuccessWithMessage(c, "Messages marked as read", nil)
}

// TransferCustomerServiceConversation reassigns one public chat conversation.
func (h *TicketHandler) TransferCustomerServiceConversation(c *gin.Context) {
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}

	var req struct {
		AssignedTo uint `json:"assigned_to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	mutation, err := h.ticketService.TransferCustomerServiceConversationForAgentWithRealtimeEvent(ticketID, agentUserID, canViewAll, req.AssignedTo)
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	if mutation != nil && h.customerServiceEvents != nil {
		h.customerServiceEvents.Publish(mutation.Event)
	}

	response.SuccessWithMessage(c, "Conversation transferred successfully", nil)
}
