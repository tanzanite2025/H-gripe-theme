package chat

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

func (h *ChatHandler) SaveMessage(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Message   struct {
			ID        string                 `json:"id" binding:"required"`
			Content   string                 `json:"content" binding:"required"`
			Role      string                 `json:"role" binding:"required,oneof=user agent system"`
			Timestamp int64                  `json:"timestamp" binding:"required"`
			AgentID   string                 `json:"agent_id"`
			Metadata  map[string]interface{} `json:"metadata"`
		} `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	message, err := h.chatService.SaveUserMessage(userID, service.ChatMessageInput{
		SessionID: req.SessionID,
		MessageID: req.Message.ID,
		Content:   req.Message.Content,
		Role:      req.Message.Role,
		Timestamp: req.Message.Timestamp,
		AgentID:   req.Message.AgentID,
		Metadata:  req.Message.Metadata,
	})
	if err != nil {
		h.writeChatError(c, err, "Failed to save message")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message_id": message.MessageID,
	})
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	sessionID := strings.TrimSpace(c.Query("session_id"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	messages, err := h.chatService.GetMessages(sessionID, userID, limit)
	if err != nil {
		h.writeChatError(c, err, "Failed to get messages")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

func (h *ChatHandler) GetUserMessages(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	messages, err := h.chatService.GetUserMessages(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

func RegisterRoutes(r *gin.RouterGroup, chatService *service.ChatService) {
	handler := NewChatHandler(chatService)

	r.POST("/messages", handler.SaveMessage)
	r.GET("/messages", handler.GetMessages)
}

func (h *ChatHandler) writeChatError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrChatInvalidSession), errors.Is(err, service.ErrChatInvalidMessage):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrChatAuthenticationRequired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrChatSessionAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrChatMessageAlreadyExists):
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Message already exists (idempotent)",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}

func authenticatedUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint)
	return userID, ok
}
