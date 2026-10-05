package admin

import (
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/shipping"

	"github.com/stretchr/testify/require"
)

func TestYanwenPublishedChannelManagementDTOContainsOnlyManagementContractFields(t *testing.T) {
	createdAt := time.Date(2026, time.October, 5, 4, 0, 0, 0, time.UTC)
	channel := shipping.YanwenPublishedChannel{
		ID:                       12,
		Environment:              shipping.YanwenPublishedChannelEnvironmentProduction,
		ProductCode:              "481",
		DisplayName:              "燕文专线",
		Countries:                `["US","CA"]`,
		PackageType:              "parcel",
		MaxWeightGrams:           2000,
		VolumetricDivisor:        8000,
		RequireReceiverTaxNumber: true,
		RequireIOSS:              true,
		RequireEORI:              false,
		Notes:                    "官网报价复核",
		Enabled:                  true,
		CreatedAt:                createdAt,
		UpdatedAt:                createdAt,
	}

	payload, err := json.Marshal(newYanwenPublishedChannelManagementDTO(channel))
	require.NoError(t, err)
	require.JSONEq(t, `{"id":12,"environment":"production","product_code":"481","display_name":"燕文专线","countries":"[\"US\",\"CA\"]","package_type":"parcel","max_weight_grams":2000,"volumetric_divisor":8000,"require_receiver_tax_number":true,"require_ioss":true,"require_eori":false,"notes":"官网报价复核","enabled":true,"created_at":"2026-10-05T04:00:00Z","updated_at":"2026-10-05T04:00:00Z"}`, string(payload))
	require.NotContains(t, string(payload), "deleted_at")
}

func TestYanwenPublishedChannelManagementDTOsReturnsAnEmptyArrayForNoChannels(t *testing.T) {
	dtos := newYanwenPublishedChannelManagementDTOs(nil)
	require.NotNil(t, dtos)
	require.Empty(t, dtos)
}
