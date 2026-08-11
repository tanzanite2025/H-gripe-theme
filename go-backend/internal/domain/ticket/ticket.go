package ticket

import (
	"commerce-platform/internal/domain/user"
	"time"

	"gorm.io/gorm"
)

// Ticket 客服工单
type Ticket struct {
	ID                 uint            `gorm:"primarykey" json:"id"`
	TicketNumber       string          `gorm:"uniqueIndex;not null" json:"ticket_number"`
	UserID             uint            `gorm:"not null;index" json:"user_id"`
	CustomerUserID     *uint           `gorm:"index" json:"customer_user_id,omitempty"`
	ConversationID     *string         `gorm:"type:varchar(64);uniqueIndex" json:"conversation_id,omitempty"`
	VisitorSessionHash string          `gorm:"type:varchar(64);index" json:"-"`
	Subject            string          `gorm:"not null" json:"subject"`
	Category           string          `gorm:"index" json:"category"`                  // order, product, shipping, other
	Priority           string          `gorm:"index;default:'medium'" json:"priority"` // low, medium, high, urgent
	Status             string          `gorm:"index;default:'open'" json:"status"`     // open, in_progress, resolved, closed
	AssignedTo         uint            `gorm:"index" json:"assigned_to"`               // 分配给的客服ID
	Messages           []TicketMessage `gorm:"foreignKey:TicketID" json:"messages"`
	User               *user.User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tags               string          `json:"tags"` // 逗号分隔的标签
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	ResolvedAt         *time.Time      `json:"resolved_at"`
	ClosedAt           *time.Time      `json:"closed_at"`
	DeletedAt          gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Ticket) TableName() string {
	return "tickets"
}

// BeforeCreate GORM钩子：创建前
func (t *Ticket) BeforeCreate(tx *gorm.DB) error {
	if t.TicketNumber == "" {
		t.TicketNumber = generateTicketNumber()
	}
	return nil
}

func generateTicketNumber() string {
	now := time.Now()
	return now.Format("TK20060102") + randomString(6)
}

func randomString(n int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
