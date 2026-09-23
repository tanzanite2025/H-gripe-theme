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

func (h *TicketHandler) EvaluateCustomerServiceRetention(c *gin.Context) {
	if h.customerServiceRetention == nil {
		apierror.RespondInternalError(c, errors.New("customer service retention service is not configured"))
		return
	}
	ids, ok := parseCustomerServiceRetentionIDs(c)
	if !ok {
		return
	}
	result, err := h.customerServiceRetention.Evaluate(ids)
	if err != nil {
		respondCustomerServiceRetentionError(c, err)
		return
	}
	response.Success(c, gin.H{"eligibility": result})
}

func (h *TicketHandler) GetCustomerServiceRetentionConfig(c *gin.Context) {
	if h.customerServiceRetention == nil {
		apierror.RespondInternalError(c, errors.New("customer service retention service is not configured"))
		return
	}
	cfg, err := h.customerServiceRetention.RuntimeConfig()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"config": cfg})
}

func (h *TicketHandler) UpdateCustomerServiceRetentionConfig(c *gin.Context) {
	if h.customerServiceRetention == nil {
		apierror.RespondInternalError(c, errors.New("customer service retention service is not configured"))
		return
	}
	var cfg service.CustomerServiceRetentionRuntimeConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if err := h.customerServiceRetention.UpdateRuntimeConfig(cfg); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"config": cfg})
}

func (h *TicketHandler) SoftDeleteCustomerServiceConversations(c *gin.Context) {
	h.mutateCustomerServiceRetention(c, false)
}

func (h *TicketHandler) PurgeCustomerServiceConversations(c *gin.Context) {
	h.mutateCustomerServiceRetention(c, true)
}

func (h *TicketHandler) mutateCustomerServiceRetention(c *gin.Context, purge bool) {
	if h.customerServiceRetention == nil {
		apierror.RespondInternalError(c, errors.New("customer service retention service is not configured"))
		return
	}
	var req struct {
		// conversation_ids is the public name used by the inbox UI. Accept
		// ticket_ids as an explicit maintenance alias because the retention
		// contract is ticket-owned at the persistence boundary.
		ConversationIDs []uint `json:"conversation_ids"`
		TicketIDs       []uint `json:"ticket_ids"`
		Reason          string `json:"reason" binding:"max=500"`
	}
	limitAdminCustomerServiceJSONBody(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAdminJSONBindError(c, err)
		return
	}
	ids := append(append([]uint(nil), req.ConversationIDs...), req.TicketIDs...)
	ids = normalizeAdminCustomerServiceRetentionIDs(ids)
	if len(ids) == 0 {
		apierror.RespondBadRequest(c, service.ErrCustomerServiceRetentionIDsRequired.Error())
		return
	}
	if len(ids) > 100 {
		apierror.RespondBadRequest(c, service.ErrCustomerServiceRetentionTooManyIDs.Error())
		return
	}
	userID := c.GetUint("user_id")
	var result []service.CustomerServiceRetentionEligibility
	var err error
	if purge {
		result, err = h.customerServiceRetention.Purge(ids, userID, req.Reason)
	} else {
		result, err = h.customerServiceRetention.SoftDelete(ids, userID, req.Reason)
	}
	if err != nil {
		respondCustomerServiceRetentionError(c, err)
		return
	}
	response.SuccessWithMessage(c, "Retention operation completed", gin.H{"eligibility": result})
}

func parseCustomerServiceRetentionIDs(c *gin.Context) ([]uint, bool) {
	var rawValues []string
	rawValues = append(rawValues, c.QueryArray("conversation_ids")...)
	rawValues = append(rawValues, c.QueryArray("ticket_ids")...)
	if len(rawValues) == 0 {
		rawValues = append(rawValues, c.Query("conversation_ids"), c.Query("ticket_ids"))
	}
	var ids []uint
	seen := make(map[uint]struct{})
	for _, rawQuery := range rawValues {
		for _, raw := range strings.Split(rawQuery, ",") {
			value := strings.TrimSpace(raw)
			if value == "" {
				continue
			}
			id, err := strconv.ParseUint(value, 10, 32)
			if err != nil || id == 0 {
				apierror.RespondBadRequest(c, "Invalid ticket_ids")
				return nil, false
			}
			parsed := uint(id)
			if _, exists := seen[parsed]; exists {
				continue
			}
			seen[parsed] = struct{}{}
			ids = append(ids, parsed)
			if len(ids) > 100 {
				apierror.RespondBadRequest(c, service.ErrCustomerServiceRetentionTooManyIDs.Error())
				return nil, false
			}
		}
	}
	if len(ids) == 0 {
		apierror.RespondBadRequest(c, "ticket_ids is required")
		return nil, false
	}
	return ids, true
}

func normalizeAdminCustomerServiceRetentionIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func respondCustomerServiceRetentionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCustomerServiceRetentionAdminRequired):
		apierror.RespondForbidden(c)
	case errors.Is(err, service.ErrCustomerServiceRetentionReasonRequired):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrCustomerServiceRetentionIDsRequired), errors.Is(err, service.ErrCustomerServiceRetentionTooManyIDs):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrCustomerServiceRetentionIneligible), errors.Is(err, service.ErrCustomerServiceRetentionWindow):
		apierror.RespondConflict(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}

// GetCustomerServiceConversationMessages returns messages for one conversation.
func (h *TicketHandler) GetCustomerServiceConversationMessages(c *gin.Context) {
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}
	messages, total, err := h.ticketService.GetCustomerServiceMessagesForAgentPage(ticketID, agentUserID, canViewAll, limit, offset)
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	items := make([]gin.H, 0, len(messages))
	for _, item := range messages {
		items = append(items, adminCustomerServiceMessageResponse(item))
	}

	response.Success(c, gin.H{
		"messages": items,
		"total":    total,
		"has_more": int64(offset)+int64(len(items)) < total,
	})
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
		UserID:      &agentUserID,
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
		h.publishCustomerServiceRealtimeMutation(mutation)
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
		h.publishCustomerServiceRealtimeMutation(mutation)
	}

	response.SuccessWithMessage(c, "Conversation transferred successfully", nil)
}

// ArchiveCustomerServiceConversation hides one conversation from the current
// staff member's active inbox without deleting any conversation facts.
func (h *TicketHandler) ArchiveCustomerServiceConversation(c *gin.Context) {
	h.setCustomerServiceConversationArchived(c, true)
}

// RestoreCustomerServiceConversation removes the current staff member's
// personal archive marker and returns the conversation to their inbox.
func (h *TicketHandler) RestoreCustomerServiceConversation(c *gin.Context) {
	h.setCustomerServiceConversationArchived(c, false)
}

// UpdateCustomerServiceConversationStatus applies an optimistic lifecycle
// command. The optional archive flag is part of the same transaction.
func (h *TicketHandler) UpdateCustomerServiceConversationStatus(c *gin.Context) {
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}
	var req struct {
		Status                string `json:"status" binding:"required,oneof=open in_progress resolved closed"`
		ExpectedStatusVersion uint   `json:"expected_status_version" binding:"required,min=1"`
		Archive               *bool  `json:"archive"`
		ReasonCode            string `json:"reason_code" binding:"required,oneof=manual_status_change operator_reopen operator_resolve operator_close_and_archive"`
	}
	limitAdminCustomerServiceJSONBody(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAdminJSONBindError(c, err)
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	mutation, statusVersion, err := h.ticketService.UpdateCustomerServiceConversationStatusForAgent(
		ticketID,
		agentUserID,
		canViewAll,
		service.CustomerServiceConversationStatusInput{
			Status:                req.Status,
			ExpectedStatusVersion: req.ExpectedStatusVersion,
			Archive:               req.Archive,
			ReasonCode:            req.ReasonCode,
		},
	)
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}
	h.publishCustomerServiceRealtimeMutation(mutation)

	message := "Conversation status updated"
	if req.Status == "closed" && req.Archive != nil && *req.Archive {
		message = "Conversation closed and archived"
	} else if req.Status == "open" {
		message = "Conversation reopened"
	}
	response.SuccessWithMessage(c, message, gin.H{
		"conversation": gin.H{
			"id":             ticketID,
			"status":         req.Status,
			"status_version": statusVersion,
		},
	})
}

// BulkArchiveCustomerServiceConversations archives the selected rows for the
// current staff inbox without deleting shared conversation history.
func (h *TicketHandler) BulkArchiveCustomerServiceConversations(c *gin.Context) {
	var req struct {
		ConversationIDs []uint `json:"conversation_ids" binding:"required,min=1,max=100,dive,gt=0"`
	}
	limitAdminCustomerServiceJSONBody(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAdminJSONBindError(c, err)
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	result, err := h.ticketService.ArchiveCustomerServiceConversationsForAgent(req.ConversationIDs, agentUserID, canViewAll)
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}
	h.publishCustomerServiceRealtimeMutation(result.Mutation)
	response.SuccessWithMessage(c, "Conversations archived", gin.H{
		"archived_count":   len(result.ArchivedConversationIDs),
		"conversation_ids": result.ArchivedConversationIDs,
	})
}

func (h *TicketHandler) setCustomerServiceConversationArchived(c *gin.Context, archived bool) {
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}
	agentUserID, canViewAll := adminCustomerServiceScope(c)
	mutation, err := h.ticketService.SetCustomerServiceConversationArchivedForAgent(ticketID, agentUserID, canViewAll, archived)
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}
	h.publishCustomerServiceRealtimeMutation(mutation)
	message := "Conversation archived"
	if !archived {
		message = "Conversation restored"
	}
	response.SuccessWithMessage(c, message, nil)
}

func (h *TicketHandler) publishCustomerServiceRealtimeMutation(mutation *service.CustomerServiceRealtimeMutation) {
	if h.customerServiceEvents == nil {
		return
	}
	for _, event := range mutation.RealtimeEvents() {
		h.customerServiceEvents.Publish(event)
	}
}
