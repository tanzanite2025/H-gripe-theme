package social

import (
	"time"

	"gorm.io/datatypes"
)

const (
	ProviderMeta    = "meta"
	ProviderX       = "x"
	ProviderYouTube = "youtube"
	ProviderReddit  = "reddit"
)

const (
	OAuthStatusDisconnected = "disconnected"
	OAuthStatusConnected    = "connected"
	OAuthStatusError        = "error"
)

const (
	OAuthSessionStatusPending  = "pending"
	OAuthSessionStatusConsumed = "consumed"
	OAuthSessionStatusError    = "error"
)

var SupportedProviders = []string{
	ProviderMeta,
	ProviderX,
	ProviderYouTube,
	ProviderReddit,
}

type OAuthConnection struct {
	ID                    uint           `gorm:"primarykey" json:"id"`
	Provider              string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"provider"`
	Status                string         `gorm:"type:varchar(24);not null;default:'disconnected';index" json:"status"`
	ProviderAccountID     string         `gorm:"type:varchar(255);not null;default:''" json:"provider_account_id"`
	ProviderAccountName   string         `gorm:"type:varchar(255);not null;default:''" json:"provider_account_name"`
	ProviderAccountURL    string         `gorm:"type:varchar(500);not null;default:''" json:"provider_account_url"`
	ProviderAccountEmail  string         `gorm:"type:varchar(320);not null;default:''" json:"provider_account_email"`
	AccessTokenEncrypted  string         `gorm:"type:text;not null;default:''" json:"-"`
	RefreshTokenEncrypted string         `gorm:"type:text;not null;default:''" json:"-"`
	TokenExpiresAt        *time.Time     `json:"-"`
	GrantedScopes         string         `gorm:"type:text;not null;default:''" json:"-"`
	ProviderResources     datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"-"`
	LastConnectedAt       *time.Time     `json:"last_connected_at,omitempty"`
	LastError             string         `gorm:"type:text;not null;default:''" json:"last_error,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

func (OAuthConnection) TableName() string {
	return "social_oauth_connections"
}

type OAuthSession struct {
	ID                    uint       `gorm:"primarykey" json:"id"`
	Provider              string     `gorm:"type:varchar(32);not null;index" json:"provider"`
	StateHash             string     `gorm:"type:varchar(128);uniqueIndex;not null" json:"-"`
	CodeVerifierEncrypted string     `gorm:"type:text;not null;default:''" json:"-"`
	RedirectURI           string     `gorm:"type:varchar(500);not null;default:''" json:"-"`
	ReturnPath            string     `gorm:"type:varchar(500);not null;default:''" json:"return_path"`
	InitiatedByUserID     uint       `gorm:"not null;default:0;index" json:"initiated_by_user_id"`
	ExpiresAt             time.Time  `gorm:"not null;index" json:"expires_at"`
	ConsumedAt            *time.Time `json:"consumed_at,omitempty"`
	Status                string     `gorm:"type:varchar(24);not null;default:'pending';index" json:"status"`
	LastError             string     `gorm:"type:text;not null;default:''" json:"last_error"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

func (OAuthSession) TableName() string {
	return "social_oauth_sessions"
}
