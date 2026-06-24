package ticket

import (
	"tanzanite/internal/domain/user"
	"time"

	"gorm.io/gorm"
)

// Ticket 客服工单
type Ticket struct {
	ID           uint            `gorm:"primarykey" json:"id"`
	TicketNumber string          `gorm:"uniqueIndex;not null" json:"ticket_number"`
	UserID       uint            `gorm:"not null;index" json:"user_id"`
	Subject      string          `gorm:"not null" json:"subject"`
	Category     string          `gorm:"index" json:"category"`                  // order, product, shipping, other
	Priority     string          `gorm:"index;default:'medium'" json:"priority"` // low, medium, high, urgent
	Status       string          `gorm:"index;default:'open'" json:"status"`     // open, in_progress, resolved, closed
	AssignedTo   uint            `gorm:"index" json:"assigned_to"`               // 分配给的客服ID
	Messages     []TicketMessage `gorm:"foreignKey:TicketID" json:"messages"`
	User         *user.User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tags         string          `json:"tags"` // 逗号分隔的标签
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	ResolvedAt   *time.Time      `json:"resolved_at"`
	ClosedAt     *time.Time      `json:"closed_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Ticket) TableName() string {
	return "tickets"
}

// TicketMessage 工单消息
type TicketMessage struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	TicketID    uint       `gorm:"not null;index" json:"ticket_id"`
	UserID      uint       `gorm:"not null" json:"user_id"`
	IsStaff     bool       `gorm:"default:false" json:"is_staff"` // 是否客服回复
	Content     string     `gorm:"type:text;not null" json:"content"`
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

// AutoReplyRule 自动回复规则
type AutoReplyRule struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	Type           string    `gorm:"type:varchar(20);not null;index" json:"type"` // welcome, keyword
	TriggerKeyword string    `gorm:"type:varchar(255);index" json:"trigger_keyword"`
	ReplyMessage   string    `gorm:"type:text;not null" json:"reply_message"`
	AgentID        string    `gorm:"type:varchar(100);index" json:"agent_id"`
	IsActive       bool      `gorm:"default:true;index" json:"is_active"`
	Priority       int       `gorm:"default:0;index" json:"priority"`
	MatchType      string    `gorm:"type:varchar(20);default:'exact'" json:"match_type"` // exact, contains
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AutoReplyRule) TableName() string {
	return "ticket_auto_replies"
}
