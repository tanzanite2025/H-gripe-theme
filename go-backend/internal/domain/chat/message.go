package chat

import (
	"time"
)

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"index;not null;size:255"` // 会话ID
	MessageID string    `json:"message_id" gorm:"uniqueIndex;not null;size:255"` // 消息唯一ID
	Content   string    `json:"content" gorm:"type:text"` // 消息内容
	Role      string    `json:"role" gorm:"size:50"` // user, agent, system
	Timestamp int64     `json:"timestamp"` // 消息时间戳（毫秒）
	UserID    *uint     `json:"user_id" gorm:"index"` // 关联用户ID（可选）
	AgentID   string    `json:"agent_id" gorm:"size:100"` // 客服ID（可选）
	Metadata  string    `json:"metadata" gorm:"type:json"` // 额外元数据（JSON格式）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// ChatSession 聊天会话模型（可选，用于会话管理）
type ChatSession struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SessionID   string    `json:"session_id" gorm:"uniqueIndex;not null;size:255"`
	UserID      *uint     `json:"user_id" gorm:"index"` // 关联用户ID（可选）
	AgentID     string    `json:"agent_id" gorm:"size:100"` // 当前服务的客服ID
	Status      string    `json:"status" gorm:"size:50;default:'active'"` // active, closed
	LastMessage string    `json:"last_message" gorm:"type:text"` // 最后一条消息
	MessageCount int      `json:"message_count" gorm:"default:0"` // 消息数量
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}
