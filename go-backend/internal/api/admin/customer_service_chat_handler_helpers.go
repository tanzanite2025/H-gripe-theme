package admin

import (
	"commerce-platform/internal/domain/ticket"
	userdomain "commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func scopeAdminCustomerServiceConversationFilters(c *gin.Context, filters service.CustomerServiceConversationListInput) (service.CustomerServiceConversationListInput, uint, bool) {
	agentUserID, canViewAll := adminCustomerServiceScope(c)
	if agentUserID > 0 && !canViewAll {
		filters.AssignedTo = &agentUserID
	}
	return filters, agentUserID, canViewAll
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
	status := input.Status
	if status == "" {
		status = "all"
	}

	identity := input.Identity
	if identity == "" {
		identity = "all"
	}

	return gin.H{
		"search":   input.Search,
		"status":   status,
		"identity": identity,
		"unread":   input.UnreadOnly,
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
	unreadCount := max(item.CustomerServiceUnreadCount, 0)
	for _, message := range item.Messages {
		if message.Content != "" {
			lastMessage = message.Content
			lastMessageTime = message.CreatedAt
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
	case "product", "order", "image", "link", "video", "faq", "config_confirm", "wheelset_selection_request":
		return value
	default:
		return "text"
	}
}

func marshalAdminCustomerServiceMessageMetadata(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	if string(payload) == "null" {
		return "", nil
	}
	return string(payload), nil
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
	return "客服"
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
