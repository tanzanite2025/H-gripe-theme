package repository

import (
	"tanzanite/internal/domain/chat"

	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// SaveMessage 保存聊天消息
func (r *ChatRepository) SaveMessage(message *chat.ChatMessage) error {
	return r.db.Create(message).Error
}

func (r *ChatRepository) SessionExists(sessionID string) (bool, error) {
	var count int64
	err := r.db.Model(&chat.ChatMessage{}).Where("session_id = ?", sessionID).Count(&count).Error
	return count > 0, err
}

func (r *ChatRepository) SessionBelongsToUser(sessionID string, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&chat.ChatMessage{}).
		Where("session_id = ? AND user_id = ?", sessionID, userID).
		Count(&count).Error
	return count > 0, err
}

// GetMessages 获取会话的聊天历史
func (r *ChatRepository) GetMessages(sessionID string, limit int) ([]chat.ChatMessage, error) {
	var messages []chat.ChatMessage

	query := r.db.Where("session_id = ?", sessionID).
		Order("timestamp ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&messages).Error
	return messages, err
}

// GetMessagesByUser 获取用户的所有聊天记录
func (r *ChatRepository) GetMessagesByUser(userID uint, limit int) ([]chat.ChatMessage, error) {
	var messages []chat.ChatMessage

	query := r.db.Where("user_id = ?", userID).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&messages).Error
	return messages, err
}

// DeleteOldMessages 删除超过指定天数的消息（数据清理）
func (r *ChatRepository) DeleteOldMessages(days int) error {
	return r.db.Where("created_at < NOW() - INTERVAL ? DAY", days).
		Delete(&chat.ChatMessage{}).Error
}

// CreateOrUpdateSession 创建或更新会话信息
func (r *ChatRepository) CreateOrUpdateSession(session *chat.ChatSession) error {
	return r.db.Save(session).Error
}

// GetSession 获取会话信息
func (r *ChatRepository) GetSession(sessionID string) (*chat.ChatSession, error) {
	var session chat.ChatSession
	err := r.db.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetUserSessions 获取用户的所有会话
func (r *ChatRepository) GetUserSessions(userID uint) ([]chat.ChatSession, error) {
	var sessions []chat.ChatSession
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&sessions).Error
	return sessions, err
}
