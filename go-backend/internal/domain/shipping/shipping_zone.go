package shipping

import (
	"time"

	"gorm.io/gorm"
)

// ShippingZone 配送区域
type ShippingZone struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Countries   string         `gorm:"type:text" json:"countries"`    // JSON数组
	States      string         `gorm:"type:text" json:"states"`       // JSON数组
	PostalCodes string         `gorm:"type:text" json:"postal_codes"` // JSON数组
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ShippingZone) TableName() string {
	return "shipping_zones"
}
