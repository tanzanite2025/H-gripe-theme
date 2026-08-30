package ticket

import (
	"commerce-platform/internal/api/v1/publicmedia"
	"commerce-platform/internal/domain/ticket"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
)

// ============ 客服消息响应 ============

func publicCustomerServiceMessageResponse(item ticket.TicketMessage, conversationID, senderName, messageType string, metadata interface{}, resolver publicmedia.Resolver) gin.H {
	attachments := publicTicketMessageAttachments(item.Attachments, resolver)
	attachmentURL := firstTicketMessageAttachment(attachments)

	if strings.TrimSpace(senderName) == "" {
		if item.IsStaff {
			senderName = "Agent"
		} else {
			senderName = "Customer"
		}
		if item.User != nil {
			senderName = displayName(item.User.FirstName, item.User.LastName, item.User.Username, item.User.Email)
		}
	}
	if strings.TrimSpace(messageType) == "" {
		messageType = item.MessageType
	}
	messageType = normalizeTicketMessageType(messageType)
	if metadata == nil {
		metadata = parseTicketMessageMetadata(item.Metadata)
	}
	metadata = publicTicketMessageMetadata(metadata, resolver)

	return gin.H{
		"id":              item.ID,
		"conversation_id": conversationID,
		"sender_id":       item.UserID,
		"sender_name":     senderName,
		"sender_email":    "",
		"message":         item.Content,
		"message_type":    messageType,
		"metadata":        metadata,
		"source":          ticketMessageSource(item.Metadata),
		"attachment_url":  attachmentURL,
		"attachments":     attachments,
		"created_at":      item.CreatedAt,
		"is_agent":        item.IsStaff,
	}
}

func normalizeTicketMessageType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "product", "order", "image", "link", "video", "faq", "config_confirm", "wheelset_selection_request":
		return value
	default:
		return "text"
	}
}

func marshalTicketMessageMetadata(value interface{}) (string, error) {
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

func parseTicketMessageAttachments(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}
	}
	var attachments []string
	if err := json.Unmarshal([]byte(value), &attachments); err != nil || attachments == nil {
		return []string{}
	}
	return attachments
}

func publicTicketMessageAttachments(value string, resolver publicmedia.Resolver) []string {
	attachments := parseTicketMessageAttachments(value)
	for index := range attachments {
		attachments[index] = publicmedia.URL(resolver, attachments[index])
	}
	return attachments
}

func firstTicketMessageAttachment(attachments []string) string {
	if len(attachments) == 0 {
		return ""
	}
	return attachments[0]
}

func ticketMessageSource(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return ""
	}
	source, _ := payload["_source"].(string)
	return strings.TrimSpace(source)
}

func marshalTicketMessageAttachments(values ...string) string {
	attachments := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			attachments = append(attachments, value)
		}
	}
	if len(attachments) == 0 {
		return ""
	}
	payload, err := json.Marshal(attachments)
	if err != nil {
		return ""
	}
	return string(payload)
}

func parseTicketMessageMetadata(value string) interface{} {
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

func publicTicketMessageMetadata(metadata interface{}, resolver publicmedia.Resolver) interface{} {
	if metadata == nil || resolver == nil {
		return metadata
	}

	switch payload := metadata.(type) {
	case map[string]interface{}:
		return publicTicketMessageMetadataObject(payload, resolver)
	case map[string]string:
		normalized := make(map[string]string, len(payload))
		for key, value := range payload {
			if isPublicTicketMessageMetadataMediaKey(key) {
				normalized[key] = publicmedia.URL(resolver, value)
				continue
			}
			normalized[key] = value
		}
		return normalized
	default:
		return metadata
	}
}

func publicTicketMessageMetadataObject(payload map[string]interface{}, resolver publicmedia.Resolver) map[string]interface{} {
	normalized := make(map[string]interface{}, len(payload))
	for key, value := range payload {
		if isPublicTicketMessageMetadataMediaKey(key) {
			if mediaURL, ok := value.(string); ok {
				normalized[key] = publicmedia.URL(resolver, mediaURL)
				continue
			}
		}
		normalized[key] = value
	}
	return normalized
}

func isPublicTicketMessageMetadataMediaKey(key string) bool {
	switch key {
	case "thumbnail", "answer_image_url":
		return true
	default:
		return false
	}
}

func displayName(firstName, lastName, username, email string) string {
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
	return "Customer"
}

func zeroToNil(value uint) interface{} {
	if value == 0 {
		return nil
	}
	return value
}
