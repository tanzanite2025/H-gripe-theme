package shipping

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	FpxChannelEnvironmentProduction = "production"
	FpxChannelEnvironmentTest       = "test"
)

// FpxChannel is the local reference for a published 4PX service code and its
// destination country scope.
// It is intentionally independent from generic shipping quote models.
type FpxChannel struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Environment string         `gorm:"size:16;not null;default:'production';uniqueIndex:idx_shipping_fpx_channel_environment_service_code,priority:1,where:deleted_at IS NULL" json:"environment"`
	ServiceCode string         `gorm:"type:varchar(80);not null;uniqueIndex:idx_shipping_fpx_channel_environment_service_code,priority:2,where:deleted_at IS NULL" json:"service_code"`
	DisplayName string         `gorm:"type:varchar(160);not null" json:"display_name"`
	Countries   string         `gorm:"type:text;not null;default:'[]'" json:"countries"`
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
	environment, err := NormalizeFpxChannelEnvironment(c.Environment)
	if err != nil {
		return err
	}
	c.Environment = environment
	c.ServiceCode = strings.TrimSpace(c.ServiceCode)
	c.DisplayName = strings.TrimSpace(c.DisplayName)
	c.Countries = NormalizeShippingServiceCollectionCountryCodes(c.Countries)
	if c.ServiceCode == "" {
		return errors.New("service code is required")
	}
	if c.DisplayName == "" {
		return errors.New("display name is required")
	}
	return nil
}

func NormalizeFpxChannelEnvironment(value string) (string, error) {
	environment := strings.ToLower(strings.TrimSpace(value))
	if environment == "" {
		environment = FpxChannelEnvironmentProduction
	}
	if environment != FpxChannelEnvironmentProduction && environment != FpxChannelEnvironmentTest {
		return "", errors.New("4PX channel environment must be production or test")
	}
	return environment, nil
}

// NormalizeShippingServiceCollectionCountryCodes keeps the published service
// contract as a canonical JSON array while accepting the CSV form used by
// manual collection inputs and older records.
func NormalizeShippingServiceCollectionCountryCodes(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "[]"
	}

	var codes []string
	if strings.HasPrefix(value, "[") {
		if err := json.Unmarshal([]byte(value), &codes); err != nil {
			codes = nil
		}
	}
	if len(codes) == 0 && !strings.HasPrefix(value, "[") {
		codes = strings.FieldsFunc(value, func(r rune) bool {
			return r == ',' || r == '\uFF0C' || r == ';' || r == '|' || r == ' ' || r == '\n' || r == '\r' || r == '\t'
		})
	}

	seen := make(map[string]struct{}, len(codes))
	normalized := make([]string, 0, len(codes))
	for _, code := range codes {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		normalized = append(normalized, code)
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func (c *FpxChannel) BeforeSave(tx *gorm.DB) error {
	return c.Validate()
}
