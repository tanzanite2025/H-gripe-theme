package shipping

import (
	"time"

	"gorm.io/gorm"
)

// ShippingRouteGroup is the storefront-facing aggregate of published
// provider routes. It contains route identity and destination coverage only;
// customer prices belong to ShippingTemplate.
type ShippingRouteGroup struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	Name            string           `gorm:"not null" json:"name"`
	Description     string           `gorm:"type:text" json:"description"`
	Countries       string           `gorm:"type:text;not null;default:'[]'" json:"countries"`
	Enabled         bool             `gorm:"not null;default:true" json:"enabled"`
	SortOrder       int              `gorm:"not null;default:0" json:"sort_order"`
	CarrierServices []CarrierService `gorm:"many2many:shipping_route_group_carrier_services;joinForeignKey:RouteGroupID;joinReferences:CarrierServiceID;" json:"carrier_services,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `gorm:"index" json:"-"`
}

type ShippingRouteGroupCarrierService struct {
	RouteGroupID     uint      `gorm:"primaryKey;column:route_group_id" json:"route_group_id"`
	CarrierServiceID uint      `gorm:"primaryKey;column:carrier_service_id" json:"carrier_service_id"`
	SortOrder        int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
}

func (ShippingRouteGroupCarrierService) TableName() string {
	return "shipping_route_group_carrier_services"
}

func (ShippingRouteGroup) TableName() string { return "shipping_route_groups" }
