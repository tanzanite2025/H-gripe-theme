package ticket

import (
	"commerce-platform/internal/domain/user"
	"time"
)

// TicketMessage 工单消息
type TicketMessage struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	TicketID    uint       `gorm:"not null;index" json:"ticket_id"`
	UserID      uint       `gorm:"not null" json:"user_id"`
	IsStaff     bool       `gorm:"default:false" json:"is_staff"` // 是否客服回复
	Content     string     `gorm:"type:text;not null" json:"content"`
	MessageType string     `gorm:"type:varchar(40);default:'text';not null;index" json:"message_type"`
	Metadata    string     `gorm:"type:text" json:"metadata"`
	Attachments string     `gorm:"type:text" json:"attachments"`     // JSON数组
	IsInternal  bool       `gorm:"default:false" json:"is_internal"` // 是否内部备注
	IsRead      bool       `gorm:"default:false" json:"is_read"`
	User        *user.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName 指定表名
func (TicketMessage) TableName() string {
	return "ticket_messages"
}
