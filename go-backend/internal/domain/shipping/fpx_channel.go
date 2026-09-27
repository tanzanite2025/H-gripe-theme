package shipping

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// FpxChannel is the minimal local reference for a 4PX service code.
// It is intentionally independent from generic shipping quote models.
type FpxChannel struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	ServiceCode string         `gorm:"type:varchar(80);not null;uniqueIndex:idx_shipping_fpx_channel_service_code,where:deleted_at IS NULL" json:"service_code"`
	DisplayName string         `gorm:"type:varchar(160);not null" json:"display_name"`
	Enabled     bool           `gorm:"not null;default:false;index" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FpxChannel) TableName() string { return "shipping_fpx_channels" }

func (c *FpxChannel) Validate() error {
	if c == nil {
		return errors.New("4PX channel is required")
	}
	c.ServiceCode = strings.TrimSpace(c.ServiceCode)
	c.DisplayName = strings.TrimSpace(c.DisplayName)
	if c.ServiceCode == "" {
		return errors.New("service code is required")
	}
	if c.DisplayName == "" {
		return errors.New("display name is required")
	}
	return nil
}

func (c *FpxChannel) BeforeSave(tx *gorm.DB) error {
	return c.Validate()
}
