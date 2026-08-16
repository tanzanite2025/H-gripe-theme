package ticket

import (
	"time"

	"gorm.io/gorm"
)

// CustomerServiceInboxState is one staff recipient's durable read cursor for
// a customer-service conversation. ticket_messages.is_read is intentionally
// not used for this multi-recipient state.
type CustomerServiceInboxState struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	TicketID          uint           `gorm:"not null;uniqueIndex:idx_customer_service_inbox_state_recipient_ticket;index" json:"ticket_id"`
	RecipientUserID   uint           `gorm:"not null;uniqueIndex:idx_customer_service_inbox_state_recipient_ticket;index" json:"recipient_user_id"`
	LastReadMessageID uint           `gorm:"not null;default:0" json:"last_read_message_id"`
	UnreadCount       int            `gorm:"not null;default:0" json:"unread_count"`
	AssignmentVersion uint           `gorm:"not null;default:1" json:"assignment_version"`
	LastReadAt        *time.Time     `json:"last_read_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CustomerServiceInboxState) TableName() string {
	return "customer_service_inbox_states"
}
