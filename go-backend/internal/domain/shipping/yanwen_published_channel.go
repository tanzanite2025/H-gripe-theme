package shipping

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	YanwenPublishedChannelEnvironmentFAT        = "fat"
	YanwenPublishedChannelEnvironmentProduction = "production"
)

// YanwenPublishedChannel is an operator-curated reference to a Yanwen
// product and destination country scope that may be published to the main
// shipping management domain.
type YanwenPublishedChannel struct {
	ID                       uint           `gorm:"primarykey" json:"id"`
	Environment              string         `gorm:"size:16;not null;default:'production';uniqueIndex:idx_shipping_yanwen_channel_environment_product_code,priority:1,where:deleted_at IS NULL" json:"environment"`
	ProductCode              string         `gorm:"type:varchar(80);not null;uniqueIndex:idx_shipping_yanwen_channel_environment_product_code,priority:2,where:deleted_at IS NULL" json:"product_code"`
	DisplayName              string         `gorm:"type:varchar(160);not null" json:"display_name"`
	Countries                string         `gorm:"type:text;not null;default:'[]'" json:"countries"`
	PackageType              string         `gorm:"type:varchar(80);not null;default:''" json:"package_type"`
	MaxWeightGrams           int            `gorm:"not null;default:0" json:"max_weight_grams"`
	VolumetricDivisor        int            `gorm:"not null;default:8000" json:"volumetric_divisor"`
	RequireReceiverTaxNumber bool           `gorm:"not null;default:false" json:"require_receiver_tax_number"`
	RequireIOSS              bool           `gorm:"not null;default:false" json:"require_ioss"`
	RequireEORI              bool           `gorm:"not null;default:false" json:"require_eori"`
	Notes                    string         `gorm:"type:text;not null;default:''" json:"notes"`
	Enabled                  bool           `gorm:"not null;default:false;index" json:"enabled"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
}

func (YanwenPublishedChannel) TableName() string { return "shipping_yanwen_published_channels" }

func (c *YanwenPublishedChannel) Validate() error {
	if c == nil {
		return errors.New("Yanwen channel is required")
	}
	environment, err := NormalizeYanwenPublishedChannelEnvironment(c.Environment)
	if err != nil {
		return err
	}
	c.Environment = environment
	c.ProductCode = strings.TrimSpace(c.ProductCode)
	c.DisplayName = strings.TrimSpace(c.DisplayName)
	c.Countries = NormalizeShippingServiceCollectionCountryCodes(c.Countries)
	c.PackageType = strings.TrimSpace(c.PackageType)
	c.Notes = strings.TrimSpace(c.Notes)
	if c.ProductCode == "" {
		return errors.New("Yanwen product code is required")
	}
	if c.DisplayName == "" {
		return errors.New("Yanwen display name is required")
	}
	if c.MaxWeightGrams < 0 {
		return errors.New("Yanwen max weight cannot be negative")
	}
	if c.VolumetricDivisor <= 0 {
		return errors.New("Yanwen volumetric divisor must be positive")
	}
	return nil
}

func NormalizeYanwenPublishedChannelEnvironment(value string) (string, error) {
	environment := strings.ToLower(strings.TrimSpace(value))
	if environment == "" {
		environment = YanwenPublishedChannelEnvironmentProduction
	}
	if environment == "test" {
		environment = YanwenPublishedChannelEnvironmentFAT
	}
	if environment != YanwenPublishedChannelEnvironmentFAT && environment != YanwenPublishedChannelEnvironmentProduction {
		return "", errors.New("Yanwen channel environment must be fat or production")
	}
	return environment, nil
}

func (c *YanwenPublishedChannel) BeforeSave(tx *gorm.DB) error {
	return c.Validate()
}
