package ops

import (
	"time"

	"gorm.io/gorm"
)

const (
	VPSProviderHostinger = "hostinger"
	VPSProviderOther     = "other"

	VPSEnvironmentProduction = "production"
	VPSEnvironmentStaging    = "staging"
	VPSEnvironmentTest       = "test"
	VPSEnvironmentLocal      = "local"

	VPSStatusActive   = "active"
	VPSStatusPending  = "pending"
	VPSStatusDisabled = "disabled"
	VPSStatusDrifted  = "drifted"
	VPSStatusError    = "error"

	VPSObservedHealthy  = "healthy"
	VPSObservedDegraded = "degraded"
	VPSObservedUnknown  = "unknown"
	VPSObservedOffline  = "offline"
)

type VPSBinding struct {
	ID                 uint           `gorm:"primarykey" json:"id"`
	Name               string         `gorm:"size:120;not null;uniqueIndex:idx_ops_vps_binding_name" json:"name"`
	Provider           string         `gorm:"size:32;not null;index" json:"provider"`
	Environment        string         `gorm:"size:32;not null;index" json:"environment"`
	ConnectorID        *uint          `gorm:"index" json:"connector_id,omitempty"`
	ProviderResourceID string         `gorm:"size:120;not null;default:''" json:"provider_resource_id"`
	Hostname           string         `gorm:"size:255;not null;default:''" json:"hostname"`
	IPv4               string         `gorm:"size:64;not null;default:''" json:"ipv4"`
	Region             string         `gorm:"size:120;not null;default:''" json:"region"`
	OperatingSystem    string         `gorm:"size:160;not null;default:''" json:"operating_system"`
	Status             string         `gorm:"size:32;not null;default:'pending';index" json:"status"`
	ObservedStatus     string         `gorm:"size:32;not null;default:'unknown';index" json:"observed_status"`
	ObservedState      string         `gorm:"size:64;not null;default:'';index" json:"observed_state"`
	ObservedSource     string         `gorm:"size:120;not null;default:''" json:"observed_source"`
	ObservedHostname   string         `gorm:"size:255;not null;default:''" json:"observed_hostname"`
	ObservedIPv4       string         `gorm:"size:64;not null;default:''" json:"observed_ipv4"`
	ObservedOS         string         `gorm:"column:observed_operating_system;size:160;not null;default:''" json:"observed_operating_system"`
	ObservedPlan       string         `gorm:"size:120;not null;default:''" json:"observed_plan"`
	ObservedRegion     string         `gorm:"size:120;not null;default:''" json:"observed_region"`
	Enabled            bool           `gorm:"not null;default:true;index" json:"enabled"`
	LastObservedAt     *time.Time     `json:"last_observed_at,omitempty"`
	LastError          string         `gorm:"type:text;not null;default:''" json:"last_error"`
	Notes              string         `gorm:"type:text;not null;default:''" json:"notes"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (VPSBinding) TableName() string {
	return "ops_vps_bindings"
}
