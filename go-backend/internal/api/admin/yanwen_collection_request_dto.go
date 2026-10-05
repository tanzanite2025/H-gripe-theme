package admin

import (
	"errors"
	"strings"

	"commerce-platform/internal/domain/shipping"
)

type yanwenPublishedChannelRequest struct {
	Environment              string  `json:"environment"`
	ProductCode              string  `json:"product_code"`
	DisplayName              string  `json:"display_name"`
	Countries                *string `json:"countries"`
	PackageType              string  `json:"package_type"`
	MaxWeightGrams           *int    `json:"max_weight_grams"`
	VolumetricDivisor        *int    `json:"volumetric_divisor"`
	RequireReceiverTaxNumber *bool   `json:"require_receiver_tax_number"`
	RequireIOSS              *bool   `json:"require_ioss"`
	RequireEORI              *bool   `json:"require_eori"`
	Notes                    *string `json:"notes"`
	Enabled                  *bool   `json:"enabled"`
}

func (r yanwenPublishedChannelRequest) newYanwenPublishedChannelDomain() shipping.YanwenPublishedChannel {
	maxWeightGrams := 0
	if r.MaxWeightGrams != nil {
		maxWeightGrams = *r.MaxWeightGrams
	}
	volumetricDivisor := 8000
	if r.VolumetricDivisor != nil {
		volumetricDivisor = *r.VolumetricDivisor
	}
	enabled := false
	if r.Enabled != nil {
		enabled = *r.Enabled
	}
	return shipping.YanwenPublishedChannel{
		Environment:              strings.TrimSpace(r.Environment),
		ProductCode:              strings.TrimSpace(r.ProductCode),
		DisplayName:              strings.TrimSpace(r.DisplayName),
		Countries:                shipping.NormalizeShippingServiceCollectionCountryCodes(yanwenRequestStringValue(r.Countries)),
		PackageType:              strings.TrimSpace(r.PackageType),
		MaxWeightGrams:           maxWeightGrams,
		VolumetricDivisor:        volumetricDivisor,
		RequireReceiverTaxNumber: yanwenRequestBoolValue(r.RequireReceiverTaxNumber),
		RequireIOSS:              yanwenRequestBoolValue(r.RequireIOSS),
		RequireEORI:              yanwenRequestBoolValue(r.RequireEORI),
		Notes:                    strings.TrimSpace(yanwenRequestStringValue(r.Notes)),
		Enabled:                  enabled,
	}
}

func (r yanwenPublishedChannelRequest) applyToYanwenPublishedChannel(channel *shipping.YanwenPublishedChannel) error {
	if strings.TrimSpace(r.Environment) != "" {
		environment, err := shipping.NormalizeYanwenPublishedChannelEnvironment(r.Environment)
		if err != nil {
			return err
		}
		if environment != channel.Environment {
			return errors.New("Yanwen channel environment cannot be changed after creation")
		}
	}
	if strings.TrimSpace(r.ProductCode) != "" {
		channel.ProductCode = strings.TrimSpace(r.ProductCode)
	}
	if strings.TrimSpace(r.DisplayName) != "" {
		channel.DisplayName = strings.TrimSpace(r.DisplayName)
	}
	if r.Countries != nil {
		channel.Countries = shipping.NormalizeShippingServiceCollectionCountryCodes(*r.Countries)
	}
	if strings.TrimSpace(r.PackageType) != "" {
		channel.PackageType = strings.TrimSpace(r.PackageType)
	}
	if r.MaxWeightGrams != nil {
		channel.MaxWeightGrams = *r.MaxWeightGrams
	}
	if r.VolumetricDivisor != nil {
		channel.VolumetricDivisor = *r.VolumetricDivisor
	}
	if r.RequireReceiverTaxNumber != nil {
		channel.RequireReceiverTaxNumber = *r.RequireReceiverTaxNumber
	}
	if r.RequireIOSS != nil {
		channel.RequireIOSS = *r.RequireIOSS
	}
	if r.RequireEORI != nil {
		channel.RequireEORI = *r.RequireEORI
	}
	if r.Notes != nil {
		channel.Notes = strings.TrimSpace(*r.Notes)
	}
	if r.Enabled != nil {
		channel.Enabled = *r.Enabled
	}
	return nil
}

func yanwenRequestStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func yanwenRequestBoolValue(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}
