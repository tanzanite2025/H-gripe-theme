package chat

import (
	"net/http"
	"strconv"

	"tanzanite/internal/domain/chat"
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatRepo *repository.ChatRepository
}

func NewChatHandler(chatRepo *repository.ChatRepository) *ChatHandler {
	return &ChatHandler{chatRepo: chatRepo}
}

// SaveMessage 保存聊天消息
// POST /api/v1/chat/messages
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

	// 获取用户ID（如果已登录）
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uint)
		userID = &id
	}

	// 创建消息对象
	message := &chat.ChatMessage{
		SessionID: req.SessionID,
		MessageID: req.Message.ID,
		Content:   req.Message.Content,
		Role:      req.Message.Role,
		Timestamp: req.Message.Timestamp,
		UserID:    userID,
		AgentID:   req.Message.AgentID,
	}

	// 保存到数据库
	if err := h.chatRepo.SaveMessage(message); err != nil {
		// 如果是重复消息（MessageID已存在），返回成功（幂等性）
		if err.Error() == "UNIQUE constraint failed: chat_messages.message_id" ||
			err.Error() == "duplicate key value violates unique constraint" {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Message already exists (idempotent)",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message_id": message.MessageID,
	})
}

// GetMessages 获取聊天历史
// GET /api/v1/chat/messages?session_id=xxx&limit=100
func (h *ChatHandler) GetMessages(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	limit := 100 // 默认返回最近100条
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	messages, err := h.chatRepo.GetMessages(sessionID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetUserMessages 获取用户的所有聊天记录（需要认证）
// GET /api/v1/chat/user/messages?limit=50
func (h *ChatHandler) GetUserMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	messages, err := h.chatRepo.GetMessagesByUser(userID.(uint), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

// RegisterRoutes 注册路由
func RegisterRoutes(r *gin.RouterGroup, chatRepo *repository.ChatRepository) {
	handler := NewChatHandler(chatRepo)

	// 公开API（游客也可以使用）
	r.POST("/messages", handler.SaveMessage)
	r.GET("/messages", handler.GetMessages)

	// 需要认证的API
	// authenticated := r.Group("")
	// authenticated.Use(middleware.AuthMiddleware())
	// authenticated.GET("/user/messages", handler.GetUserMessages)
}
