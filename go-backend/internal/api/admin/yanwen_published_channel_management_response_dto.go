package admin

import (
	"time"

	"commerce-platform/internal/domain/shipping"
)

// YanwenPublishedChannelManagementDTO is the full channel contract for the
// Yanwen administration page. It keeps the HTTP response independent from
// the persisted domain entity while retaining Yanwen-only operational fields.
type YanwenPublishedChannelManagementDTO struct {
	ID                       uint      `json:"id"`
	Environment              string    `json:"environment"`
	ProductCode              string    `json:"product_code"`
	DisplayName              string    `json:"display_name"`
	Countries                string    `json:"countries"`
	PackageType              string    `json:"package_type"`
	MaxWeightGrams           int       `json:"max_weight_grams"`
	VolumetricDivisor        int       `json:"volumetric_divisor"`
	RequireReceiverTaxNumber bool      `json:"require_receiver_tax_number"`
	RequireIOSS              bool      `json:"require_ioss"`
	RequireEORI              bool      `json:"require_eori"`
	Notes                    string    `json:"notes"`
	Enabled                  bool      `json:"enabled"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func newYanwenPublishedChannelManagementDTO(
	channel shipping.YanwenPublishedChannel,
) YanwenPublishedChannelManagementDTO {
	return YanwenPublishedChannelManagementDTO{
		ID:                       channel.ID,
		Environment:              channel.Environment,
		ProductCode:              channel.ProductCode,
		DisplayName:              channel.DisplayName,
		Countries:                channel.Countries,
		PackageType:              channel.PackageType,
		MaxWeightGrams:           channel.MaxWeightGrams,
		VolumetricDivisor:        channel.VolumetricDivisor,
		RequireReceiverTaxNumber: channel.RequireReceiverTaxNumber,
		RequireIOSS:              channel.RequireIOSS,
		RequireEORI:              channel.RequireEORI,
		Notes:                    channel.Notes,
		Enabled:                  channel.Enabled,
		CreatedAt:                channel.CreatedAt,
		UpdatedAt:                channel.UpdatedAt,
	}
}

func newYanwenPublishedChannelManagementDTOs(
	channels []shipping.YanwenPublishedChannel,
) []YanwenPublishedChannelManagementDTO {
	dtos := make([]YanwenPublishedChannelManagementDTO, 0, len(channels))
	for _, channel := range channels {
		dtos = append(dtos, newYanwenPublishedChannelManagementDTO(channel))
	}
	return dtos
}
