package ops

import (
	"time"

	"gorm.io/gorm"
)

const (
	ProjectEnvironmentProduction = "production"
	ProjectEnvironmentStaging    = "staging"
	ProjectEnvironmentTest       = "test"
	ProjectEnvironmentLocal      = "local"

	ProjectStatusActive   = "active"
	ProjectStatusPending  = "pending"
	ProjectStatusDisabled = "disabled"
	ProjectStatusDrifted  = "drifted"
	ProjectStatusError    = "error"

	ProjectHealthHealthy  = "healthy"
	ProjectHealthDegraded = "degraded"
	ProjectHealthUnknown  = "unknown"
	ProjectHealthOffline  = "offline"
)

type ProjectBinding struct {
	ID                      uint           `gorm:"primarykey" json:"id"`
	Name                    string         `gorm:"size:120;not null;uniqueIndex:idx_ops_project_binding_name" json:"name"`
	VPSBindingID            uint           `gorm:"not null;index" json:"vps_binding_id"`
	ConnectorID             *uint          `gorm:"index" json:"connector_id,omitempty"`
	ProviderResourceID      string         `gorm:"size:120;not null;default:''" json:"provider_resource_id"`
	Environment             string         `gorm:"size:32;not null;index" json:"environment"`
	ComposeSource           string         `gorm:"size:255;not null;default:''" json:"compose_source"`
	ComposeProjectName      string         `gorm:"size:120;not null;default:''" json:"compose_project_name"`
	GatewayNetwork          string         `gorm:"size:120;not null;default:''" json:"gateway_network"`
	GatewayAlias            string         `gorm:"size:120;not null;default:''" json:"gateway_alias"`
	Services                string         `gorm:"type:text;not null;default:''" json:"services"`
	Networks                string         `gorm:"type:text;not null;default:''" json:"networks"`
	Volumes                 string         `gorm:"type:text;not null;default:''" json:"volumes"`
	CurrentImageTag         string         `gorm:"size:160;not null;default:''" json:"current_image_tag"`
	CurrentCommitSHA        string         `gorm:"size:80;not null;default:''" json:"current_commit_sha"`
	Status                  string         `gorm:"size:32;not null;default:'pending';index" json:"status"`
	HealthStatus            string         `gorm:"size:32;not null;default:'unknown';index" json:"health_status"`
	ObservedState           string         `gorm:"size:64;not null;default:'';index" json:"observed_state"`
	ObservedSource          string         `gorm:"size:120;not null;default:''" json:"observed_source"`
	ObservedContainerCount  int            `gorm:"column:observed_container_count;not null;default:0" json:"observed_container_count"`
	ObservedRunningCount    int            `gorm:"column:observed_running_container_count;not null;default:0" json:"observed_running_container_count"`
	ObservedHealthyCount    int            `gorm:"column:observed_healthy_container_count;not null;default:0" json:"observed_healthy_container_count"`
	Enabled                 bool           `gorm:"not null;default:true;index" json:"enabled"`
	LastDeploymentAt        *time.Time     `json:"last_deployment_at,omitempty"`
	LastCheckedAt           *time.Time     `json:"last_checked_at,omitempty"`
	LastError               string         `gorm:"type:text;not null;default:''" json:"last_error"`
	BackupPolicy            string         `gorm:"type:text;not null;default:''" json:"backup_policy"`
	RestoreNotes            string         `gorm:"type:text;not null;default:''" json:"restore_notes"`
	QuickBuyRateLimitPolicy string         `gorm:"type:text;not null;default:''" json:"quick_buy_rate_limit_policy"`
	Notes                   string         `gorm:"type:text;not null;default:''" json:"notes"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProjectBinding) TableName() string {
	return "ops_project_bindings"
}

type ProjectBindingView struct {
	ProjectBinding
	VPSName        string `json:"vps_name"`
	VPSProvider    string `json:"vps_provider"`
	VPSHostname    string `json:"vps_hostname"`
	VPSIPv4        string `json:"vps_ipv4"`
	VPSConnectorID *uint  `json:"vps_connector_id,omitempty"`
}
