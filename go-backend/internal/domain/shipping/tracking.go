package shipping

import (
	"time"

	"gorm.io/gorm"
)

// TrackingProviderConfig 物流追踪服务商配置，例如 17TRACK、AfterShip 等
type TrackingProviderConfig struct {
	ID                     uint           `gorm:"primarykey" json:"id"`
	ProviderCode           string         `gorm:"type:varchar(80);not null;index" json:"provider_code"`
	ProviderName           string         `gorm:"type:varchar(160);not null" json:"provider_name"`
	Environment            string         `gorm:"type:varchar(40);default:'production';not null" json:"environment"`
	BaseURL                string         `gorm:"type:text" json:"base_url"`
	APIKey                 string         `gorm:"type:text" json:"api_key"`
	WebhookSecret          string         `gorm:"type:text" json:"webhook_secret"`
	WebhookEnabled         bool           `gorm:"default:false;not null" json:"webhook_enabled"`
	AutoRegister           bool           `gorm:"default:false;not null" json:"auto_register"`
	PollingEnabled         bool           `gorm:"default:false;not null" json:"polling_enabled"`
	PollingIntervalMinutes int            `gorm:"default:60;not null" json:"polling_interval_minutes"`
	RequestTimeoutSeconds  int            `gorm:"default:15;not null" json:"request_timeout_seconds"`
	Enabled                bool           `gorm:"default:true;not null;index" json:"enabled"`
	SortOrder              int            `gorm:"default:0;not null;index" json:"sort_order"`
	Description            string         `gorm:"type:text" json:"description"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TrackingProviderConfig) TableName() string {
	return "shipping_tracking_providers"
}

// TrackingCarrierMapping maps local carriers or carrier services to provider-specific carrier codes.
type TrackingCarrierMapping struct {
	ID                  uint                    `gorm:"primarykey" json:"id"`
	ProviderID          uint                    `gorm:"not null;index" json:"provider_id"`
	Scope               string                  `gorm:"type:varchar(40);default:'carrier';not null;index" json:"scope"` // carrier, carrier_service
	CarrierID           *uint                   `gorm:"index" json:"carrier_id"`
	CarrierServiceID    *uint                   `gorm:"index" json:"carrier_service_id"`
	ProviderCarrierCode string                  `gorm:"type:varchar(120);not null;index" json:"provider_carrier_code"`
	ProviderCarrierName string                  `gorm:"type:varchar(160)" json:"provider_carrier_name"`
	Enabled             bool                    `gorm:"default:true;not null;index" json:"enabled"`
	Priority            int                     `gorm:"default:0;not null;index" json:"priority"`
	Description         string                  `gorm:"type:text" json:"description"`
	Provider            *TrackingProviderConfig `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	Carrier             *Carrier                `gorm:"foreignKey:CarrierID" json:"carrier,omitempty"`
	CarrierService      *CarrierService         `gorm:"foreignKey:CarrierServiceID" json:"carrier_service,omitempty"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
	DeletedAt           gorm.DeletedAt          `gorm:"index" json:"-"`
}

func (TrackingCarrierMapping) TableName() string {
	return "shipping_tracking_carrier_mappings"
}

// TrackingShipment tracks the sync lifecycle for an order tracking number.
// Order fields remain the source for provider/carrier selection; this table stores operational status only.
type TrackingShipment struct {
	ID                       uint                    `gorm:"primarykey" json:"id"`
	OrderID                  uint                    `gorm:"not null;index;uniqueIndex:idx_shipping_tracking_shipments_order" json:"order_id"`
	TrackingProviderID       uint                    `gorm:"not null;index" json:"tracking_provider_id"`
	TrackingNumber           string                  `gorm:"type:varchar(120);not null;index" json:"tracking_number"`
	ProviderCarrierCode      string                  `gorm:"type:varchar(120);not null;index" json:"provider_carrier_code"`
	CarrierID                *uint                   `gorm:"index" json:"carrier_id"`
	CarrierServiceID         *uint                   `gorm:"index" json:"carrier_service_id"`
	TrackingCarrierMappingID *uint                   `gorm:"index" json:"tracking_carrier_mapping_id"`
	RegistrationStatus       string                  `gorm:"type:varchar(40);default:'pending';not null;index" json:"registration_status"` // pending, registered, failed
	SyncStatus               string                  `gorm:"type:varchar(40);default:'pending';not null;index" json:"sync_status"`         // pending, syncing, synced, failed
	EventCount               int                     `gorm:"default:0;not null" json:"event_count"`
	LastEventAt              *time.Time              `gorm:"index" json:"last_event_at"`
	LastSyncedAt             *time.Time              `gorm:"index" json:"last_synced_at"`
	NextSyncAt               *time.Time              `gorm:"index" json:"next_sync_at"`
	LastError                string                  `gorm:"type:text" json:"last_error"`
	Enabled                  bool                    `gorm:"default:true;not null;index" json:"enabled"`
	Provider                 *TrackingProviderConfig `gorm:"foreignKey:TrackingProviderID" json:"provider,omitempty"`
	Carrier                  *Carrier                `gorm:"foreignKey:CarrierID" json:"carrier,omitempty"`
	CarrierService           *CarrierService         `gorm:"foreignKey:CarrierServiceID" json:"carrier_service,omitempty"`
	Mapping                  *TrackingCarrierMapping `gorm:"foreignKey:TrackingCarrierMappingID" json:"mapping,omitempty"`
	CreatedAt                time.Time               `json:"created_at"`
	UpdatedAt                time.Time               `json:"updated_at"`
	DeletedAt                gorm.DeletedAt          `gorm:"index" json:"-"`
}

func (TrackingShipment) TableName() string {
	return "shipping_tracking_shipments"
}

// TrackingEvent 物流追踪事件
type TrackingEvent struct {
	ID                  uint      `gorm:"primarykey" json:"id"`
	OrderID             uint      `gorm:"not null;index" json:"order_id"`
	TrackingNumber      string    `gorm:"index" json:"tracking_number"`
	ProviderCarrierCode string    `json:"provider_carrier_code"`
	Status              string    `json:"status"`
	Location            string    `json:"location"`
	Description         string    `gorm:"type:text" json:"description"`
	EventTime           time.Time `json:"event_time"`
	CreatedAt           time.Time `json:"created_at"`
}

// TableName 指定表名
func (TrackingEvent) TableName() string {
	return "tracking_events"
}
