package admin

import (
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/ticket"
	userdomain "commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ListCustomerServiceConversations returns the admin chat inbox conversation list.
// Admin/manager users can see every public chat conversation; support users only
// see conversations assigned to their own backend user id.
func (h *TicketHandler) ListCustomerServiceConversations(c *gin.Context) {
	params := pagination.ParsePagination(c)
	agentUserID, canViewAll := adminCustomerServiceScope(c)

	filters, ok := parseAdminCustomerServiceConversationFilters(c)
	if !ok {
		return
	}

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

func (h *TicketHandler) GetCustomerServiceRegionAnalytics(c *gin.Context) {
	if h.customerServiceContext == nil {
		apierror.RespondInternalError(c, errors.New("customer service context service is not configured"))
		return
	}

	timezoneOffsetMinutes, err := strconv.Atoi(c.DefaultQuery("tz_offset_minutes", "0"))
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid timezone offset")
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	analytics, err := h.customerServiceContext.RegionAnalyticsForAgent(service.CustomerServiceRegionAnalyticsInput{
		Date:                  strings.TrimSpace(c.Query("date")),
		TimezoneOffsetMinutes: timezoneOffsetMinutes,
		AgentUserID:           agentUserID,
		CanViewAll:            canViewAll,
	})
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"analytics": analytics})
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
		Message       string   `json:"message" binding:"required"`
		AttachmentURL string   `json:"attachment_url"`
		Attachments   []string `json:"attachments"`
	}
	limitAdminCustomerServiceJSONBody(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAdminJSONBindError(c, err)
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
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
		MessageType: "text",
		Attachments: string(attachmentsJSON),
		IsRead:      false,
		IsInternal:  false,
	}
	if err := h.ticketService.AddCustomerServiceAgentMessage(msg, agentUserID, canViewAll); err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	messagePayload := adminCustomerServiceMessageResponse(*msg)
	conversation, _ := h.ticketService.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll)
	h.publishAdminCustomerServiceEvent(
		service.CustomerServiceEventMessageCreated,
		conversation,
		ticketID,
		adminCustomerServiceRealtimeActor(agentUserID),
		messagePayload,
	)

	response.Created(c, gin.H{"message": messagePayload})
}

func (h *TicketHandler) SendCustomerServiceConversationTyping(c *gin.Context) {
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}

	var req struct {
		IsTyping *bool `json:"is_typing"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	conversation, err := h.ticketService.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll)
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	isTyping := true
	if req.IsTyping != nil {
		isTyping = *req.IsTyping
	}

	h.publishAdminCustomerServiceEvent(
		service.CustomerServiceEventTyping,
		conversation,
		ticketID,
		adminCustomerServiceRealtimeActor(agentUserID),
		gin.H{
			"is_typing":    isTyping,
			"display_name": adminCustomerServiceAssigneeName(agentUserID),
			"expires_at":   time.Now().UTC().Add(5 * time.Second),
		},
	)

	response.Success(c, gin.H{"typing": isTyping})
}

// MarkCustomerServiceConversationMessagesRead marks customer messages as read in one conversation.
func (h *TicketHandler) MarkCustomerServiceConversationMessagesRead(c *gin.Context) {
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	if err := h.ticketService.MarkCustomerServiceMessagesReadForAgent(ticketID, agentUserID, canViewAll); err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	conversation, _ := h.ticketService.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll)
	h.publishAdminCustomerServiceEvent(
		service.CustomerServiceEventMessagesRead,
		conversation,
		ticketID,
		adminCustomerServiceRealtimeActor(agentUserID),
		gin.H{
			"reader_kind":     "agent",
			"read_by_user_id": agentUserID,
		},
	)

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
	if err := h.ticketService.TransferCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll, req.AssignedTo); err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	conversation, _ := h.ticketService.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll)
	h.publishAdminCustomerServiceEvent(
		service.CustomerServiceEventAssigned,
		conversation,
		ticketID,
		adminCustomerServiceRealtimeActor(agentUserID),
		gin.H{
			"assigned_to":         req.AssignedTo,
			"assigned_to_name":    adminCustomerServiceAssigneeName(req.AssignedTo),
			"assigned_by_user_id": agentUserID,
			"status":              "in_progress",
			"display_status":      adminCustomerServiceDisplayStatus("in_progress"),
		},
	)

	response.SuccessWithMessage(c, "Conversation transferred successfully", nil)
}

func parseAdminCustomerServiceConversationID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid conversation ID")
		return 0, false
	}
	return uint(id), true
}

func parseAdminCustomerServiceConversationFilters(c *gin.Context) (service.CustomerServiceConversationListInput, bool) {
	input := service.CustomerServiceConversationListInput{
		Status:     normalizeAdminCustomerServiceFilterValue(c.Query("status")),
		Identity:   normalizeAdminCustomerServiceFilterValue(c.Query("identity")),
		Search:     strings.TrimSpace(c.Query("search")),
		UnreadOnly: parseAdminCustomerServiceBoolQuery(c.Query("unread")),
	}

	if !validAdminCustomerServiceStatusFilter(input.Status) {
		apierror.RespondBadRequest(c, "Invalid customer-service conversation status filter")
		return input, false
	}
	if !validAdminCustomerServiceIdentityFilter(input.Identity) {
		apierror.RespondBadRequest(c, "Invalid customer-service customer identity filter")
		return input, false
	}

	assignedToRaw := normalizeAdminCustomerServiceFilterValue(c.Query("assigned_to"))
	if assignedToRaw != "" {
		assignedTo, err := strconv.ParseUint(assignedToRaw, 10, 32)
		if err != nil {
			apierror.RespondBadRequest(c, "Invalid assigned customer-service agent filter")
			return input, false
		}
		if assignedTo > 0 {
			assignedToValue := uint(assignedTo)
			input.AssignedTo = &assignedToValue
		}
	}

	groupIDRaw := normalizeAdminCustomerServiceFilterValue(c.Query("group_id"))
	if groupIDRaw != "" {
		groupID, err := strconv.ParseUint(groupIDRaw, 10, 32)
		if err != nil || groupID == 0 {
			apierror.RespondBadRequest(c, "Invalid customer-service group filter")
			return input, false
		}
		groupIDValue := uint(groupID)
		input.GroupID = &groupIDValue
	}

	return input, true
}

func normalizeAdminCustomerServiceFilterValue(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "all" {
		return ""
	}
	return value
}

func parseAdminCustomerServiceBoolQuery(value string) bool {
	switch normalizeAdminCustomerServiceFilterValue(value) {
	case "1", "true", "yes", "y", "unread":
		return true
	default:
		return false
	}
}

func validAdminCustomerServiceStatusFilter(status string) bool {
	switch status {
	case "", "pending", "open", "active", "in_progress", "closed", "resolved":
		return true
	default:
		return false
	}
}

func validAdminCustomerServiceIdentityFilter(identity string) bool {
	switch identity {
	case "", "account", "member", "user", "anonymous", "visitor", "guest":
		return true
	default:
		return false
	}
}

func adminCustomerServiceConversationFilterResponse(input service.CustomerServiceConversationListInput) gin.H {
	assignedTo := "all"
	if input.AssignedTo != nil && *input.AssignedTo > 0 {
		assignedTo = strconv.FormatUint(uint64(*input.AssignedTo), 10)
	}
	groupID := "all"
	if input.GroupID != nil && *input.GroupID > 0 {
		groupID = strconv.FormatUint(uint64(*input.GroupID), 10)
	}
	status := input.Status
	if status == "" {
		status = "all"
	}
	identity := input.Identity
	if identity == "" {
		identity = "all"
	}

	return gin.H{
		"search":      input.Search,
		"status":      status,
		"identity":    identity,
		"assigned_to": assignedTo,
		"group_id":    groupID,
		"unread":      input.UnreadOnly,
	}
}

func adminCustomerServiceGroupsResponse(groups []userdomain.AgentGroup) []gin.H {
	items := make([]gin.H, 0, len(groups))
	for _, group := range groups {
		items = append(items, adminCustomerServiceGroupResponse(group))
	}
	return items
}

func adminCustomerServiceAgentGroupsResponse(groups []userdomain.AgentGroup) []gin.H {
	items := make([]gin.H, 0, len(groups))
	for _, group := range groups {
		if group.Status != "" && group.Status != "active" {
			continue
		}
		items = append(items, adminCustomerServiceGroupResponse(group))
	}
	return items
}

func adminCustomerServiceGroupResponse(group userdomain.AgentGroup) gin.H {
	return gin.H{
		"id":          group.ID,
		"code":        group.Code,
		"name":        group.Name,
		"description": group.Description,
		"status":      group.Status,
		"sort_order":  group.SortOrder,
	}
}

func adminCustomerServiceAgentGroupIDs(groups []userdomain.AgentGroup) []uint {
	ids := make([]uint, 0, len(groups))
	for _, group := range groups {
		if group.ID > 0 && (group.Status == "" || group.Status == "active") {
			ids = append(ids, group.ID)
		}
	}
	return ids
}

func adminCustomerServicePrimaryAgentGroup(groups []userdomain.AgentGroup) interface{} {
	for _, group := range groups {
		if group.ID > 0 && (group.Status == "" || group.Status == "active") {
			return adminCustomerServiceGroupResponse(group)
		}
	}
	return nil
}

func adminCustomerServiceScope(c *gin.Context) (uint, bool) {
	value, _ := c.Get("user_id")
	userID, _ := value.(uint)

	roleValue, _ := c.Get("role")
	role := auth.RoleUser
	if rawRole, ok := roleValue.(string); ok {
		role = auth.NormalizeRole(rawRole)
	}

	return userID, role == auth.RoleAdmin || role == auth.RoleManager
}

func respondAdminCustomerServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCustomerServiceAgentAccessDenied):
		apierror.RespondForbidden(c)
	case service.IsRecordNotFound(err):
		apierror.RespondNotFound(c, "Conversation")
	default:
		apierror.RespondInternalError(c, err)
	}
}

func adminCustomerServiceConversationResponse(item ticket.Ticket, summary *service.CustomerServiceConversationSummary) gin.H {
	lastMessage := ""
	lastMessageTime := item.UpdatedAt
	unreadCount := 0
	for _, message := range item.Messages {
		if message.Content != "" {
			lastMessage = message.Content
			lastMessageTime = message.CreatedAt
		}
		if !message.IsStaff && !message.IsRead {
			unreadCount++
		}
	}

	conversationID := ""
	if item.ConversationID != nil {
		conversationID = strings.TrimSpace(*item.ConversationID)
	}
	if conversationID == "" && strings.HasPrefix(item.Tags, "conversation_id:") {
		conversationID = strings.TrimPrefix(item.Tags, "conversation_id:")
	}

	customerName := adminCustomerDisplayName(item)
	customerSummary := adminCustomerServiceFallbackSummary(item, customerName)
	if summary != nil {
		customerSummary = *summary
		if strings.TrimSpace(summary.DisplayName) != "" {
			customerName = strings.TrimSpace(summary.DisplayName)
		}
	}

	return gin.H{
		"id":                item.ID,
		"ticket_id":         item.ID,
		"conversation_id":   conversationID,
		"customer_user_id":  item.CustomerUserID,
		"customer_name":     customerName,
		"customer_summary":  customerSummary,
		"assigned_to":       item.AssignedTo,
		"status":            item.Status,
		"display_status":    adminCustomerServiceDisplayStatus(item.Status),
		"unread_count":      unreadCount,
		"last_message":      lastMessage,
		"last_message_time": lastMessageTime,
		"created_at":        item.CreatedAt,
		"updated_at":        item.UpdatedAt,
		"ticket_number":     item.TicketNumber,
		"visitor_anonymous": item.CustomerUserID == nil,
	}
}

func adminCustomerServiceFallbackSummary(item ticket.Ticket, customerName string) service.CustomerServiceConversationSummary {
	summary := service.CustomerServiceConversationSummary{
		Type:          "visitor",
		Identity:      "visitor",
		IdentityLabel: "游客",
		DisplayName:   strings.TrimSpace(customerName),
		RegionLabel:   "未知区域",
		RegionStatus:  "unknown",
	}
	if summary.DisplayName == "" {
		summary.DisplayName = "匿名客户"
	}
	if item.CustomerUserID != nil && *item.CustomerUserID > 0 {
		summary.Type = "member"
		summary.Identity = "member"
		summary.IdentityLabel = "会员"
	}
	return summary
}

func adminCustomerServiceMessageResponse(item ticket.TicketMessage) gin.H {
	attachmentURL := ""
	attachments := []string{}
	if err := json.Unmarshal([]byte(item.Attachments), &attachments); err == nil && len(attachments) > 0 {
		attachmentURL = attachments[0]
	}

	return gin.H{
		"id":              item.ID,
		"conversation_id": item.TicketID,
		"sender_id":       item.UserID,
		"sender_name":     adminMessageSenderName(item),
		"message":         item.Content,
		"content":         item.Content,
		"message_type":    normalizeAdminCustomerServiceMessageType(item.MessageType),
		"metadata":        parseAdminCustomerServiceMessageMetadata(item.Metadata),
		"source":          adminCustomerServiceMessageSource(item.Metadata),
		"attachment_url":  attachmentURL,
		"attachments":     attachments,
		"created_at":      item.CreatedAt,
		"is_read":         item.IsRead,
		"is_agent":        item.IsStaff,
	}
}

func normalizeAdminCustomerServiceMessageType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "product", "order", "image", "link", "faq", "config_confirm":
		return value
	default:
		return "text"
	}
}

func parseAdminCustomerServiceMessageMetadata(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil
	}
	return payload
}

func adminCustomerServiceMessageSource(value string) string {
	payload, ok := parseAdminCustomerServiceMessageMetadata(value).(map[string]interface{})
	if !ok {
		return ""
	}
	source, _ := payload["_source"].(string)
	return strings.TrimSpace(source)
}

func adminCustomerDisplayName(item ticket.Ticket) string {
	if item.CustomerUserID == nil {
		return "匿名客户"
	}
	if item.User != nil {
		return adminDisplayName(item.User.FirstName, item.User.LastName, item.User.Username, item.User.Email)
	}
	return "客户 " + strconv.FormatUint(uint64(*item.CustomerUserID), 10)
}

func adminCustomerServiceAssigneeName(userID uint) string {
	if userID == 0 {
		return "未分配"
	}
	return "用户 " + strconv.FormatUint(uint64(userID), 10)
}

func adminMessageSenderName(item ticket.TicketMessage) string {
	if item.User != nil {
		return adminDisplayName(item.User.FirstName, item.User.LastName, item.User.Username, item.User.Email)
	}
	if item.IsStaff {
		return "客服"
	}
	return "客户"
}

func adminDisplayName(firstName, lastName, username, email string) string {
	fullName := strings.TrimSpace(strings.TrimSpace(firstName) + " " + strings.TrimSpace(lastName))
	if fullName != "" {
		return fullName
	}
	if strings.TrimSpace(username) != "" {
		return strings.TrimSpace(username)
	}
	if strings.TrimSpace(email) != "" {
		return strings.TrimSpace(email)
	}
	return "客户"
}

func adminCustomerServiceDisplayStatus(status string) string {
	switch status {
	case "in_progress":
		return "active"
	case "open":
		return "pending"
	case "resolved", "closed":
		return "closed"
	default:
		return "pending"
	}
}
