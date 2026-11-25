package ticket

import (
	"time"

	"commerce-platform/internal/domain/user"
)

// TicketMessage 工单消息
type TicketMessage struct {
	ID                               uint                                                  `gorm:"primarykey;index:idx_ticket_messages_ticket_created_id,priority:3" json:"id"`
	TicketID                         uint                                                  `gorm:"not null;index;index:idx_ticket_messages_ticket_created_id,priority:1" json:"ticket_id"`
	ClientMessageID                  *string                                               `gorm:"type:varchar(64);uniqueIndex:idx_ticket_messages_ticket_client_message_id,priority:2" json:"-"`
	ClientMessagePayloadHash         string                                                `gorm:"type:varchar(64);not null;default:''" json:"-"`
	UserID                           *uint                                                 `gorm:"index" json:"user_id"`
	IsStaff                          bool                                                  `gorm:"default:false" json:"is_staff"` // 是否客服回复
	Content                          string                                                `gorm:"type:text;not null" json:"content"`
	MessageType                      string                                                `gorm:"type:varchar(40);default:'text';not null;index" json:"message_type"`
	Metadata                         string                                                `gorm:"type:text" json:"metadata"`
	Attachments                      string                                                `gorm:"type:text" json:"attachments"` // JSON数组
	AttachmentReferenceFormat        string                                                `gorm:"type:varchar(32);not null;default:'legacy_urls'" json:"-"`
	RequestedMediaAttachmentBindings []CustomerServiceMessageMediaAttachmentBindingRequest `gorm:"-" json:"-"`
	IsInternal                       bool                                                  `gorm:"default:false" json:"is_internal"` // 是否内部备注
	IsRead                           bool                                                  `gorm:"default:false" json:"is_read"`
	User                             *user.User                                            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt                        time.Time                                             `gorm:"index:idx_ticket_messages_ticket_created_id,priority:2" json:"created_at"`
}

// TableName 指定表名
func (TicketMessage) TableName() string {
	return "ticket_messages"
}
