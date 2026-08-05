package ticket

import "time"

// AutoReplyRule 自动回复规则
type AutoReplyRule struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	Type            string    `gorm:"type:varchar(20);not null;index" json:"type"` // welcome, keyword
	TriggerKeyword  string    `gorm:"type:varchar(255);index" json:"trigger_keyword"`
	ReplyMessage    string    `gorm:"type:text;not null" json:"reply_message"`
	AgentID         string    `gorm:"type:varchar(100);index" json:"agent_id"`
	GroupID         *uint     `gorm:"index" json:"group_id"`
	Locale          string    `gorm:"type:varchar(20);not null;default:'*';index" json:"locale"`
	MessageType     string    `gorm:"type:varchar(40);not null;default:'text';index" json:"message_type"`
	Metadata        string    `gorm:"type:text;not null;default:''" json:"metadata"`
	Attachments     string    `gorm:"type:text;not null;default:''" json:"attachments"`
	IsActive        bool      `gorm:"default:true;index" json:"is_active"`
	Priority        int       `gorm:"default:0;index" json:"priority"`
	MatchType       string    `gorm:"type:varchar(20);default:'exact'" json:"match_type"` // exact, contains
	CooldownSeconds int       `gorm:"default:0" json:"cooldown_seconds"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AutoReplyRule) TableName() string {
	return "ticket_auto_replies"
}
