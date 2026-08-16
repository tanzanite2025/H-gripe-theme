package ticket

import (
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const publicCustomerServiceAttachmentMaxRequestBytes = 6 << 20

func (h *Handler) EnsurePublicCustomerServiceConversation(c *gin.Context) {
	var req struct {
		AgentID string `json:"agent_id"`
		Locale  string `json:"locale"`
	}
	if c.Request != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
			return
		}
	}

	agentID := parseCustomerServiceAgentID(req.AgentID)
	owner := h.publicCustomerOwner(c)
	t, err := h.ticketService.GetOrCreatePublicCustomerServiceConversation(
		owner,
		agentID,
	)
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}

	conversationID := publicConversationID(t)
	responseData := gin.H{
		"conversation_id": conversationID,
		"lastAgentId":     zeroToNil(t.AssignedTo),
	}

	_, _, autoReplyMessage, autoReplyErr := h.ticketService.GetWelcomeMessage(
		conversationID,
		owner,
		agentID,
		publicCustomerServiceLocale(c, req.Locale),
	)
	if autoReplyErr != nil {
		log.Printf("customer-service welcome auto reply failed: conversation=%s error=%v", conversationID, autoReplyErr)
	} else if autoReplyMessage != nil {
		autoReplyPayload := publicCustomerServiceMessageResponse(*autoReplyMessage, conversationID, "", "", nil)
		h.publishPublicCustomerServiceMessageCreated(
			t,
			autoReplyMessage.ID,
			autoReplyMessage.CreatedAt,
			service.CustomerServiceRealtimeActor{Kind: "system"},
		)
		responseData["auto_reply"] = gin.H{
			"message": autoReplyPayload,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"hasConversation": true,
		"conversation_id": conversationID,
		"lastAgentId":     zeroToNil(t.AssignedTo),
		"data":            responseData,
	})
}

func (h *Handler) HasPublicCustomerServiceConversation(c *gin.Context) {
	hasConversation, conversationID, lastAgentID, err := h.ticketService.HasPublicCustomerServiceConversation(h.existingPublicCustomerOwner(c))
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hasConversation": hasConversation,
		"conversation_id": conversationID,
		"lastAgentId":     zeroToNil(lastAgentID),
	})
}

func (h *Handler) SendPublicCustomerServiceMessage(c *gin.Context) {
	var req struct {
		ConversationID string      `json:"conversation_id"`
		Message        string      `json:"message" binding:"required"`
		SenderType     string      `json:"sender_type"`
		SenderName     string      `json:"sender_name" binding:"required"`
		SenderEmail    string      `json:"sender_email"`
		AgentID        string      `json:"agent_id"`
		Locale         string      `json:"locale"`
		MessageType    string      `json:"message_type"`
		Metadata       interface{} `json:"metadata"`
		AttachmentURL  string      `json:"attachment_url"`
		Attachments    []string    `json:"attachments"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, ticketMessageJSONMaxBytes)
	if err := c.ShouldBindJSON(&req); err != nil {
		respondPublicCustomerServiceJSONBindError(c, err)
		return
	}

	conversationID := strings.TrimSpace(req.ConversationID)
	message := strings.TrimSpace(req.Message)
	senderName := strings.TrimSpace(req.SenderName)
	if message == "" || senderName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] missing required parameters"})
		return
	}

	owner := h.publicCustomerOwner(c)
	messageType := normalizeTicketMessageType(req.MessageType)
	metadata, err := marshalTicketMessageMetadata(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] invalid message metadata"})
		return
	}
	attachmentValues, err := h.sanitizeTicketMessageAttachments(
		mergeTicketAttachmentInputs(req.AttachmentURL, req.Attachments),
		publicCustomerServiceMessageAttachmentLimit(c),
	)
	if err != nil {
		respondPublicCustomerServiceAttachmentError(c, err)
		return
	}
	attachments := marshalTicketMessageAttachments(attachmentValues...)

	emailCaptured := strings.TrimSpace(req.SenderEmail) != ""

	t, msg, err := h.ticketService.AddPublicCustomerServiceMessage(
		conversationID,
		owner,
		message,
		parseCustomerServiceAgentID(req.AgentID),
		messageType,
		metadata,
		attachments,
	)
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}
	if msg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] AddPublicCustomerServiceMessage returned nil message"})
		return
	}

	conversationID = publicConversationID(t)
	h.touchCustomerServiceVisitorProfile(c, owner, req.SenderEmail, "public_chat")
	response := publicCustomerServiceMessageResponse(*msg, conversationID, senderName, "", nil)
	h.publishPublicCustomerServiceMessageCreated(
		t,
		msg.ID,
		msg.CreatedAt,
		publicCustomerServiceRealtimeActor(owner),
	)
	if emailCaptured {
		h.publishPublicCustomerServiceEventToAudience(
			service.CustomerServiceEventContextUpdated,
			t,
			publicCustomerServiceRealtimeActor(owner),
			service.CustomerServiceRealtimeAudienceBackoffice,
			gin.H{"source": "public_chat_email"},
		)
	}

	// The backend owns keyword matching after the customer message is
	// persisted. The legacy public endpoint remains idempotent for older
	// clients during rollout.
	_, ruleID, autoReplyMessage, autoReplyErr := h.ticketService.MatchKeywordMessage(
		conversationID,
		message,
		owner,
		parseCustomerServiceAgentID(req.AgentID),
		publicCustomerServiceLocale(c, req.Locale),
	)
	if autoReplyErr != nil {
		log.Printf("customer-service auto reply failed: conversation=%s error=%v", conversationID, autoReplyErr)
	} else if autoReplyMessage != nil {
		autoReplyPayload := publicCustomerServiceMessageResponse(*autoReplyMessage, conversationID, "", "", nil)
		h.publishPublicCustomerServiceMessageCreated(
			t,
			autoReplyMessage.ID,
			autoReplyMessage.CreatedAt,
			service.CustomerServiceRealtimeActor{Kind: "system"},
		)
		response["auto_reply"] = gin.H{
			"rule_id": ruleID,
			"message": autoReplyPayload,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message_id":      msg.ID,
		"conversation_id": conversationID,
		"data":            response,
	})
}

func (h *Handler) UploadPublicCustomerServiceAttachment(c *gin.Context) {
	if h.mediaService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "customer-service attachment storage is unavailable"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, publicCustomerServiceAttachmentMaxRequestBytes)
	if err := c.Request.ParseMultipartForm(2 << 20); err != nil {
		status := http.StatusBadRequest
		code := "invalid_upload"
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			status = http.StatusRequestEntityTooLarge
			code = upload.CodeFileTooLarge
		}
		c.JSON(status, gin.H{"success": false, "message": err.Error(), "code": code})
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}

	conversationID := strings.TrimSpace(c.PostForm("conversation_id"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] conversation_id is required"})
		return
	}

	owner := h.existingPublicCustomerOwner(c)
	if _, err := h.ticketService.GetPublicCustomerServiceConversation(conversationID, owner); err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "file is required"})
		return
	}
	if err := upload.ValidateFile(file, upload.SuggestionImageRule); err != nil {
		c.JSON(upload.HTTPStatus(err), gin.H{
			"success": false,
			"message": err.Error(),
			"code":    upload.ErrorCode(err),
		})
		return
	}

	source := strings.ToLower(strings.TrimSpace(c.PostForm("source")))
	if source != "camera" {
		source = "library"
	}

	uploaderID := uint(0)
	if owner.UserID != nil {
		uploaderID = *owner.UserID
	}

	asset, err := h.mediaService.UploadAsset(c.Request.Context(), service.MediaUploadInput{
		File:       file,
		MediaType:  "image",
		Alt:        "Customer service attachment",
		UploaderID: uploaderID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to upload customer-service attachment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"asset": gin.H{
			"id":                asset.ID,
			"url":               asset.URL,
			"original_filename": asset.OriginalFilename,
			"mime_type":         asset.MimeType,
			"size":              asset.Size,
			"source":            source,
		},
	})
}

func (h *Handler) GetWelcomeMessage(c *gin.Context) {
	conversationID := strings.TrimSpace(c.Query("conversation_id"))
	agentID := parseCustomerServiceAgentID(c.Query("agent_id"))
	owner := h.publicCustomerOwner(c)

	if conversationID == "" {
		t, err := h.ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agentID)
		if err != nil {
			writePublicCustomerServiceError(c, err)
			return
		}
		conversationID = publicConversationID(t)
	}

	reply, alreadySent, msg, err := h.ticketService.GetWelcomeMessage(conversationID, owner, agentID, publicCustomerServiceLocale(c))
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}
	if msg != nil {
		if conversation, err := h.ticketService.GetPublicCustomerServiceConversation(conversationID, owner); err == nil {
			h.publishPublicCustomerServiceMessageCreated(
				conversation,
				msg.ID,
				msg.CreatedAt,
				service.CustomerServiceRealtimeActor{Kind: "system"},
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"conversation_id": conversationID,
		"data": gin.H{
			"message":      reply,
			"already_sent": alreadySent,
		},
	})
}

func (h *Handler) MatchKeywordMessage(c *gin.Context) {
	var req struct {
		Message        string `json:"message" binding:"required"`
		ConversationID string `json:"conversation_id" binding:"required"`
		AgentID        string `json:"agent_id"`
		Locale         string `json:"locale"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}
	req.ConversationID = strings.TrimSpace(req.ConversationID)
	if req.ConversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] conversation_id is required"})
		return
	}

	owner := h.existingPublicCustomerOwner(c)
	reply, ruleID, msg, err := h.ticketService.MatchKeywordMessage(
		req.ConversationID,
		req.Message,
		owner,
		parseCustomerServiceAgentID(req.AgentID),
		publicCustomerServiceLocale(c, req.Locale),
	)
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}
	if msg != nil {
		if conversation, err := h.ticketService.GetPublicCustomerServiceConversation(req.ConversationID, owner); err == nil {
			h.publishPublicCustomerServiceMessageCreated(
				conversation,
				msg.ID,
				msg.CreatedAt,
				service.CustomerServiceRealtimeActor{Kind: "system"},
			)
		}
	}

	if reply == "" {
		c.JSON(http.StatusOK, gin.H{
			"success":         true,
			"conversation_id": req.ConversationID,
			"data": gin.H{
				"reply": "",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"conversation_id": req.ConversationID,
		"data": gin.H{
			"reply":   reply,
			"rule_id": ruleID,
		},
	})
}

func publicCustomerServiceLocale(c *gin.Context, explicit ...string) string {
	if len(explicit) > 0 && strings.TrimSpace(explicit[0]) != "" {
		return locales.ResolveSupported(explicit[0])
	}

	if value := strings.TrimSpace(c.Query("locale")); value != "" {
		return locales.ResolveSupported(value)
	}
	if value, err := c.Cookie("locale"); err == nil && strings.TrimSpace(value) != "" {
		if resolved := locales.ResolveSupported(value); resolved != "" {
			return resolved
		}
	}

	header := strings.TrimSpace(c.GetHeader("Accept-Language"))
	if header == "" {
		return ""
	}

	for _, part := range strings.Split(header, ",") {
		value := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if resolved := locales.ResolveSupported(value); resolved != "" {
			return resolved
		}
	}
	return ""
}

func (h *Handler) GetPublicCustomerServiceMessages(c *gin.Context) {
	conversationID := strings.TrimSpace(c.Param("conversation_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] missing conversation id"})
		return
	}

	messages, err := h.ticketService.GetPublicCustomerServiceMessages(conversationID, h.existingPublicCustomerOwner(c), limit, offset)
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}
	if messages == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] GetPublicCustomerServiceMessages returned nil"})
		return
	}

	items := make([]gin.H, 0, len(messages))
	for _, item := range messages {
		items = append(items, publicCustomerServiceMessageResponse(item, conversationID, "", "", nil))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"total":   len(items),
	})
}
