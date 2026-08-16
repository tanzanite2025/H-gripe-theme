package ops

import "time"

const (
	ConnectorOAuthStatusPending  = "pending"
	ConnectorOAuthStatusConsumed = "consumed"
	ConnectorOAuthStatusError    = "error"
)

type ConnectorOAuthSession struct {
	ID                    uint       `gorm:"primarykey" json:"id"`
	Provider              string     `gorm:"size:32;not null;index" json:"provider"`
	ConnectorID           *uint      `gorm:"index" json:"connector_id,omitempty"`
	StateHash             string     `gorm:"size:128;not null;uniqueIndex" json:"-"`
	CodeVerifierEncrypted string     `gorm:"type:text;not null;default:''" json:"-"`
	ClientID              string     `gorm:"size:255;not null;default:''" json:"-"`
	RedirectURI           string     `gorm:"size:500;not null;default:''" json:"-"`
	ReturnPath            string     `gorm:"size:500;not null;default:''" json:"-"`
	CreatedByUserID       uint       `gorm:"not null;default:0" json:"created_by_user_id"`
	ExpiresAt             time.Time  `gorm:"not null;index" json:"expires_at"`
	ConsumedAt            *time.Time `json:"consumed_at,omitempty"`
	Status                string     `gorm:"size:32;not null;default:'pending';index" json:"status"`
	LastError             string     `gorm:"type:text;not null;default:''" json:"last_error"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

func (ConnectorOAuthSession) TableName() string {
	return "ops_connector_oauth_sessions"
}
