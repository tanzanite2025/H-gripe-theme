package ops

import (
	"time"

	"gorm.io/gorm"
)

const (
	NetworkOwnerVPS       = "vps"
	NetworkOwnerProject   = "project"
	NetworkOwnerDomain    = "domain"
	NetworkOwnerConnector = "connector"
	NetworkOwnerManual    = "manual"
	NetworkOwnerOther     = "other"

	NetworkManagedByCloudflare = "cloudflare"
	NetworkManagedByHostinger  = "hostinger"
	NetworkManagedByOSFirewall = "os_firewall"
	NetworkManagedByManual     = "manual"
	NetworkManagedByOther      = "other"

	NetworkSourceManual        = "manual"
	NetworkSourceDomainBinding = "domain_binding"
	NetworkSourceFirewallRule  = "firewall_rule"
	NetworkSourceProvider      = "provider"

	NetworkScopeEdge       = "edge"
	NetworkScopeDNS        = "dns"
	NetworkScopeOrigin     = "origin"
	NetworkScopeOSFirewall = "os_firewall"
	NetworkScopeGateway    = "gateway"

	NetworkDirectionIngress = "ingress"
	NetworkDirectionEgress  = "egress"

	NetworkProtocolTCP  = "tcp"
	NetworkProtocolUDP  = "udp"
	NetworkProtocolICMP = "icmp"
	NetworkProtocolAll  = "all"

	NetworkStateOpen       = "open"
	NetworkStateClosed     = "closed"
	NetworkStateRestricted = "restricted"
	NetworkStateUnknown    = "unknown"
	NetworkStateDrifted    = "drifted"
	NetworkStateError      = "error"

	NetworkStatusActive   = "active"
	NetworkStatusPending  = "pending"
	NetworkStatusDisabled = "disabled"
	NetworkStatusDrifted  = "drifted"
	NetworkStatusError    = "error"

	NetworkItemKindRule      = "rule"
	NetworkItemKindDomainDNS = "domain_dns"
)

type NetworkRule struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	Name             string         `gorm:"size:160;not null;index" json:"name"`
	Environment      string         `gorm:"size:32;not null;index" json:"environment"`
	VPSBindingID     *uint          `gorm:"index" json:"vps_binding_id,omitempty"`
	ProjectBindingID *uint          `gorm:"index" json:"project_binding_id,omitempty"`
	DomainBindingID  *uint          `gorm:"index" json:"domain_binding_id,omitempty"`
	ConnectorID      *uint          `gorm:"index" json:"connector_id,omitempty"`
	OwnerKind        string         `gorm:"size:32;not null;index" json:"owner_kind"`
	OwnerID          uint           `gorm:"not null;default:0;index" json:"owner_id"`
	ManagedBy        string         `gorm:"size:64;not null;index" json:"managed_by"`
	SourceKind       string         `gorm:"size:64;not null;default:'';index" json:"source_kind"`
	Scope            string         `gorm:"size:64;not null;index" json:"scope"`
	Direction        string         `gorm:"size:16;not null;default:'ingress';index" json:"direction"`
	Protocol         string         `gorm:"size:16;not null;default:'tcp'" json:"protocol"`
	Ports            string         `gorm:"size:120;not null;default:''" json:"ports"`
	SourceCIDR       string         `gorm:"size:120;not null;default:''" json:"source_cidr"`
	Target           string         `gorm:"size:255;not null;default:''" json:"target"`
	DesiredState     string         `gorm:"size:32;not null;default:'unknown'" json:"desired_state"`
	ObservedState    string         `gorm:"size:32;not null;default:'unknown'" json:"observed_state"`
	EffectiveState   string         `gorm:"size:32;not null;default:'unknown'" json:"effective_state"`
	Status           string         `gorm:"size:32;not null;default:'pending';index" json:"status"`
	Enabled          bool           `gorm:"not null;default:true;index" json:"enabled"`
	LastObservedAt   *time.Time     `json:"last_observed_at,omitempty"`
	LastError        string         `gorm:"type:text;not null;default:''" json:"last_error"`
	Notes            string         `gorm:"type:text;not null;default:''" json:"notes"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (NetworkRule) TableName() string {
	return "ops_network_rules"
}

type NetworkSummaryCounts struct {
	Total             int            `json:"total"`
	Enabled           int            `json:"enabled"`
	Attention         int            `json:"attention"`
	Unknown           int            `json:"unknown"`
	ExplicitRuleCount int            `json:"explicit_rule_count"`
	InferredItemCount int            `json:"inferred_item_count"`
	VPSCount          int            `json:"vps_count"`
	ManagedBy         map[string]int `json:"managed_by"`
	Scopes            map[string]int `json:"scopes"`
}

type NetworkSummaryItem struct {
	Key              string     `json:"key"`
	Kind             string     `json:"kind"`
	ID               uint       `json:"id"`
	Name             string     `json:"name"`
	Environment      string     `json:"environment"`
	OwnerKind        string     `json:"owner_kind"`
	OwnerID          uint       `json:"owner_id"`
	OwnerName        string     `json:"owner_name"`
	VPSBindingID     *uint      `json:"vps_binding_id,omitempty"`
	VPSName          string     `json:"vps_name,omitempty"`
	ProjectBindingID *uint      `json:"project_binding_id,omitempty"`
	ProjectName      string     `json:"project_name,omitempty"`
	DomainBindingID  *uint      `json:"domain_binding_id,omitempty"`
	DomainName       string     `json:"domain_name,omitempty"`
	ConnectorID      *uint      `json:"connector_id,omitempty"`
	ConnectorName    string     `json:"connector_name,omitempty"`
	ManagedBy        string     `json:"managed_by"`
	SourceKind       string     `json:"source_kind"`
	Scope            string     `json:"scope"`
	Direction        string     `json:"direction"`
	Protocol         string     `json:"protocol"`
	Ports            string     `json:"ports"`
	SourceCIDR       string     `json:"source_cidr"`
	Target           string     `json:"target"`
	DesiredState     string     `json:"desired_state"`
	ObservedState    string     `json:"observed_state"`
	EffectiveState   string     `json:"effective_state"`
	Status           string     `json:"status"`
	Enabled          bool       `json:"enabled"`
	LastObservedAt   *time.Time `json:"last_observed_at,omitempty"`
	LastError        string     `json:"last_error,omitempty"`
	Notes            string     `json:"notes,omitempty"`
}

type NetworkSummary struct {
	Environment string               `json:"environment"`
	GeneratedAt time.Time            `json:"generated_at"`
	Summary     NetworkSummaryCounts `json:"summary"`
	Items       []NetworkSummaryItem `json:"items"`
}
