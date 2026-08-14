package ops

import (
	"time"

	"gorm.io/gorm"
)

const (
	ConnectorProviderCloudflare = "cloudflare"
	ConnectorProviderHostinger  = "hostinger"
	ConnectorProviderGitHub     = "github"
	ConnectorProviderGHCR       = "ghcr"
	ConnectorProviderOther      = "other"

	ConnectorEnvironmentProduction = "production"
	ConnectorEnvironmentStaging    = "staging"
	ConnectorEnvironmentTest       = "test"
	ConnectorEnvironmentLocal      = "local"

	ConnectorAuthNone        = "none"
	ConnectorAuthAPIToken    = "api_token"
	ConnectorAuthAPIKey      = "api_key"
	ConnectorAuthBearer      = "bearer"
	ConnectorAuthBasic       = "basic"
	ConnectorAuthEnvironment = "environment"
	ConnectorAuthManual      = "manual"

	ConnectorStatusActive   = "active"
	ConnectorStatusPending  = "pending"
	ConnectorStatusDisabled = "disabled"
	ConnectorStatusError    = "error"

	ConnectorTestSuccess = "success"
	ConnectorTestFailed  = "failed"
)

type Connector struct {
	ID                   uint           `gorm:"primarykey" json:"id"`
	Name                 string         `gorm:"size:120;not null;uniqueIndex:idx_ops_connector_name" json:"name"`
	Provider             string         `gorm:"size:32;not null;index" json:"provider"`
	Environment          string         `gorm:"size:32;not null;index" json:"environment"`
	Endpoint             string         `gorm:"size:500;not null;default:''" json:"endpoint"`
	AuthType             string         `gorm:"size:32;not null;default:'api_token'" json:"auth_type"`
	CredentialRef        string         `gorm:"size:160;not null;default:''" json:"credential_ref"`
	CredentialsEncrypted string         `gorm:"type:text;not null;default:''" json:"-"`
	CredentialFields     string         `gorm:"type:text;not null;default:''" json:"-"`
	Scopes               string         `gorm:"type:text;not null;default:''" json:"scopes"`
	Status               string         `gorm:"size:32;not null;default:'pending';index" json:"status"`
	Enabled              bool           `gorm:"not null;default:true;index" json:"enabled"`
	LastTestStatus       string         `gorm:"size:32;not null;default:''" json:"last_test_status"`
	LastTestedAt         *time.Time     `json:"last_tested_at"`
	LastError            string         `gorm:"type:text;not null;default:''" json:"last_error"`
	Notes                string         `gorm:"type:text;not null;default:''" json:"notes"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Connector) TableName() string {
	return "ops_connectors"
}

type ConnectorView struct {
	ID                   uint       `json:"id"`
	Name                 string     `json:"name"`
	Provider             string     `json:"provider"`
	Environment          string     `json:"environment"`
	Endpoint             string     `json:"endpoint"`
	AuthType             string     `json:"auth_type"`
	CredentialRef        string     `json:"credential_ref"`
	CredentialConfigured bool       `json:"credential_configured"`
	CredentialFields     []string   `json:"credential_fields"`
	Scopes               string     `json:"scopes"`
	Status               string     `json:"status"`
	Enabled              bool       `json:"enabled"`
	LastTestStatus       string     `json:"last_test_status"`
	LastTestedAt         *time.Time `json:"last_tested_at,omitempty"`
	LastError            string     `json:"last_error,omitempty"`
	Notes                string     `json:"notes"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type ConnectorTestResult struct {
	ConnectorID          uint      `json:"connector_id"`
	Success              bool      `json:"success"`
	StatusCode           int       `json:"status_code,omitempty"`
	Message              string    `json:"message"`
	CheckedAt            time.Time `json:"checked_at"`
	CredentialConfigured bool      `json:"credential_configured"`
}
