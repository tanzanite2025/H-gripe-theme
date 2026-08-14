package ops

import (
	"time"

	"gorm.io/gorm"
)

const (
	DomainRoleCanonical    = "canonical"
	DomainRoleAlias        = "alias"
	DomainRoleAdmin        = "admin"
	DomainRoleRedirect     = "redirect"
	DomainRoleVerification = "verification"
	DomainRoleInternal     = "internal"

	DomainEnvironmentProduction = "production"
	DomainEnvironmentStaging    = "staging"
	DomainEnvironmentTest       = "test"
	DomainEnvironmentLocal      = "local"

	DomainProviderCloudflare = "cloudflare"
	DomainProviderHostinger  = "hostinger"
	DomainProviderOther      = "other"

	DomainProxyProxied = "proxied"
	DomainProxyDNSOnly = "dns_only"
	DomainProxyUnknown = "unknown"

	DomainTLSFullStrict = "full_strict"
	DomainTLSFull       = "full"
	DomainTLSFlexible   = "flexible"
	DomainTLSOff        = "off"
	DomainTLSUnknown    = "unknown"

	DomainStatusActive   = "active"
	DomainStatusPending  = "pending"
	DomainStatusDisabled = "disabled"
	DomainStatusDrifted  = "drifted"
	DomainStatusError    = "error"

	DomainObservedUnknown = "unknown"
	DomainObservedMatched = "matched"
	DomainObservedDrifted = "drifted"
	DomainObservedError   = "error"
)

type DomainBinding struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	Domain           string         `gorm:"size:255;not null;uniqueIndex:idx_ops_domain_binding_domain" json:"domain"`
	ConnectorID      *uint          `gorm:"index" json:"connector_id,omitempty"`
	ProjectBindingID *uint          `gorm:"index" json:"project_binding_id,omitempty"`
	Role             string         `gorm:"size:32;not null;index" json:"role"`
	Environment      string         `gorm:"size:32;not null;index" json:"environment"`
	Provider         string         `gorm:"size:32;not null" json:"provider"`
	Zone             string         `gorm:"size:255;not null;default:''" json:"zone"`
	Target           string         `gorm:"size:255;not null;default:''" json:"target"`
	ProxyMode        string         `gorm:"size:32;not null;default:'unknown'" json:"proxy_mode"`
	TLSMode          string         `gorm:"size:32;not null;default:'unknown'" json:"tls_mode"`
	RedirectTarget   string         `gorm:"size:255;not null;default:''" json:"redirect_target"`
	Status           string         `gorm:"size:32;not null;default:'pending';index" json:"status"`
	ObservedStatus   string         `gorm:"size:32;not null;default:'unknown';index" json:"observed_status"`
	ObservedTarget   string         `gorm:"size:255;not null;default:''" json:"observed_target"`
	ObservedProxy    string         `gorm:"column:observed_proxy_mode;size:32;not null;default:'unknown'" json:"observed_proxy_mode"`
	ObservedTLS      string         `gorm:"column:observed_tls_mode;size:32;not null;default:'unknown'" json:"observed_tls_mode"`
	ObservedSource   string         `gorm:"size:120;not null;default:''" json:"observed_source"`
	LastObservedAt   *time.Time     `json:"last_observed_at,omitempty"`
	ObservedError    string         `gorm:"type:text;not null;default:''" json:"observed_error,omitempty"`
	Enabled          bool           `gorm:"not null;default:true;index" json:"enabled"`
	Notes            string         `gorm:"type:text;not null;default:''" json:"notes"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DomainBinding) TableName() string {
	return "ops_domain_bindings"
}
