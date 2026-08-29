package security

import (
	"time"

	"gorm.io/gorm"
)

const (
	IPBlockRuleSourceManual         = "manual"
	IPBlockRuleSourceVisitorProfile = "visitor_profile"
	IPBlockRuleSourceCommercialBot  = "commercial_crawler"
	IPBlockRuleSourceRiskAutomation = "risk_automation"

	IPBlockRuleStatusActive   = "active"
	IPBlockRuleStatusExpired  = "expired"
	IPBlockRuleStatusDisabled = "disabled"
)

// IPBlockRule is the durable source of truth for application-wide IP/CIDR
// enforcement. Different detectors can create rules through the same service
// while retaining their own source and reference metadata.
type IPBlockRule struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CIDR            string         `gorm:"column:cidr;size:120;not null;index" json:"cidr"`
	Source          string         `gorm:"size:64;not null;default:'manual';index" json:"source"`
	SourceReference string         `gorm:"size:160;not null;default:'';index" json:"source_reference"`
	Reason          string         `gorm:"size:500;not null;default:''" json:"reason"`
	ExpiresAt       *time.Time     `gorm:"index" json:"expires_at,omitempty"`
	Enabled         bool           `gorm:"not null;default:true;index" json:"enabled"`
	CreatedBy       *uint          `gorm:"index" json:"created_by,omitempty"`
	DisabledBy      *uint          `gorm:"index" json:"disabled_by,omitempty"`
	DisabledAt      *time.Time     `json:"disabled_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (IPBlockRule) TableName() string {
	return "global_ip_block_rules"
}
